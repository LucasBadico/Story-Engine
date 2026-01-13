package relations

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/llm"
)

type Phase6RelationNormalizeUseCase struct {
	logger       *logger.Logger
	summaryModel llm.RouterModel
}

func NewPhase6RelationNormalizeUseCase(logger *logger.Logger) *Phase6RelationNormalizeUseCase {
	return &Phase6RelationNormalizeUseCase{logger: logger}
}

func (uc *Phase6RelationNormalizeUseCase) SetSummaryModel(model llm.RouterModel) {
	uc.summaryModel = model
}

type Phase6RelationTypeDefinition struct {
	Mirror             string `json:"mirror"`
	Symmetric          bool   `json:"symmetric"`
	PreferredDirection string `json:"preferred_direction"`
	Semantics          string `json:"semantics"`
}

//go:embed prompts/phase6_custom_relation_summary.prompt
var phase6CustomSummaryPromptTemplate string

type Phase6ResolvedRef struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

type Phase6ExistingRelation struct {
	SourceID     string `json:"source_id"`
	TargetID     string `json:"target_id"`
	RelationType string `json:"relation_type"`
	ContextType  string `json:"context_type,omitempty"`
	ContextID    string `json:"context_id,omitempty"`
}

type Phase6RelationNormalizeInput struct {
	RequestID                      string                                  `json:"request_id"`
	Context                        Phase5Context                           `json:"context"`
	Relations                      []Phase5RelationCandidate               `json:"relations"`
	RefMap                         map[string]Phase6ResolvedRef            `json:"ref_map,omitempty"`
	SuggestedRelationsBySourceType map[string]Phase5PerEntityRelationMap   `json:"suggested_relations_by_source_type"`
	RelationTypes                  map[string]Phase6RelationTypeDefinition `json:"relation_types"`
	ExistingRelations              []Phase6ExistingRelation                `json:"existing_relations,omitempty"`
}

type Phase6RelationNormalizeOutput struct {
	Relations []Phase6NormalizedRelation `json:"relations"`
}

type Phase6NormalizedRelation struct {
	Source       Phase6NormalizedNode `json:"source"`
	Target       Phase6NormalizedNode `json:"target"`
	RelationType string               `json:"relation_type"`
	Direction    string               `json:"direction"`
	CreateMirror bool                 `json:"create_mirror"`
	Confidence   float64              `json:"confidence,omitempty"`
	Polarity     string               `json:"polarity,omitempty"`
	Implicit     bool                 `json:"implicit,omitempty"`
	Evidence     Phase5Evidence       `json:"evidence,omitempty"`
	Status       string               `json:"status"`
	Dedup        Phase6Dedup          `json:"dedup"`
	Summary      string               `json:"summary,omitempty"`
	MirrorOf     *string              `json:"mirror_of,omitempty"`
}

type Phase6NormalizedNode struct {
	Ref  string `json:"ref"`
	ID   string `json:"id,omitempty"`
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

type Phase6Dedup struct {
	IsDuplicate bool   `json:"is_duplicate"`
	Reason      string `json:"reason,omitempty"`
}

func (uc *Phase6RelationNormalizeUseCase) Execute(ctx context.Context, input Phase6RelationNormalizeInput) (Phase6RelationNormalizeOutput, error) {
	if strings.TrimSpace(input.RequestID) == "" {
		return Phase6RelationNormalizeOutput{}, errors.New("request_id is required")
	}
	if len(input.Relations) == 0 {
		return Phase6RelationNormalizeOutput{Relations: []Phase6NormalizedRelation{}}, nil
	}

	parallelism := getRelationNormalizeParallelism(uc.logger)
	if parallelism < 1 {
		parallelism = 1
	}

	normalizedBuckets := make([][]Phase6NormalizedRelation, len(input.Relations))
	sem := make(chan struct{}, parallelism)
	var wg sync.WaitGroup
	for idx, rel := range input.Relations {
		idx := idx
		rel := rel
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()

			normalizedRel, ok := uc.normalizePhase6Relation(ctx, input, rel)
			if !ok {
				return
			}

			items := []Phase6NormalizedRelation{normalizedRel}
			if strings.HasPrefix(strings.ToLower(normalizedRel.RelationType), "custom:") {
				mirror := normalizedRel
				mirror.Source, mirror.Target = normalizedRel.Target, normalizedRel.Source
				mirror.CreateMirror = false
				mirror.Direction = "source_to_target"
				key := phase6RelationKey(normalizedRel)
				mirror.MirrorOf = &key
				items = append(items, mirror)
			}

			normalizedBuckets[idx] = items
		}()
	}
	wg.Wait()

	normalized := make([]Phase6NormalizedRelation, 0, len(input.Relations))
	for _, items := range normalizedBuckets {
		if len(items) == 0 {
			continue
		}
		normalized = append(normalized, items...)
	}

	return Phase6RelationNormalizeOutput{Relations: normalized}, nil
}

func (uc *Phase6RelationNormalizeUseCase) normalizePhase6Relation(ctx context.Context, input Phase6RelationNormalizeInput, rel Phase5RelationCandidate) (Phase6NormalizedRelation, bool) {
	sourceType := normalizeEntityType(rel.Source.Type)
	targetType := normalizeEntityType(rel.Target.Type)
	if sourceType == "" || targetType == "" || strings.TrimSpace(rel.RelationType) == "" {
		return Phase6NormalizedRelation{}, false
	}

	relationType := strings.TrimSpace(rel.RelationType)
	custom := strings.HasPrefix(strings.ToLower(relationType), "custom:")
	if custom {
		trimmed := strings.TrimPrefix(relationType, "custom:")
		trimmed = strings.TrimSpace(trimmed)
		if trimmed != "" {
			if _, ok := input.RelationTypes[trimmed]; ok {
				relationType = trimmed
				custom = false
			}
		}
	}

	def, hasDef := input.RelationTypes[relationType]
	if !hasDef && !custom {
		relationType = fmt.Sprintf("custom:%s", relationType)
		custom = true
	}

	if !custom && !isRelationAllowed(input.SuggestedRelationsBySourceType, sourceType, targetType, relationType) {
		return Phase6NormalizedRelation{}, false
	}

	source := Phase6NormalizedNode{
		Ref:  rel.Source.Ref,
		Type: sourceType,
	}
	target := Phase6NormalizedNode{
		Ref:  rel.Target.Ref,
		Type: targetType,
	}
	if resolved, ok := input.RefMap[rel.Source.Ref]; ok {
		source.ID = resolved.ID
		source.Name = resolved.Name
	}
	if resolved, ok := input.RefMap[rel.Target.Ref]; ok {
		target.ID = resolved.ID
		target.Name = resolved.Name
	}

	status := "ready"
	if source.ID == "" || target.ID == "" {
		status = "pending_entities"
	}

	dedup := Phase6Dedup{}
	if status == "ready" && isDuplicate(input.ExistingRelations, source.ID, target.ID, relationType, input.Context) {
		dedup.IsDuplicate = true
		dedup.Reason = "existing_relation"
	}

	direction := "source_to_target"
	createMirror := false
	if hasDef {
		if strings.TrimSpace(def.PreferredDirection) != "" {
			direction = def.PreferredDirection
		}
		if strings.TrimSpace(def.Mirror) != "" {
			createMirror = true
		}
	}
	if custom {
		createMirror = false
	}

	summary := uc.buildRelationSummary(ctx, relationType, def, source, target, rel)

	return Phase6NormalizedRelation{
		Source:       source,
		Target:       target,
		RelationType: relationType,
		Direction:    direction,
		CreateMirror: createMirror,
		Confidence:   rel.Confidence,
		Polarity:     rel.Polarity,
		Implicit:     rel.Implicit,
		Evidence:     rel.Evidence,
		Status:       status,
		Dedup:        dedup,
		Summary:      summary,
	}, true
}

func isDuplicate(existing []Phase6ExistingRelation, sourceID, targetID, relationType string, ctx Phase5Context) bool {
	if len(existing) == 0 {
		return false
	}
	for _, rel := range existing {
		if rel.SourceID != sourceID || rel.TargetID != targetID || rel.RelationType != relationType {
			continue
		}
		if rel.ContextType != "" && rel.ContextType != ctx.Type {
			continue
		}
		if rel.ContextID != "" && rel.ContextID != ctx.ID {
			continue
		}
		return true
	}
	return false
}

func phase6RelationKey(rel Phase6NormalizedRelation) string {
	return fmt.Sprintf("%s|%s|%s", rel.Source.Ref, rel.Target.Ref, rel.RelationType)
}

func (uc *Phase6RelationNormalizeUseCase) buildRelationSummary(ctx context.Context, relationType string, def Phase6RelationTypeDefinition, source Phase6NormalizedNode, target Phase6NormalizedNode, rel Phase5RelationCandidate) string {
	sourceName := displayName(source)
	targetName := displayName(target)
	if sourceName == "" || targetName == "" {
		return ""
	}

	if strings.HasPrefix(strings.ToLower(relationType), "custom:") {
		if uc.summaryModel != nil {
			summary, err := uc.generateCustomSummary(ctx, relationType, sourceName, targetName, rel)
			if err == nil && strings.TrimSpace(summary) != "" {
				return summary
			}
		}
		return fmt.Sprintf("%s is related to %s.", sourceName, targetName)
	}

	semantics := strings.TrimSpace(def.Semantics)
	if semantics == "" {
		if template, ok := relationSummaryTemplates[relationType]; ok {
			return fmt.Sprintf(template, sourceName, targetName)
		}
		return ""
	}

	semantics = strings.ReplaceAll(semantics, "Source", sourceName)
	semantics = strings.ReplaceAll(semantics, "source", sourceName)
	semantics = strings.ReplaceAll(semantics, "Target", targetName)
	semantics = strings.ReplaceAll(semantics, "target", targetName)
	if !strings.HasSuffix(semantics, ".") {
		semantics += "."
	}
	return semantics
}

func displayName(node Phase6NormalizedNode) string {
	if strings.TrimSpace(node.Name) != "" {
		return strings.TrimSpace(node.Name)
	}
	if node.ID != "" {
		return node.ID
	}
	return ""
}

var relationSummaryTemplates = map[string]string{
	"parent_of":       "%s is a parent of %s.",
	"child_of":        "%s is a child of %s.",
	"sibling_of":      "%s and %s are siblings.",
	"spouse_of":       "%s and %s are spouses or partners.",
	"ally_of":         "%s and %s are allies.",
	"enemy_of":        "%s and %s are enemies.",
	"member_of":       "%s is a member of %s.",
	"has_member":      "%s has %s as a member.",
	"leader_of":       "%s leads %s.",
	"led_by":          "%s is led by %s.",
	"located_in":      "%s is located in %s.",
	"contains":        "%s contains %s.",
	"owns":            "%s owns %s.",
	"owned_by":        "%s is owned by %s.",
	"mentor_of":       "%s mentors %s.",
	"mentored_by":     "%s is mentored by %s.",
	"participated_in": "%s participated in %s.",
	"has_participant": "%s has %s as a participant.",
	"triggered":       "%s triggered %s.",
	"triggered_by":    "%s was triggered by %s.",
	"resulted_in":     "%s resulted in %s.",
	"resulted_from":   "%s resulted from %s.",
	"prevented":       "%s prevented %s.",
	"prevented_by":    "%s was prevented by %s.",
	"revealed":        "%s revealed %s.",
	"revealed_by":     "%s was revealed by %s.",
}

type phase6SummaryModelOutput struct {
	Summary string `json:"summary"`
}

func (uc *Phase6RelationNormalizeUseCase) generateCustomSummary(ctx context.Context, relationType, sourceName, targetName string, rel Phase5RelationCandidate) (string, error) {
	if uc.summaryModel == nil {
		return "", errors.New("summary model not configured")
	}

	prompt := buildPhase6CustomSummaryPrompt(relationType, sourceName, targetName, rel)
	raw, err := uc.summaryModel.Generate(ctx, prompt)
	if err != nil {
		if uc.logger != nil {
			uc.logger.Error("phase6 custom summary model failed", "error", err)
		}
		return "", err
	}

	parsed, err := parsePhase6CustomSummaryOutput(raw)
	if err != nil {
		if uc.logger != nil {
			uc.logger.Error("phase6 custom summary parse failed", "error", err)
		}
		return "", err
	}

	return strings.TrimSpace(parsed.Summary), nil
}

func buildPhase6CustomSummaryPrompt(relationType, sourceName, targetName string, rel Phase5RelationCandidate) string {
	payload := map[string]interface{}{
		"relation_type": relationType,
		"source_name":   sourceName,
		"target_name":   targetName,
		"evidence": map[string]string{
			"span_id": rel.Evidence.SpanID,
			"quote":   rel.Evidence.Quote,
		},
	}
	encoded, _ := json.Marshal(payload)
	prompt := strings.TrimSpace(phase6CustomSummaryPromptTemplate)
	prompt = strings.ReplaceAll(prompt, "{{input_json}}", string(encoded))
	return strings.TrimSpace(prompt) + "\n"
}

func parsePhase6CustomSummaryOutput(raw string) (phase6SummaryModelOutput, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return phase6SummaryModelOutput{}, errors.New("empty summary output")
	}

	clean = stripCodeFences(clean)

	var output phase6SummaryModelOutput
	if err := json.Unmarshal([]byte(clean), &output); err == nil {
		return output, nil
	}

	if sliced := extractFirstJSONObject(clean); sliced != "" {
		if err := json.Unmarshal([]byte(sliced), &output); err == nil {
			return output, nil
		}
	}

	return phase6SummaryModelOutput{}, errors.New("invalid summary output JSON")
}
