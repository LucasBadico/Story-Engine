import type { Beat, ContentBlock, Scene } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { PathResolver } from "../../fileRenamer/PathResolver";

export class BeatHandler {
	readonly entityType = "beat";

	async pull(id: string, context: SyncContext): Promise<Beat> {
		const beat = await context.apiClient.getBeat(id);
		const scene: Scene = await context.apiClient.getScene(beat.scene_id);
		const story = await context.apiClient.getStory(scene.story_id);
		const folderPath = context.fileManager.getStoryFolderPath(story.title);
		const beatsFolder = `${folderPath}/02-beats`;
		await context.fileManager.ensureFolderExists(beatsFolder);
		const pathResolver = new PathResolver(folderPath);
		const filePath = pathResolver.getBeatPath(beat);
		const contentBlocks: ContentBlock[] = await context.apiClient.getContentBlocksByBeat(beat.id);

		await context.fileManager.writeBeatFile(beat, filePath, story.title, contentBlocks);
		return beat;
	}

	async push(_entity: Beat, _context: SyncContext): Promise<void> {
		// TODO: Implement beat push logic
	}

	async delete(id: string, context: SyncContext): Promise<void> {
		await context.apiClient.deleteBeat(id);
	}
}

