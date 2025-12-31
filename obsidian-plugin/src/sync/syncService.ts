import { Notice, TFile, TFolder } from "obsidian";
import { StoryEngineClient } from "../api/client";
import { FileManager } from "./fileManager";
import { MarkdownParser } from "./markdownParser";
import {
	StoryEngineSettings,
} from "../types";

export class SyncService {
	constructor(
		private apiClient: StoryEngineClient,
		private fileManager: FileManager,
		private settings: StoryEngineSettings
	) {}

	// Pull story from service to Obsidian (Service → Obsidian)
	async pullStory(storyId: string): Promise<void> {
		try {
			const storyData = await this.apiClient.getStoryWithHierarchy(storyId);
			const folderPath = this.fileManager.getStoryFolderPath(
				storyData.story.title
			);

			// Write story metadata
			await this.fileManager.writeStoryMetadata(
				storyData.story,
				folderPath
			);

			// Ensure chapters folder exists
			const chaptersFolderPath = `${folderPath}/chapters`;
			await this.fileManager.ensureFolderExists(chaptersFolderPath);

			// Write chapter files
			for (const chapterWithContent of storyData.chapters) {
				const chapterFileName = `Chapter-${chapterWithContent.chapter.number}.md`;
				const chapterFilePath = `${chaptersFolderPath}/${chapterFileName}`;
				await this.fileManager.writeChapterFile(
					chapterWithContent,
					chapterFilePath
				);
			}

			// Check if version changed and create snapshot if needed
			const existingMetadata = await this.fileManager
				.readStoryMetadata(folderPath)
				.catch(() => null);

			if (
				existingMetadata &&
				existingMetadata.frontmatter.version !== undefined &&
				existingMetadata.frontmatter.version !== storyData.story.version_number
			) {
				await this.createVersionSnapshot(
					folderPath,
					existingMetadata.frontmatter.version
				);
			}

			new Notice(`Story "${storyData.story.title}" synced successfully`);
		} catch (err) {
			const errorMessage =
				err instanceof Error ? err.message : "Failed to sync story";
			new Notice(`Error syncing story: ${errorMessage}`, 5000);
			throw err;
		}
	}

	// Pull all stories
	async pullAllStories(): Promise<void> {
		if (!this.settings.tenantId) {
			throw new Error("Tenant ID is required");
		}

		const stories = await this.apiClient.listStories(this.settings.tenantId);

		for (const story of stories) {
			try {
				await this.pullStory(story.id);
			} catch (err) {
				console.error(`Failed to sync story ${story.id}:`, err);
			}
		}

		new Notice(`Synced ${stories.length} stories`);
	}

	// Push story from Obsidian to service (Obsidian → Service)
	async pushStory(folderPath: string): Promise<void> {
		try {
			// Read story metadata
			const { frontmatter: storyFrontmatter, content: storyContent } =
				await this.fileManager.readStoryMetadata(folderPath);

			if (!storyFrontmatter.id) {
				throw new Error("Story metadata missing ID");
			}

			const storyId = storyFrontmatter.id;

			// Update story metadata if needed
			const titleMatch = storyContent.match(/^#\s+(.+)$/m);
			if (titleMatch) {
				const newTitle = titleMatch[1].trim();
				await this.apiClient.updateStory(storyId, newTitle);
			}

			// Read and sync chapters
			const vault = this.fileManager.getVault();
			const chaptersFolder = vault.getAbstractFileByPath(
				`${folderPath}/chapters`
			);

			if (chaptersFolder && chaptersFolder instanceof TFolder) {
				const chapterFiles = chaptersFolder.children.filter(
					(file) => file instanceof TFile && file.extension === "md"
				) as TFile[];

				for (const chapterFile of chapterFiles) {
					await this.syncChapterFromFile(chapterFile.path, storyId);
				}
			}

			new Notice("Story pushed to service successfully");
		} catch (err) {
			const errorMessage =
				err instanceof Error ? err.message : "Failed to push story";
			new Notice(`Error pushing story: ${errorMessage}`, 5000);
			throw err;
		}
	}

	// Sync chapter from markdown file
	private async syncChapterFromFile(
		filePath: string,
		storyId: string
	): Promise<void> {
		const { frontmatter, content } = await this.fileManager.readChapterFile(
			filePath
		);

		// Extract chapter title and number from content
		const titleMatch = content.match(/^#\s+Chapter\s+\d+:\s*(.+)$/m);
		const numberMatch = content.match(/^#\s+Chapter\s+(\d+):/m);

		if (!titleMatch || !numberMatch) {
			throw new Error(`Invalid chapter format in ${filePath}`);
		}

		const chapterTitle = titleMatch[1].trim();
		const chapterNumber = parseInt(numberMatch[1], 10);

		// Create or update chapter
		let chapterId: string;
		if (frontmatter.id) {
			// Update existing chapter
			chapterId = frontmatter.id;
			await this.apiClient.updateChapter(frontmatter.id, {
				title: chapterTitle,
				number: chapterNumber,
			});
		} else {
			// Create new chapter
			const newChapter = await this.apiClient.createChapter(storyId, {
				title: chapterTitle,
				number: chapterNumber,
			});
			chapterId = newChapter.id;
			// Update frontmatter with new ID
			await this.fileManager.updateFrontmatter(filePath, {
				id: newChapter.id,
			});
		}

		// Parse and sync scenes and beats
		const parsedScenes = MarkdownParser.parseChapterMarkdown(content);

		for (const parsedScene of parsedScenes) {
			const scene = parsedScene.scene;
			if (!scene.order_num) continue;

			// Create or update scene
			let sceneId: string;
			if (scene.id) {
				sceneId = scene.id;
				await this.apiClient.updateScene(sceneId, {
					order_num: scene.order_num,
					goal: scene.goal || "",
					time_ref: scene.time_ref || "",
				});
			} else {
				const newScene = await this.apiClient.createScene({
					story_id: storyId,
					chapter_id: chapterId,
					order_num: scene.order_num,
					goal: scene.goal || "",
					time_ref: scene.time_ref || "",
				});
				sceneId = newScene.id;
				// Note: We can't update the markdown file's scene ID inline easily
				// This would require re-parsing and updating the content
			}

			// Sync beats
			for (const beat of parsedScene.beats) {
				if (!beat.order_num || !beat.type) continue;

				if (beat.id) {
					await this.apiClient.updateBeat(beat.id, {
						order_num: beat.order_num,
						type: beat.type,
						intent: beat.intent || "",
						outcome: beat.outcome || "",
					});
				} else {
					await this.apiClient.createBeat({
						scene_id: sceneId,
						order_num: beat.order_num,
						type: beat.type,
						intent: beat.intent || "",
						outcome: beat.outcome || "",
					});
				}
			}
		}
	}

	// Bidirectional sync (pull then push)
	async syncStory(storyId: string): Promise<void> {
		// First pull latest from service
		await this.pullStory(storyId);

		// Then push local changes back
		const story = await this.apiClient.getStory(storyId);
		const folderPath = this.fileManager.getStoryFolderPath(story.title);
		await this.pushStory(folderPath);
	}

	// Create version snapshot
	async createVersionSnapshot(
		storyFolder: string,
		versionNumber: number
	): Promise<void> {
		const versionsFolder = `${storyFolder}/versions`;
		await this.fileManager.ensureFolderExists(versionsFolder);

		const versionFolder = `${versionsFolder}/v${versionNumber}`;
		await this.fileManager.ensureFolderExists(versionFolder);

		// Copy metadata.md
		const metadataSource = `${storyFolder}/metadata.md`;
		const metadataDest = `${versionFolder}/metadata.md`;

		const vault = this.fileManager.getVault();
		const metadataFile = vault.getAbstractFileByPath(metadataSource);
		if (metadataFile instanceof TFile) {
			const content = await vault.read(metadataFile);
			await vault.create(metadataDest, content);
		}

		// Copy chapters folder (excluding versions folder)
		const chaptersSource = `${storyFolder}/chapters`;
		const chaptersDest = `${versionFolder}/chapters`;

		const chaptersFolder = vault.getAbstractFileByPath(chaptersSource);
		if (chaptersFolder && chaptersFolder instanceof TFolder) {
			await this.fileManager.ensureFolderExists(chaptersDest);

			for (const file of chaptersFolder.children) {
				if (file instanceof TFile) {
					const content = await vault.read(file);
					const destPath = `${chaptersDest}/${file.name}`;
					await vault.create(destPath, content);
				}
			}
		}
	}
}

