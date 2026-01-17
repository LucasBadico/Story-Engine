import type {
	SyncContext,
	SyncOperation,
	SyncResult,
	SyncOperationType,
	SyncWarning,
} from "../types/sync";
import { EntityRegistry } from "./EntityRegistry";
import { StoryHandler } from "../handlers/story/StoryHandler";
import { ChapterHandler } from "../handlers/chapter/ChapterHandler";
import { SceneHandler } from "../handlers/scene/SceneHandler";
import { BeatHandler } from "../handlers/beat/BeatHandler";
import { ContentBlockHandler } from "../handlers/content/ContentBlockHandler";
import { WorldHandler } from "../handlers/world/WorldHandler";
import { CharacterHandler } from "../handlers/world/CharacterHandler";
import { LocationHandler } from "../handlers/world/LocationHandler";
import { FactionHandler } from "../handlers/world/FactionHandler";
import { ArtifactHandler } from "../handlers/world/ArtifactHandler";
import { EventHandler } from "../handlers/world/EventHandler";
import { LoreHandler } from "../handlers/world/LoreHandler";
import { ArchetypeHandler } from "../handlers/world/ArchetypeHandler";
import { TraitHandler } from "../handlers/world/TraitHandler";
import { ContentsGenerator } from "../generators/ContentsGenerator";
import { PushPlanner } from "../push/PushPlanner";
import { PushExecutor } from "../push/PushExecutor";
import { ContentCitationService } from "../push/ContentCitationService";
import { getFrontmatterId } from "../utils/frontmatterHelpers";
import { ConflictResolver } from "../conflict/ConflictResolver";
import { BackupManager } from "../backup/BackupManager";

export class SyncOrchestrator {
	private readonly context: SyncContext;
	private readonly registry: EntityRegistry;
	private readonly contentsGenerator: ContentsGenerator;
	private readonly pushPlanner: PushPlanner;
	private readonly pushExecutor: PushExecutor;
	private readonly conflictResolver: ConflictResolver;
	private readonly backupManager: BackupManager;
	private readonly contentCitationService: ContentCitationService;
	private warningBuffer: SyncWarning[] = [];

	constructor(context: SyncContext) {
		this.context = context;
		this.registry = new EntityRegistry();
		this.contentsGenerator = new ContentsGenerator();
		this.pushPlanner = new PushPlanner();
		this.contentCitationService = new ContentCitationService(context);
		this.pushExecutor = new PushExecutor(context.apiClient, this.contentCitationService);
		this.conflictResolver = new ConflictResolver(context.app, context);
		this.backupManager = new BackupManager(context.app);
		const originalEmitter = this.context.emitWarning;
		this.context.emitWarning = (warning: SyncWarning) => {
			this.warningBuffer.push(warning);
			originalEmitter?.(warning);
		};
		this.registerDefaultHandlers();
	}

	/**
	 * Get the ConflictResolver instance
	 */
	getConflictResolver(): ConflictResolver {
		return this.conflictResolver;
	}

	dispose(): void {
		// Future: flush pending tasks or watchers
	}

	async run(operation: SyncOperation): Promise<SyncResult> {
		this.warningBuffer = [];
		try {
			switch (operation.type) {
				case "pull_story":
					return this.attachWarnings(await this.handlePullStory(operation));
				case "pull_all_stories":
					return this.attachWarnings(await this.handlePullAllStories(operation));
				case "push_story":
					return this.attachWarnings(await this.handlePushStory(operation));
				case "pull_chapter":
					return this.attachWarnings(await this.handlePullChapter(operation));
				case "pull_scene":
					return this.attachWarnings(await this.handlePullScene(operation));
				case "pull_beat":
					return this.attachWarnings(await this.handlePullBeat(operation));
				case "pull_content_block":
					return this.attachWarnings(await this.handlePullContentBlock(operation));
				case "pull_world":
					return this.attachWarnings(await this.handlePullWorld(operation));
				case "pull_character":
					return this.attachWarnings(await this.handlePullCharacter(operation));
				case "pull_location":
					return this.attachWarnings(await this.handlePullLocation(operation));
				case "pull_faction":
					return this.attachWarnings(await this.handlePullFaction(operation));
				case "pull_artifact":
					return this.attachWarnings(await this.handlePullArtifact(operation));
				case "pull_event":
					return this.attachWarnings(await this.handlePullEvent(operation));
				case "pull_lore":
					return this.attachWarnings(await this.handlePullLore(operation));
				case "pull_archetype":
					return this.attachWarnings(await this.handlePullArchetype(operation));
				case "pull_trait":
					return this.attachWarnings(await this.handlePullTrait(operation));
			}

			throw new Error(`Unhandled sync operation: ${(operation as any).type}`);
		} catch (error) {
			return this.attachWarnings({
				success: false,
				errors: [
					{
						code: "sync_v2_unhandled_error",
						message:
							error instanceof Error
								? error.message
								: "Unknown error while executing Sync V2 operation",
						details: error,
					},
				],
			});
		}
	}

	private attachWarnings(result: SyncResult): SyncResult {
		if (this.warningBuffer.length === 0) {
			return result;
		}
		const combined = [...(result.warnings ?? []), ...this.warningBuffer];
		return {
			...result,
			warnings: combined,
		};
	}

	private async handlePullStory(operation: Extract<SyncOperation, { type: "pull_story" }>): Promise<SyncResult> {
		const { storyId } = operation.payload;
		const handler = this.registry.get<StoryHandler>("story");
		if (!handler) {
			return this.missingHandler("story");
		}
		await this.backupStoryBeforePull(storyId);
		await handler.pull(storyId, this.context);
		return {
			success: true,
			message: "Story pulled successfully (handlers in progress).",
		};
	}

	private async handlePullAllStories(
		operation: Extract<SyncOperation, { type: "pull_all_stories" }>
	): Promise<SyncResult> {
		const stories = await this.context.apiClient.listStories();
		return {
			success: false,
			message: "Sync V2 bulk pull is not yet implemented.",
			errors: [
				{
					code: "sync_v2_not_implemented",
					message: "Bulk pull still pending implementation",
					details: {
						requestedWorlds: operation.payload.includeWorlds ?? false,
						storyCount: stories.length,
					},
					recoverable: true,
				},
			],
		};
	}

	private async handlePushStory(operation: Extract<SyncOperation, { type: "push_story" }>): Promise<SyncResult> {
		const { folderPath } = operation.payload;
		await this.backupStoryFolder(folderPath, "push");
		try {
			// Pass configured ID field name to readStoryMetadata
			const idField = this.context.settings.frontmatterIdField;
			const metadata = await this.context.fileManager.readStoryMetadata(folderPath, idField);
			const storyId = metadata.frontmatter.id;
			if (!storyId) {
				return {
					success: false,
					errors: [
						{
							code: "sync_v2_missing_story_id",
							message: `Could not determine story ID from story.md frontmatter (using field: ${idField || "id"}).`,
							details: { folderPath, idField: idField || "id" },
							recoverable: false,
						},
					],
				};
			}

			const localContentsPath = `${folderPath}/story.contents.md`;
			const localContents = await this.context.fileManager.readFile(localContentsPath);

			const remoteStory = await this.context.apiClient.getStoryWithHierarchy(storyId);
			const remoteContents = this.contentsGenerator.generateStoryContents({
				story: remoteStory.story,
				chapters: remoteStory.chapters,
				options: { 
					syncedAt: this.context.timestamp(),
					idField: this.context.settings.frontmatterIdField,
				},
			});

			const plan = this.pushPlanner.buildPlan(remoteContents, localContents);

			if (plan.actions.length === 0) {
				const hasWarnings =
					plan.unsupportedOperations.length > 0 || plan.untrackedSegments.length > 0;
				return {
					success: true,
					message: hasWarnings
						? "Push skipped: only unsupported edits detected (placeholders or free text)."
						: "Push skipped: no structural changes detected.",
					stats: { skipped: 0 },
					warnings: plan.warnings,
				};
			}

			const execution = await this.pushExecutor.execute(plan.actions, {
				worldId: remoteStory.story.world_id ?? undefined,
			});
			const success = execution.errors.length === 0;
			const message = success
				? `Push applied ${execution.applied} structural updates.`
				: `Push applied ${execution.applied} updates. ${execution.errors.length} actions failed.`;

			return {
				success,
				message,
				stats: {
					updated: execution.applied,
					skipped: plan.actions.length - execution.applied,
				},
				errors: execution.errors.length ? execution.errors : undefined,
				warnings: plan.warnings,
			};
		} catch (error) {
			return {
				success: false,
				errors: [
					{
						code: "sync_v2_push_failed",
						message:
							error instanceof Error ? error.message : "Unknown error while pushing story.",
						details: {
							folderPath,
						},
						recoverable: false,
					},
				],
			};
		}
	}

	private async backupStoryBeforePull(storyId: string): Promise<void> {
		if (this.context.backupMode === "off") {
			return;
		}

		try {
			const story = await this.context.apiClient.getStory(storyId);
			if (!story?.title) {
				return;
			}
			const folderPath = this.context.fileManager.getStoryFolderPath(story.title);
			await this.backupStoryFolder(folderPath, "pull");
		} catch (error) {
			this.context.emitWarning?.({
				code: "backup_failed",
				message: `Failed to backup story before pull: ${error instanceof Error ? error.message : String(error)}`,
				details: error,
			});
		}
	}

	private async backupStoryFolder(folderPath: string, operation: "pull" | "push"): Promise<void> {
		if (this.context.backupMode === "off") {
			return;
		}

		try {
			const filesToBackup = this.backupManager.getStoryFilesForBackup(folderPath);
			const result = await this.backupManager.createBackup(filesToBackup, operation);
			if (!result.success) {
				this.context.emitWarning?.({
					code: "backup_partial",
					message: `Backup parcial: ${result.filesCopied.length} arquivos salvos, ${result.errors.length} erros`,
					details: result.errors,
				});
			}
		} catch (error) {
			this.context.emitWarning?.({
				code: "backup_failed",
				message: `Failed to create backup: ${error instanceof Error ? error.message : String(error)}`,
				details: error,
			});
		}
	}

	private notImplemented(operation: SyncOperationType): SyncResult {
		return {
			success: false,
			errors: [
				{
					code: "sync_v2_not_implemented",
					message: `No handler registered for operation "${operation}".`,
					recoverable: true,
				},
			],
		};
	}

	private missingHandler(entityType: string): SyncResult {
		return {
			success: false,
			errors: [
				{
					code: "sync_v2_missing_handler",
					message: `No handler registered for entity "${entityType}".`,
					recoverable: false,
				},
			],
		};
	}

	private registerDefaultHandlers(): void {
		this.registry.register("story", new StoryHandler());
		this.registry.register("chapter", new ChapterHandler());
		this.registry.register("scene", new SceneHandler());
		this.registry.register("beat", new BeatHandler());
		this.registry.register("content_block", new ContentBlockHandler());
		this.registry.register("world", new WorldHandler());
		this.registry.register("character", new CharacterHandler());
		this.registry.register("location", new LocationHandler());
		this.registry.register("faction", new FactionHandler());
		this.registry.register("artifact", new ArtifactHandler());
		this.registry.register("event", new EventHandler());
		this.registry.register("lore", new LoreHandler());
		this.registry.register("archetype", new ArchetypeHandler());
		this.registry.register("trait", new TraitHandler());
	}

	private async handlePullChapter(operation: Extract<SyncOperation, { type: "pull_chapter" }>): Promise<SyncResult> {
		const handler = this.registry.get<ChapterHandler>("chapter");
		if (!handler) {
			return this.missingHandler("chapter");
		}
		await handler.pull(operation.payload.chapterId, this.context);
		return { success: true, message: "Chapter pulled successfully." };
	}

	private async handlePullScene(operation: Extract<SyncOperation, { type: "pull_scene" }>): Promise<SyncResult> {
		const handler = this.registry.get<SceneHandler>("scene");
		if (!handler) {
			return this.missingHandler("scene");
		}
		await handler.pull(operation.payload.sceneId, this.context);
		return { success: true, message: "Scene pulled successfully." };
	}

	private async handlePullBeat(operation: Extract<SyncOperation, { type: "pull_beat" }>): Promise<SyncResult> {
		const handler = this.registry.get<BeatHandler>("beat");
		if (!handler) {
			return this.missingHandler("beat");
		}
		await handler.pull(operation.payload.beatId, this.context);
		return { success: true, message: "Beat pulled successfully." };
	}

	private async handlePullContentBlock(
		operation: Extract<SyncOperation, { type: "pull_content_block" }>
	): Promise<SyncResult> {
		const handler = this.registry.get<ContentBlockHandler>("content_block");
		if (!handler) {
			return this.missingHandler("content_block");
		}
		await handler.pull(operation.payload.contentBlockId, this.context);
		return { success: true, message: "Content block pulled successfully." };
	}

	private async handlePullWorld(operation: Extract<SyncOperation, { type: "pull_world" }>): Promise<SyncResult> {
		const handler = this.registry.get<WorldHandler>("world");
		if (!handler) {
			return this.missingHandler("world");
		}
		await handler.pull(operation.payload.worldId, this.context);
		return { success: true, message: "World pulled successfully." };
	}

	private async handlePullCharacter(
		operation: Extract<SyncOperation, { type: "pull_character" }>
	): Promise<SyncResult> {
		const handler = this.registry.get<CharacterHandler>("character");
		if (!handler) {
			return this.missingHandler("character");
		}
		await handler.pull(operation.payload.entityId, this.context);
		return { success: true, message: "Character pulled successfully." };
	}

	private async handlePullLocation(
		operation: Extract<SyncOperation, { type: "pull_location" }>
	): Promise<SyncResult> {
		const handler = this.registry.get<LocationHandler>("location");
		if (!handler) {
			return this.missingHandler("location");
		}
		await handler.pull(operation.payload.entityId, this.context);
		return { success: true, message: "Location pulled successfully." };
	}

	private async handlePullFaction(
		operation: Extract<SyncOperation, { type: "pull_faction" }>
	): Promise<SyncResult> {
		const handler = this.registry.get<FactionHandler>("faction");
		if (!handler) {
			return this.missingHandler("faction");
		}
		await handler.pull(operation.payload.entityId, this.context);
		return { success: true, message: "Faction pulled successfully." };
	}

	private async handlePullArtifact(
		operation: Extract<SyncOperation, { type: "pull_artifact" }>
	): Promise<SyncResult> {
		const handler = this.registry.get<ArtifactHandler>("artifact");
		if (!handler) {
			return this.missingHandler("artifact");
		}
		await handler.pull(operation.payload.entityId, this.context);
		return { success: true, message: "Artifact pulled successfully." };
	}

	private async handlePullEvent(operation: Extract<SyncOperation, { type: "pull_event" }>): Promise<SyncResult> {
		const handler = this.registry.get<EventHandler>("event");
		if (!handler) {
			return this.missingHandler("event");
		}
		await handler.pull(operation.payload.entityId, this.context);
		return { success: true, message: "Event pulled successfully." };
	}

	private async handlePullLore(operation: Extract<SyncOperation, { type: "pull_lore" }>): Promise<SyncResult> {
		const handler = this.registry.get<LoreHandler>("lore");
		if (!handler) {
			return this.missingHandler("lore");
		}
		await handler.pull(operation.payload.entityId, this.context);
		return { success: true, message: "Lore pulled successfully." };
	}

	private async handlePullArchetype(
		operation: Extract<SyncOperation, { type: "pull_archetype" }>
	): Promise<SyncResult> {
		const handler = this.registry.get<ArchetypeHandler>("archetype");
		if (!handler) {
			return this.missingHandler("archetype");
		}
		await handler.pull(operation.payload.entityId, this.context);
		return { success: true, message: "Archetype pulled successfully." };
	}

	private async handlePullTrait(
		operation: Extract<SyncOperation, { type: "pull_trait" }>
	): Promise<SyncResult> {
		const handler = this.registry.get<TraitHandler>("trait");
		if (!handler) {
			return this.missingHandler("trait");
		}
		await handler.pull(operation.payload.entityId, this.context);
		return { success: true, message: "Trait pulled successfully." };
	}
}

