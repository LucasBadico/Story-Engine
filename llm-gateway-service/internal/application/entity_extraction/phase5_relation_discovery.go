package entity_extraction

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/llm"
)

type Phase5RelationDiscoveryUseCase struct {
	model  llm.RouterModel
	logger *logger.Logger
}

func NewPhase5RelationDiscoveryUseCase(model llm.RouterModel, logger *logger.Logger) *Phase5RelationDiscoveryUseCase {
	return &Phase5RelationDiscoveryUseCase{
		model:  model,
		logger: logger,
	}
}

type Phase5TextSpec struct {
	Mode          string       `json:"mode"` // full_text|spans
	Text          *string      `json:"text,omitempty"`
	GlobalSummary []string     `json:"global_summary,omitempty"`
	Spans         []Phase5Span `json:"spans,omitempty"`
}

type Phase5Span struct {
	SpanID string `json:"span_id"`
	Start  int    `json:"start"`
	End    int    `json:"end"`
	Text   string `json:"text"`
}

type Phase5Context struct {
	Type        string  `json:"type"` // story|world|scene|beat
	ID          string  `json:"id"`
	POVRef      *string `json:"pov_ref,omitempty"`
	LocationRef *string `json:"location_ref,omitempty"`
}

type Phase5EntityFinding struct {
	Ref      string   `json:"ref"`
	Type     string   `json:"type"`
	Name     string   `json:"name"`
	Summary  string   `json:"summary"`
	Mentions []string `json:"mentions,omitempty"`
}

type Phase5Match struct {
	Ref           string  `json:"ref"`
	Type          string  `json:"type"`
	ID            string  `json:"id"`
	CanonicalName string  `json:"canonical_name"`
	Similarity    float64 `json:"similarity"`
}

type Phase5ConfirmedMatch struct {
	FindingRef string      `json:"finding_ref"`
	Match      Phase5Match `json:"match"`
}

type Phase5RelationConstraints struct {
	MinConfidence    float64 `json:"min_confidence"`
	AllowImplicit    bool    `json:"allow_implicit"`
	RequiresEvidence bool    `json:"requires_evidence"`
}

type Phase5RelationConstraintSpec struct {
	PairCandidates []string                   `json:"pair_candidates"`
	Description    string                     `json:"description"`
	Contexts       []string                   `json:"contexts,omitempty"`
	Signals        []string                   `json:"signals,omitempty"`
	AntiSignals    []string                   `json:"anti_signals,omitempty"`
	Constraints    *Phase5RelationConstraints `json:"constraints,omitempty"`
}

type Phase5PerEntityRelationMap struct {
	EntityType string                                  `json:"entity_type"`
	Version    int                                     `json:"version"`
	Relations  map[string]Phase5RelationConstraintSpec `json:"relations"`
}

type Phase5RelationDiscoveryInput struct {
	RequestID                      string                                `json:"request_id"`
	Context                        Phase5Context                         `json:"context"`
	Text                           Phase5TextSpec                        `json:"text"`
	EntityFindings                 []Phase5EntityFinding                 `json:"entity_findings"`
	ConfirmedMatches               []Phase5ConfirmedMatch                `json:"confirmed_matches,omitempty"`
	SuggestedRelationsBySourceType map[string]Phase5PerEntityRelationMap `json:"suggested_relations_by_source_type"`
	RelationTypeSemantics          map[string]string                     `json:"relation_type_semantics,omitempty"`
}

type Phase5RelationDiscoveryOutput struct {
	Relations []Phase5RelationCandidate `json:"relations"`
}

type Phase5RelationCandidate struct {
	Source       Phase5RelationNode `json:"source"`
	Target       Phase5RelationNode `json:"target"`
	RelationType string             `json:"relation_type"`
	Polarity     string             `json:"polarity"`
	Implicit     bool               `json:"implicit"`
	Confidence   float64            `json:"confidence"`
	Evidence     Phase5Evidence     `json:"evidence"`
}

type Phase5RelationNode struct {
	Ref  string `json:"ref"`
	Type string `json:"type"`
}

type Phase5Evidence struct {
	SpanID string `json:"span_id"`
	Quote  string `json:"quote"`
}

//go:embed prompts/phase5_relation_discovery.prompt
var phase5RelationDiscoveryPromptTemplate string

func (uc *Phase5RelationDiscoveryUseCase) Execute(ctx context.Context, input Phase5RelationDiscoveryInput) (Phase5RelationDiscoveryOutput, error) {
	if uc.model == nil {
		return Phase5RelationDiscoveryOutput{}, errors.New("model is required")
	}
	if strings.TrimSpace(input.RequestID) == "" {
		return Phase5RelationDiscoveryOutput{}, errors.New("request_id is required")
	}
	if err := validatePhase5Input(input); err != nil {
		return Phase5RelationDiscoveryOutput{}, err
	}

	prompt, err := buildPhase5RelationDiscoveryPrompt(input)
	if err != nil {
		return Phase5RelationDiscoveryOutput{}, err
	}

	raw, err := uc.model.Generate(ctx, prompt)
	if err != nil {
		if uc.logger != nil {
			uc.logger.Error("phase5 relation discovery model failed", "error", err)
		}
		return Phase5RelationDiscoveryOutput{}, err
	}

	parsed, err := parsePhase5RelationDiscoveryOutput(raw)
	if err != nil {
		if uc.logger != nil {
			uc.logger.Error("phase5 relation discovery parse failed", "error", err)
		}
		return Phase5RelationDiscoveryOutput{}, err
	}

	validated := validatePhase5Relations(input, parsed.Relations)
	return Phase5RelationDiscoveryOutput{Relations: validated}, nil
}

func validatePhase5Input(input Phase5RelationDiscoveryInput) error {
	if strings.TrimSpace(input.Text.Mode) == "" {
		return errors.New("text.mode is required")
	}
	mode := strings.ToLower(strings.TrimSpace(input.Text.Mode))
	if mode != "full_text" && mode != "spans" {
		return fmt.Errorf("unsupported text mode: %s", input.Text.Mode)
	}
	if mode == "full_text" {
		if input.Text.Text == nil || strings.TrimSpace(*input.Text.Text) == "" {
			return errors.New("text.text is required for full_text mode")
		}
	}
	if mode == "spans" {
		if len(input.Text.Spans) == 0 {
			return errors.New("text.spans is required for spans mode")
		}
	}
	if len(input.EntityFindings) == 0 {
		return errors.New("entity_findings is required")
	}
	if len(input.SuggestedRelationsBySourceType) == 0 {
		return errors.New("suggested_relations_by_source_type is required")
	}
	return nil
}

func buildPhase5RelationDiscoveryPrompt(input Phase5RelationDiscoveryInput) (string, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	prompt := phase5RelationDiscoveryPromptTemplate
	prompt = strings.ReplaceAll(prompt, "{{input_json}}", string(payload))
	prompt = strings.ReplaceAll(prompt, "{{suggested_relations_block}}", renderSuggestedRelationsBlock(input.SuggestedRelationsBySourceType))
	return strings.TrimSpace(prompt) + "\n", nil
}

func parsePhase5RelationDiscoveryOutput(raw string) (Phase5RelationDiscoveryOutput, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return Phase5RelationDiscoveryOutput{}, errors.New("empty relation discovery output")
	}

	clean = stripCodeFences(clean)

	var output Phase5RelationDiscoveryOutput
	if err := json.Unmarshal([]byte(clean), &output); err == nil {
		return output, nil
	}

	if sliced := extractFirstJSONObject(clean); sliced != "" {
		if err := json.Unmarshal([]byte(sliced), &output); err == nil {
			return output, nil
		}
	}

	return Phase5RelationDiscoveryOutput{}, errors.New("invalid relation discovery output JSON")
}

func validatePhase5Relations(input Phase5RelationDiscoveryInput, relations []Phase5RelationCandidate) []Phase5RelationCandidate {
	if len(relations) == 0 {
		return relations
	}

	mode := strings.ToLower(strings.TrimSpace(input.Text.Mode))
	spanIDs := map[string]struct{}{}
	if mode == "spans" {
		for _, span := range input.Text.Spans {
			if span.SpanID == "" {
				continue
			}
			spanIDs[span.SpanID] = struct{}{}
		}
	}

	valid := make([]Phase5RelationCandidate, 0, len(relations))
	for _, rel := range relations {
		sourceType := normalizeEntityType(rel.Source.Type)
		targetType := normalizeEntityType(rel.Target.Type)
		if sourceType == "" || targetType == "" || rel.RelationType == "" {
			continue
		}

		if mode == "spans" {
			if strings.TrimSpace(rel.Evidence.SpanID) == "" {
				continue
			}
			if _, ok := spanIDs[rel.Evidence.SpanID]; !ok {
				continue
			}
		}
		if strings.TrimSpace(rel.Evidence.Quote) == "" {
			continue
		}

		if !isRelationAllowed(input.SuggestedRelationsBySourceType, sourceType, targetType, rel.RelationType) {
			continue
		}

		if rel.Confidence < 0 {
			rel.Confidence = 0
		}
		if rel.Confidence > 1 {
			rel.Confidence = 1
		}

		valid = append(valid, rel)
	}

	return valid
}

func normalizeEntityType(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	switch trimmed {
	case "organization", "group":
		return "faction"
	default:
		return trimmed
	}
}

func isRelationAllowed(
	allowed map[string]Phase5PerEntityRelationMap,
	sourceType string,
	targetType string,
	relationType string,
) bool {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(relationType)), "custom:") {
		return true
	}
	sourceMap, ok := allowed[sourceType]
	if !ok {
		return false
	}
	spec, ok := sourceMap.Relations[relationType]
	if !ok {
		return false
	}
	if len(spec.PairCandidates) == 0 {
		return true
	}
	for _, candidate := range spec.PairCandidates {
		if normalizeEntityType(candidate) == targetType {
			return true
		}
	}
	return false
}

func renderSuggestedRelationsBlock(allowed map[string]Phase5PerEntityRelationMap) string {
	if len(allowed) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(allowed))
	for sourceType, relMap := range allowed {
		if len(relMap.Relations) == 0 {
			continue
		}
		relations := make([]string, 0, len(relMap.Relations))
		for relType, spec := range relMap.Relations {
			pairs := strings.Join(spec.PairCandidates, ", ")
			if pairs != "" {
				relations = append(relations, fmt.Sprintf("%s -> [%s]", relType, pairs))
			} else {
				relations = append(relations, relType)
			}
		}
		if len(relations) == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s: %s", sourceType, strings.Join(relations, "; ")))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, "\n")
}
