import type { StoryWithHierarchy } from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import type { ReconcileResult } from "../../../diff/ContentsReconciler";
import { FileRenamer } from "../../../fileRenamer/FileRenamer";
import { PathResolver } from "../../../fileRenamer/PathResolver";

/**
 * Service responsible for renaming entity files when reorder operations are detected.
 * Extracted from StoryHandler for better separation of concerns.
 */
export class StoryRenameService {
	constructor(
		private readonly fileRenamerFactory: (context: SyncContext) => FileRenamer = (ctx) =>
			new FileRenamer(ctx)
	) {}

	/**
	 * Handle file renames based on reorder operations from the diff.
	 */
	async handleReorders(
		operations: ReconcileResult["diff"]["operations"],
		story: StoryWithHierarchy,
		folderPath: string,
		context: SyncContext
	): Promise<void> {
		const renamer = this.fileRenamerFactory(context);
		const pathResolver = new PathResolver(folderPath);

		const sceneMap = new Map(
			story.chapters.flatMap((chapter) =>
				chapter.scenes.map((scene) => [
					scene.scene.id,
					{ scene: scene.scene, chapterOrder: chapter.chapter.number ?? 0 },
				] as const)
			)
		);
		const beatMap = new Map(
			story.chapters.flatMap((chapter) =>
				chapter.scenes.flatMap((scene) =>
					scene.beats.map((beat) => [
						beat.id,
						{
							beat,
							chapterOrder: chapter.chapter.number ?? 0,
							sceneOrder: scene.scene.order_num ?? 0,
						},
					] as const)
				)
			)
		);
		const chapterMap = new Map(
			story.chapters.map((chapter) => [chapter.chapter.id, chapter.chapter] as const)
		);

		for (const op of operations) {
			if (op.kind !== "reordered") continue;

			if (op.fenceType === "chapter") {
				await this.handleChapterReorder(op, chapterMap, pathResolver, renamer);
			} else if (op.fenceType === "scene") {
				await this.handleSceneReorder(op, sceneMap, pathResolver, renamer);
			} else if (op.fenceType === "beat") {
				await this.handleBeatReorder(op, beatMap, pathResolver, renamer);
			} else if (op.fenceType === "content") {
				await this.handleContentReorder(op, pathResolver, renamer, context);
			}
		}
	}

	private async handleChapterReorder(
		op: ReconcileResult["diff"]["operations"][0],
		chapterMap: Map<string, StoryWithHierarchy["chapters"][0]["chapter"]>,
		pathResolver: PathResolver,
		renamer: FileRenamer
	): Promise<void> {
		const chapter = chapterMap.get(op.fenceId);
		if (!chapter || op.metadata?.newOrder === undefined || op.metadata.oldOrder === undefined) {
			return;
		}
		const oldPath = pathResolver.getChapterPath(chapter, { order: op.metadata.oldOrder });
		const newPath = pathResolver.getChapterPath(chapter, { order: op.metadata.newOrder });
		if (oldPath === newPath) return;
		await this.renameSafely(renamer, oldPath, newPath, "chapter");
	}

	private async handleSceneReorder(
		op: ReconcileResult["diff"]["operations"][0],
		sceneMap: Map<
			string,
			{ scene: StoryWithHierarchy["chapters"][0]["scenes"][0]["scene"]; chapterOrder: number }
		>,
		pathResolver: PathResolver,
		renamer: FileRenamer
	): Promise<void> {
		const sceneEntry = sceneMap.get(op.fenceId);
		if (!sceneEntry || op.metadata?.newOrder === undefined || op.metadata.oldOrder === undefined) {
			return;
		}
		const oldPath = pathResolver.getScenePath(sceneEntry.scene, {
			order: op.metadata.oldOrder,
			chapterOrder: sceneEntry.chapterOrder,
		});
		const newPath = pathResolver.getScenePath(sceneEntry.scene, {
			order: op.metadata.newOrder,
			chapterOrder: sceneEntry.chapterOrder,
		});
		if (oldPath === newPath) return;
		await this.renameSafely(renamer, oldPath, newPath, "scene");
	}

	private async handleBeatReorder(
		op: ReconcileResult["diff"]["operations"][0],
		beatMap: Map<
			string,
			{
				beat: StoryWithHierarchy["chapters"][0]["scenes"][0]["beats"][0];
				chapterOrder: number;
				sceneOrder: number;
			}
		>,
		pathResolver: PathResolver,
		renamer: FileRenamer
	): Promise<void> {
		const beatEntry = beatMap.get(op.fenceId);
		if (!beatEntry || op.metadata?.newOrder === undefined || op.metadata.oldOrder === undefined) {
			return;
		}
		const oldPath = pathResolver.getBeatPath(beatEntry.beat, {
			order: op.metadata.oldOrder,
			chapterOrder: beatEntry.chapterOrder,
			sceneOrder: beatEntry.sceneOrder,
		});
		const newPath = pathResolver.getBeatPath(beatEntry.beat, {
			order: op.metadata.newOrder,
			chapterOrder: beatEntry.chapterOrder,
			sceneOrder: beatEntry.sceneOrder,
		});
		if (oldPath === newPath) return;
		await this.renameSafely(renamer, oldPath, newPath, "beat");
	}

	private async handleContentReorder(
		op: ReconcileResult["diff"]["operations"][0],
		pathResolver: PathResolver,
		renamer: FileRenamer,
		context: SyncContext
	): Promise<void> {
		if (op.metadata?.newOrder === undefined || op.metadata.oldOrder === undefined) {
			return;
		}
		try {
			const contentBlock = await context.apiClient.getContentBlock(op.fenceId);
			if (!contentBlock) return;
			const oldPath = pathResolver.getContentBlockPath(contentBlock, {
				order: op.metadata.oldOrder,
			});
			const newPath = pathResolver.getContentBlockPath(contentBlock, {
				order: op.metadata.newOrder,
			});
			if (oldPath === newPath) return;
			await this.renameSafely(renamer, oldPath, newPath, "content block");
		} catch (err) {
			console.warn(`[Sync V2] Failed to get content block for rename`, err);
		}
	}

	private async renameSafely(
		renamer: FileRenamer,
		oldPath: string,
		newPath: string,
		entity: string
	): Promise<void> {
		try {
			await renamer.rename({ oldPath, newPath });
		} catch (err) {
			console.warn(`[Sync V2] Failed to rename ${entity} file`, err);
		}
	}
}
