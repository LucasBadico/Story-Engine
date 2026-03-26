import type { ContentBlock, StoryWithHierarchy } from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import { PathResolver } from "../../../fileRenamer/PathResolver";
import { ContentsGenerator } from "../../../generators/ContentsGenerator";
import { OutlineGenerator } from "../../../generators/OutlineGenerator";
import { writeRelationsFile } from "../../../relations/relationsFileWriter";

/**
 * Service responsible for file I/O operations during story sync.
 * Extracted from StoryHandler for better separation of concerns.
 */
export class StoryFileService {
	private readonly contentsGenerator = new ContentsGenerator();
	private readonly outlineGenerator = new OutlineGenerator();
	/**
	 * Read a file silently, returning null if it doesn't exist or on error.
	 */
	async readFileSilently(context: SyncContext, path: string): Promise<string | null> {
		try {
			return await context.fileManager.readFile(path);
		} catch {
			return null;
		}
	}

	/**
	 * Write individual entity files (chapters, scenes, beats) as placeholders.
	 */
	async writeIndividualEntityFiles(
		story: StoryWithHierarchy,
		folderPath: string,
		context: SyncContext
	): Promise<void> {
		const pathResolver = new PathResolver(folderPath);
		let worldFolderPath: string | undefined;
		if (story.story.world_id) {
			try {
				const world = await context.apiClient.getWorld(story.story.world_id);
				if (world?.name) {
					worldFolderPath = context.fileManager.getWorldFolderPath(world.name);
				}
			} catch {
				// ignore world lookup errors
			}
		}
		const chaptersFolderPath = `${folderPath}/00-chapters`;
		const scenesFolderPath = `${folderPath}/01-scenes`;
		const beatsFolderPath = `${folderPath}/02-beats`;
		await context.fileManager.ensureFolderExists(chaptersFolderPath);
		await context.fileManager.ensureFolderExists(scenesFolderPath);
		await context.fileManager.ensureFolderExists(beatsFolderPath);

		const contentBlocksByScene = new Map<string, ContentBlock[]>();
		const contentBlocksByBeat = new Map<string, ContentBlock[]>();

		for (const chapterWithContent of story.chapters) {
			const chapter = chapterWithContent.chapter;
			const chapterOrder = chapter.number ?? 0;
			const chapterBasePath = pathResolver.getChapterPath(chapter);
			await context.fileManager.ensureFolderExists(chapterBasePath.replace(/\/[^/]+$/, ""));
			await context.fileManager.writeChapterFile(
				chapterWithContent,
				chapterBasePath,
				story.story.title,
				undefined,
				undefined,
				undefined,
				{ linkMode: "full_path", storyFolderPath: folderPath }
			);

			const chapterOutlinePath = chapterBasePath.replace(/\.md$/, ".outline.md");
			const chapterContentsPath = chapterBasePath.replace(/\.md$/, ".contents.md");
			const chapterRelationsPath = chapterBasePath.replace(/\.md$/, ".relations.md");

			const chapterOutline = this.outlineGenerator.generateChapterOutline(chapterWithContent, {
				syncedAt: context.timestamp(),
				showHelpBox: context.settings.showHelpBox,
				idField: context.settings.frontmatterIdField,
				storyFolderPath: folderPath,
			});
			await context.fileManager.writeFile(chapterOutlinePath, chapterOutline);

			const sceneContentBlocks = new Map<string, ContentBlock[]>();
			const beatContentBlocks = new Map<string, ContentBlock[]>();

			for (const sceneWrapper of chapterWithContent.scenes) {
				const scene = sceneWrapper.scene;
				const sceneOrder = scene.order_num ?? 0;
				const scenePath = pathResolver.getScenePath(scene, { chapterOrder });
				const sceneBlocks = await context.apiClient.getContentBlocksByScene(scene.id);
				contentBlocksByScene.set(scene.id, sceneBlocks);
				sceneContentBlocks.set(scene.id, sceneBlocks);
				await context.fileManager.writeSceneFile(
					sceneWrapper,
					scenePath,
					story.story.title,
					sceneBlocks,
					[],
					{
						linkMode: "full_path",
						storyFolderPath: folderPath,
						chapterOrder,
					}
				);

				for (const beat of sceneWrapper.beats) {
					const beatPath = pathResolver.getBeatPath(beat, {
						chapterOrder,
						sceneOrder,
					});
					const beatBlocks = await context.apiClient.getContentBlocksByBeat(beat.id);
					contentBlocksByBeat.set(beat.id, beatBlocks);
					beatContentBlocks.set(beat.id, beatBlocks);
					await context.fileManager.writeBeatFile(
						beat,
						beatPath,
						story.story.title,
						beatBlocks,
						{
							linkMode: "full_path",
							storyFolderPath: folderPath,
							chapterOrder,
							sceneOrder,
						}
					);

					const beatContentsPath = beatPath.replace(/\.md$/, ".contents.md");
					const beatRelationsPath = beatPath.replace(/\.md$/, ".relations.md");
					const beatContents = this.contentsGenerator.generateBeatContents(
						beat,
						beatContentBlocks,
						{ syncedAt: context.timestamp(), idField: context.settings.frontmatterIdField }
					);
					await context.fileManager.writeFile(beatContentsPath, beatContents);

					await writeRelationsFile({
						entity: {
							id: beat.id,
							name: `Beat ${beat.order_num ?? 0}: ${beat.intent || "Untitled"}`,
							type: "beat",
							worldId: story.story.world_id ?? undefined,
						},
						outputPath: beatRelationsPath,
						context,
						worldFolderPath,
					});
				}

				const sceneOutlinePath = scenePath.replace(/\.md$/, ".outline.md");
				const sceneContentsPath = scenePath.replace(/\.md$/, ".contents.md");
				const sceneRelationsPath = scenePath.replace(/\.md$/, ".relations.md");

				const sceneOutline = this.outlineGenerator.generateSceneOutline(sceneWrapper, {
					syncedAt: context.timestamp(),
					showHelpBox: context.settings.showHelpBox,
					idField: context.settings.frontmatterIdField,
					storyFolderPath: folderPath,
				});
				await context.fileManager.writeFile(sceneOutlinePath, sceneOutline);

				const sceneContents = this.contentsGenerator.generateSceneContents(
					sceneWrapper,
					sceneContentBlocks,
					beatContentBlocks,
					{ syncedAt: context.timestamp(), idField: context.settings.frontmatterIdField }
				);
				await context.fileManager.writeFile(sceneContentsPath, sceneContents);

				await writeRelationsFile({
					entity: {
						id: scene.id,
						name: `Scene ${scene.order_num ?? 0}: ${scene.goal || "Untitled"}`,
						type: "scene",
						worldId: story.story.world_id ?? undefined,
					},
					outputPath: sceneRelationsPath,
					context,
					worldFolderPath,
				});
			}

			const chapterContents = this.contentsGenerator.generateChapterContents(
				chapterWithContent,
				new Map(),
				sceneContentBlocks,
				beatContentBlocks,
				{ syncedAt: context.timestamp(), idField: context.settings.frontmatterIdField }
			);
			await context.fileManager.writeFile(chapterContentsPath, chapterContents);

			await writeRelationsFile({
				entity: {
					id: chapter.id,
					name: `Chapter ${chapter.number}: ${chapter.title}`,
					type: "chapter",
					worldId: story.story.world_id ?? undefined,
				},
				outputPath: chapterRelationsPath,
				context,
				worldFolderPath,
			});
		}

		const writtenBlocks = new Set<string>();
		const allBlocks = [...contentBlocksByScene.values(), ...contentBlocksByBeat.values()].flat();
		for (const block of allBlocks) {
			if (writtenBlocks.has(block.id)) {
				continue;
			}
			writtenBlocks.add(block.id);
			const blockPath = pathResolver.getContentBlockPath(block);
			const blockFolder = blockPath.replace(/\/[^/]+$/, "");
			await context.fileManager.ensureFolderExists(blockFolder);
			await context.fileManager.writeContentBlockFile(block, blockPath, story.story.title);
		}
	}
}
