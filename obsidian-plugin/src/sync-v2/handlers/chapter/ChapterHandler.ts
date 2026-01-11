import type { ChapterWithContent, SceneWithBeats } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { PathResolver } from "../../fileRenamer/PathResolver";

export class ChapterHandler {
	readonly entityType = "chapter";

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

		await context.fileManager.writeChapterFile(data, filePath, story.title);
		return data;
	}

	async push(_entity: ChapterWithContent, _context: SyncContext): Promise<void> {
		// TODO: Implement chapter push logic (parse chapter files and update API)
	}

	async delete(id: string, context: SyncContext): Promise<void> {
		await context.apiClient.deleteChapter(id);
	}
}

