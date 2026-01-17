import type { ChapterWithContent, ContentBlock, SceneWithBeats } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { PathResolver } from "../../fileRenamer/PathResolver";
import { ContentsGenerator } from "../../generators/ContentsGenerator";
import { OutlineGenerator } from "../../generators/OutlineGenerator";
import { writeRelationsFile } from "../../relations/relationsFileWriter";

export class ChapterHandler {
	readonly entityType = "chapter";
	private readonly contentsGenerator = new ContentsGenerator();
	private readonly outlineGenerator = new OutlineGenerator();

	async pull(id: string, context: SyncContext): Promise<ChapterWithContent> {
		const chapter = await context.apiClient.getChapter(id);
		const scenes = await context.apiClient.getScenes(chapter.id);
		const scenesWithBeats: SceneWithBeats[] = await Promise.all(
			scenes.map(async (scene) => ({
				scene,
				beats: await context.apiClient.getBeats(scene.id),
			}))
		);

		const data: ChapterWithContent = {
			chapter,
			scenes: scenesWithBeats,
		};

		const story = await context.apiClient.getStory(chapter.story_id);
		const folderPath = context.fileManager.getStoryFolderPath(story.title);
		const chaptersFolder = `${folderPath}/00-chapters`;
		await context.fileManager.ensureFolderExists(chaptersFolder);
		const pathResolver = new PathResolver(folderPath);
		const filePath = pathResolver.getChapterPath(chapter);

		await context.fileManager.writeChapterFile(
			data,
			filePath,
			story.title,
			undefined,
			undefined,
			undefined,
			{ linkMode: "full_path", storyFolderPath: folderPath }
		);

		const outlinePath = filePath.replace(/\.md$/, ".outline.md");
		const contentsPath = filePath.replace(/\.md$/, ".contents.md");
		const relationsPath = filePath.replace(/\.md$/, ".relations.md");

		const outline = this.outlineGenerator.generateChapterOutline(data, {
			syncedAt: context.timestamp(),
			showHelpBox: context.settings.showHelpBox,
			idField: context.settings.frontmatterIdField,
			storyFolderPath: folderPath,
		});
		await context.fileManager.writeFile(outlinePath, outline);

		const sceneContentBlocks = new Map<string, ContentBlock[]>();
		const beatContentBlocks = new Map<string, ContentBlock[]>();
		for (const sceneWrapper of data.scenes) {
			const sceneBlocks = await context.apiClient.getContentBlocksByScene(sceneWrapper.scene.id);
			sceneContentBlocks.set(sceneWrapper.scene.id, sceneBlocks);
			for (const beat of sceneWrapper.beats) {
				const beatBlocks = await context.apiClient.getContentBlocksByBeat(beat.id);
				beatContentBlocks.set(beat.id, beatBlocks);
			}
		}
		const contents = this.contentsGenerator.generateChapterContents(
			data,
			new Map(),
			sceneContentBlocks,
			beatContentBlocks,
			{ syncedAt: context.timestamp(), idField: context.settings.frontmatterIdField }
		);
		await context.fileManager.writeFile(contentsPath, contents);

		let worldFolderPath: string | undefined;
		if (story.world_id) {
			try {
				const world = await context.apiClient.getWorld(story.world_id);
				if (world?.name) {
					worldFolderPath = context.fileManager.getWorldFolderPath(world.name);
				}
			} catch {
				// ignore world lookup errors
			}
		}
		await writeRelationsFile({
			entity: {
				id: chapter.id,
				name: `Chapter ${chapter.number}: ${chapter.title}`,
				type: "chapter",
				worldId: story.world_id ?? undefined,
			},
			outputPath: relationsPath,
			context,
			worldFolderPath,
		});
		return data;
	}

	async push(_entity: ChapterWithContent, _context: SyncContext): Promise<void> {
		// TODO: Implement chapter push logic (parse chapter files and update API)
	}

	async delete(id: string, context: SyncContext): Promise<void> {
		await context.apiClient.deleteChapter(id);
	}
}

