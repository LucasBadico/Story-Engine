package entity_extraction

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
)

type relationFindingRef struct {
	ref   string
	index int
}

func (u *EntityAndRelationshipsExtractor) extractRelations(
	ctx context.Context,
	input EntityAndRelationshipsExtractorInput,
	phase2Output Phase2EntryOutput,
	phase3Output Phase3MatchOutput,
) ([]Phase8RelationResult, error) {
	requestID := strings.TrimSpace(input.RequestID)
	if requestID == "" {
		requestID = uuid.NewString()
	}

	findingRefs, entityFindings := buildPhase5Findings(phase2Output.Findings)
	confirmedMatches, refMap := buildPhase5MatchesAndRefMap(findingRefs, phase3Output.Results)
	textSpec := buildPhase5TextSpec(phase2Output.Findings, strings.TrimSpace(input.Text))

	phase5Input := Phase5RelationDiscoveryInput{
		RequestID:                      requestID,
		Context:                        buildPhase5Context(input),
		Text:                           textSpec,
		EntityFindings:                 entityFindings,
		ConfirmedMatches:               confirmedMatches,
		SuggestedRelationsBySourceType: input.SuggestedRelations,
		RelationTypeSemantics:          buildRelationTypeSemantics(input.RelationTypes, input.RelationTypeSemantics),
	}

	phase5Output, err := u.relationDiscovery.Execute(ctx, phase5Input)
	if err != nil {
		return nil, err
	}

	phase6Output, err := u.relationNormalize.Execute(ctx, Phase6RelationNormalizeInput{
		RequestID:                      requestID,
		Context:                        phase5Input.Context,
		Relations:                      phase5Output.Relations,
		RefMap:                         refMap,
		SuggestedRelationsBySourceType: input.SuggestedRelations,
		RelationTypes:                  input.RelationTypes,
	})
	if err != nil {
		return nil, err
	}

	matchOutput, err := u.relationMatcher.Execute(ctx, Phase7RelationMatchInput{
		TenantID:      input.TenantID,
		Relations:     phase6Output.Relations,
		SourceTypes:   defaultRelationMatchSourceTypes(),
		MaxMatches:    input.MaxRelationMatches,
		MinSimilarity: input.RelationMatchMinSim,
	})
	if err != nil {
		return nil, err
	}

	return buildPhase8Relations(phase6Output.Relations, matchOutput), nil
}

func buildPhase5Context(input EntityAndRelationshipsExtractorInput) Phase5Context {
	ctxType := "world"
	ctxID := input.WorldID.String()
	if strings.TrimSpace(input.Context) != "" {
		// Keep world context but pass extra string into ID suffix for now.
		ctxID = input.WorldID.String()
	}
	return Phase5Context{
		Type: ctxType,
		ID:   ctxID,
	}
}

func buildPhase5Findings(findings []Phase2EntityFinding) (map[string]relationFindingRef, []Phase5EntityFinding) {
	refs := make(map[string]relationFindingRef)
	out := make([]Phase5EntityFinding, 0, len(findings))
	perTypeCount := map[string]int{}

	for _, finding := range findings {
		entityType := strings.TrimSpace(finding.EntityType)
		if entityType == "" || strings.TrimSpace(finding.Name) == "" {
			continue
		}
		index := perTypeCount[entityType]
		perTypeCount[entityType]++
		ref := fmt.Sprintf("finding:%s:%d", entityType, index)

		mentions := collectMentions(finding.Occurrences)
		out = append(out, Phase5EntityFinding{
			Ref:      ref,
			Type:     entityType,
			Name:     finding.Name,
			Summary:  strings.TrimSpace(finding.Summary),
			Mentions: mentions,
		})
		key := buildFindingKey(entityType, finding.Name)
		refs[key] = relationFindingRef{ref: ref, index: index}
	}

	return refs, out
}

func buildPhase5MatchesAndRefMap(
	findingRefs map[string]relationFindingRef,
	results []Phase3MatchResult,
) ([]Phase5ConfirmedMatch, map[string]Phase6ResolvedRef) {
	confirmed := make([]Phase5ConfirmedMatch, 0, len(results))
	refMap := make(map[string]Phase6ResolvedRef)

	for _, result := range results {
		key := buildFindingKey(result.EntityType, result.Name)
		ref, ok := findingRefs[key]
		if !ok {
			continue
		}
		findingRef := ref.ref
		matchRef := fmt.Sprintf("match:%s:%d", strings.TrimSpace(result.EntityType), ref.index)

		refMap[findingRef] = Phase6ResolvedRef{
			ID:   "",
			Name: result.Name,
		}

		if result.Match == nil {
			continue
		}

		match := result.Match.Candidate
		refMap[matchRef] = Phase6ResolvedRef{
			ID:   match.SourceID.String(),
			Name: fallbackString(match.EntityName, result.Name),
		}

		confirmed = append(confirmed, Phase5ConfirmedMatch{
			FindingRef: findingRef,
			Match: Phase5Match{
				Ref:           matchRef,
				Type:          strings.TrimSpace(result.EntityType),
				ID:            match.SourceID.String(),
				CanonicalName: fallbackString(match.EntityName, result.Name),
				Similarity:    match.Similarity,
			},
		})
	}

	return confirmed, refMap
}

func buildPhase5TextSpec(findings []Phase2EntityFinding, fullText string) Phase5TextSpec {
	spans := make([]Phase5Span, 0)
	seen := map[string]struct{}{}
	spanIndex := 0

	for _, finding := range findings {
		for _, occurrence := range finding.Occurrences {
			text := strings.TrimSpace(occurrence.Evidence)
			if text == "" {
				continue
			}
			key := fmt.Sprintf("%d:%d:%s", occurrence.StartOffset, occurrence.EndOffset, text)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			spanIndex++
			spans = append(spans, Phase5Span{
				SpanID: fmt.Sprintf("span:%d", spanIndex),
				Start:  occurrence.StartOffset,
				End:    occurrence.EndOffset,
				Text:   text,
			})
		}
	}

	return Phase5TextSpec{
		Mode:  "spans",
		Text:  nil,
		Spans: spans,
	}
}

func buildPhase8Relations(relations []Phase6NormalizedRelation, matches Phase7RelationMatchOutput) []Phase8RelationResult {
	if len(relations) == 0 {
		return []Phase8RelationResult{}
	}

	matchMap := map[int][]Phase7RelationMatchCandidate{}
	for _, match := range matches.Matches {
		matchMap[match.RelationIndex] = match.Matches
	}

	out := make([]Phase8RelationResult, 0, len(relations))
	for idx, relation := range relations {
		out = append(out, Phase8RelationResult{
			Source:       relation.Source,
			Target:       relation.Target,
			RelationType: relation.RelationType,
			Direction:    relation.Direction,
			Summary:      relation.Summary,
			Confidence:   relation.Confidence,
			Polarity:     relation.Polarity,
			Implicit:     relation.Implicit,
			Evidence:     relation.Evidence,
			Status:       relation.Status,
			Dedup:        relation.Dedup,
			CreateMirror: relation.CreateMirror,
			MirrorOf:     relation.MirrorOf,
			Matches:      matchMap[idx],
		})
	}

	return out
}

func defaultRelationMatchSourceTypes() []memory.SourceType {
	return []memory.SourceType{
		memory.SourceTypeStory,
		memory.SourceTypeChapter,
		memory.SourceTypeScene,
		memory.SourceTypeBeat,
		memory.SourceTypeContentBlock,
	}
}

func buildRelationTypeSemantics(
	relationTypes map[string]Phase6RelationTypeDefinition,
	overrides map[string]string,
) map[string]string {
	result := map[string]string{}
	for relType, def := range relationTypes {
		if strings.TrimSpace(def.Semantics) != "" {
			result[relType] = strings.TrimSpace(def.Semantics)
		}
	}
	for relType, semantics := range overrides {
		if strings.TrimSpace(semantics) != "" {
			result[relType] = strings.TrimSpace(semantics)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func collectMentions(occurrences []Phase2EntityOccurrence) []string {
	mentions := make([]string, 0, len(occurrences))
	seen := map[string]struct{}{}
	for _, occ := range occurrences {
		mention := strings.TrimSpace(occ.Evidence)
		if mention == "" {
			continue
		}
		if _, ok := seen[mention]; ok {
			continue
		}
		seen[mention] = struct{}{}
		mentions = append(mentions, mention)
	}
	return mentions
}

func buildFindingKey(entityType string, name string) string {
	return fmt.Sprintf("%s|%s", strings.TrimSpace(entityType), strings.ToLower(strings.TrimSpace(name)))
}

func fallbackString(primary, secondary string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return secondary
}
