import type { ContentBlock } from "../../types";
import type { SyncContext } from "../types/sync";
import { RelationsGenerator } from "../generators/RelationsGenerator";
import {
	mapRelationsToGeneratorInput,
	type RelationTargetResolver,
} from "./mappers";

interface RelationsEntityMeta {
	id: string;
	name: string;
	type: string;
	worldId?: string;
}

interface WriteRelationsOptions {
	entity: RelationsEntityMeta;
	outputPath: string;
	context: SyncContext;
	worldFolderPath?: string;
}

export async function writeRelationsFile({
	entity,
	outputPath,
	context,
	worldFolderPath,
}: WriteRelationsOptions): Promise<void> {
	const relationsGenerator = new RelationsGenerator();
	try {
		const relationsResponse = await context.apiClient.listRelationsByTarget({
			targetType: entity.type,
			targetId: entity.id,
		});

		const entityCache = new Map<string, { name: string; type: string; id: string }>();
		const resolvedRelations = await Promise.all(
			relationsResponse.data.map(async (relation: any) => {
				return resolveRelationTarget(relation, entityCache, context);
			})
		);

		const entityMap = new Map(
			resolvedRelations.map((r) => [`${r.targetType}:${r.targetId}`, r] as const)
		);

		const resolveTarget: RelationTargetResolver = (relation) => {
			const key = `${relation.source_type}:${relation.source_id}` as `${string}:${string}`;
			const resolved = entityMap.get(key);
			if (!resolved) return null;
			return {
				targetId: resolved.targetId,
				targetName: resolved.targetName,
				summary: resolved.summary,
			};
		};

		const mappedRelations = relationsResponse.data.map((rel: any) => ({
			...rel,
			target_type: rel.source_type,
			target_id: rel.source_id,
		}));

		const input = mapRelationsToGeneratorInput({
			entity,
			relations: mappedRelations,
			resolveTarget,
			options: {
				syncedAt: context.timestamp(),
				showHelpBox: context.settings.showHelpBox,
				idField: context.settings.frontmatterIdField,
				worldFolderPath,
			},
		});

		const relationsContent = relationsGenerator.generate(input);
		await context.fileManager.writeFile(outputPath, relationsContent);
	} catch (error) {
		console.warn("[Sync V2] Failed to generate relations file", {
			entity,
			error,
		});
		await context.fileManager.writeFile(outputPath, renderRelationsPlaceholder(entity.name));
	}
}

async function resolveRelationTarget(
	relation: { source_type: string; source_id: string; relation_type: string; context?: string },
	entityCache: Map<string, { name: string; type: string; id: string }>,
	context: SyncContext
): Promise<{
	targetType: string;
	targetId: string;
	targetName: string;
	relationType: string;
	summary?: string;
}> {
	try {
		let targetName = relation.source_id;
		let targetId = relation.source_id;
		let targetType = relation.source_type;

		const cacheKey = `${relation.source_type}:${relation.source_id}`;
		if (entityCache.has(cacheKey)) {
			const cached = entityCache.get(cacheKey)!;
			targetName = cached.name;
			targetId = cached.id;
			targetType = cached.type;
		} else {
			const resolved = await fetchEntityName(relation.source_type, relation.source_id, context);
			if (resolved) {
				targetName = resolved.name;
				targetId = resolved.id;
				entityCache.set(cacheKey, { name: resolved.name, type: relation.source_type, id: resolved.id });
			}
		}

		return {
			targetType,
			targetId,
			targetName,
			relationType: relation.relation_type,
			summary: relation.context,
		};
	} catch (error) {
		console.warn("[Sync V2] Failed to resolve relation target", {
			relation,
			error,
		});
		return {
			targetType: relation.source_type,
			targetId: relation.source_id,
			targetName: relation.source_id,
			relationType: relation.relation_type,
			summary: relation.context,
		};
	}
}

async function fetchEntityName(
	entityType: string,
	entityId: string,
	context: SyncContext
): Promise<{ id: string; name: string } | null> {
	switch (entityType) {
		case "character": {
			const char = await context.apiClient.getCharacter(entityId);
			return { id: char.id, name: char.name };
		}
		case "location": {
			const loc = await context.apiClient.getLocation(entityId);
			return { id: loc.id, name: loc.name };
		}
		case "faction": {
			const faction = await context.apiClient.getFaction(entityId);
			return { id: faction.id, name: faction.name };
		}
		case "artifact": {
			const artifact = await context.apiClient.getArtifact(entityId);
			return { id: artifact.id, name: artifact.name };
		}
		case "event": {
			const event = await context.apiClient.getEvent(entityId);
			return { id: event.id, name: event.name };
		}
		case "lore": {
			const lore = await context.apiClient.getLore(entityId);
			return { id: lore.id, name: lore.name };
		}
		case "world": {
			const world = await context.apiClient.getWorld(entityId);
			return { id: world.id, name: world.name };
		}
		case "chapter": {
			const chapter = await context.apiClient.getChapter(entityId);
			return { id: chapter.id, name: `Chapter ${chapter.number}: ${chapter.title}` };
		}
		case "scene": {
			const scene = await context.apiClient.getScene(entityId);
			return { id: scene.id, name: `Scene ${scene.order_num ?? 0}: ${scene.goal || "Untitled"}` };
		}
		case "beat": {
			const beat = await context.apiClient.getBeat(entityId);
			return { id: beat.id, name: `Beat ${beat.order_num ?? 0}: ${beat.intent || "Untitled"}` };
		}
		case "content_block": {
			const block: ContentBlock = await context.apiClient.getContentBlock(entityId);
			const name =
				block.metadata?.title ||
				block.kind ||
				block.type ||
				block.content?.slice(0, 40) ||
				"Content Block";
			return { id: block.id, name };
		}
		default:
			return null;
	}
}

function renderRelationsPlaceholder(name: string): string {
	return [`# ${name} - Relations`, "", "_Relations will be populated when synced._", ""].join("\n");
}
