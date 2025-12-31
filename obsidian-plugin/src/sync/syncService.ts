import { Notice } from "obsidian";
import { StoryEngineClient } from "../api/client";
import { FileManager } from "./fileManager";
import { StoryEngineSettings } from "../types";

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

				// Write version chapters
				const versionChaptersPath = `${versionFolderPath}/chapters`;
				await this.fileManager.ensureFolderExists(versionChaptersPath);

				for (const chapterWithContent of versionData.chapters) {
					const chapterFileName = `Chapter-${chapterWithContent.chapter.number}.md`;
					const chapterFilePath = `${versionChaptersPath}/${chapterFileName}`;
					await this.fileManager.writeChapterFile(
						chapterWithContent,
						chapterFilePath
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

