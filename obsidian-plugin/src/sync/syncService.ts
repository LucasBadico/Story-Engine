import { App, Notice, TFile } from "obsidian";
import { StoryEngineClient } from "../api/client";
import { FileManager } from "./fileManager";
import { StoryEngineSettings, ProseBlock } from "../types";
import { parseChapterProse, compareProseBlocks, ParsedParagraph, ProseBlockComparison } from "./proseBlockParser";
import { ConflictModal, ConflictResolutionResult } from "../views/modals/ConflictModal";

export class SyncService {
	constructor(
		private apiClient: StoryEngineClient,
		private fileManager: FileManager,
		private settings: StoryEngineSettings,
		private app: App
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
				const existingVersionFolder = this.fileManager.getVault().getAbstractFileByPath(
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

			// Push prose blocks for each chapter
			for (const chapterFilePath of chapterFiles) {
				await this.pushChapterProseBlocks(chapterFilePath, folderPath);
			}

			new Notice(`Story "${storyFrontmatter.title}" pushed successfully`);
		} catch (err) {
			const errorMessage =
				err instanceof Error ? err.message : "Failed to push story";
			new Notice(`Error pushing story: ${errorMessage}`, 5000);
			throw err;
		}
	}

	// Push prose blocks from a chapter file
	async pushChapterProseBlocks(chapterFilePath: string, storyFolderPath: string): Promise<void> {
		// Read chapter file
		const file = this.fileManager.getVault().getAbstractFileByPath(chapterFilePath);
		if (!(file instanceof TFile)) {
			throw new Error(`Chapter file not found: ${chapterFilePath}`);
		}

		const chapterContent = await this.fileManager.getVault().read(file);
		const frontmatter = this.fileManager.parseFrontmatter(chapterContent);

		if (!frontmatter.id) {
			throw new Error("Chapter metadata missing ID");
		}

		const chapterId = frontmatter.id;
		const proseBlocksFolderPath = `${storyFolderPath}/prose-blocks`;

		// Parse paragraphs from Prose section
		const paragraphs = parseChapterProse(chapterContent);

		// Get all prose blocks from API for this chapter
		const remoteProseBlocks = await this.apiClient.getProseBlocks(chapterId);
		const remoteProseBlocksMap = new Map<string, ProseBlock>();
		for (const pb of remoteProseBlocks) {
			remoteProseBlocksMap.set(pb.id, pb);
		}

		// Process each paragraph
		const updatedParagraphs: string[] = [];
		let newOrderNum = 1;

		for (const paragraph of paragraphs) {
			let localProseBlock: ProseBlock | null = null;
			let remoteProseBlock: ProseBlock | null = null;

			// If paragraph has a link, read the local file
			if (paragraph.linkName) {
				const proseBlockFilePath = `${proseBlocksFolderPath}/${paragraph.linkName}.md`;
				localProseBlock = await this.fileManager.readProseBlockFromFile(proseBlockFilePath);

				// Find corresponding remote prose block
				if (localProseBlock) {
					remoteProseBlock = remoteProseBlocksMap.get(localProseBlock.id) || null;
				}
			}

			// Compare and determine status
			const status = compareProseBlocks(paragraph, localProseBlock, remoteProseBlock);

			let finalProseBlock: ProseBlock;
			let needsUpdate = false;

			switch (status) {
				case "new": {
					// Create new prose block
					finalProseBlock = await this.apiClient.createProseBlock(chapterId, {
						order_num: newOrderNum++,
						kind: "final",
						content: paragraph.content,
					});

					// Create file
					const fileName = this.fileManager.generateProseBlockFileName(finalProseBlock);
					const filePath = `${proseBlocksFolderPath}/${fileName}`;
					await this.fileManager.writeProseBlockFile(
						finalProseBlock,
						filePath,
						undefined
					);

					// Add link to paragraph
					const linkName = fileName.replace(/\.md$/, "");
					updatedParagraphs.push(`[[${linkName}|${paragraph.content}]]`);
					break;
				}

				case "unchanged": {
					// Check if order_num needs update
					if (localProseBlock && localProseBlock.order_num !== newOrderNum) {
						finalProseBlock = await this.apiClient.updateProseBlock(localProseBlock.id, {
							order_num: newOrderNum++,
						});
						needsUpdate = true;
					} else {
						finalProseBlock = localProseBlock!;
						newOrderNum++;
					}

					// Keep original paragraph with link
					const linkName = paragraph.linkName!;
					updatedParagraphs.push(`[[${linkName}|${paragraph.content}]]`);
					break;
				}

				case "local_modified": {
					// Update prose block with local content
					finalProseBlock = await this.apiClient.updateProseBlock(localProseBlock!.id, {
						content: paragraph.content,
						order_num: newOrderNum++,
					});

					// Update local file
					const fileName = this.fileManager.generateProseBlockFileName(finalProseBlock);
					const filePath = `${proseBlocksFolderPath}/${fileName}`;
					await this.fileManager.writeProseBlockFile(finalProseBlock, filePath, undefined);

					const linkName = fileName.replace(/\.md$/, "");
					updatedParagraphs.push(`[[${linkName}|${paragraph.content}]]`);
					break;
				}

				case "remote_modified": {
					// Update local file with remote content
					finalProseBlock = remoteProseBlock!;
					const fileName = this.fileManager.generateProseBlockFileName(finalProseBlock);
					const filePath = `${proseBlocksFolderPath}/${fileName}`;
					await this.fileManager.writeProseBlockFile(finalProseBlock, filePath, undefined);

					const linkName = fileName.replace(/\.md$/, "");
					updatedParagraphs.push(`[[${linkName}|${finalProseBlock.content}]]`);
					new Notice(`Prose block updated from remote: ${linkName}`, 3000);
					newOrderNum++;
					break;
				}

				case "conflict": {
					// Show conflict modal
					const resolution = await this.resolveConflict(
						localProseBlock!,
						remoteProseBlock!
					);

					let resolvedContent: string;
					if (resolution.resolution === "local") {
						resolvedContent = paragraph.content;
					} else if (resolution.resolution === "remote") {
						resolvedContent = remoteProseBlock!.content;
					} else {
						resolvedContent = resolution.mergedContent || paragraph.content;
					}

					// Update prose block with resolved content
					finalProseBlock = await this.apiClient.updateProseBlock(localProseBlock!.id, {
						content: resolvedContent,
						order_num: newOrderNum++,
					});

					// Update local file
					const fileName = this.fileManager.generateProseBlockFileName(finalProseBlock);
					const filePath = `${proseBlocksFolderPath}/${fileName}`;
					await this.fileManager.writeProseBlockFile(finalProseBlock, filePath, undefined);

					const linkName = fileName.replace(/\.md$/, "");
					updatedParagraphs.push(`[[${linkName}|${resolvedContent}]]`);
					break;
				}
			}

			if (needsUpdate && finalProseBlock) {
				const fileName = this.fileManager.generateProseBlockFileName(finalProseBlock);
				const filePath = `${proseBlocksFolderPath}/${fileName}`;
				await this.fileManager.writeProseBlockFile(finalProseBlock, filePath, undefined);
			}
		}

		// Update chapter file with new paragraph content
		await this.updateChapterProseSection(chapterContent, updatedParagraphs, file);
	}

	// Resolve conflict using modal
	private async resolveConflict(
		localProseBlock: ProseBlock,
		remoteProseBlock: ProseBlock
	): Promise<ConflictResolutionResult> {
		return new Promise((resolve) => {
			const modal = new ConflictModal(
				this.app,
				localProseBlock,
				remoteProseBlock,
				async (result) => {
					resolve(result);
				}
			);
			modal.open();
		});
	}

	// Update the Prose section in chapter file
	private async updateChapterProseSection(
		originalContent: string,
		updatedParagraphs: string[],
		file: TFile
	): Promise<void> {
		// Extract frontmatter
		const frontmatterMatch = originalContent.match(/^---\n([\s\S]*?)\n---/);
		const frontmatter = frontmatterMatch ? frontmatterMatch[0] : "";

		// Extract content after frontmatter
		const bodyStart = frontmatterMatch ? frontmatterMatch[0].length : 0;
		const bodyContent = originalContent.substring(bodyStart).trim();

		// Find Prose section and replace it
		const proseSectionMatch = bodyContent.match(/([\s\S]*?##\s+Prose\s*\n\n)([\s\S]*?)(?=\n##|\n*$)/);
		
		if (!proseSectionMatch) {
			// No Prose section found, add it
			const newProseSection = `\n\n## Prose\n\n${updatedParagraphs.join("\n\n")}\n\n`;
			const updatedContent = `${frontmatter}\n${bodyContent}${newProseSection}`;
			await this.fileManager.getVault().modify(file, updatedContent);
			return;
		}

		// Replace Prose section content
		const beforeProse = proseSectionMatch[1];
		const newProseContent = updatedParagraphs.join("\n\n");
		const afterProse = bodyContent.substring(proseSectionMatch.index! + proseSectionMatch[0].length);

		const updatedBody = `${beforeProse}${newProseContent}\n\n${afterProse}`;
		const updatedContent = `${frontmatter}\n${updatedBody}`;

		await this.fileManager.getVault().modify(file, updatedContent);
	}
}

