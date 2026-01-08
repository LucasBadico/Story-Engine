package entity_extraction

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	"github.com/story-engine/llm-gateway-service/internal/ports/llm"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

type Phase3MatchUseCase struct {
	chunkRepo repositories.ChunkRepository
	docRepo   repositories.DocumentRepository
	embedder  embeddings.Embedder
	model     llm.RouterModel
	logger    *logger.Logger
}

func NewPhase3MatchUseCase(
	chunkRepo repositories.ChunkRepository,
	docRepo repositories.DocumentRepository,
	embedder embeddings.Embedder,
	model llm.RouterModel,
	logger *logger.Logger,
) *Phase3MatchUseCase {
	return &Phase3MatchUseCase{
		chunkRepo: chunkRepo,
		docRepo:   docRepo,
		embedder:  embedder,
		model:     model,
		logger:    logger,
	}
}

type Phase3MatchInput struct {
	TenantID      uuid.UUID
	WorldID       *uuid.UUID
	Findings      []Phase2EntityFinding
	Context       string
	MinSimilarity float64
	MaxCandidates int
	EventLogger   ExtractionEventLogger
}

type Phase3MatchOutput struct {
	Results []Phase3MatchResult
}

type Phase3MatchResult struct {
	EntityType string
	Name       string
	Summary    string
	Candidates []Phase3MatchCandidate
	Match      *Phase3ConfirmedMatch
}

type Phase3MatchCandidate struct {
	ChunkID    uuid.UUID
	DocumentID uuid.UUID
	SourceType memory.SourceType
	SourceID   uuid.UUID
	EntityName string
	Summary    string
	Similarity float64
}

type Phase3ConfirmedMatch struct {
	Candidate Phase3MatchCandidate
	Reason    string
}

//go:embed prompts/phase3_entity_match.prompt
var phase3EntityMatchPromptTemplate string

//go:embed prompts/phase3_entity_match_repair.prompt
var phase3EntityMatchRepairPromptTemplate string

func (u *Phase3MatchUseCase) Execute(ctx context.Context, input Phase3MatchInput) (Phase3MatchOutput, error) {
	if input.TenantID == uuid.Nil {
		return Phase3MatchOutput{}, errors.New("tenant id is required")
	}
	if u.chunkRepo == nil || u.docRepo == nil || u.embedder == nil {
		return Phase3MatchOutput{}, errors.New("repositories and embedder are required")
	}

	minSimilarity := input.MinSimilarity
	if minSimilarity <= 0 {
		minSimilarity = 0.75
	}
	maxCandidates := input.MaxCandidates
	if maxCandidates <= 0 {
		maxCandidates = 5
	}

	results := make([]Phase3MatchResult, 0, len(input.Findings))
	var resultsMu sync.Mutex

	parallelism := getMatchParallelism(u.logger)
	sem := make(chan struct{}, parallelism)
	var wg sync.WaitGroup

	for _, finding := range input.Findings {
		finding := finding
		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()

			result, err := u.matchFinding(ctx, input, finding, minSimilarity, maxCandidates)
			if err != nil {
				u.logger.Error("phase3 match failed", "entity_type", finding.EntityType, "error", err)
				result = Phase3MatchResult{
					EntityType: finding.EntityType,
					Name:       finding.Name,
					Summary:    finding.Summary,
				}
			}

			resultsMu.Lock()
			results = append(results, result)
			resultsMu.Unlock()
		}()
	}

	wg.Wait()
	return Phase3MatchOutput{Results: results}, nil
}

func getMatchParallelism(log *logger.Logger) int {
	cpuCount := runtime.NumCPU()
	parallelism := cpuCount
	if parallelism < 1 {
		parallelism = 1
	}

	if value := strings.TrimSpace(os.Getenv("ENTITY_EXTRACT_PARALLELISM")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			parallelism = parsed
		}
	}

	if log != nil && parallelism > cpuCount {
		log.Warn(
			"entity extract parallelism exceeds CPU count",
			"parallelism", parallelism,
			"cpu_count", cpuCount,
		)
	}

	if parallelism < 1 {
		parallelism = 1
	}

	return parallelism
}

func (u *Phase3MatchUseCase) matchFinding(
	ctx context.Context,
	input Phase3MatchInput,
	finding Phase2EntityFinding,
	minSimilarity float64,
	maxCandidates int,
) (Phase3MatchResult, error) {
	query := strings.TrimSpace(finding.Summary)
	if query == "" {
		query = strings.TrimSpace(finding.Name)
	}
	if query == "" {
		return Phase3MatchResult{
			EntityType: finding.EntityType,
			Name:       finding.Name,
			Summary:    finding.Summary,
		}, nil
	}

	sourceType, ok := mapEntityTypeToSourceType(finding.EntityType)
	if !ok {
		u.logger.Warn("unsupported entity type for phase3 match", "entity_type", finding.EntityType)
		return Phase3MatchResult{
			EntityType: finding.EntityType,
			Name:       finding.Name,
			Summary:    finding.Summary,
		}, nil
	}

	embedding, err := u.embedder.EmbedText(query)
	if err != nil {
		return Phase3MatchResult{}, err
	}

	filters := &repositories.SearchFilters{
		SourceTypes: []memory.SourceType{sourceType},
		ChunkTypes:  []string{"summary"},
	}
	if input.WorldID != nil {
		filters.WorldIDs = []uuid.UUID{*input.WorldID}
	}

	scored, err := u.chunkRepo.SearchSimilar(ctx, input.TenantID, embedding, maxCandidates, nil, filters)
	if err != nil {
		return Phase3MatchResult{}, err
	}

	candidates := make([]Phase3MatchCandidate, 0, len(scored))
	for _, entry := range scored {
		if entry == nil || entry.Chunk == nil {
			continue
		}
		similarity := 1 - entry.Distance
		if similarity < minSimilarity {
			continue
		}
		doc, err := u.docRepo.GetByID(ctx, entry.Chunk.DocumentID)
		if err != nil {
			u.logger.Error("phase3 match: failed to load document", "document_id", entry.Chunk.DocumentID, "error", err)
			continue
		}

		summary := strings.TrimSpace(entry.Chunk.Content)
		if summary == "" && entry.Chunk.EmbedText != nil {
			summary = strings.TrimSpace(*entry.Chunk.EmbedText)
		}

		entityName := ""
		if entry.Chunk.EntityName != nil {
			entityName = strings.TrimSpace(*entry.Chunk.EntityName)
		}

		candidates = append(candidates, Phase3MatchCandidate{
			ChunkID:    entry.Chunk.ID,
			DocumentID: entry.Chunk.DocumentID,
			SourceType: doc.SourceType,
			SourceID:   doc.SourceID,
			EntityName: entityName,
			Summary:    summary,
			Similarity: similarity,
		})
	}

	match := (*Phase3ConfirmedMatch)(nil)
	if len(candidates) > 0 {
		match = u.confirmMatch(ctx, finding, candidates, input.Context)
	}

	eventLogger := normalizeEventLogger(input.EventLogger)
	if match == nil {
		emitEvent(ctx, eventLogger, ExtractionEvent{
			Type:    "match.none",
			Phase:   "matcher",
			Message: fmt.Sprintf("no match for %s: %s", finding.EntityType, finding.Name),
			Data: map[string]interface{}{
				"entity_type": finding.EntityType,
				"name":        finding.Name,
				"candidates":  len(candidates),
			},
		})
	} else {
		emitEvent(ctx, eventLogger, ExtractionEvent{
			Type:    "match.found",
			Phase:   "matcher",
			Message: fmt.Sprintf("match for %s: %s", finding.EntityType, finding.Name),
			Data: map[string]interface{}{
				"entity_type": finding.EntityType,
				"name":        finding.Name,
				"source_type": match.Candidate.SourceType,
				"source_id":   match.Candidate.SourceID.String(),
				"similarity":  match.Candidate.Similarity,
			},
		})
	}

	return Phase3MatchResult{
		EntityType: finding.EntityType,
		Name:       finding.Name,
		Summary:    finding.Summary,
		Candidates: candidates,
		Match:      match,
	}, nil
}

func (u *Phase3MatchUseCase) confirmMatch(
	ctx context.Context,
	finding Phase2EntityFinding,
	candidates []Phase3MatchCandidate,
	context string,
) *Phase3ConfirmedMatch {
	if u.model == nil || len(candidates) == 0 {
		return nil
	}

	prompt := buildPhase3MatchPrompt(finding, candidates, context)
	raw, err := u.model.Generate(ctx, prompt)
	if err != nil {
		u.logger.Error("phase3 matcher model failed", "error", err)
		return nil
	}

	u.logger.Info(fmt.Sprintf("====model answer====\n%s\n===================", raw))

	parsed, err := parsePhase3MatchOutput(raw)
	if err != nil {
		u.logger.Error("failed to parse phase3 match output", "error", err)
		repaired, repairErr := u.repairMatchOutput(ctx, raw, finding, candidates, context)
		if repairErr != nil {
			return nil
		}
		parsed = repaired
	}

	if parsed.Match == nil {
		return nil
	}
	if parsed.Match.Index < 0 || parsed.Match.Index >= len(candidates) {
		return nil
	}

	reason := strings.TrimSpace(parsed.Match.Reason)
	if reason == "" {
		reason = strings.TrimSpace(parsed.Reason)
	}

	return &Phase3ConfirmedMatch{
		Candidate: candidates[parsed.Match.Index],
		Reason:    reason,
	}
}

func (u *Phase3MatchUseCase) repairMatchOutput(
	ctx context.Context,
	raw string,
	finding Phase2EntityFinding,
	candidates []Phase3MatchCandidate,
	context string,
) (phase3MatchModelOutput, error) {
	if strings.TrimSpace(raw) == "" {
		return phase3MatchModelOutput{}, errors.New("empty match output")
	}

	repairPrompt := buildPhase3MatchRepairPrompt(raw, finding, candidates, context)
	repairedRaw, err := u.model.Generate(ctx, repairPrompt)
	if err != nil {
		u.logger.Error("phase3 matcher repair failed", "error", err)
		return phase3MatchModelOutput{}, err
	}

	u.logger.Info(fmt.Sprintf("====model answer====\n%s\n===================", repairedRaw))

	return parsePhase3MatchOutput(repairedRaw)
}

type phase3MatchModelOutput struct {
	Match  *phase3MatchSelection `json:"match"`
	Reason string                `json:"reason"`
}

type phase3MatchSelection struct {
	Index  int    `json:"index"`
	Reason string `json:"reason"`
}

func parsePhase3MatchOutput(raw string) (phase3MatchModelOutput, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return phase3MatchModelOutput{}, errors.New("empty match output")
	}

	clean = stripCodeFences(clean)

	var output phase3MatchModelOutput
	if err := json.Unmarshal([]byte(clean), &output); err == nil {
		return output, nil
	}

	if sliced := extractFirstJSONObject(clean); sliced != "" {
		if err := json.Unmarshal([]byte(sliced), &output); err == nil {
			return output, nil
		}
	}

	return phase3MatchModelOutput{}, errors.New("invalid match output JSON")
}

func buildPhase3MatchPrompt(finding Phase2EntityFinding, candidates []Phase3MatchCandidate, context string) string {
	prompt := phase3EntityMatchPromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{entity_type}}", strings.TrimSpace(finding.EntityType))
	prompt = strings.ReplaceAll(prompt, "{{entity_name}}", strings.TrimSpace(finding.Name))
	prompt = strings.ReplaceAll(prompt, "{{entity_summary}}", strings.TrimSpace(finding.Summary))
	prompt = strings.ReplaceAll(prompt, "{{context_if_any}}", strings.TrimSpace(context))
	prompt = strings.ReplaceAll(prompt, "{{candidates_block}}", renderPhase3CandidatesBlock(candidates))
	return strings.TrimSpace(prompt) + "\n"
}

func buildPhase3MatchRepairPrompt(rawOutput string, finding Phase2EntityFinding, candidates []Phase3MatchCandidate, context string) string {
	prompt := phase3EntityMatchRepairPromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{raw_output}}", strings.TrimSpace(rawOutput))
	prompt = strings.ReplaceAll(prompt, "{{entity_type}}", strings.TrimSpace(finding.EntityType))
	prompt = strings.ReplaceAll(prompt, "{{entity_name}}", strings.TrimSpace(finding.Name))
	prompt = strings.ReplaceAll(prompt, "{{entity_summary}}", strings.TrimSpace(finding.Summary))
	prompt = strings.ReplaceAll(prompt, "{{context_if_any}}", strings.TrimSpace(context))
	prompt = strings.ReplaceAll(prompt, "{{candidates_block}}", renderPhase3CandidatesBlock(candidates))
	return strings.TrimSpace(prompt) + "\n"
}

func renderPhase3CandidatesBlock(candidates []Phase3MatchCandidate) string {
	if len(candidates) == 0 {
		return "none"
	}

	lines := make([]string, 0, len(candidates))
	for i, candidate := range candidates {
		name := strings.TrimSpace(candidate.EntityName)
		if name == "" {
			name = "(unknown)"
		}
		summary := strings.TrimSpace(candidate.Summary)
		if summary == "" {
			summary = "(no summary)"
		}
		lines = append(lines, fmt.Sprintf(
			"- index: %d | source_type: %s | source_id: %s | name: %s | similarity: %.3f\n  summary: %s",
			i, candidate.SourceType, candidate.SourceID, name, candidate.Similarity, summary,
		))
	}

	return strings.Join(lines, "\n")
}

func mapEntityTypeToSourceType(entityType string) (memory.SourceType, bool) {
	switch strings.ToLower(strings.TrimSpace(entityType)) {
	case "character":
		return memory.SourceTypeCharacter, true
	case "location":
		return memory.SourceTypeLocation, true
	case "artefact":
		return memory.SourceTypeArtifact, true
	case "faction":
		return memory.SourceTypeFaction, true
	case "event":
		return memory.SourceTypeEvent, true
	default:
		return "", false
	}
}
