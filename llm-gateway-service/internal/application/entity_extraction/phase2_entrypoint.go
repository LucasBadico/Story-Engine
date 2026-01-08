package entity_extraction

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/llm"
)

type EntityMatchFunc func(ctx context.Context, entityType string, existingName string, newName string, context string) (bool, error)

type Phase2EntryUseCase struct {
	extractors  map[string]*Phase2EntityExtractorUseCase
	logger      *logger.Logger
	entityMatch EntityMatchFunc
}

func NewPhase2EntryUseCase(model llm.RouterModel, logger *logger.Logger, matcher EntityMatchFunc) *Phase2EntryUseCase {
	extractors := map[string]*Phase2EntityExtractorUseCase{
		"character": NewPhase2CharacterExtractorUseCase(model, logger),
		"location":  NewPhase2LocationExtractorUseCase(model, logger),
		"artefact":  NewPhase2ArtefactExtractorUseCase(model, logger),
		"faction":   NewPhase2FactionExtractorUseCase(model, logger),
	}

	return &Phase2EntryUseCase{
		extractors:  extractors,
		logger:      logger,
		entityMatch: matcher,
	}
}

type Phase2EntryInput struct {
	Context               string
	MaxCandidatesPerChunk int
	Chunks                []Phase2RoutedChunk
}

type Phase2RoutedChunk struct {
	ParagraphID int
	ChunkID     int
	StartOffset int
	EndOffset   int
	Text        string
	Types       []string
}

type Phase2EntityFinding struct {
	EntityType  string
	Name        string
	Summary     string
	Occurrences []Phase2EntityOccurrence
}

type Phase2EntityOccurrence struct {
	ParagraphID int
	ChunkID     int
	StartOffset int
	EndOffset   int
	Evidence    string
}

type Phase2EntryOutput struct {
	Findings []Phase2EntityFinding
}

func (u *Phase2EntryUseCase) Execute(ctx context.Context, input Phase2EntryInput) (Phase2EntryOutput, error) {
	if len(input.Chunks) == 0 {
		return Phase2EntryOutput{}, errors.New("chunks are required")
	}

	maxCandidates := input.MaxCandidatesPerChunk
	if maxCandidates <= 0 {
		maxCandidates = 5
	}

	findingsByType := map[string]map[string]*Phase2EntityFinding{}

	for _, chunk := range input.Chunks {
		text := strings.TrimSpace(chunk.Text)
		if text == "" {
			continue
		}
		if len(chunk.Types) == 0 {
			continue
		}

		results := make([]Phase2EntityExtractorOutput, 0, len(chunk.Types))
		var resultsMu sync.Mutex
		var wg sync.WaitGroup

		for _, entityType := range chunk.Types {
			extractor, ok := u.extractors[entityType]
			if !ok {
				u.logger.Warn("unknown phase2 entity type", "entity_type", entityType)
				continue
			}

			alreadyFound := collectExistingEntities(findingsByType, entityType)

			wg.Add(1)
			go func(entityType string, extractor *Phase2EntityExtractorUseCase, known []Phase2KnownEntity) {
				defer wg.Done()

				output, err := extractor.Execute(ctx, Phase2EntityExtractorInput{
					Text:          text,
					Context:       input.Context,
					AlreadyFound:  known,
					MaxCandidates: maxCandidates,
				})
				if err != nil {
					u.logger.Error("phase2 extractor failed", "entity_type", entityType, "error", err)
					return
				}

				resultsMu.Lock()
				results = append(results, output)
				resultsMu.Unlock()
			}(entityType, extractor, alreadyFound)
		}

		wg.Wait()

		for _, output := range results {
			for _, candidate := range output.Candidates {
				start := candidate.StartOffset + chunk.StartOffset
				end := candidate.EndOffset + chunk.StartOffset

				u.mergeFinding(ctx, findingsByType, output.EntityType, candidate.Name, Phase2EntityOccurrence{
					ParagraphID: chunk.ParagraphID,
					ChunkID:     chunk.ChunkID,
					StartOffset: start,
					EndOffset:   end,
					Evidence:    candidate.Evidence,
				}, candidate.Summary, input.Context)
			}
		}
	}

	findings := make([]Phase2EntityFinding, 0)
	for _, byName := range findingsByType {
		for _, finding := range byName {
			findings = append(findings, *finding)
		}
	}

	return Phase2EntryOutput{Findings: findings}, nil
}

func (u *Phase2EntryUseCase) mergeFinding(
	ctx context.Context,
	store map[string]map[string]*Phase2EntityFinding,
	entityType string,
	name string,
	occurrence Phase2EntityOccurrence,
	summary string,
	context string,
) {
	if name == "" {
		return
	}

	normalizedName := normalizeEntityName(name)
	if normalizedName == "" {
		return
	}

	byName, ok := store[entityType]
	if !ok {
		byName = map[string]*Phase2EntityFinding{}
		store[entityType] = byName
	}

	if existing := findMatchingFinding(ctx, u.entityMatch, entityType, normalizedName, name, byName, context); existing != nil {
		if !occurrenceAlreadyPresent(existing.Occurrences, occurrence) {
			existing.Occurrences = append(existing.Occurrences, occurrence)
		}
		if strings.TrimSpace(summary) != "" {
			existing.Summary = summary
		}
		return
	}

	byName[normalizedName] = &Phase2EntityFinding{
		EntityType: entityType,
		Name:       name,
		Summary:    strings.TrimSpace(summary),
		Occurrences: []Phase2EntityOccurrence{
			occurrence,
		},
	}
}

func findMatchingFinding(
	ctx context.Context,
	matcher EntityMatchFunc,
	entityType string,
	normalizedName string,
	rawName string,
	existing map[string]*Phase2EntityFinding,
	context string,
) *Phase2EntityFinding {
	if finding, ok := existing[normalizedName]; ok {
		return finding
	}

	if matcher == nil {
		return nil
	}

	for _, finding := range existing {
		same, err := matcher(ctx, entityType, finding.Name, rawName, context)
		if err != nil {
			continue
		}
		if same {
			return finding
		}
	}

	return nil
}

func occurrenceAlreadyPresent(existing []Phase2EntityOccurrence, incoming Phase2EntityOccurrence) bool {
	for _, occurrence := range existing {
		if occurrence.ParagraphID == incoming.ParagraphID &&
			occurrence.ChunkID == incoming.ChunkID &&
			occurrence.StartOffset == incoming.StartOffset &&
			occurrence.EndOffset == incoming.EndOffset {
			return true
		}
	}
	return false
}

func normalizeEntityName(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.Trim(normalized, "\"'`")
	return normalized
}

func collectExistingEntities(store map[string]map[string]*Phase2EntityFinding, entityType string) []Phase2KnownEntity {
	byName, ok := store[entityType]
	if !ok {
		return nil
	}
	entities := make([]Phase2KnownEntity, 0, len(byName))
	for _, finding := range byName {
		if strings.TrimSpace(finding.Name) == "" {
			continue
		}
		entities = append(entities, Phase2KnownEntity{
			Name:    finding.Name,
			Summary: finding.Summary,
		})
	}
	return entities
}
