import type { ContentBlock, StoryWithHierarchy } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { OutlineGenerator } from "../../generators/OutlineGenerator";
import { ContentsGenerator } from "../../generators/ContentsGenerator";
import { ContentsReconciler } from "../../diff/ContentsReconciler";
import { OutlineReconciler } from "../../diff/OutlineReconciler";
import { FileRenamer } from "../../fileRenamer/FileRenamer";
import { PathResolver } from "../../fileRenamer/PathResolver";
import { RelationsGenerator } from "../../generators/RelationsGenerator";
import {
	mapRelationsToGeneratorInput,
} from "../../relations/mappers";
import type { RelationTargetResolver } from "../../relations/mappers";
import { RelationsPushHandler } from "../../push/RelationsPushHandler";
import { ConflictResolver } from "../../conflict/ConflictResolver";
import { parseFrontmatter } from "../../utils/detectEntityMentions";

export class StoryHandler {
	readonly entityType = "story";
	private readonly contentBlockCache = new Map<string, ContentBlock>();

	constructor(
		private readonly outlineGenerator = new OutlineGenerator(),
		private readonly contentsGenerator = new ContentsGenerator(),
		private readonly contentsReconciler = new ContentsReconciler(),
		private readonly outlineReconciler = new OutlineReconciler(),
		private readonly relationsGenerator = new RelationsGenerator(),
		private readonly relationsPushHandler = new RelationsPushHandler(),
		private readonly fileRenamerFactory: (context: SyncContext) => FileRenamer = (context) =>
			new FileRenamer(context),
		private readonly conflictResolverFactory: (context: SyncContext) => ConflictResolver = (context) =>
			new ConflictResolver(context.app, context)
	) {}

	async pull(id: string, context: SyncContext): Promise<StoryWithHierarchy> {
		const story = await context.apiClient.getStoryWithHierarchy(id);
		const folderPath = context.fileManager.getStoryFolderPath(story.story.title);
		await context.fileManager.ensureFolderExists(folderPath);
		
		const conflictResolver = this.conflictResolverFactory(context);

		// Check for conflicts before writing story metadata
		const storyMetadataPath = `${folderPath}/story.md`;
		let existingStoryMetadata: string | null = null;
		let localStoryTimestamp: string | undefined;
		try {
			existingStoryMetadata = await context.fileManager.readFile(storyMetadataPath);
			const parsed = parseFrontmatter(existingStoryMetadata);
			// Use updated_at from frontmatter if available, otherwise use synced_at
			localStoryTimestamp = (parsed.updated_at as string | undefined) || (parsed.synced_at as string | undefined);
		} catch {
			existingStoryMetadata = null;
		}

		// Detect conflict for story metadata
		if (existingStoryMetadata && localStoryTimestamp) {
			const conflict = conflictResolver.detectConflict(
				"story",
				story.story.id,
				storyMetadataPath,
				{ updated_at: localStoryTimestamp },
				{ updated_at: story.story.updated_at },
				localStoryTimestamp,
				story.story.updated_at
			);

			if (conflict) {
				const resolution = await conflictResolver.resolve(conflict);
				if (!resolution.success) {
					context.emitWarning?.({
						code: "conflict_resolution_failed",
						message: `Failed to resolve conflict for story ${story.story.id}: ${resolution.error || "Unknown error"}`,
						filePath: storyMetadataPath,
						severity: "warning",
					});
				} else if (!resolution.resolution.autoResolved) {
					// Manual resolution required - log warning
					context.emitWarning?.({
						code: "conflict_requires_manual_resolution",
						message: `Conflict detected for story ${story.story.id}. Manual resolution may be required.`,
						filePath: storyMetadataPath,
						severity: "warning",
						details: conflict,
					});
				}
			}
		}

		await context.fileManager.writeStoryMetadata(
			story.story,
			folderPath,
			story.chapters
		);

		const outlinePath = `${folderPath}/story.outline.md`;
		const outlineGenerated = this.outlineGenerator.generateStoryOutline(story, {
			syncedAt: context.timestamp(),
			showHelpBox: context.settings.showHelpBox,
			idField: context.settings.frontmatterIdField,
		});
		let existingOutline: string | null = null;
		let localOutlineTimestamp: string | undefined;
		try {
			existingOutline = await context.fileManager.readFile(outlinePath);
			const parsed = parseFrontmatter(existingOutline);
			localOutlineTimestamp = parsed.synced_at as string | undefined;
		} catch {
			existingOutline = null;
		}

		// Detect conflict for outline
		if (existingOutline && localOutlineTimestamp) {
			const conflict = conflictResolver.detectConflict(
				"story-outline",
				story.story.id,
				outlinePath,
				{ synced_at: localOutlineTimestamp, content: existingOutline },
				{ updated_at: story.story.updated_at, content: outlineGenerated },
				localOutlineTimestamp,
				story.story.updated_at
			);

			if (conflict) {
				const resolution = await conflictResolver.resolve(conflict);
				if (!resolution.success) {
					context.emitWarning?.({
						code: "conflict_resolution_failed",
						message: `Failed to resolve conflict for story outline ${story.story.id}: ${resolution.error || "Unknown error"}`,
						filePath: outlinePath,
						severity: "warning",
					});
				} else if (!resolution.resolution.autoResolved) {
					context.emitWarning?.({
						code: "conflict_requires_manual_resolution",
						message: `Conflict detected for story outline ${story.story.id}. Manual resolution may be required.`,
						filePath: outlinePath,
						severity: "warning",
						details: conflict,
					});
				}
			}
		}

		const outlineMerged = this.outlineReconciler.reconcile(existingOutline, outlineGenerated);
		const contentsPath = `${folderPath}/story.contents.md`;
		let existingContents: string | null = null;
		let localContentsTimestamp: string | undefined;
		try {
			existingContents = await context.fileManager.readFile(contentsPath);
			const parsed = parseFrontmatter(existingContents);
			localContentsTimestamp = parsed.synced_at as string | undefined;
		} catch {
			existingContents = null;
		}

		const generatedContents = this.contentsGenerator.generateStoryContents({
			story: story.story,
			chapters: story.chapters,
			options: { 
				syncedAt: context.timestamp(),
				idField: context.settings.frontmatterIdField,
			},
		});

		// Detect conflict for contents
		if (existingContents && localContentsTimestamp) {
			const conflict = conflictResolver.detectConflict(
				"story-contents",
				story.story.id,
				contentsPath,
				{ synced_at: localContentsTimestamp, content: existingContents },
				{ updated_at: story.story.updated_at, content: generatedContents },
				localContentsTimestamp,
				story.story.updated_at
			);

			if (conflict) {
				const resolution = await conflictResolver.resolve(conflict);
				if (!resolution.success) {
					context.emitWarning?.({
						code: "conflict_resolution_failed",
						message: `Failed to resolve conflict for story contents ${story.story.id}: ${resolution.error || "Unknown error"}`,
						filePath: contentsPath,
						severity: "warning",
					});
				} else if (!resolution.resolution.autoResolved) {
					context.emitWarning?.({
						code: "conflict_requires_manual_resolution",
						message: `Conflict detected for story contents ${story.story.id}. Manual resolution may be required.`,
						filePath: contentsPath,
						severity: "warning",
						details: conflict,
					});
				}
			}
		}

		const reconciled = this.contentsReconciler.reconcile(existingContents, generatedContents);
		if (reconciled.warnings.length) {
			reconciled.warnings.forEach((warning) =>
				context.emitWarning?.({
					...warning,
					filePath: contentsPath,
				})
			);
		}

		await context.fileManager.writeFile(outlinePath, outlineMerged);
		await context.fileManager.writeFile(contentsPath, reconciled.mergedContent);

		await this.handleReorders(
			reconciled.diff.operations,
			story,
			folderPath,
			context
		);

		// Generate relations file (citations are only for world entities)
		await this.generateRelations(story, folderPath, context);

		return story;
	}

	async push(entity: StoryWithHierarchy, context: SyncContext): Promise<void> {
		// Push relations if relations file exists
		const folderPath = context.fileManager.getStoryFolderPath(entity.story.title);
		const relationsFilePath = `${folderPath}/story.relations.md`;

		try {
			await context.fileManager.readFile(relationsFilePath);
			// File exists, push relations
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
			// File doesn't exist or error reading - skip push
			if (error?.message?.includes("missing") || error?.code === "ENOENT") {
				// File doesn't exist, that's fine - no relations to push
				return;
			}
			// Other error - log warning
			context.emitWarning?.({
				code: "relations_push_error",
				message: `Failed to push relations: ${error}`,
				filePath: relationsFilePath,
			});
		}

		// TODO: parse outline/contents and push updates upstream
	}

	async delete(_id: string, _context: SyncContext): Promise<void> {
		// TODO: remove story directory
	}

	private async handleReorders(
		operations: ReturnType<ContentsReconciler["reconcile"]>["diff"]["operations"],
		story: StoryWithHierarchy,
		folderPath: string,
		context: SyncContext
	): Promise<void> {
		const renamer = this.fileRenamerFactory(context);
		const pathResolver = new PathResolver(folderPath);
		const sceneMap = new Map(
			story.chapters.flatMap((chapter) => chapter.scenes.map((scene) => [scene.scene.id, scene.scene] as const))
		);
		const beatMap = new Map(
			story.chapters.flatMap((chapter) =>
				chapter.scenes.flatMap((scene) => scene.beats.map((beat) => [beat.id, beat] as const))
			)
		);

		const chapterMap = new Map(story.chapters.map((chapter) => [chapter.chapter.id, chapter.chapter] as const));

		for (const op of operations) {
			if (op.kind !== "reordered") continue;
			if (op.fenceType === "chapter") {
				const chapter = chapterMap.get(op.fenceId);
				if (!chapter || op.metadata?.newOrder === undefined || op.metadata.oldOrder === undefined) {
					continue;
				}
				const oldPath = pathResolver.getChapterPath(chapter, { order: op.metadata.oldOrder });
				const newPath = pathResolver.getChapterPath(chapter, { order: op.metadata.newOrder });
				if (oldPath === newPath) continue;
				await this.safeRename(renamer, oldPath, newPath, "chapter");
			} else if (op.fenceType === "scene") {
				const scene = sceneMap.get(op.fenceId);
				if (!scene || op.metadata?.newOrder === undefined || op.metadata.oldOrder === undefined) continue;
				const oldPath = pathResolver.getScenePath(scene, { order: op.metadata.oldOrder });
				const newPath = pathResolver.getScenePath(scene, { order: op.metadata.newOrder });
				if (oldPath === newPath) continue;
				await this.safeRename(renamer, oldPath, newPath, "scene");
			} else if (op.fenceType === "beat") {
				const beat = beatMap.get(op.fenceId);
				if (!beat || op.metadata?.newOrder === undefined || op.metadata.oldOrder === undefined) continue;
				const oldPath = pathResolver.getBeatPath(beat, { order: op.metadata.oldOrder });
				const newPath = pathResolver.getBeatPath(beat, { order: op.metadata.newOrder });
				if (oldPath === newPath) continue;
				await this.safeRename(renamer, oldPath, newPath, "beat");
			} else if (op.fenceType === "content") {
				if (op.metadata?.newOrder === undefined || op.metadata.oldOrder === undefined) {
					continue;
				}
				const block = await this.getContentBlock(op.fenceId, context);
				if (!block) continue;
				const oldPath = pathResolver.getContentBlockPath(block, { order: op.metadata.oldOrder });
				const newPath = pathResolver.getContentBlockPath(block, { order: op.metadata.newOrder });
				if (oldPath === newPath) continue;
				await this.safeRename(renamer, oldPath, newPath, "content");
			}
		}
	}

	private async getContentBlock(id: string, context: SyncContext): Promise<ContentBlock | null> {
		if (this.contentBlockCache.has(id)) {
			return this.contentBlockCache.get(id)!;
		}
		try {
			const block = await context.apiClient.getContentBlock(id);
			this.contentBlockCache.set(id, block);
			return block;
		} catch (error) {
			console.warn("[Sync V2] Failed to load content block", { id, error });
			return null;
		}
	}

	private async safeRename(
		renamer: FileRenamer,
		oldPath: string,
		newPath: string,
		entity: string
	): Promise<void> {
		try {
			await renamer.rename({
				oldPath,
				newPath,
			});
		} catch (err) {
			console.warn(`[Sync V2] Failed to rename ${entity} file`, err);
		}
	}

	private async generateRelations(
		story: StoryWithHierarchy,
		folderPath: string,
		context: SyncContext
	): Promise<void> {
		try {
			// Fetch relations where story is the target (other entities relate to this story)
			const relationsResponse = await context.apiClient.listRelationsByTarget({
				targetType: "story",
				targetId: story.story.id,
			});

			// Resolve all target names asynchronously
			const resolvedRelations = await Promise.all(
				relationsResponse.data.map(async (relation) => {
					try {
						let targetName = relation.source_id;
						let targetId = relation.source_id;

						switch (relation.source_type) {
							case "character": {
								const char = await context.apiClient.getCharacter(relation.source_id);
								targetName = char.name;
								targetId = char.id;
								break;
							}
							case "location": {
								const loc = await context.apiClient.getLocation(relation.source_id);
								targetName = loc.name;
								targetId = loc.id;
								break;
							}
							case "faction": {
								const faction = await context.apiClient.getFaction(relation.source_id);
								targetName = faction.name;
								targetId = faction.id;
								break;
							}
							case "artifact": {
								const artifact = await context.apiClient.getArtifact(relation.source_id);
								targetName = artifact.name;
								targetId = artifact.id;
								break;
							}
							case "event": {
								const event = await context.apiClient.getEvent(relation.source_id);
								targetName = event.name;
								targetId = event.id;
								break;
							}
							case "lore": {
								const lore = await context.apiClient.getLore(relation.source_id);
								targetName = lore.name;
								targetId = lore.id;
								break;
							}
						}

						return {
							targetType: relation.source_type,
							targetId,
							targetName,
							relationType: relation.relation_type,
							summary: relation.context,
						};
					} catch (error) {
						console.warn(`[Sync V2] Failed to resolve target for relation`, {
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
				})
			);

			// Build resolver that uses pre-resolved data
			// Map relations by source_type:source_id to find resolved names
			const entityMap = new Map<string, typeof resolvedRelations[0]>();
			relationsResponse.data.forEach((rel, idx) => {
				const key = `${rel.source_type}:${rel.source_id}`;
				entityMap.set(key, resolvedRelations[idx]);
			});
			const resolveTarget: RelationTargetResolver = (relation) => {
				const key = `${relation.source_type}:${relation.source_id}`;
				const resolved = entityMap.get(key);
				if (!resolved) return null;

				return {
					targetId: resolved.targetId,
					targetName: resolved.targetName,
					summary: resolved.summary,
				};
			};

			// Fetch world name if world_id is available
			let worldName: string | undefined;
			if (story.story.world_id) {
				try {
					const world = await context.apiClient.getWorld(story.story.world_id);
					worldName = world.name;
				} catch (error) {
					console.warn("[Sync V2] Failed to fetch world name", {
						worldId: story.story.world_id,
						error,
					});
				}
			}

			const input = mapRelationsToGeneratorInput({
				entity: {
					id: story.story.id,
					name: story.story.title,
					type: "story",
					worldId: story.story.world_id ?? undefined,
					worldName,
				},
				relations: relationsResponse.data,
				resolveTarget,
				options: {
					syncedAt: context.timestamp(),
					showHelpBox: context.settings.showHelpBox,
					idField: context.settings.frontmatterIdField,
				},
			});

			const relationsContent = this.relationsGenerator.generate(input);
			await context.fileManager.writeFile(`${folderPath}/story.relations.md`, relationsContent);
		} catch (error) {
			console.warn("[Sync V2] Failed to generate relations file", { storyId: story.story.id, error });
		}
	}

}

