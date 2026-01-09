package relation

import (
	"fmt"
	"strings"

	"github.com/story-engine/main-service/internal/core/relation"
)

// SummaryGenerator generates synthetic summaries for relations without LLM
type SummaryGenerator struct{}

// NewSummaryGenerator creates a new SummaryGenerator
func NewSummaryGenerator() *SummaryGenerator {
	return &SummaryGenerator{}
}

// GenerateSummary generates a summary string for a relation using entity IDs
// Format: "<source_type>:<source_id> <relation_type> <target_type>:<target_id>"
func (g *SummaryGenerator) GenerateSummary(rel *relation.EntityRelation) string {
	summary := fmt.Sprintf("%s:%s %s %s:%s",
		rel.SourceType, rel.SourceID.String(),
		rel.RelationType,
		rel.TargetType, rel.TargetID.String())

	if rel.ContextType != nil && rel.ContextID != nil {
		summary += fmt.Sprintf(" (context: %s)", *rel.ContextType)
	}

	return strings.TrimSpace(summary)
}

// GenerateSummaryWithNames generates a summary using entity names
// Format: "<source_name> <relation_type> <target_name>"
// Example: "John parent_of Mary" or "The Red Keep located_in King's Landing"
func (g *SummaryGenerator) GenerateSummaryWithNames(rel *relation.EntityRelation, sourceName, targetName string) string {
	summary := fmt.Sprintf("%s %s %s", sourceName, rel.RelationType, targetName)

	if rel.ContextType != nil && rel.ContextID != nil {
		summary += fmt.Sprintf(" (context: %s)", *rel.ContextType)
	}

	return strings.TrimSpace(summary)
}
