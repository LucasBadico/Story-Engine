import type { StoryWithHierarchy } from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import { RelationsGenerator } from "../../../generators/RelationsGenerator";
import { RelationsPushHandler } from "../../../push/RelationsPushHandler";
import {
	mapRelationsToGeneratorInput,
	type RelationTargetResolver,
} from "../../../relations/mappers";

/**
 * Service responsible for generating and pushing story relations.
 * Extracted from StoryHandler for better separation of concerns.
 */
export class StoryRelationsService {
	constructor(
		private readonly relationsGenerator = new RelationsGenerator(),
		private readonly relationsPushHandler = new RelationsPushHandler()
	) {}

	/**
	 * Generate the story.relations.md file by fetching relations from API
	 * and resolving entity names.
	 */
	async generateRelationsFile(
		story: StoryWithHierarchy,
		folderPath: string,
		context: SyncContext
	): Promise<void> {
		try {
			let worldFolderPath: string | undefined;
			if (story.story.world_id) {
				try {
					const world = await context.apiClient.getWorld(story.story.world_id);
					if (world?.name) {
						worldFolderPath = context.fileManager.getWorldFolderPath(world.name);
					}
				} catch {
					// Ignore world lookup errors for relations generation
				}
			}
			// Fetch relations for this story
			const relationsResponse = await context.apiClient.listRelationsByTarget({
				targetType: "story",
				targetId: story.story.id,
			});

			// Build entity cache for efficient lookup
			const entityCache = new Map<string, { name: string; type: string }>();

			// Resolve all target names
			const resolvedRelations = await Promise.all(
				relationsResponse.data.map(async (relation) => {
					return this.resolveRelationTarget(relation, entityCache, context);
				})
			);

			// Build resolver that uses pre-resolved data
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

			// Map relations for the generator - note we use source_type/source_id as targets
			// because these are relations where the story is the target
			const mappedRelations = relationsResponse.data.map((rel) => ({
				...rel,
				target_type: rel.source_type,
				target_id: rel.source_id,
			}));

			const input = mapRelationsToGeneratorInput({
				entity: {
					id: story.story.id,
					name: story.story.title,
					type: "story",
					worldId: story.story.world_id ?? undefined,
				},
				relations: mappedRelations,
				resolveTarget,
				options: {
					syncedAt: context.timestamp(),
					showHelpBox: context.settings.showHelpBox,
					idField: context.settings.frontmatterIdField,
					worldFolderPath,
				},
			});

			const relationsContent = this.relationsGenerator.generate(input);
			await context.fileManager.writeFile(`${folderPath}/story.relations.md`, relationsContent);
		} catch (error) {
			console.warn("[Sync V2] Failed to generate story relations file", {
				storyId: story.story.id,
				error,
			});
			// Fallback to placeholder
			await context.fileManager.writeFile(
				`${folderPath}/story.relations.md`,
				this.renderRelationsPlaceholder(story)
			);
		}
	}

	/**
	 * Push local relations changes to the API.
	 */
	async pushRelations(
		entity: StoryWithHierarchy,
		folderPath: string,
		context: SyncContext
	): Promise<void> {
		const relationsFilePath = `${folderPath}/story.relations.md`;

		try {
			await context.fileManager.readFile(relationsFilePath);
			const result = await this.relationsPushHandler.pushRelations(
				relationsFilePath,
				"story",
				entity.story.id,
				context,
				entity.story.world_id ?? undefined
			);

			if (result.warnings.length > 0) {
				result.warnings.forEach((warning) =>
					context.emitWarning?.({
						code: "relations_push_warning",
						message: warning,
						filePath: relationsFilePath,
					})
				);
			}
		} catch (error: any) {
			if (!error?.message?.includes("missing") && error?.code !== "ENOENT") {
				context.emitWarning?.({
					code: "relations_push_error",
					message: `Failed to push relations: ${error}`,
					filePath: relationsFilePath,
				});
			}
		}
	}

	private async resolveRelationTarget(
		relation: { source_type: string; source_id: string; relation_type: string; context?: string },
		entityCache: Map<string, { name: string; type: string }>,
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
				targetType = cached.type;
			} else {
				const resolved = await this.fetchEntityName(relation.source_type, relation.source_id, context);
				if (resolved) {
					targetName = resolved.name;
					targetId = resolved.id;
					entityCache.set(cacheKey, { name: resolved.name, type: relation.source_type });
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
			console.warn("[Sync V2] Failed to resolve target for story relation", {
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

	private async fetchEntityName(
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
			default:
				return null;
		}
	}

	private renderRelationsPlaceholder(story: StoryWithHierarchy): string {
		return [
			`# ${story.story.title} - Relations`,
			"",
			"_Relations will be populated when synced with the server._",
			"",
		].join("\n");
	}
}
