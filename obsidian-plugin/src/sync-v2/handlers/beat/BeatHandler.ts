import type { Beat, ContentBlock, Scene } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { PathResolver } from "../../fileRenamer/PathResolver";
import { ContentsGenerator } from "../../generators/ContentsGenerator";
import { writeRelationsFile } from "../../relations/relationsFileWriter";

export class BeatHandler {
	readonly entityType = "beat";
	private readonly contentsGenerator = new ContentsGenerator();

	async pull(id: string, context: SyncContext): Promise<Beat> {
		const beat = await context.apiClient.getBeat(id);
		const scene: Scene = await context.apiClient.getScene(beat.scene_id);
		const chapterOrder =
			scene.chapter_id ? (await context.apiClient.getChapter(scene.chapter_id)).number ?? 0 : 0;
		const story = await context.apiClient.getStory(scene.story_id);
		const folderPath = context.fileManager.getStoryFolderPath(story.title);
		const beatsFolder = `${folderPath}/02-beats`;
		await context.fileManager.ensureFolderExists(beatsFolder);
		const pathResolver = new PathResolver(folderPath);
		const filePath = pathResolver.getBeatPath(beat, {
			chapterOrder,
			sceneOrder: scene.order_num ?? 0,
		});
		const contentBlocks: ContentBlock[] = await context.apiClient.getContentBlocksByBeat(beat.id);

		await context.fileManager.writeBeatFile(beat, filePath, story.title, contentBlocks, {
			linkMode: "full_path",
			storyFolderPath: folderPath,
			chapterOrder,
			sceneOrder: scene.order_num ?? 0,
		});

		const contentsPath = filePath.replace(/\.md$/, ".contents.md");
		const relationsPath = filePath.replace(/\.md$/, ".relations.md");

		const beatContentBlocks = new Map<string, ContentBlock[]>();
		beatContentBlocks.set(beat.id, contentBlocks);
		const contents = this.contentsGenerator.generateBeatContents(
			beat,
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
				id: beat.id,
				name: `Beat ${beat.order_num ?? 0}: ${beat.intent || "Untitled"}`,
				type: "beat",
				worldId: story.world_id ?? undefined,
			},
			outputPath: relationsPath,
			context,
			worldFolderPath,
		});
		return beat;
	}

	async push(_entity: Beat, _context: SyncContext): Promise<void> {
		// TODO: Implement beat push logic
	}

	async delete(id: string, context: SyncContext): Promise<void> {
		await context.apiClient.deleteBeat(id);
	}
}

