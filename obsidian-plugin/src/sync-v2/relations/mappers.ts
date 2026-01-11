import type { EntityRelation } from "../types/relations";
import type {
	CitationsGeneratorInput,
	ParsedCitationEntry,
	ParsedRelationEntry,
	RelationsEntityMeta,
	RelationsGeneratorInput,
	GeneratorMetadata,
} from "../types/generators";

export interface RelationTargetInfo {
	targetId: string;
	targetName: string;
	contextLabel?: string;
	summary?: string;
}

export type RelationTargetResolver = (relation: EntityRelation) => RelationTargetInfo | null;

export interface CitationSourceInfo {
	storyId: string;
	storyTitle: string;
	sourceTitle: string;
	sourceType: ParsedCitationEntry["sourceType"];
	chapterTitle?: string;
	summary?: string;
}

export type CitationSourceResolver = (relation: EntityRelation) => CitationSourceInfo | null;

export function mapRelationsToGeneratorInput({
	entity,
	relations,
	resolveTarget,
	options,
}: {
	entity: RelationsEntityMeta;
	relations: EntityRelation[];
	resolveTarget?: RelationTargetResolver;
	options?: GeneratorMetadata;
}): RelationsGeneratorInput {
	const parsed: ParsedRelationEntry[] = [];

	relations.forEach((relation) => {
		const resolved = resolveTarget?.(relation);
		if (!resolved) {
			if (!resolveTarget) {
				parsed.push({
					targetType: relation.target_type,
					targetId: relation.target_id,
					targetName: relation.target_id,
					relationType: relation.relation_type,
					summary: relation.context,
				});
			}
			return;
		}

		parsed.push({
			targetType: relation.target_type,
			targetId: resolved.targetId,
			targetName: resolved.targetName,
			relationType: relation.relation_type,
			summary: resolved.summary ?? relation.context,
			contextLabel: resolved.contextLabel,
		});
	});

	return {
		entity,
		relations: parsed,
		options,
	};
}

export function mapCitationsToGeneratorInput({
	entity,
	relations,
	resolveSource,
	options,
}: {
	entity: RelationsEntityMeta;
	relations: EntityRelation[];
	resolveSource: CitationSourceResolver;
	options?: GeneratorMetadata;
}): CitationsGeneratorInput {
	const citations: ParsedCitationEntry[] = [];

	relations.forEach((relation) => {
		const source = resolveSource(relation);
		if (!source) {
			return;
		}

		citations.push({
			storyId: source.storyId,
			storyTitle: source.storyTitle,
			relationType: relation.relation_type,
			sourceType: source.sourceType,
			sourceId: relation.source_id,
			sourceTitle: source.sourceTitle,
			chapterTitle: source.chapterTitle,
			summary: source.summary ?? relation.context,
		});
	});

	return {
		entity,
		citations,
		options,
	};
}

