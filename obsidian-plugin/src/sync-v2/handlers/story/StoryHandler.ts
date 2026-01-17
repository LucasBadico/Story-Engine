import type { StoryWithHierarchy } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { OutlineGenerator } from "../../generators/OutlineGenerator";
import { ContentsGenerator } from "../../generators/ContentsGenerator";
import { ContentsReconciler } from "../../diff/ContentsReconciler";
import { OutlineReconciler } from "../../diff/OutlineReconciler";
import { FileRenamer } from "../../fileRenamer/FileRenamer";
import { RelationsGenerator } from "../../generators/RelationsGenerator";
import { RelationsPushHandler } from "../../push/RelationsPushHandler";
import { OutlinePushHandler } from "../../push/OutlinePushHandler";
import { PushPlanner } from "../../push/PushPlanner";
import { PushExecutor } from "../../push/PushExecutor";
import { ContentCitationService } from "../../push/ContentCitationService";
import { StoryRelationsService } from "./services/StoryRelationsService";
import { StoryConflictService } from "./services/StoryConflictService";
import { StoryRenameService } from "./services/StoryRenameService";
import { StoryFileService } from "./services/StoryFileService";

export class StoryHandler {
	readonly entityType = "story";
	private readonly relationsService: StoryRelationsService;
	private readonly conflictService: StoryConflictService;
	private readonly renameService: StoryRenameService;
	private readonly fileService: StoryFileService;

	constructor(
		private readonly outlineGenerator = new OutlineGenerator(),
		private readonly contentsGenerator = new ContentsGenerator(),
		private readonly contentsReconciler = new ContentsReconciler(),
		private readonly outlineReconciler = new OutlineReconciler(),
		private readonly relationsGenerator = new RelationsGenerator(),
		private readonly relationsPushHandler = new RelationsPushHandler(),
		private readonly fileRenamerFactory: (context: SyncContext) => FileRenamer = (ctx) =>
			new FileRenamer(ctx)
	) {
		// Initialize services with injected dependencies
		this.relationsService = new StoryRelationsService(relationsGenerator, relationsPushHandler);
		this.conflictService = new StoryConflictService();
		this.renameService = new StoryRenameService(fileRenamerFactory);
		this.fileService = new StoryFileService();
	}

	async pull(id: string, context: SyncContext): Promise<StoryWithHierarchy> {
		const story = await context.apiClient.getStoryWithHierarchy(id);
		const folderPath = context.fileManager.getStoryFolderPath(story.story.title);

		await context.fileManager.ensureFolderExists(folderPath);
		await context.fileManager.writeStoryMetadata(story.story, folderPath, story.chapters, undefined, undefined, undefined, {
			linkMode: "full_path",
		});

		// Check conflicts for story.md
		await this.conflictService.checkConflicts(
			`${folderPath}/story.md`,
			story.story,
			context
		);

		const outlinePath = `${folderPath}/story.outline.md`;
		const contentsPath = `${folderPath}/story.contents.md`;

		// Generate content
		const outlineGenerated = this.outlineGenerator.generateStoryOutline(story, {
			syncedAt: context.timestamp(),
			showHelpBox: context.settings.showHelpBox,
			idField: context.settings.frontmatterIdField,
			storyFolderPath: folderPath,
		});
		const contentsGenerated = this.contentsGenerator.generateStoryContents({
			story: story.story,
			chapters: story.chapters,
			options: {
				syncedAt: context.timestamp(),
				idField: context.settings.frontmatterIdField,
			},
		});

		// Check conflicts for outline and contents
		await this.conflictService.checkConflicts(outlinePath, story.story, context, {
			localContentPath: outlinePath,
			remoteContent: outlineGenerated,
			entityType: "story-outline",
		});

		await this.conflictService.checkConflicts(contentsPath, story.story, context, {
			localContentPath: contentsPath,
			remoteContent: contentsGenerated,
			entityType: "story-contents",
		});

		// Reconcile with existing local content
		const existingOutline = await this.fileService.readFileSilently(context, outlinePath);
		const existingContents = await this.fileService.readFileSilently(context, contentsPath);

		const outlineMerged = this.outlineReconciler.reconcile(existingOutline, outlineGenerated);
		const contentsReconciled = this.contentsReconciler.reconcile(existingContents, contentsGenerated);

		// Emit warnings from reconciliation
		if (contentsReconciled.warnings.length) {
			contentsReconciled.warnings.forEach((warning) =>
				context.emitWarning?.({
					...warning,
					filePath: contentsPath,
				})
			);
		}

		// Write files
		await context.fileManager.writeFile(outlinePath, outlineMerged);
		await context.fileManager.writeFile(contentsPath, contentsReconciled.mergedContent);

		// Handle file renames from reordering
		await this.renameService.handleReorders(contentsReconciled.diff.operations, story, folderPath, context);

		// Generate relations file
		await this.relationsService.generateRelationsFile(story, folderPath, context);

		// Write individual entity files
		await this.fileService.writeIndividualEntityFiles(story, folderPath, context);

		return story;
	}

	async push(entity: StoryWithHierarchy, context: SyncContext): Promise<void> {
		const folderPath = context.fileManager.getStoryFolderPath(entity.story.title);

		// Push relations using the service
		await this.relationsService.pushRelations(entity, folderPath, context);

		// Push outline changes (chapter reorders)
		const outlineFilePath = `${folderPath}/story.outline.md`;
		try {
			const outlinePushHandler = new OutlinePushHandler();
			const outlineResult = await outlinePushHandler.analyzeOutline(
				outlineFilePath,
				entity.story.id,
				context
			);

			for (const action of outlineResult.actions) {
				if (action.type === "chapter_reorder") {
					await context.apiClient.updateChapter(action.chapterId, {
						number: action.newOrder,
					});
				}
			}

			if (outlineResult.warnings.length > 0) {
				outlineResult.warnings.forEach((warning) =>
					context.emitWarning?.({
						code: "outline_push_warning",
						message: warning,
						filePath: outlineFilePath,
					})
				);
			}
		} catch (error: any) {
			if (!error?.message?.includes("missing") && error?.code !== "ENOENT") {
				context.emitWarning?.({
					code: "outline_push_error",
					message: `Failed to push outline: ${error}`,
					filePath: outlineFilePath,
				});
			}
		}

		// Push contents changes (scene/beat reorders and content updates)
		const contentsFilePath = `${folderPath}/story.contents.md`;
		try {
			const localContents = await context.fileManager.readFile(contentsFilePath);
			const remoteContents = this.contentsGenerator.generateStoryContents({
				story: entity.story,
				chapters: entity.chapters,
				options: {
					syncedAt: context.timestamp(),
					idField: context.settings.frontmatterIdField,
				},
			});

			const planner = new PushPlanner();
			const plan = planner.buildPlan(remoteContents, localContents);

			if (plan.actions.length > 0) {
				const citationService = new ContentCitationService(context);
				const executor = new PushExecutor(context.apiClient, citationService);
				const result = await executor.execute(plan.actions, {
					worldId: entity.story.world_id ?? undefined,
				});

				if (result.errors.length > 0) {
					result.errors.forEach((error) =>
						context.emitWarning?.({
							...error,
							filePath: contentsFilePath,
						})
					);
				}
			}

			if (plan.warnings.length > 0) {
				plan.warnings.forEach((warning) =>
					context.emitWarning?.({
						...warning,
						filePath: contentsFilePath,
					})
				);
			}
		} catch (error: any) {
			if (!error?.message?.includes("missing") && error?.code !== "ENOENT") {
				context.emitWarning?.({
					code: "contents_push_error",
					message: `Failed to push contents: ${error}`,
					filePath: contentsFilePath,
				});
			}
		}
	}

	async delete(_id: string, _context: SyncContext): Promise<void> {
		// TODO: remove story directory
	}
}
