package entity_extraction

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/core/memory"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
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
	eventLogger := normalizeEventLogger(input.EventLogger)
	requestID := strings.TrimSpace(input.RequestID)
	if requestID == "" {
		requestID = uuid.NewString()
	}

	findingRefs, entityFindings := buildPhase5Findings(phase2Output.Findings)
	confirmedMatches, refMap := buildPhase5MatchesAndRefMap(findingRefs, phase3Output.Results)
	pairs := buildRelationDiscoveryPairs(entityFindings)

	emitEvent(ctx, eventLogger, ExtractionEvent{
		Type:    "phase.start",
		Phase:   "relation.discovery",
		Message: "relation discovery started",
		Data: map[string]interface{}{
			"pairs": len(pairs),
		},
	})

	relations := make([]Phase5RelationCandidate, 0)
	seenRelations := map[string]struct{}{}
	var relationsMu sync.Mutex
	var errMu sync.Mutex
	var firstErr error

	parallelism := getRelationDiscoveryParallelism(u.logger)
	sem := make(chan struct{}, parallelism)
	var wg sync.WaitGroup

	for _, pair := range pairs {
		pair := pair
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()

			emitEvent(ctx, eventLogger, ExtractionEvent{
				Type:    "relation.discovery.batch",
				Phase:   "relation.discovery",
				Message: "relation discovery batch",
				Data: map[string]interface{}{
					"source_type": pair.SourceType,
					"target_type": pair.TargetType,
				},
			})

			suggested, ok := input.SuggestedRelations[pair.SourceType]
			if !ok {
				return
			}

			filteredPhase2 := filterPhase2FindingsByTypes(phase2Output.Findings, pair.SourceType, pair.TargetType)
			textSpec := buildPhase5TextSpec(filteredPhase2, strings.TrimSpace(input.Text))
			if len(textSpec.Spans) == 0 {
				return
			}

			filteredFindings, refSet := filterPhase5FindingsByTypes(entityFindings, pair.SourceType, pair.TargetType)
			if len(filteredFindings) == 0 {
				return
			}

			filteredMatches := filterPhase5ConfirmedMatchesByRefs(confirmedMatches, refSet)

			phase5Input := Phase5RelationDiscoveryInput{
				RequestID:                      requestID,
				Context:                        buildPhase5Context(input),
				Text:                           textSpec,
				EntityFindings:                 filteredFindings,
				ConfirmedMatches:               filteredMatches,
				SuggestedRelationsBySourceType: map[string]Phase5PerEntityRelationMap{pair.SourceType: suggested},
				RelationTypeSemantics:          buildRelationTypeSemantics(input.RelationTypes, input.RelationTypeSemantics),
			}

			phase5Output, err := u.relationDiscovery.Execute(ctx, phase5Input)
			if err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				errMu.Unlock()
				return
			}

			relationsMu.Lock()
			for _, relation := range phase5Output.Relations {
				key := buildPhase5RelationKey(relation)
				if _, ok := seenRelations[key]; ok {
					continue
				}
				seenRelations[key] = struct{}{}
				relations = append(relations, relation)
			}
			relationsMu.Unlock()
		}()
	}

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	emitEvent(ctx, eventLogger, ExtractionEvent{
		Type:    "phase.done",
		Phase:   "relation.discovery",
		Message: "relation discovery finished",
		Data: map[string]interface{}{
			"relations": len(relations),
		},
	})

	if len(relations) == 0 {
		return []Phase8RelationResult{}, nil
	}

	emitEvent(ctx, eventLogger, ExtractionEvent{
		Type:    "phase.start",
		Phase:   "relation.normalize",
		Message: "relation normalization started",
		Data: map[string]interface{}{
			"relations": len(relations),
		},
	})
	phase6Output, err := u.relationNormalize.Execute(ctx, Phase6RelationNormalizeInput{
		RequestID:                      requestID,
		Context:                        buildPhase5Context(input),
		Relations:                      relations,
		RefMap:                         refMap,
		SuggestedRelationsBySourceType: input.SuggestedRelations,
		RelationTypes:                  input.RelationTypes,
	})
	if err != nil {
		return nil, err
	}
	emitEvent(ctx, eventLogger, ExtractionEvent{
		Type:    "phase.done",
		Phase:   "relation.normalize",
		Message: "relation normalization finished",
		Data: map[string]interface{}{
			"relations": len(phase6Output.Relations),
		},
	})

	emitEvent(ctx, eventLogger, ExtractionEvent{
		Type:    "phase.start",
		Phase:   "relation.match",
		Message: "relation match started",
		Data: map[string]interface{}{
			"relations": len(phase6Output.Relations),
		},
	})
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
	emitEvent(ctx, eventLogger, ExtractionEvent{
		Type:    "phase.done",
		Phase:   "relation.match",
		Message: "relation match finished",
		Data: map[string]interface{}{
			"matched_relations": len(matchOutput.Matches),
		},
	})

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

type relationDiscoveryPair struct {
	SourceType string
	TargetType string
}

func buildRelationDiscoveryPairs(findings []Phase5EntityFinding) []relationDiscoveryPair {
	present := map[string]struct{}{}
	for _, finding := range findings {
		if strings.TrimSpace(finding.Type) == "" {
			continue
		}
		present[finding.Type] = struct{}{}
	}

	typesOrder := []string{"character", "location", "faction", "event", "artefact"}
	pairs := make([]relationDiscoveryPair, 0)
	for _, sourceType := range typesOrder {
		if _, ok := present[sourceType]; !ok {
			continue
		}
		for _, targetType := range typesOrder {
			if _, ok := present[targetType]; !ok {
				continue
			}
			pairs = append(pairs, relationDiscoveryPair{
				SourceType: sourceType,
				TargetType: targetType,
			})
		}
	}
	return pairs
}

func filterPhase2FindingsByTypes(findings []Phase2EntityFinding, sourceType, targetType string) []Phase2EntityFinding {
	out := make([]Phase2EntityFinding, 0, len(findings))
	for _, finding := range findings {
		entityType := strings.TrimSpace(finding.EntityType)
		if entityType == sourceType || entityType == targetType {
			out = append(out, finding)
		}
	}
	return out
}

func filterPhase5FindingsByTypes(
	findings []Phase5EntityFinding,
	sourceType, targetType string,
) ([]Phase5EntityFinding, map[string]struct{}) {
	out := make([]Phase5EntityFinding, 0, len(findings))
	refs := make(map[string]struct{})
	for _, finding := range findings {
		entityType := strings.TrimSpace(finding.Type)
		if entityType != sourceType && entityType != targetType {
			continue
		}
		out = append(out, finding)
		refs[finding.Ref] = struct{}{}
	}
	return out, refs
}

func filterPhase5ConfirmedMatchesByRefs(
	matches []Phase5ConfirmedMatch,
	refs map[string]struct{},
) []Phase5ConfirmedMatch {
	if len(matches) == 0 || len(refs) == 0 {
		return nil
	}
	out := make([]Phase5ConfirmedMatch, 0, len(matches))
	for _, match := range matches {
		if _, ok := refs[match.FindingRef]; !ok {
			continue
		}
		out = append(out, match)
	}
	return out
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

func buildPhase5RelationKey(relation Phase5RelationCandidate) string {
	return fmt.Sprintf(
		"%s|%s|%s|%s|%s",
		relation.Source.Ref,
		relation.Target.Ref,
		relation.RelationType,
		relation.Evidence.SpanID,
		strings.TrimSpace(relation.Evidence.Quote),
	)
}

func buildPhase8Relations(relations []Phase6NormalizedRelation, matches Phase7RelationMatchOutput) []Phase8RelationResult {
	if len(relations) == 0 {
		return []Phase8RelationResult{}
	}

	matchMap := map[int][]Phase7RelationMatchCandidate{}
	for _, match := range matches.Matches {
		matchMap[match.RelationIndex] = match.Matches
	}

	ordered := orderRelationsWithMirrors(relations)
	out := make([]Phase8RelationResult, 0, len(relations))
	for _, idx := range ordered {
		relation := relations[idx]
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

func orderRelationsWithMirrors(relations []Phase6NormalizedRelation) []int {
	ordered := make([]int, 0, len(relations))
	mirrorBuckets := map[string][]int{}
	seen := make(map[int]struct{}, len(relations))

	for idx, rel := range relations {
		if rel.MirrorOf != nil && strings.TrimSpace(*rel.MirrorOf) != "" {
			key := strings.TrimSpace(*rel.MirrorOf)
			mirrorBuckets[key] = append(mirrorBuckets[key], idx)
			continue
		}
		ordered = append(ordered, idx)
		seen[idx] = struct{}{}
	}

	for _, idx := range ordered {
		key := phase6RelationKey(relations[idx])
		mirrors := mirrorBuckets[key]
		if len(mirrors) == 0 {
			continue
		}
		for _, mirrorIdx := range mirrors {
			if _, ok := seen[mirrorIdx]; ok {
				continue
			}
			seen[mirrorIdx] = struct{}{}
			ordered = append(ordered, mirrorIdx)
		}
	}

	for key, mirrors := range mirrorBuckets {
		_ = key
		for _, mirrorIdx := range mirrors {
			if _, ok := seen[mirrorIdx]; ok {
				continue
			}
			seen[mirrorIdx] = struct{}{}
			ordered = append(ordered, mirrorIdx)
		}
	}

	return ordered
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

func getRelationDiscoveryParallelism(log *logger.Logger) int {
	parallelism := 2
	cpuCount := runtime.NumCPU()

	if value := strings.TrimSpace(os.Getenv("RELATION_DISCOVERY_PARALLELISM")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			parallelism = parsed
		}
	} else if value := strings.TrimSpace(os.Getenv("ENTITY_EXTRACT_PARALLELISM")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			parallelism = parsed
		}
	}

	if log != nil && parallelism > cpuCount {
		log.Warn(
			"relation discovery parallelism exceeds CPU count",
			"parallelism", parallelism,
			"cpu_count", cpuCount,
		)
	}

	if parallelism < 1 {
		parallelism = 1
	}

	return parallelism
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
