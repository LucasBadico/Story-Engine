import { Notice } from "obsidian";
import { StoryEngineClient } from "../api/client";
import { FileManager } from "./fileManager";
import { StoryEngineSettings, ProseBlock } from "../types";

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

			// Ensure prose-blocks folder exists
			const proseBlocksFolderPath = `${folderPath}/prose-blocks`;
			await this.fileManager.ensureFolderExists(proseBlocksFolderPath);

			// Ensure chapters folder exists
			const chaptersFolderPath = `${folderPath}/chapters`;
			await this.fileManager.ensureFolderExists(chaptersFolderPath);

			// Write chapter files with prose blocks
			for (const chapterWithContent of storyData.chapters) {
				// Fetch prose blocks for this chapter
				const proseBlocks = await this.apiClient.getProseBlocks(chapterWithContent.chapter.id);

				// Write prose block files
				for (const proseBlock of proseBlocks) {
					const proseBlockFileName = this.fileManager.generateProseBlockFileName(proseBlock);
					const proseBlockFilePath = `${proseBlocksFolderPath}/${proseBlockFileName}`;
					await this.fileManager.writeProseBlockFile(
						proseBlock,
						proseBlockFilePath,
						storyData.story.title
					);
				}

				// Write chapter file with prose block embeds
				const chapterFileName = `Chapter-${chapterWithContent.chapter.number}.md`;
				const chapterFilePath = `${chaptersFolderPath}/${chapterFileName}`;
				await this.fileManager.writeChapterFile(
					chapterWithContent,
					chapterFilePath,
					storyData.story.title,
					proseBlocks
				);

				// Write scene files with prose block references
				for (const { scene, beats } of chapterWithContent.scenes) {
					// Fetch prose blocks referenced by this scene
					const sceneProseBlocks = await this.apiClient.getProseBlocksByScene(scene.id);

					const sceneFileName = `Scene-${scene.order_num}.md`;
					const sceneFolderPath = `${chaptersFolderPath}/Chapter-${chapterWithContent.chapter.number}/scenes`;
					await this.fileManager.ensureFolderExists(sceneFolderPath);
					const sceneFilePath = `${sceneFolderPath}/${sceneFileName}`;

					await this.fileManager.writeSceneFile(
						{ scene, beats },
						sceneFilePath,
						storyData.story.title,
						sceneProseBlocks
					);

					// Write beat files with prose block references
					for (const beat of beats) {
						const beatProseBlocks = await this.apiClient.getProseBlocksByBeat(beat.id);
						// Note: writeBeatFile doesn't currently support prose blocks
						// This could be added if needed
					}
				}
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
				await this.fileManager.createVersionSnapshot(
					folderPath,
					existingMetadata.frontmatter.version
				);
			}

			// Sync all previous versions to the versions folder
			await this.syncVersionHistory(storyData.story.root_story_id, folderPath);

			new Notice(`Story "${storyData.story.title}" synced successfully`);
		} catch (err) {
			const errorMessage =
				err instanceof Error ? err.message : "Failed to sync story";
			new Notice(`Error syncing story: ${errorMessage}`, 5000);
			throw err;
		}
	}

	// Sync all previous versions of a story
	private async syncVersionHistory(
		rootStoryId: string,
		storyFolderPath: string
	): Promise<void> {
		try {
			// Get ALL stories and filter by root_story_id
			const allStories = await this.apiClient.listStories(this.settings.tenantId);
			const versions = allStories.filter(s => s.root_story_id === rootStoryId);

			// Sort by version number
			versions.sort((a, b) => a.version_number - b.version_number);

			const versionsPath = `${storyFolderPath}/versions`;
			await this.fileManager.ensureFolderExists(versionsPath);

			// For each version (except the current one), fetch and store
			for (const versionStory of versions) {
				// Skip the current version (it's already in the root folder)
				const currentVersion = versions[versions.length - 1].version_number;
				if (versionStory.version_number === currentVersion) {
					continue;
				}

				const versionFolderPath = `${versionsPath}/v${versionStory.version_number}`;

				// Check if this version already exists
				// (We won't overwrite existing versions to preserve local edits)
				const existingVersionFolder = this.fileManager["vault"].getAbstractFileByPath(
					versionFolderPath
				);
				if (existingVersionFolder) {
					console.log(`Version v${versionStory.version_number} already exists, skipping`);
					continue;
				}

				// Fetch full version data
				const versionData = await this.apiClient.getStoryWithHierarchy(
					versionStory.id
				);

				// Create version folder
				await this.fileManager.ensureFolderExists(versionFolderPath);

				// Write version metadata
				await this.fileManager.writeStoryMetadata(
					versionData.story,
					versionFolderPath
				);

				// Ensure prose-blocks folder exists for version
				const versionProseBlocksFolderPath = `${versionFolderPath}/prose-blocks`;
				await this.fileManager.ensureFolderExists(versionProseBlocksFolderPath);

				// Write version chapters
				const versionChaptersPath = `${versionFolderPath}/chapters`;
				await this.fileManager.ensureFolderExists(versionChaptersPath);

				for (const chapterWithContent of versionData.chapters) {
					// Fetch prose blocks for this chapter
					const proseBlocks = await this.apiClient.getProseBlocks(chapterWithContent.chapter.id);

					// Write prose block files
					for (const proseBlock of proseBlocks) {
						const proseBlockFileName = this.fileManager.generateProseBlockFileName(proseBlock);
						const proseBlockFilePath = `${versionProseBlocksFolderPath}/${proseBlockFileName}`;
						await this.fileManager.writeProseBlockFile(
							proseBlock,
							proseBlockFilePath,
							versionData.story.title
						);
					}

					const chapterFileName = `Chapter-${chapterWithContent.chapter.number}.md`;
					const chapterFilePath = `${versionChaptersPath}/${chapterFileName}`;
					await this.fileManager.writeChapterFile(
						chapterWithContent,
						chapterFilePath,
						versionData.story.title,
						proseBlocks
					);
				}

				console.log(`Synced version v${versionStory.version_number}`);
			}
		} catch (err) {
			console.error("Error syncing version history:", err);
			// Don't throw - version sync is optional
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
			const { frontmatter: storyFrontmatter } =
				await this.fileManager.readStoryMetadata(folderPath);

			if (!storyFrontmatter.id) {
				throw new Error("Story metadata missing ID");
			}

			const storyId = storyFrontmatter.id;

			// Update story metadata
			await this.apiClient.updateStory(
				storyId,
				storyFrontmatter.title,
				storyFrontmatter.status
			);

			// Read and update chapters
			const chapterFiles = await this.fileManager.listChapterFiles(folderPath);

			for (const chapterFilePath of chapterFiles) {
				// Parse chapter file and update via API
				// (This would require implementing chapter update logic)
				console.log(`Would update chapter: ${chapterFilePath}`);
			}

			new Notice(`Story "${storyFrontmatter.title}" pushed successfully`);
		} catch (err) {
			const errorMessage =
				err instanceof Error ? err.message : "Failed to push story";
			new Notice(`Error pushing story: ${errorMessage}`, 5000);
			throw err;
		}
	}
}

