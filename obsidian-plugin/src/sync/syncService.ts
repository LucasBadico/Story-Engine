import { App, Notice, TFile, TFolder } from "obsidian";
import { StoryEngineClient } from "../api/client";
import { FileManager } from "./fileManager";
import { StoryEngineSettings, ContentBlock, ContentAnchor, Scene, Beat, Chapter, SceneWithBeats } from "../types";
import { parseHierarchicalProse, parseChapterProse, compareContentBlocks, ParsedParagraph, ContentBlockComparison, HierarchicalProse, parseSceneBeatList, ParsedSceneBeatListItem, parseChapterList, ParsedChapterListItem, ParsedChapterList, parseBeatList, ParsedBeatList, ParsedBeatListItem, parseOrphanScenesList, parseOrphanBeatsList, parseStoryProse, parseSceneProse, parseBeatProse } from "./contentBlockParser";
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

			// IMPORTANT: Read existing metadata BEFORE writing to detect version changes
			const existingMetadata = await this.fileManager
				.readStoryMetadata(folderPath)
				.catch(() => null);

			// Fetch scenes and beats without chapter_id (orphans)
			const allScenes = await this.apiClient.getScenesByStory(storyId);
			const orphanScenes: SceneWithBeats[] = [];
			
			for (const scene of allScenes) {
				if (!scene.chapter_id) {
					const beats = await this.apiClient.getBeats(scene.id);
					orphanScenes.push({ scene, beats });
				}
			}
			
			// Sort orphan scenes by order_num
			orphanScenes.sort((a, b) => a.scene.order_num - b.scene.order_num);
			
			// Fetch beats without scene_id (orphan beats)
			const allBeats = await this.apiClient.getBeatsByStory(storyId);
			const orphanBeats: Beat[] = [];
			const sceneIdSet = new Set(allScenes.map(s => s.id));
			
			for (const beat of allBeats) {
				// Check if beat's scene_id doesn't exist or is invalid
				if (!beat.scene_id || !sceneIdSet.has(beat.scene_id)) {
					orphanBeats.push(beat);
				}
			}
			
			// Sort orphan beats by order_num
			orphanBeats.sort((a, b) => a.order_num - b.order_num);
			
			// Fetch prose blocks for all chapters to include in story.md
			const chapterContentData = new Map<string, { contentBlocks: ContentBlock[], contentBlockRefs: ContentAnchor[] }>();
			for (const chapterWithContent of storyData.chapters) {
				const contentBlocks = await this.apiClient.getContentBlocks(chapterWithContent.chapter.id);
				const contentBlockRefs: ContentAnchor[] = [];
				for (const contentBlock of contentBlocks) {
					const refs = await this.apiClient.getContentAnchors(contentBlock.id);
					contentBlockRefs.push(...refs);
				}
				chapterContentData.set(chapterWithContent.chapter.id, { contentBlocks, contentBlockRefs });
			}
			
			// Write story metadata with chapters list, orphan scenes, orphan beats, and prose data
			await this.fileManager.writeStoryMetadata(
				storyData.story,
				folderPath,
				storyData.chapters,
				orphanScenes,
				orphanBeats,
				chapterContentData
			);

			// Ensure contents folder and type subfolders exist
			const contentsFolderPath = `${folderPath}/03-contents`;
			await this.fileManager.ensureFolderExists(contentsFolderPath);
			// Create type subfolders
			for (const typeFolder of ["00-texts", "01-images", "02-videos", "03-audios", "04-embeds", "05-links"]) {
				await this.fileManager.ensureFolderExists(`${contentsFolderPath}/${typeFolder}`);
			}

			// Ensure chapters folder exists
			const chaptersFolderPath = `${folderPath}/00-chapters`;
			await this.fileManager.ensureFolderExists(chaptersFolderPath);

			// Write chapter files with content blocks
			for (const chapterWithContent of storyData.chapters) {
				// Get content blocks from the already fetched data
				const contentData = chapterContentData.get(chapterWithContent.chapter.id);
				const contentBlocks = contentData?.contentBlocks || [];
				const contentBlockRefs = contentData?.contentBlockRefs || [];

				// Write content block files to their type-specific subfolders
				for (const contentBlock of contentBlocks) {
					const contentBlockFileName = this.fileManager.generateContentBlockFileName(contentBlock);
					const typeFolderPath = this.fileManager.getContentBlockFolderPath(folderPath, contentBlock.type || "text");
					await this.fileManager.ensureFolderExists(typeFolderPath);
					const contentBlockFilePath = `${typeFolderPath}/${contentBlockFileName}`;
					await this.fileManager.writeContentBlockFile(
						contentBlock,
						contentBlockFilePath,
						storyData.story.title
					);
				}

				// Write chapter file with hierarchical prose structure
				const chapterFileName = `Chapter-${chapterWithContent.chapter.number}.md`;
				const chapterFilePath = `${chaptersFolderPath}/${chapterFileName}`;
				await this.fileManager.writeChapterFile(
					chapterWithContent,
					chapterFilePath,
					storyData.story.title,
					contentBlocks,
					contentBlockRefs,
					orphanScenes // Include orphan scenes for easy association
				);

				// Write scene files with prose block references (flat structure)
				const scenesFolderPath = `${folderPath}/01-scenes`;
				await this.fileManager.ensureFolderExists(scenesFolderPath);
				
				for (const { scene, beats } of chapterWithContent.scenes) {
					// Fetch prose blocks referenced by this scene
					const sceneContentBlocks = await this.apiClient.getContentBlocksByScene(scene.id);

					const sceneFileName = this.fileManager.generateSceneFileName(scene);
					const sceneFilePath = `${scenesFolderPath}/${sceneFileName}`;

					await this.fileManager.writeSceneFile(
						{ scene, beats },
						sceneFilePath,
						storyData.story.title,
						sceneContentBlocks,
						orphanBeats // Include orphan beats for easy association
					);

					// Write beat files with prose block references (flat structure)
					const beatsFolderPath = `${folderPath}/02-beats`;
					await this.fileManager.ensureFolderExists(beatsFolderPath);
					
					for (const beat of beats) {
						const beatContentBlocks = await this.apiClient.getContentBlocksByBeat(beat.id);
						const beatFileName = this.fileManager.generateBeatFileName(beat);
						const beatFilePath = `${beatsFolderPath}/${beatFileName}`;
						await this.fileManager.writeBeatFile(beat, beatFilePath, storyData.story.title, beatContentBlocks);
					}
				}
			}
			
			// Write orphan scene files (scenes without chapter_id)
			const scenesFolderPath = `${folderPath}/01-scenes`;
			await this.fileManager.ensureFolderExists(scenesFolderPath);
			
			for (const { scene, beats } of orphanScenes) {
				// Fetch prose blocks referenced by this scene
				const sceneContentBlocks = await this.apiClient.getContentBlocksByScene(scene.id);

				const sceneFileName = this.fileManager.generateSceneFileName(scene);
				const sceneFilePath = `${scenesFolderPath}/${sceneFileName}`;

				await this.fileManager.writeSceneFile(
					{ scene, beats },
					sceneFilePath,
					storyData.story.title,
					sceneContentBlocks,
					orphanBeats // Include orphan beats for easy association
				);

				// Write beat files with prose block references (flat structure)
				const beatsFolderPath = `${folderPath}/02-beats`;
				await this.fileManager.ensureFolderExists(beatsFolderPath);
				
				for (const beat of beats) {
					const beatContentBlocks = await this.apiClient.getContentBlocksByBeat(beat.id);
					const beatFileName = this.fileManager.generateBeatFileName(beat);
					const beatFilePath = `${beatsFolderPath}/${beatFileName}`;
					await this.fileManager.writeBeatFile(beat, beatFilePath, storyData.story.title, beatContentBlocks);
				}
			}
			
			// Write orphan beat files (beats without scene_id)
			const beatsFolderPath = `${folderPath}/02-beats`;
			await this.fileManager.ensureFolderExists(beatsFolderPath);
			
			for (const beat of orphanBeats) {
				const beatContentBlocks = await this.apiClient.getContentBlocksByBeat(beat.id);
				const beatFileName = this.fileManager.generateBeatFileName(beat);
				const beatFilePath = `${beatsFolderPath}/${beatFileName}`;
				await this.fileManager.writeBeatFile(beat, beatFilePath, storyData.story.title, beatContentBlocks);
			}

			// Check if version changed and create snapshot if needed (existingMetadata was read BEFORE writing)
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
			const allStories = await this.apiClient.listStories();
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

				// Ensure contents folder and type subfolders exist for version
				const versionContentsFolderPath = `${versionFolderPath}/03-contents`;
				await this.fileManager.ensureFolderExists(versionContentsFolderPath);
				for (const typeFolder of ["00-texts", "01-images", "02-videos", "03-audios", "04-embeds", "05-links"]) {
					await this.fileManager.ensureFolderExists(`${versionContentsFolderPath}/${typeFolder}`);
				}

				// Write version chapters
				const versionChaptersPath = `${versionFolderPath}/00-chapters`;
				await this.fileManager.ensureFolderExists(versionChaptersPath);

				for (const chapterWithContent of versionData.chapters) {
					// Fetch prose blocks for this chapter
					const contentBlocks = await this.apiClient.getContentBlocks(chapterWithContent.chapter.id);

					// Fetch prose block references for all prose blocks
					const contentBlockRefs: ContentAnchor[] = [];
					for (const contentBlock of contentBlocks) {
						const refs = await this.apiClient.getContentAnchors(contentBlock.id);
						contentBlockRefs.push(...refs);
					}

					// Write prose block files
					for (const contentBlock of contentBlocks) {
						const contentBlockFileName = this.fileManager.generateContentBlockFileName(contentBlock);
						const typeFolderPath = this.fileManager.getContentBlockFolderPath(versionFolderPath, contentBlock.type || "text");
						await this.fileManager.ensureFolderExists(typeFolderPath);
						const contentBlockFilePath = `${typeFolderPath}/${contentBlockFileName}`;
						await this.fileManager.writeContentBlockFile(
							contentBlock,
							contentBlockFilePath,
							versionData.story.title
						);
					}

					const chapterFileName = `Chapter-${chapterWithContent.chapter.number}.md`;
					const chapterFilePath = `${versionChaptersPath}/${chapterFileName}`;
					await this.fileManager.writeChapterFile(
						chapterWithContent,
						chapterFilePath,
						versionData.story.title,
						contentBlocks,
						contentBlockRefs
					);

					// Write scene files (flat structure)
					const versionScenesPath = `${versionFolderPath}/01-scenes`;
					await this.fileManager.ensureFolderExists(versionScenesPath);
					
					for (const { scene, beats } of chapterWithContent.scenes) {
						const sceneContentBlocks = await this.apiClient.getContentBlocksByScene(scene.id);
						const sceneFileName = this.fileManager.generateSceneFileName(scene);
						const sceneFilePath = `${versionScenesPath}/${sceneFileName}`;
						await this.fileManager.writeSceneFile(
							{ scene, beats },
							sceneFilePath,
							versionData.story.title,
							sceneContentBlocks
						);

						// Write beat files (flat structure)
						const versionBeatsPath = `${versionFolderPath}/02-beats`;
						await this.fileManager.ensureFolderExists(versionBeatsPath);
						
						for (const beat of beats) {
							const beatContentBlocks = await this.apiClient.getContentBlocksByBeat(beat.id);
							const beatFileName = this.fileManager.generateBeatFileName(beat);
							const beatFilePath = `${versionBeatsPath}/${beatFileName}`;
							await this.fileManager.writeBeatFile(beat, beatFilePath, versionData.story.title, beatContentBlocks);
						}
					}
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
		// Only validate tenant ID in remote mode
		if (this.settings.mode === "remote" && !this.settings.tenantId) {
			throw new Error("Tenant ID is required");
		}

		// Pull worlds (for display/metadata)
		try {
			await this.apiClient.getWorlds();
		} catch (err) {
			console.error("Failed to load worlds:", err);
			// Don't fail the entire sync if worlds fail
		}

		const stories = await this.apiClient.listStories();

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

			// Read story file to parse chapter list and orphan lists
			const storyFilePath = `${folderPath}/story.md`;
			const storyFile = this.fileManager.getVault().getAbstractFileByPath(storyFilePath);
			if (storyFile instanceof TFile) {
				const storyContent = await this.fileManager.getVault().read(storyFile);
				const chapterList = parseChapterList(storyContent);
				
				// Process chapter list to update order
				if (chapterList.items.length > 0) {
					await this.processChapterList(chapterList, storyId);
				}
				
				// Parse and process orphan scenes list
				const orphanScenesList = parseOrphanScenesList(storyContent);
				if (orphanScenesList.items.length > 0) {
					await this.processOrphanScenesList(orphanScenesList, storyId);
				}

				// Parse and process orphan beats list
				const orphanBeatsList = parseOrphanBeatsList(storyContent);
				if (orphanBeatsList.items.length > 0) {
					await this.processOrphanBeatsList(orphanBeatsList, storyId);
				}

				// Parse and process prose blocks at story level
				const storyProse = parseStoryProse(storyContent);
				if (storyProse.sections.length > 0) {
					await this.pushStoryContentBlocks(storyFilePath, folderPath, storyId);
				}
			}

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
				await this.pushChapterContentBlocks(chapterFilePath, folderPath);
			}

			// Process scene files to update beat lists and prose blocks
			// Scenes are stored in ${folderPath}/01-scenes (not in chapters/01-scenes)
			const sceneFiles = await this.fileManager.listStorySceneFiles(folderPath);
			
			for (const sceneFilePath of sceneFiles) {
				await this.pushSceneBeats(sceneFilePath, storyId);
				await this.pushSceneContentBlocks(sceneFilePath, folderPath);
			}

			// Process beat files to update prose blocks
			// Beats are stored in ${folderPath}/02-beats
			const beatFiles = await this.fileManager.listStoryBeatFiles(folderPath);
			
			for (const beatFilePath of beatFiles) {
				await this.pushBeatContentBlocks(beatFilePath, folderPath);
			}

			new Notice(`Story "${storyFrontmatter.title}" pushed successfully`);

			// Pull story after push to sync any server-side changes and update local files
			try {
				new Notice(`Syncing story "${storyFrontmatter.title}" from service...`);
				await this.pullStory(storyId);
				new Notice(`Story "${storyFrontmatter.title}" synced successfully`);
			} catch (pullErr) {
				const pullErrorMessage =
					pullErr instanceof Error ? pullErr.message : "Failed to sync story after push";
				new Notice(`Warning: ${pullErrorMessage}`, 5000);
				// Don't throw - push was successful, sync failure is just a warning
			}
		} catch (err) {
			const errorMessage =
				err instanceof Error ? err.message : "Failed to push story";
			new Notice(`Error pushing story: ${errorMessage}`, 5000);
			throw err;
		}
	}

	// Push prose blocks from a chapter file (hierarchical structure)
	async pushChapterContentBlocks(chapterFilePath: string, storyFolderPath: string): Promise<void> {
		// Read chapter file
		const file = this.fileManager.getVault().getAbstractFileByPath(chapterFilePath);
		if (!(file instanceof TFile)) {
			throw new Error(`Chapter file not found: ${chapterFilePath}`);
		}

		const chapterContent = await this.fileManager.getVault().read(file);
		const frontmatter = this.fileManager.parseFrontmatter(chapterContent);

		if (!frontmatter.id || !frontmatter.story_id) {
			throw new Error("Chapter metadata missing ID or story_id");
		}

		const chapterId = frontmatter.id;
		const storyId = frontmatter.story_id;
		const contentsFolderPath = `${storyFolderPath}/03-contents`;

		// Parse and process the "## Scenes & Beats" list first
		const sceneBeatList = parseSceneBeatList(chapterContent);
		await this.processSceneBeatList(sceneBeatList, chapterId, storyId);

		// Get all prose blocks from API for this chapter
		const remoteContentBlocks = await this.apiClient.getContentBlocks(chapterId);
		const remoteContentBlocksMap = new Map<string, ContentBlock>();
		for (const pb of remoteContentBlocks) {
			remoteContentBlocksMap.set(pb.id, pb);
		}

		// Get existing scenes and beats for this chapter (after processing list)
		const existingScenes = await this.apiClient.getScenes(chapterId);
		const sceneMap = new Map<string, Scene>(); // Map by link name
		const sceneIdMap = new Map<string, Scene>(); // Map by ID
		for (const scene of existingScenes) {
			const fileName = this.fileManager.generateSceneFileName(scene);
			const linkName = fileName.replace(/\.md$/, "");
			sceneMap.set(linkName, scene);
			sceneIdMap.set(scene.id, scene);
		}

		// Get all beats for scenes in this chapter
		const beatMap = new Map<string, Beat>(); // Map by link name
		const beatIdMap = new Map<string, Beat>(); // Map by ID
		for (const scene of existingScenes) {
			const beats = await this.apiClient.getBeats(scene.id);
			for (const beat of beats) {
				const fileName = this.fileManager.generateBeatFileName(beat);
				const linkName = fileName.replace(/\.md$/, "");
				beatMap.set(linkName, beat);
				beatIdMap.set(beat.id, beat);
			}
		}

		// Parse hierarchical structure from Chapter section
		const hierarchical = parseHierarchicalProse(chapterContent);

		// Process sections hierarchically
		const updatedSections: string[] = [];
		let currentScene: Scene | null = null;
		let currentBeat: Beat | null = null;
		let proseOrderNum = 1;
		let sceneOrderNum = existingScenes.length > 0 
			? Math.max(...existingScenes.map(s => s.order_num)) + 1 
			: 1;

		for (const section of hierarchical.sections) {
			if (section.type === "scene" && section.scene) {
				const { scene: parsedScene } = section;
				
				if (parsedScene.linkName) {
					// Scene exists - find it
					currentScene = sceneMap.get(parsedScene.linkName) || null;
					if (!currentScene) {
						// Try to find by ID if linkName is actually an ID
						currentScene = sceneIdMap.get(parsedScene.linkName) || null;
					}
					
					if (currentScene) {
						// Update scene if needed
						if (parsedScene.goal !== currentScene.goal || parsedScene.timeRef !== currentScene.time_ref) {
							currentScene = await this.apiClient.updateScene(currentScene.id, {
								goal: parsedScene.goal,
								time_ref: parsedScene.timeRef,
							});
						}
					}
				} else {
					// Create new scene
					currentScene = await this.apiClient.createScene({
						story_id: storyId,
						chapter_id: chapterId,
						order_num: sceneOrderNum++,
						goal: parsedScene.goal,
						time_ref: parsedScene.timeRef,
					});
					
					// Update maps
					const sceneFileName = this.fileManager.generateSceneFileName(currentScene);
					const sceneLinkName = sceneFileName.replace(/\.md$/, "");
					sceneMap.set(sceneLinkName, currentScene);
					sceneIdMap.set(currentScene.id, currentScene);
					
					// Write scene file (scenes are written separately during pull, so we skip here)
					// The scene will be written properly on next pull
				}

				// Format scene header with link and prefix
				if (currentScene) {
					const sceneFileName = this.fileManager.generateSceneFileName(currentScene);
					const sceneLinkName = sceneFileName.replace(/\.md$/, "");
					const sceneDisplayText = currentScene.time_ref
						? `${currentScene.goal} - ${currentScene.time_ref}`
						: currentScene.goal;
					updatedSections.push(`## Scene: [[${sceneLinkName}|${sceneDisplayText}]]`);
				}
				
				currentBeat = null; // Reset beat when new scene starts
			} else if (section.type === "beat" && section.beat) {
				const { beat: parsedBeat } = section;
				
				if (!currentScene) {
					throw new Error("Beat found without a parent scene");
				}
				
				if (parsedBeat.linkName) {
					// Beat exists - find it
					currentBeat = beatMap.get(parsedBeat.linkName) || null;
					if (!currentBeat) {
						// Try to find by ID if linkName is actually an ID
						currentBeat = beatIdMap.get(parsedBeat.linkName) || null;
					}
					
					if (currentBeat) {
						// Update beat if needed
						if (parsedBeat.intent !== currentBeat.intent || parsedBeat.outcome !== currentBeat.outcome) {
							currentBeat = await this.apiClient.updateBeat(currentBeat.id, {
								intent: parsedBeat.intent,
								outcome: parsedBeat.outcome,
							});
						}
					}
				} else {
					// Create new beat
					const existingBeats = await this.apiClient.getBeats(currentScene.id);
					const beatOrderNum = existingBeats.length > 0
						? Math.max(...existingBeats.map(b => b.order_num)) + 1
						: 1;
					
					currentBeat = await this.apiClient.createBeat({
						scene_id: currentScene.id,
						order_num: beatOrderNum,
						type: "setup", // Default type
						intent: parsedBeat.intent,
						outcome: parsedBeat.outcome,
					});
					
					// Update maps
					const beatFileName = this.fileManager.generateBeatFileName(currentBeat);
					const beatLinkName = beatFileName.replace(/\.md$/, "");
					beatMap.set(beatLinkName, currentBeat);
					beatIdMap.set(currentBeat.id, currentBeat);
					
					// Write beat file (beats are written separately during pull, so we skip here)
					// The beat will be written properly on next pull
				}

				// Format beat header with link and prefix
				if (currentBeat) {
					const beatFileName = this.fileManager.generateBeatFileName(currentBeat);
					const beatLinkName = beatFileName.replace(/\.md$/, "");
					const beatDisplayText = currentBeat.outcome
						? `${currentBeat.intent} -> ${currentBeat.outcome}`
						: currentBeat.intent;
					updatedSections.push(`### Beat: [[${beatLinkName}|${beatDisplayText}]]`);
				}
			} else if (section.type === "prose" && section.prose) {
				const { prose: paragraph } = section;
				
				// Process prose block
				let localContentBlock: ContentBlock | null = null;
				let remoteContentBlock: ContentBlock | null = null;

				// If paragraph has a link, read the local file
				if (paragraph.linkName) {
					// Search for file in type subfolders
					const contentBlockFilePath = await this.findContentBlockFileByLinkName(contentsFolderPath, paragraph.linkName);
					if (contentBlockFilePath) {
						localContentBlock = await this.fileManager.readContentBlockFromFile(contentBlockFilePath);
					}

					// If file not found by exact name, try to find by searching all prose block files
					if (!localContentBlock) {
						localContentBlock = await this.findContentBlockByContent(contentsFolderPath, paragraph.content);
					}

					// Find corresponding remote prose block
					if (localContentBlock) {
						remoteContentBlock = remoteContentBlocksMap.get(localContentBlock.id) || null;
					} else {
						// If local file not found, try to find remote by content match
						const normalizedContent = paragraph.content.trim();
						for (const [id, remotePB] of remoteContentBlocksMap.entries()) {
							if (remotePB.content.trim() === normalizedContent) {
								remoteContentBlock = remotePB;
								break;
							}
						}
					}
				} else {
					// No link - check if there's a local file with matching content
					localContentBlock = await this.findContentBlockByContent(contentsFolderPath, paragraph.content);
					
					// Also check remote prose blocks for content match
					const normalizedContent = paragraph.content.trim();
					for (const [id, remotePB] of remoteContentBlocksMap.entries()) {
						if (remotePB.content.trim() === normalizedContent) {
							remoteContentBlock = remotePB;
							// If we found a remote match, try to find local file by ID
							if (!localContentBlock) {
								// Search for file with this ID in frontmatter
								localContentBlock = await this.findContentBlockById(contentsFolderPath, remotePB.id);
							}
							break;
						}
					}
				}

				// Compare and determine status
				const status = compareContentBlocks(paragraph, localContentBlock, remoteContentBlock);

				let finalContentBlock: ContentBlock;

				switch (status) {
					case "new": {
						// Create new prose block
						finalContentBlock = await this.apiClient.createContentBlock(chapterId, {
							order_num: proseOrderNum++,
							kind: "final",
							content: paragraph.content,
						});

						// Create file in correct type subfolder
						const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);

						// Create references if needed
						if (currentScene) {
							await this.apiClient.createContentAnchor(finalContentBlock.id, "scene", currentScene.id);
						}
						if (currentBeat) {
							await this.apiClient.createContentAnchor(finalContentBlock.id, "beat", currentBeat.id);
						}

						// Add link to paragraph
						const linkName = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
						updatedSections.push(`[[${linkName}|${paragraph.content}]]`);
						break;
					}

					case "unchanged": {
						// Use remoteContentBlock if localContentBlock doesn't exist
						if (!localContentBlock && remoteContentBlock) {
							finalContentBlock = remoteContentBlock;
							// Create local file in correct type subfolder since it doesn't exist
							const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
							await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);
						} else if (localContentBlock) {
							// Check if order_num needs update
							if (localContentBlock.order_num !== proseOrderNum) {
								finalContentBlock = await this.apiClient.updateContentBlock(localContentBlock.id, {
									order_num: proseOrderNum++,
								});
								const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
								await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);
							} else {
								finalContentBlock = localContentBlock;
								proseOrderNum++;
							}
						} else {
							// Should not happen, but handle gracefully
							finalContentBlock = remoteContentBlock!;
							proseOrderNum++;
						}

						// Update references if needed
						if (finalContentBlock) {
							const existingRefs = await this.apiClient.getContentAnchors(finalContentBlock.id);
							const hasSceneRef = existingRefs.some(r => r.entity_type === "scene" && r.entity_id === currentScene?.id);
							const hasBeatRef = existingRefs.some(r => r.entity_type === "beat" && r.entity_id === currentBeat?.id);
							
							if (currentScene && !hasSceneRef) {
								await this.apiClient.createContentAnchor(finalContentBlock.id, "scene", currentScene.id);
							}
							if (currentBeat && !hasBeatRef) {
								await this.apiClient.createContentAnchor(finalContentBlock.id, "beat", currentBeat.id);
							}
						}

						// Add link - use existing linkName if available, otherwise generate from file
						if (paragraph.linkName) {
							updatedSections.push(`[[${paragraph.linkName}|${paragraph.content}]]`);
						} else {
							// No linkName - generate from finalContentBlock
							const linkName = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
							updatedSections.push(`[[${linkName}|${paragraph.content}]]`);
						}
						break;
					}

					case "local_modified": {
						// Update prose block with local content
						finalContentBlock = await this.apiClient.updateContentBlock(localContentBlock!.id, {
							content: paragraph.content,
							order_num: proseOrderNum++,
						});

						// Update local file in correct type subfolder
						const filePathLocalMod = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePathLocalMod, undefined);

						// Update references if needed
						if (finalContentBlock) {
							const existingRefs = await this.apiClient.getContentAnchors(finalContentBlock.id);
							const hasSceneRef = existingRefs.some(r => r.entity_type === "scene" && r.entity_id === currentScene?.id);
							const hasBeatRef = existingRefs.some(r => r.entity_type === "beat" && r.entity_id === currentBeat?.id);
							
							if (currentScene && !hasSceneRef) {
								await this.apiClient.createContentAnchor(finalContentBlock.id, "scene", currentScene.id);
							}
							if (currentBeat && !hasBeatRef) {
								await this.apiClient.createContentAnchor(finalContentBlock.id, "beat", currentBeat.id);
							}
						}

						const linkNameLocalMod = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
						updatedSections.push(`[[${linkNameLocalMod}|${paragraph.content}]]`);
						break;
					}

					case "remote_modified": {
						// Update local file with remote content in correct type subfolder
						finalContentBlock = remoteContentBlock!;
						const filePathRemoteMod = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePathRemoteMod, undefined);

						// Update references if needed
						if (finalContentBlock) {
							const existingRefs = await this.apiClient.getContentAnchors(finalContentBlock.id);
							const hasSceneRef = existingRefs.some(r => r.entity_type === "scene" && r.entity_id === currentScene?.id);
							const hasBeatRef = existingRefs.some(r => r.entity_type === "beat" && r.entity_id === currentBeat?.id);
							
							if (currentScene && !hasSceneRef) {
								await this.apiClient.createContentAnchor(finalContentBlock.id, "scene", currentScene.id);
							}
							if (currentBeat && !hasBeatRef) {
								await this.apiClient.createContentAnchor(finalContentBlock.id, "beat", currentBeat.id);
							}
						}

						const linkNameRemoteMod = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
						updatedSections.push(`[[${linkNameRemoteMod}|${finalContentBlock.content}]]`);
						new Notice(`Prose block updated from remote: ${linkNameRemoteMod}`, 3000);
						proseOrderNum++;
						break;
					}

					case "conflict": {
						// Show conflict modal
						const resolution = await this.resolveConflict(localContentBlock!, remoteContentBlock!);

						let resolvedContent: string;
						if (resolution.resolution === "local") {
							resolvedContent = paragraph.content;
						} else if (resolution.resolution === "remote") {
							resolvedContent = remoteContentBlock!.content;
						} else {
							resolvedContent = resolution.mergedContent || paragraph.content;
						}

						// Update prose block with resolved content
						finalContentBlock = await this.apiClient.updateContentBlock(localContentBlock!.id, {
							content: resolvedContent,
							order_num: proseOrderNum++,
						});

						// Update local file in correct type subfolder
						const filePathConflict = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePathConflict, undefined);

						// Update references if needed
						if (finalContentBlock) {
							const existingRefs = await this.apiClient.getContentAnchors(finalContentBlock.id);
							const hasSceneRef = existingRefs.some(r => r.entity_type === "scene" && r.entity_id === currentScene?.id);
							const hasBeatRef = existingRefs.some(r => r.entity_type === "beat" && r.entity_id === currentBeat?.id);
							
							if (currentScene && !hasSceneRef) {
								await this.apiClient.createContentAnchor(finalContentBlock.id, "scene", currentScene.id);
							}
							if (currentBeat && !hasBeatRef) {
								await this.apiClient.createContentAnchor(finalContentBlock.id, "beat", currentBeat.id);
							}
						}

						const linkNameConflict = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
						updatedSections.push(`[[${linkNameConflict}|${resolvedContent}]]`);
						break;
					}
				}
			}
		}

		// Update chapter file with new hierarchical content and scene/beat list
		await this.updateChapterFile(chapterContent, updatedSections, file, frontmatter, existingScenes, beatMap, remoteContentBlocks, chapterId);
	}

	// Update chapter file with both scene/beat list and chapter content
	private async updateChapterFile(
		originalContent: string,
		updatedSections: string[],
		file: TFile,
		frontmatter: Record<string, string>,
		scenes: Scene[],
		beatMap: Map<string, Beat>,
		contentBlocks: ContentBlock[],
		chapterId: string
	): Promise<void> {
		// Get prose block references to determine which scenes/02-beats have prose
		const allContentAnchors: ContentAnchor[] = [];
		for (const contentBlock of contentBlocks) {
			const refs = await this.apiClient.getContentAnchors(contentBlock.id);
			allContentAnchors.push(...refs);
		}

		// Create maps for quick lookup
		const proseRefsByScene = new Map<string, ContentAnchor[]>();
		const proseRefsByBeat = new Map<string, ContentAnchor[]>();
		for (const ref of allContentAnchors) {
			if (ref.entity_type === "scene") {
				if (!proseRefsByScene.has(ref.entity_id)) {
					proseRefsByScene.set(ref.entity_id, []);
				}
				proseRefsByScene.get(ref.entity_id)!.push(ref);
			} else if (ref.entity_type === "beat") {
				if (!proseRefsByBeat.has(ref.entity_id)) {
					proseRefsByBeat.set(ref.entity_id, []);
				}
				proseRefsByBeat.get(ref.entity_id)!.push(ref);
			}
		}

		// Generate updated scene/beat list
		const sceneBeatListItems: string[] = [];
		for (const scene of scenes.sort((a, b) => a.order_num - b.order_num)) {
			const sceneFileName = this.fileManager.generateSceneFileName(scene);
			const sceneLinkName = sceneFileName.replace(/\.md$/, "");
			const sceneDisplayText = scene.time_ref 
				? `${scene.goal} - ${scene.time_ref}`
				: scene.goal;
			
			// Check if scene has prose blocks (not associated with beats)
			const sceneProseRefs = proseRefsByScene.get(scene.id) || [];
			const sceneContentBlockIds = new Set(sceneProseRefs.map((r: ContentAnchor) => r.content_block_id));
			// Check if any prose block is not also associated with a beat
			const hasSceneProse = Array.from(sceneContentBlockIds).some((contentBlockId: string) => {
				const blockRefs = allContentAnchors.filter((r: ContentAnchor) => r.content_block_id === contentBlockId);
				return !blockRefs.some((r: ContentAnchor) => r.entity_type === "beat");
			});
			const sceneMarker = hasSceneProse ? "+" : "-";
			
			sceneBeatListItems.push(`${sceneMarker} [[${sceneLinkName}|Scene ${scene.order_num}: ${sceneDisplayText}]]`);
			
			// Get beats for this scene
			const sceneBeats: Beat[] = [];
			for (const [linkName, beat] of beatMap.entries()) {
				if (beat.scene_id === scene.id) {
					sceneBeats.push(beat);
				}
			}
			
			for (const beat of sceneBeats.sort((a, b) => a.order_num - b.order_num)) {
				const beatFileName = this.fileManager.generateBeatFileName(beat);
				const beatLinkName = beatFileName.replace(/\.md$/, "");
				const beatDisplayText = beat.outcome
					? `${beat.intent} -> ${beat.outcome}`
					: beat.intent;
				
				// Check if beat has prose blocks
				const beatProseRefs = proseRefsByBeat.get(beat.id) || [];
				const hasBeatProse = beatProseRefs.length > 0;
				const beatMarker = hasBeatProse ? "+" : "-";
				
				sceneBeatListItems.push(`\t${beatMarker} [[${beatLinkName}|Beat ${beat.order_num}: ${beatDisplayText}]]`);
			}
		}

		// Extract frontmatter
		const frontmatterMatch = originalContent.match(/^---\n([\s\S]*?)\n---/);
		const frontmatterText = frontmatterMatch ? frontmatterMatch[0] : "";

		// Extract content after frontmatter
		const bodyStart = frontmatterMatch ? frontmatterMatch[0].length : 0;
		const bodyContent = originalContent.substring(bodyStart).trim();

		// Get chapter number and title from frontmatter
		const chapterNumber = frontmatter.number || "1";
		const chapterTitle = frontmatter.title || "Untitled";

		// Update scene/beat list section
		const listSectionMatch = bodyContent.match(/([\s\S]*?##\s+Scenes\s+&\s+Beats\s*\n+)([\s\S]*?)(?=\n##|$)/);
		const updatedListSection = `## Scenes & Beats\n\n${sceneBeatListItems.join("\n")}\n\n`;

		let updatedBody: string;
		if (listSectionMatch) {
			// Replace existing list section
			const beforeList = listSectionMatch[1];
			const afterList = bodyContent.substring(listSectionMatch.index! + listSectionMatch[0].length);
			updatedBody = `${beforeList}${updatedListSection}${afterList}`;
		} else {
			// Add list section after title
			const titleMatch = bodyContent.match(/(#\s+[^\n]+\n+)([\s\S]*)/);
			if (titleMatch) {
				updatedBody = `${titleMatch[1]}${updatedListSection}${titleMatch[2]}`;
			} else {
				updatedBody = `${updatedListSection}${bodyContent}`;
			}
		}

		// Update chapter section
		const chapterHeaderPattern = `##\\s+Chapter\\s+${chapterNumber}:\\s+[^\\n]+`;
		const chapterSectionMatch = updatedBody.match(new RegExp(`([\\s\\S]*?${chapterHeaderPattern}\\s*\\n+)([\\s\\S]*?)(?=\\n##\\s+Chapter\\s+\\d+:|$)`, "i"));
		
		if (!chapterSectionMatch) {
			// No Chapter section found, add it
			const newChapterSection = `\n\n## Chapter ${chapterNumber}: ${chapterTitle}\n\n${updatedSections.join("\n\n")}\n\n`;
			updatedBody = `${updatedBody}${newChapterSection}`;
		} else {
			// Replace Chapter section content
			const beforeChapter = chapterSectionMatch[1];
			const afterChapter = updatedBody.substring(chapterSectionMatch.index! + chapterSectionMatch[0].length);
			const newChapterContent = updatedSections.join("\n\n");
			updatedBody = `${beforeChapter}${newChapterContent}\n\n${afterChapter}`;
		}

		const updatedContent = `${frontmatterText}\n${updatedBody}`;
		await this.fileManager.getVault().modify(file, updatedContent);
	}

	// Process the "## Chapters, Scenes & Beats" list and update order
	private async processChapterList(
		list: ParsedChapterList,
		storyId: string
	): Promise<void> {
		// Get existing chapters for this story
		const existingChapters = await this.apiClient.getChapters(storyId);
		const chapterMap = new Map<string, Chapter>(); // Map by link name
		const chapterIdMap = new Map<string, Chapter>(); // Map by ID
		
		for (const chapter of existingChapters) {
			const fileName = `Chapter-${chapter.number}.md`;
			const linkName = fileName.replace(/\.md$/, "");
			chapterMap.set(linkName, chapter);
			chapterIdMap.set(chapter.id, chapter);
		}

		let currentChapter: Chapter | null = null;
		let currentScene: Scene | null = null;
		let chapterOrderNum = 1; // Start from 1, will increment as we process
		const sceneOrderNums = new Map<string, number>(); // Track order_num per chapter
		const beatOrderNums = new Map<string, number>(); // Track order_num per scene

		// Process list items - order matters!
		for (const item of list.items) {
			if (item.type === "chapter") {
				// Parse chapter display text
				// Can be: "Chapter N: title" or just "title"
				let title: string;
				let chapterNumber: number | null = null;

				const chapterMatch = item.displayText.match(/Chapter\s+(\d+):\s*(.+)/);
				if (chapterMatch) {
					// Format: "Chapter N: title"
					chapterNumber = parseInt(chapterMatch[1], 10);
					title = chapterMatch[2].trim();
				} else {
					// Format: just "title"
					title = item.displayText.trim();
				}

				// Reset scene order for new chapter
				currentScene = null;
				sceneOrderNums.set("current", 1);

				if (item.linkName) {
					// Chapter exists - find it
					currentChapter = chapterMap.get(item.linkName) || null;
					if (!currentChapter) {
						// Try to find by ID if linkName is actually an ID
						currentChapter = chapterIdMap.get(item.linkName) || null;
					}
					
					if (currentChapter) {
						// Check if order_num needs to be updated
						const needsOrderUpdate = currentChapter.number !== chapterOrderNum;
						const needsTitleUpdate = title !== currentChapter.title;
						
						if (needsOrderUpdate || needsTitleUpdate) {
							// Update chapter number and title
							currentChapter = await this.apiClient.updateChapter(currentChapter.id, {
								number: chapterOrderNum,
								title: needsTitleUpdate ? title : undefined,
							});
							
							// Update maps
							const fileName = `Chapter-${currentChapter.number}.md`;
							const linkName = fileName.replace(/\.md$/, "");
							chapterMap.set(linkName, currentChapter);
							chapterIdMap.set(currentChapter.id, currentChapter);
						}
					}
				} else {
					// Create new chapter (if title is provided)
					if (title) {
						currentChapter = await this.apiClient.createChapter(storyId, {
							number: chapterOrderNum,
							title,
							status: "draft",
						});
						
						// Update maps
						const fileName = `Chapter-${currentChapter.number}.md`;
						const linkName = fileName.replace(/\.md$/, "");
						chapterMap.set(linkName, currentChapter);
						chapterIdMap.set(currentChapter.id, currentChapter);
					}
				}
				
				// Initialize scene order for this chapter
				if (currentChapter) {
					sceneOrderNums.set(currentChapter.id, 1);
				}
				
				chapterOrderNum++;
			} else if (item.type === "scene" && currentChapter) {
				// Parse scene display text
				// Can be: "Scene N: goal - timeRef", "goal - timeRef", or just "goal"
				let goal: string;
				let timeRef: string = "";

				const sceneMatch = item.displayText.match(/Scene\s+\d+:\s*(.+)/);
				if (sceneMatch) {
					// Format: "Scene N: goal - timeRef" or "Scene N: goal"
					const sceneText = sceneMatch[1].trim();
					const parts = sceneText.split(/\s*-\s*/);
					goal = parts[0].trim();
					timeRef = parts.length > 1 ? parts.slice(1).join(" - ").trim() : "";
				} else {
					// Format: "goal - timeRef" or just "goal"
					const parts = item.displayText.split(/\s*-\s*/);
					goal = parts[0].trim();
					timeRef = parts.length > 1 ? parts.slice(1).join(" - ").trim() : "";
				}

				// Get current scene order_num for this chapter
				const currentSceneOrderNum = sceneOrderNums.get(currentChapter.id) || 1;

				// Reset currentScene before processing
				currentScene = null;

				if (item.linkName) {
					// Scene exists - find it
					const existingScenes = await this.apiClient.getScenes(currentChapter.id);
					const sceneMap = new Map<string, Scene>();
					const sceneIdMap = new Map<string, Scene>();
					for (const scene of existingScenes) {
						const fileName = this.fileManager.generateSceneFileName(scene);
						const linkName = fileName.replace(/\.md$/, "");
						sceneMap.set(linkName, scene);
						sceneIdMap.set(scene.id, scene);
					}

					currentScene = sceneMap.get(item.linkName) || null;
					if (!currentScene) {
						currentScene = sceneIdMap.get(item.linkName) || null;
					}
					
					if (currentScene) {
						// Check if order_num or content needs to be updated
						const needsOrderUpdate = currentScene.order_num !== currentSceneOrderNum;
						const needsContentUpdate = goal !== currentScene.goal || timeRef !== currentScene.time_ref;
						const needsChapterUpdate = currentScene.chapter_id !== currentChapter.id;
						
						if (needsOrderUpdate || needsContentUpdate || needsChapterUpdate) {
							currentScene = await this.apiClient.updateScene(currentScene.id, {
								goal,
								time_ref: timeRef,
								order_num: currentSceneOrderNum,
								chapter_id: currentChapter.id,
							});
						}
						
						// Initialize beat order for this scene (important for empty scenes)
						// Check if scene already has beats, if not, start from 1
						const existingBeats = await this.apiClient.getBeats(currentScene.id);
						if (existingBeats.length === 0) {
							beatOrderNums.set(currentScene.id, 1);
						} else {
							// Scene has beats, use max order_num + 1 for new beats
							const maxOrderNum = Math.max(...existingBeats.map(b => b.order_num));
							beatOrderNums.set(currentScene.id, maxOrderNum + 1);
						}
					}
				} else {
					// Create new scene
					currentScene = await this.apiClient.createScene({
						story_id: storyId,
						chapter_id: currentChapter.id,
						order_num: currentSceneOrderNum,
						goal,
						time_ref: timeRef,
					});
					
					// Initialize beat order for new scene (starts at 1)
					beatOrderNums.set(currentScene.id, 1);
				}
				
				// Increment scene order for next scene in this chapter
				sceneOrderNums.set(currentChapter.id, currentSceneOrderNum + 1);
			} else if (item.type === "beat" && currentScene) {
				// Parse beat display text
				// Can be: "Beat N: intent -> outcome", "intent -> outcome", or just "intent"
				let intent: string;
				let outcome: string = "";

				const beatMatch = item.displayText.match(/Beat\s+\d+:\s*(.+)/);
				if (beatMatch) {
					// Format: "Beat N: intent -> outcome" or "Beat N: intent"
					const beatText = beatMatch[1].trim();
					const parts = beatText.split(/\s*->\s*/);
					intent = parts[0].trim();
					outcome = parts.length > 1 ? parts.slice(1).join(" -> ").trim() : "";
				} else {
					// Format: "intent -> outcome" or just "intent"
					const parts = item.displayText.split(/\s*->\s*/);
					intent = parts[0].trim();
					outcome = parts.length > 1 ? parts.slice(1).join(" -> ").trim() : "";
				}

				// Get current beat order_num for this scene
				const currentBeatOrderNum = beatOrderNums.get(currentScene.id) || 1;

				if (item.linkName) {
					// Beat exists - find it
					const existingBeats = await this.apiClient.getBeats(currentScene.id);
					const beatMap = new Map<string, Beat>();
					const beatIdMap = new Map<string, Beat>();
					for (const beat of existingBeats) {
						const fileName = this.fileManager.generateBeatFileName(beat);
						const linkName = fileName.replace(/\.md$/, "");
						beatMap.set(linkName, beat);
						beatIdMap.set(beat.id, beat);
					}

					let currentBeat = beatMap.get(item.linkName) || null;
					if (!currentBeat) {
						currentBeat = beatIdMap.get(item.linkName) || null;
					}
					
					if (currentBeat) {
						// Check if order_num or content needs to be updated
						const needsOrderUpdate = currentBeat.order_num !== currentBeatOrderNum;
						const needsContentUpdate = intent !== currentBeat.intent || outcome !== currentBeat.outcome;
						const needsSceneUpdate = currentBeat.scene_id !== currentScene.id;
						
						if (needsOrderUpdate || needsContentUpdate || needsSceneUpdate) {
							currentBeat = await this.apiClient.updateBeat(currentBeat.id, {
								intent,
								outcome,
								order_num: currentBeatOrderNum,
								scene_id: currentScene.id,
							});
						}
					}
				} else {
					// Create new beat
					await this.apiClient.createBeat({
						scene_id: currentScene.id,
						order_num: currentBeatOrderNum,
						type: "setup", // Default type
						intent,
						outcome,
					});
				}
				
				// Increment beat order for next beat in this scene
				beatOrderNums.set(currentScene.id, currentBeatOrderNum + 1);
			}
		}
	}

	// Process the "## Beats" list from a scene file and update beat order
	private async pushSceneBeats(sceneFilePath: string, storyId: string): Promise<void> {
		// Read scene file
		const file = this.fileManager.getVault().getAbstractFileByPath(sceneFilePath);
		if (!(file instanceof TFile)) {
			return;
		}

		const sceneContent = await this.fileManager.getVault().read(file);
		const frontmatter = this.fileManager.parseFrontmatter(sceneContent);

		if (!frontmatter.id) {
			return;
		}

		const sceneId = frontmatter.id;

		// Parse beat list from scene file
		const beatList = parseBeatList(sceneContent);
		
		if (beatList.items.length === 0) {
			return;
		}

		// Get existing beats for this scene
		const existingBeats = await this.apiClient.getBeats(sceneId);
		const beatMap = new Map<string, Beat>(); // Map by link name
		const beatIdMap = new Map<string, Beat>(); // Map by ID
		
		for (const beat of existingBeats) {
			const fileName = this.fileManager.generateBeatFileName(beat);
			const linkName = fileName.replace(/\.md$/, "");
			beatMap.set(linkName, beat);
			beatIdMap.set(beat.id, beat);
		}

		let beatOrderNum = 1; // Start from 1, will increment as we process

		// Process list items - order matters!
		for (const item of beatList.items) {
			// Parse beat display text
			// Can be: "Beat N: intent -> outcome", "intent -> outcome", or just "intent"
			let intent: string;
			let outcome: string = "";

			const beatMatch = item.displayText.match(/Beat\s+\d+:\s*(.+)/);
			if (beatMatch) {
				// Format: "Beat N: intent -> outcome" or "Beat N: intent"
				const beatText = beatMatch[1].trim();
				const parts = beatText.split(/\s*->\s*/);
				intent = parts[0].trim();
				outcome = parts.length > 1 ? parts.slice(1).join(" -> ").trim() : "";
			} else {
				// Format: "intent -> outcome" or just "intent"
				const parts = item.displayText.split(/\s*->\s*/);
				intent = parts[0].trim();
				outcome = parts.length > 1 ? parts.slice(1).join(" -> ").trim() : "";
			}

			if (item.linkName) {
				// Beat exists - find it
				let currentBeat = beatMap.get(item.linkName) || null;
				if (!currentBeat) {
					// Try to find by ID if linkName is actually an ID
					currentBeat = beatIdMap.get(item.linkName) || null;
				}
				
				if (currentBeat) {
					// Check if order_num or content needs to be updated
					const needsOrderUpdate = currentBeat.order_num !== beatOrderNum;
					const needsContentUpdate = intent !== currentBeat.intent || outcome !== currentBeat.outcome;
					
					if (needsOrderUpdate || needsContentUpdate) {
						currentBeat = await this.apiClient.updateBeat(currentBeat.id, {
							intent,
							outcome,
							order_num: beatOrderNum,
						});
					}
				}
			} else {
				// Create new beat
				await this.apiClient.createBeat({
					scene_id: sceneId,
					order_num: beatOrderNum,
					type: "setup", // Default type
					intent,
					outcome,
				});
			}
			
			beatOrderNum++;
		}
	}

	// Push prose blocks from a story file with hierarchical structure
	// Format: # Story: title, ## Chapter: title, ### Scene: title, #### Beat: title
	private async pushStoryContentBlocks(storyFilePath: string, storyFolderPath: string, storyId: string): Promise<void> {
		const file = this.fileManager.getVault().getAbstractFileByPath(storyFilePath);
		if (!(file instanceof TFile)) {
			throw new Error(`Story file not found: ${storyFilePath}`);
		}

		const storyContent = await this.fileManager.getVault().read(file);
		const contentsFolderPath = `${storyFolderPath}/03-contents`;

		// Parse hierarchical prose from story content
		const storyProse = parseStoryProse(storyContent);
		
		if (storyProse.sections.length === 0) {
			return;
		}

		// Get existing chapters, scenes, beats
		const existingChapters = await this.apiClient.getChapters(storyId);
		const chapterByTitle = new Map<string, Chapter>();
		for (const ch of existingChapters) {
			chapterByTitle.set(ch.title.toLowerCase(), ch);
		}

		// Track current context
		let currentChapter: Chapter | null = null;
		let currentScene: Scene | null = null;
		let currentBeat: Beat | null = null;
		let proseOrderNum = 1;

		// Maps for scenes and beats
		const sceneMap = new Map<string, Scene>();
		const beatMap = new Map<string, Beat>();

		// Cache for remote content blocks by chapter ID (avoids N+1 queries)
		const remoteContentBlocksCache = new Map<string, Map<string, ContentBlock>>();
		const getRemoteContentBlocksMap = async (chapterId: string): Promise<Map<string, ContentBlock>> => {
			if (!remoteContentBlocksCache.has(chapterId)) {
				const blocks = await this.apiClient.getContentBlocks(chapterId);
				const map = new Map<string, ContentBlock>();
				for (const pb of blocks) {
					map.set(pb.id, pb);
				}
				remoteContentBlocksCache.set(chapterId, map);
			}
			return remoteContentBlocksCache.get(chapterId)!;
		};

		// Process each section
		for (const section of storyProse.sections) {
			if (section.type === "scene" && section.scene) {
				const { scene: parsedScene } = section;
				
				// If no current chapter, create a default one for orphan scenes
				if (!currentChapter) {
					currentChapter = chapterByTitle.get("story prose") || null;
					if (!currentChapter) {
						currentChapter = await this.apiClient.createChapter(storyId, {
							number: 9999,
							title: "Story Prose",
							status: "draft",
						});
						chapterByTitle.set("story prose", currentChapter);
					}
				}

				// Find or create scene
				if (parsedScene.linkName) {
					currentScene = sceneMap.get(parsedScene.linkName) || null;
				}
				
				if (!currentScene && parsedScene.goal) {
					// Try to find by goal
					const allScenes = await this.apiClient.getScenes(currentChapter.id);
					currentScene = allScenes.find(s => s.goal === parsedScene.goal) || null;
				}

				if (!currentScene) {
					// Create new scene
					const existingScenes = await this.apiClient.getScenes(currentChapter.id);
					const sceneOrderNum = existingScenes.length > 0
						? Math.max(...existingScenes.map(s => s.order_num)) + 1
						: 1;

					currentScene = await this.apiClient.createScene({
						story_id: storyId,
						chapter_id: currentChapter.id,
						order_num: sceneOrderNum,
						goal: parsedScene.goal,
						time_ref: parsedScene.timeRef,
					});
					// Scene file will be created during pull
				}

				if (currentScene) {
					const sceneFileName = this.fileManager.generateSceneFileName(currentScene);
					const sceneLinkName = sceneFileName.replace(/\.md$/, "");
					sceneMap.set(sceneLinkName, currentScene);
				}

				currentBeat = null;
				proseOrderNum = 1;

			} else if (section.type === "beat" && section.beat) {
				const { beat: parsedBeat } = section;
				
				if (!currentScene) {
					// Can't create beat without scene - skip
					continue;
				}

				// Find or create beat
				if (parsedBeat.linkName) {
					currentBeat = beatMap.get(parsedBeat.linkName) || null;
				}

				if (!currentBeat && parsedBeat.intent) {
					const allBeats = await this.apiClient.getBeats(currentScene.id);
					currentBeat = allBeats.find(b => b.intent === parsedBeat.intent) || null;
				}

				if (!currentBeat) {
					const existingBeats = await this.apiClient.getBeats(currentScene.id);
					const beatOrderNum = existingBeats.length > 0
						? Math.max(...existingBeats.map(b => b.order_num)) + 1
						: 1;

					currentBeat = await this.apiClient.createBeat({
						scene_id: currentScene.id,
						order_num: beatOrderNum,
						type: "setup",
						intent: parsedBeat.intent,
						outcome: parsedBeat.outcome,
					});
					// Beat file will be created during pull
				}

				if (currentBeat) {
					const beatFileName = this.fileManager.generateBeatFileName(currentBeat);
					const beatLinkName = beatFileName.replace(/\.md$/, "");
					beatMap.set(beatLinkName, currentBeat);
				}

			} else if (section.type === "prose" && section.prose) {
				const { prose: paragraph } = section;
				
				// Need a chapter for prose blocks
				if (!currentChapter) {
					currentChapter = chapterByTitle.get("story prose") || null;
					if (!currentChapter) {
						currentChapter = await this.apiClient.createChapter(storyId, {
							number: 9999,
							title: "Story Prose",
							status: "draft",
						});
						chapterByTitle.set("story prose", currentChapter);
					}
				}

				// Get remote prose blocks for current chapter (using cache)
				const remoteContentBlocksMap = await getRemoteContentBlocksMap(currentChapter.id);

				let localContentBlock: ContentBlock | null = null;
				let remoteContentBlock: ContentBlock | null = null;

				if (paragraph.linkName) {
					// Search for file in type subfolders
					const contentBlockFilePath = await this.findContentBlockFileByLinkName(contentsFolderPath, paragraph.linkName);
					if (contentBlockFilePath) {
						localContentBlock = await this.fileManager.readContentBlockFromFile(contentBlockFilePath);
					}

					if (!localContentBlock) {
						localContentBlock = await this.findContentBlockByContent(contentsFolderPath, paragraph.content);
					}

					if (localContentBlock) {
						remoteContentBlock = remoteContentBlocksMap.get(localContentBlock.id) || null;
					} else {
						const normalizedContent = paragraph.content.trim();
						for (const [, remotePB] of remoteContentBlocksMap.entries()) {
							if (remotePB.content.trim() === normalizedContent) {
								remoteContentBlock = remotePB;
								break;
							}
						}
					}
				} else {
					localContentBlock = await this.findContentBlockByContent(contentsFolderPath, paragraph.content);
					
					const normalizedContent = paragraph.content.trim();
					for (const [, remotePB] of remoteContentBlocksMap.entries()) {
						if (remotePB.content.trim() === normalizedContent) {
							remoteContentBlock = remotePB;
							if (!localContentBlock) {
								localContentBlock = await this.findContentBlockById(contentsFolderPath, remotePB.id);
							}
							break;
						}
					}
				}

				const status = compareContentBlocks(paragraph, localContentBlock, remoteContentBlock);
				let finalContentBlock: ContentBlock;

				switch (status) {
					case "new": {
						finalContentBlock = await this.apiClient.createContentBlock(currentChapter.id, {
							order_num: proseOrderNum++,
							kind: "final",
							content: paragraph.content,
						});

						// Create file in correct type subfolder
						const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);

						// Create references
						if (currentScene) {
							await this.apiClient.createContentAnchor(finalContentBlock.id, "scene", currentScene.id);
						}
						if (currentBeat) {
							await this.apiClient.createContentAnchor(finalContentBlock.id, "beat", currentBeat.id);
						}
						break;
					}

					case "unchanged": {
						if (!localContentBlock && remoteContentBlock) {
							finalContentBlock = remoteContentBlock;
							const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
							await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);
						} else if (localContentBlock) {
							finalContentBlock = localContentBlock;
							proseOrderNum++;
						} else {
							finalContentBlock = remoteContentBlock!;
							proseOrderNum++;
						}
						break;
					}

					case "local_modified": {
						finalContentBlock = await this.apiClient.updateContentBlock(localContentBlock!.id, {
							content: paragraph.content,
							order_num: proseOrderNum++,
						});

						const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);
						break;
					}

					case "remote_modified": {
						finalContentBlock = remoteContentBlock!;
						const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);
						proseOrderNum++;
						break;
					}

					case "conflict": {
						const resolution = await this.resolveConflict(localContentBlock!, remoteContentBlock!);
						let resolvedContent = resolution.resolution === "local" 
							? paragraph.content 
							: resolution.resolution === "remote" 
								? remoteContentBlock!.content 
								: resolution.mergedContent || paragraph.content;

						finalContentBlock = await this.apiClient.updateContentBlock(localContentBlock!.id, {
							content: resolvedContent,
							order_num: proseOrderNum++,
						});

						const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);
						break;
					}
				}
			}
		}
	}

	// Push prose blocks from a scene file (scene-level prose, not inside chapters)
	private async pushSceneContentBlocks(sceneFilePath: string, storyFolderPath: string): Promise<void> {
		// Read scene file
		const file = this.fileManager.getVault().getAbstractFileByPath(sceneFilePath);
		if (!(file instanceof TFile)) {
			return;
		}

		const sceneContent = await this.fileManager.getVault().read(file);
		const frontmatter = this.fileManager.parseFrontmatter(sceneContent);

		if (!frontmatter.id || !frontmatter.story_id) {
			return;
		}

		const sceneId = frontmatter.id;
		const storyId = frontmatter.story_id;
		const contentsFolderPath = `${storyFolderPath}/03-contents`;

		// Parse prose blocks from scene content
		const sceneProse = parseSceneProse(sceneContent);
		
		if (sceneProse.sections.length === 0) {
			return; // No prose blocks to process
		}

		// Get or create a temporary chapter for scene-level prose blocks
		const chapters = await this.apiClient.getChapters(storyId);
		let tempChapter = chapters.find(c => c.title === "Scene-Level Prose");
		
		if (!tempChapter) {
			tempChapter = await this.apiClient.createChapter(storyId, {
				number: 9998, // High number to keep it at the end
				title: "Scene-Level Prose",
				status: "draft",
			});
		}

		// Get prose blocks referenced to this scene
		const remoteContentBlocks = await this.apiClient.getContentBlocks(tempChapter.id);
		const remoteContentBlocksMap = new Map<string, ContentBlock>();
		for (const pb of remoteContentBlocks) {
			remoteContentBlocksMap.set(pb.id, pb);
		}

		// Get beats for this scene
		const existingBeats = await this.apiClient.getBeats(sceneId);
		const beatMap = new Map<string, Beat>();
		const beatIdMap = new Map<string, Beat>();
		for (const beat of existingBeats) {
			const fileName = this.fileManager.generateBeatFileName(beat);
			const linkName = fileName.replace(/\.md$/, "");
			beatMap.set(linkName, beat);
			beatIdMap.set(beat.id, beat);
		}

		// Process prose sections
		let proseOrderNum = 1;
		let currentBeat: Beat | null = null;
		const updatedSections: string[] = [];

		for (const section of sceneProse.sections) {
			if (section.type === "beat" && section.beat) {
				const { beat: parsedBeat } = section;
				
				if (parsedBeat.linkName) {
					currentBeat = beatMap.get(parsedBeat.linkName) || null;
					if (!currentBeat) {
						currentBeat = beatIdMap.get(parsedBeat.linkName) || null;
					}
				}
				
				if (currentBeat) {
					const beatFileName = this.fileManager.generateBeatFileName(currentBeat);
					const beatLinkName = beatFileName.replace(/\.md$/, "");
					const beatDisplayText = currentBeat.outcome
						? `${currentBeat.intent} -> ${currentBeat.outcome}`
						: currentBeat.intent;
					updatedSections.push(`### Beat: [[${beatLinkName}|${beatDisplayText}]]`);
				}
			} else if (section.type === "prose" && section.prose) {
				const { prose: paragraph } = section;
				
				let localContentBlock: ContentBlock | null = null;
				let remoteContentBlock: ContentBlock | null = null;

				if (paragraph.linkName) {
					// Search for file in type subfolders
					const contentBlockFilePath = await this.findContentBlockFileByLinkName(contentsFolderPath, paragraph.linkName);
					if (contentBlockFilePath) {
						localContentBlock = await this.fileManager.readContentBlockFromFile(contentBlockFilePath);
					}

					if (!localContentBlock) {
						localContentBlock = await this.findContentBlockByContent(contentsFolderPath, paragraph.content);
					}

					if (localContentBlock) {
						remoteContentBlock = remoteContentBlocksMap.get(localContentBlock.id) || null;
					} else {
						const normalizedContent = paragraph.content.trim();
						for (const [id, remotePB] of remoteContentBlocksMap.entries()) {
							if (remotePB.content.trim() === normalizedContent) {
								remoteContentBlock = remotePB;
								break;
							}
						}
					}
				} else {
					localContentBlock = await this.findContentBlockByContent(contentsFolderPath, paragraph.content);
					
					const normalizedContent = paragraph.content.trim();
					for (const [id, remotePB] of remoteContentBlocksMap.entries()) {
						if (remotePB.content.trim() === normalizedContent) {
							remoteContentBlock = remotePB;
							if (!localContentBlock) {
								localContentBlock = await this.findContentBlockById(contentsFolderPath, remotePB.id);
							}
							break;
						}
					}
				}

				const status = compareContentBlocks(paragraph, localContentBlock, remoteContentBlock);

				let finalContentBlock: ContentBlock;

				switch (status) {
					case "new": {
						finalContentBlock = await this.apiClient.createContentBlock(tempChapter!.id, {
							order_num: proseOrderNum++,
							kind: "final",
							content: paragraph.content,
						});

						// Create file in correct type subfolder
						const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);

						// Create reference to scene
						await this.apiClient.createContentAnchor(finalContentBlock.id, "scene", sceneId);
						
						// Also create reference to beat if we're under one
						if (currentBeat) {
							await this.apiClient.createContentAnchor(finalContentBlock.id, "beat", currentBeat.id);
						}

						const linkName = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
						updatedSections.push(`[[${linkName}|${paragraph.content}]]`);
						break;
					}

					case "unchanged": {
						if (!localContentBlock && remoteContentBlock) {
							finalContentBlock = remoteContentBlock;
							const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
							await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);
						} else if (localContentBlock) {
							if (localContentBlock.order_num !== proseOrderNum) {
								finalContentBlock = await this.apiClient.updateContentBlock(localContentBlock.id, {
									order_num: proseOrderNum++,
								});
								const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
								await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);
							} else {
								finalContentBlock = localContentBlock;
								proseOrderNum++;
							}
						} else {
							finalContentBlock = remoteContentBlock!;
							proseOrderNum++;
						}

						if (paragraph.linkName) {
							updatedSections.push(`[[${paragraph.linkName}|${paragraph.content}]]`);
						} else {
							const linkName = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
							updatedSections.push(`[[${linkName}|${paragraph.content}]]`);
						}
						break;
					}

					case "local_modified": {
						finalContentBlock = await this.apiClient.updateContentBlock(localContentBlock!.id, {
							content: paragraph.content,
							order_num: proseOrderNum++,
						});

						const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);

						const linkName = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
						updatedSections.push(`[[${linkName}|${paragraph.content}]]`);
						break;
					}

					case "remote_modified": {
						finalContentBlock = remoteContentBlock!;
						const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);

						const linkName = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
						updatedSections.push(`[[${linkName}|${finalContentBlock.content}]]`);
						new Notice(`Scene prose block updated from remote: ${linkName}`, 3000);
						proseOrderNum++;
						break;
					}

					case "conflict": {
						const resolution = await this.resolveConflict(localContentBlock!, remoteContentBlock!);

						let resolvedContent: string;
						if (resolution.resolution === "local") {
							resolvedContent = paragraph.content;
						} else if (resolution.resolution === "remote") {
							resolvedContent = remoteContentBlock!.content;
						} else {
							resolvedContent = resolution.mergedContent || paragraph.content;
						}

						finalContentBlock = await this.apiClient.updateContentBlock(localContentBlock!.id, {
							content: resolvedContent,
							order_num: proseOrderNum++,
						});

						const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);

						const linkName = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
						updatedSections.push(`[[${linkName}|${resolvedContent}]]`);
						break;
					}
				}
			}
		}

		// Update scene file with links if there are updated sections
		if (updatedSections.length > 0) {
			// Find the insertion point in the scene file
			const frontmatterMatch = sceneContent.match(/^---\n([\s\S]*?)\n---/);
			const frontmatterEnd = frontmatterMatch ? frontmatterMatch[0].length : 0;
			const afterFrontmatter = sceneContent.substring(frontmatterEnd).trim();
			
			// Find the ## Beats section to insert before it
			const beatsSectionMatch = afterFrontmatter.match(/\n##\s+Beats\s*\n/);
			const insertionPoint = beatsSectionMatch 
				? frontmatterEnd + afterFrontmatter.indexOf(beatsSectionMatch[0])
				: sceneContent.length;
			
			// Build updated content with prose sections
			const beforeProse = sceneContent.substring(0, insertionPoint).trimEnd();
			const afterProse = sceneContent.substring(insertionPoint);
			
			const updatedContent = `${beforeProse}\n\n${updatedSections.join("\n\n")}\n${afterProse}`;
			await this.fileManager.getVault().modify(file, updatedContent);
		}
	}

	// Push prose blocks from a beat file
	private async pushBeatContentBlocks(beatFilePath: string, storyFolderPath: string): Promise<void> {
		// Read beat file
		const file = this.fileManager.getVault().getAbstractFileByPath(beatFilePath);
		if (!(file instanceof TFile)) {
			return;
		}

		const beatContent = await this.fileManager.getVault().read(file);
		const frontmatter = this.fileManager.parseFrontmatter(beatContent);

		if (!frontmatter.id || !frontmatter.scene_id) {
			return;
		}

		const beatId = frontmatter.id;
		const sceneId = frontmatter.scene_id;
		const contentsFolderPath = `${storyFolderPath}/03-contents`;

		// Parse prose blocks from beat content
		const beatProse = parseBeatProse(beatContent);
		
		if (beatProse.sections.length === 0) {
			return; // No prose blocks to process
		}

		// Get the scene to find story_id and chapter_id
		const scene = await this.apiClient.getScene(sceneId);
		if (!scene) {
			return;
		}

		const storyId = scene.story_id;

		// Get or create a temporary chapter for beat-level prose blocks
		const chapters = await this.apiClient.getChapters(storyId);
		let tempChapter = chapters.find(c => c.title === "Beat-Level Prose");
		
		if (!tempChapter) {
			tempChapter = await this.apiClient.createChapter(storyId, {
				number: 9997, // High number to keep it at the end
				title: "Beat-Level Prose",
				status: "draft",
			});
		}

		// Get existing remote prose blocks
		const remoteContentBlocks = await this.apiClient.getContentBlocks(tempChapter.id);
		const remoteContentBlocksMap = new Map<string, ContentBlock>();
		for (const pb of remoteContentBlocks) {
			remoteContentBlocksMap.set(pb.id, pb);
		}

		// Process prose sections
		let proseOrderNum = 1;
		const updatedSections: string[] = [];

		// Add beat header
		const beat = await this.apiClient.getBeat(beatId);
		if (beat) {
			const beatFileName = this.fileManager.generateBeatFileName(beat);
			const beatLinkName = beatFileName.replace(/\.md$/, "");
			const beatDisplayText = beat.outcome
				? `${beat.intent} -> ${beat.outcome}`
				: beat.intent;
			updatedSections.push(`## Beat: [[${beatLinkName}|${beatDisplayText}]]`);
		}

		for (const section of beatProse.sections) {
			if (section.type === "prose" && section.prose) {
				const { prose: paragraph } = section;
				
				let localContentBlock: ContentBlock | null = null;
				let remoteContentBlock: ContentBlock | null = null;

				if (paragraph.linkName) {
					// Search for file in type subfolders
					const contentBlockFilePath = await this.findContentBlockFileByLinkName(contentsFolderPath, paragraph.linkName);
					if (contentBlockFilePath) {
						localContentBlock = await this.fileManager.readContentBlockFromFile(contentBlockFilePath);
					}

					if (!localContentBlock) {
						localContentBlock = await this.findContentBlockByContent(contentsFolderPath, paragraph.content);
					}

					if (localContentBlock) {
						remoteContentBlock = remoteContentBlocksMap.get(localContentBlock.id) || null;
					} else {
						const normalizedContent = paragraph.content.trim();
						for (const [, remotePB] of remoteContentBlocksMap.entries()) {
							if (remotePB.content.trim() === normalizedContent) {
								remoteContentBlock = remotePB;
								break;
							}
						}
					}
				} else {
					localContentBlock = await this.findContentBlockByContent(contentsFolderPath, paragraph.content);
					
					const normalizedContent = paragraph.content.trim();
					for (const [, remotePB] of remoteContentBlocksMap.entries()) {
						if (remotePB.content.trim() === normalizedContent) {
							remoteContentBlock = remotePB;
							if (!localContentBlock) {
								localContentBlock = await this.findContentBlockById(contentsFolderPath, remotePB.id);
							}
							break;
						}
					}
				}

				const status = compareContentBlocks(paragraph, localContentBlock, remoteContentBlock);

				let finalContentBlock: ContentBlock;

				switch (status) {
					case "new": {
						finalContentBlock = await this.apiClient.createContentBlock(tempChapter!.id, {
							order_num: proseOrderNum++,
							kind: "final",
							content: paragraph.content,
						});

						// Create file in correct type subfolder
						const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);

						// Create reference to scene and beat
						await this.apiClient.createContentAnchor(finalContentBlock.id, "scene", sceneId);
						await this.apiClient.createContentAnchor(finalContentBlock.id, "beat", beatId);

						const linkName = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
						updatedSections.push(`[[${linkName}|${paragraph.content}]]`);
						break;
					}

					case "unchanged": {
						if (!localContentBlock && remoteContentBlock) {
							finalContentBlock = remoteContentBlock;
							const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
							await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);
							
							const linkName = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
							updatedSections.push(`[[${linkName}|${remoteContentBlock.content}]]`);
						} else if (localContentBlock) {
							finalContentBlock = localContentBlock;
							const linkName = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
							updatedSections.push(`[[${linkName}|${paragraph.content}]]`);
							proseOrderNum++;
						} else {
							proseOrderNum++;
							continue;
						}
						break;
					}

					case "local_modified": {
						finalContentBlock = await this.apiClient.updateContentBlock(localContentBlock!.id, {
							content: paragraph.content,
							order_num: proseOrderNum++,
						});

						const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);

						const linkName = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
						updatedSections.push(`[[${linkName}|${paragraph.content}]]`);
						break;
					}

					case "remote_modified": {
						finalContentBlock = remoteContentBlock!;
						const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);
						proseOrderNum++;

						const linkName = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
						updatedSections.push(`[[${linkName}|${remoteContentBlock!.content}]]`);
						break;
					}

					case "conflict": {
						const resolution = await this.resolveConflict(localContentBlock!, remoteContentBlock!);

						let resolvedContent: string;
						if (resolution.resolution === "local") {
							resolvedContent = paragraph.content;
						} else if (resolution.resolution === "remote") {
							resolvedContent = remoteContentBlock!.content;
						} else {
							resolvedContent = resolution.mergedContent || paragraph.content;
						}

						finalContentBlock = await this.apiClient.updateContentBlock(localContentBlock!.id, {
							content: resolvedContent,
							order_num: proseOrderNum++,
						});

						const filePath = await this.getContentBlockFilePathFromContents(contentsFolderPath, finalContentBlock);
						await this.fileManager.writeContentBlockFile(finalContentBlock, filePath, undefined);

						const linkName = this.fileManager.generateContentBlockFileName(finalContentBlock).replace(/\.md$/, "");
						updatedSections.push(`[[${linkName}|${resolvedContent}]]`);
						break;
					}
				}
			}
		}

		// Update beat file with links if there are updated sections (more than just the header)
		if (updatedSections.length > 1) {
			// Find the insertion point in the beat file (after # Beat header and intent/outcome)
			const frontmatterMatch = beatContent.match(/^---\n([\s\S]*?)\n---/);
			const frontmatterEnd = frontmatterMatch ? frontmatterMatch[0].length : 0;
			const afterFrontmatter = beatContent.substring(frontmatterEnd).trim();
			
			// Find the ## Beat: section to replace it
			const beatSectionMatch = afterFrontmatter.match(/##\s+Beat:\s*.+[\s\S]*/);
			
			if (beatSectionMatch) {
				// Replace the entire ## Beat section
				const beforeBeatSection = beatContent.substring(0, frontmatterEnd + afterFrontmatter.indexOf(beatSectionMatch[0]));
				const updatedContent = `${beforeBeatSection.trimEnd()}\n\n${updatedSections.join("\n\n")}\n`;
				await this.fileManager.getVault().modify(file, updatedContent);
			} else {
				// Append ## Beat section
				const updatedContent = `${beatContent.trimEnd()}\n\n${updatedSections.join("\n\n")}\n`;
				await this.fileManager.getVault().modify(file, updatedContent);
			}
		}
	}

	// Process the "## Orphan Scenes" list and update orphan scenes order
	private async processOrphanScenesList(
		list: { items: ParsedSceneBeatListItem[] },
		storyId: string
	): Promise<void> {
		// Get all scenes for this story
		const allScenes = await this.apiClient.getScenesByStory(storyId);
		const sceneMap = new Map<string, Scene>(); // Map by link name
		const sceneIdMap = new Map<string, Scene>(); // Map by ID
		
		// Filter to only orphan scenes (without chapter_id)
		const orphanScenes = allScenes.filter(s => !s.chapter_id);
		
		for (const scene of orphanScenes) {
			const fileName = this.fileManager.generateSceneFileName(scene);
			const linkName = fileName.replace(/\.md$/, "");
			sceneMap.set(linkName, scene);
			sceneIdMap.set(scene.id, scene);
		}

		// Get all beats for orphan scenes
		const beatMap = new Map<string, Beat>(); // Map by link name
		const beatIdMap = new Map<string, Beat>(); // Map by ID
		for (const scene of orphanScenes) {
			const beats = await this.apiClient.getBeats(scene.id);
			for (const beat of beats) {
				const fileName = this.fileManager.generateBeatFileName(beat);
				const linkName = fileName.replace(/\.md$/, "");
				beatMap.set(linkName, beat);
				beatIdMap.set(beat.id, beat);
			}
		}

		let currentScene: Scene | null = null;
		let sceneOrderNum = 1; // Start from 1, will increment as we process
		const beatOrderNums = new Map<string, number>(); // Track order_num per scene

		// Process list items - order matters!
		for (const item of list.items) {
			if (item.type === "scene") {
				// Parse scene display text
				// Can be: "Scene N: goal - timeRef", "goal - timeRef", or just "goal"
				let goal: string;
				let timeRef: string = "";

				const sceneMatch = item.displayText.match(/Scene\s+\d+:\s*(.+)/);
				if (sceneMatch) {
					// Format: "Scene N: goal - timeRef" or "Scene N: goal"
					const sceneText = sceneMatch[1].trim();
					const parts = sceneText.split(/\s*-\s*/);
					goal = parts[0].trim();
					timeRef = parts.length > 1 ? parts.slice(1).join(" - ").trim() : "";
				} else {
					// Format: "goal - timeRef" or just "goal"
					const parts = item.displayText.split(/\s*-\s*/);
					goal = parts[0].trim();
					timeRef = parts.length > 1 ? parts.slice(1).join(" - ").trim() : "";
				}

				if (item.linkName) {
					// Scene exists - find it
					currentScene = sceneMap.get(item.linkName) || null;
					if (!currentScene) {
						// Try to find by ID if linkName is actually an ID
						currentScene = sceneIdMap.get(item.linkName) || null;
					}
					
					if (currentScene) {
						// Check if order_num or content needs to be updated
						const needsOrderUpdate = currentScene.order_num !== sceneOrderNum;
						const needsContentUpdate = goal !== currentScene.goal || timeRef !== currentScene.time_ref;
						
						if (needsOrderUpdate || needsContentUpdate) {
							currentScene = await this.apiClient.updateScene(currentScene.id, {
								goal,
								time_ref: timeRef,
								order_num: sceneOrderNum,
							});
						}
						
						// Initialize beat order for this scene
						const existingBeats = await this.apiClient.getBeats(currentScene.id);
						if (existingBeats.length === 0) {
							beatOrderNums.set(currentScene.id, 1);
						} else {
							const maxOrderNum = Math.max(...existingBeats.map(b => b.order_num));
							beatOrderNums.set(currentScene.id, maxOrderNum + 1);
						}
					}
				} else {
					// Create new orphan scene
					currentScene = await this.apiClient.createScene({
						story_id: storyId,
						chapter_id: null, // Orphan scene
						order_num: sceneOrderNum,
						goal,
						time_ref: timeRef,
					});
					
					// Update maps
					const sceneFileName = this.fileManager.generateSceneFileName(currentScene);
					const sceneLinkName = sceneFileName.replace(/\.md$/, "");
					sceneMap.set(sceneLinkName, currentScene);
					sceneIdMap.set(currentScene.id, currentScene);
					
					// Initialize beat order for new scene (starts at 1)
					beatOrderNums.set(currentScene.id, 1);
				}
				
				sceneOrderNum++;
			} else if (item.type === "beat" && currentScene) {
				// Parse beat display text
				// Can be: "Beat N: intent -> outcome", "intent -> outcome", or just "intent"
				let intent: string;
				let outcome: string = "";

				const beatMatch = item.displayText.match(/Beat\s+\d+:\s*(.+)/);
				if (beatMatch) {
					// Format: "Beat N: intent -> outcome" or "Beat N: intent"
					const beatText = beatMatch[1].trim();
					const parts = beatText.split(/\s*->\s*/);
					intent = parts[0].trim();
					outcome = parts.length > 1 ? parts.slice(1).join(" -> ").trim() : "";
				} else {
					// Format: "intent -> outcome" or just "intent"
					const parts = item.displayText.split(/\s*->\s*/);
					intent = parts[0].trim();
					outcome = parts.length > 1 ? parts.slice(1).join(" -> ").trim() : "";
				}

				// Get current beat order_num for this scene
				const currentBeatOrderNum = beatOrderNums.get(currentScene.id) || 1;

				if (item.linkName) {
					// Beat exists - find it
					let currentBeat = beatMap.get(item.linkName) || null;
					if (!currentBeat) {
						// Try to find by ID if linkName is actually an ID
						currentBeat = beatIdMap.get(item.linkName) || null;
					}
					
					if (currentBeat) {
						// Check if order_num or content needs to be updated
						const needsOrderUpdate = currentBeat.order_num !== currentBeatOrderNum;
						const needsContentUpdate = intent !== currentBeat.intent || outcome !== currentBeat.outcome;
						const needsSceneUpdate = currentBeat.scene_id !== currentScene.id;
						
						if (needsOrderUpdate || needsContentUpdate || needsSceneUpdate) {
							currentBeat = await this.apiClient.updateBeat(currentBeat.id, {
								intent,
								outcome,
								order_num: currentBeatOrderNum,
								scene_id: currentScene.id,
							});
						}
					}
				} else {
					// Create new beat
					await this.apiClient.createBeat({
						scene_id: currentScene.id,
						order_num: currentBeatOrderNum,
						type: "setup", // Default type
						intent,
						outcome,
					});
				}
				
				beatOrderNums.set(currentScene.id, currentBeatOrderNum + 1);
			}
		}
	}

	// Process the "## Orphan Beats" list and update orphan beats order
	private async processOrphanBeatsList(
		list: { items: ParsedBeatListItem[] },
		storyId: string
	): Promise<void> {
		// Get all beats for this story
		const allBeats = await this.apiClient.getBeatsByStory(storyId);
		const allScenes = await this.apiClient.getScenesByStory(storyId);
		const sceneIdSet = new Set(allScenes.map(s => s.id));
		
		// Filter to only orphan beats (without scene_id or with invalid scene_id)
		const orphanBeats = allBeats.filter(b => !b.scene_id || !sceneIdSet.has(b.scene_id));
		
		const beatMap = new Map<string, Beat>(); // Map by link name
		const beatIdMap = new Map<string, Beat>(); // Map by ID
		
		for (const beat of orphanBeats) {
			const fileName = this.fileManager.generateBeatFileName(beat);
			const linkName = fileName.replace(/\.md$/, "");
			beatMap.set(linkName, beat);
			beatIdMap.set(beat.id, beat);
		}

		let beatOrderNum = 1; // Start from 1, will increment as we process

		// Process list items - order matters!
		for (const item of list.items) {
			// Parse beat display text
			// Can be: "Beat N: intent -> outcome", "intent -> outcome", or just "intent"
			let intent: string;
			let outcome: string = "";

			const beatMatch = item.displayText.match(/Beat\s+\d+:\s*(.+)/);
			if (beatMatch) {
				// Format: "Beat N: intent -> outcome" or "Beat N: intent"
				const beatText = beatMatch[1].trim();
				const parts = beatText.split(/\s*->\s*/);
				intent = parts[0].trim();
				outcome = parts.length > 1 ? parts.slice(1).join(" -> ").trim() : "";
			} else {
				// Format: "intent -> outcome" or just "intent"
				const parts = item.displayText.split(/\s*->\s*/);
				intent = parts[0].trim();
				outcome = parts.length > 1 ? parts.slice(1).join(" -> ").trim() : "";
			}

			if (item.linkName) {
				// Beat exists - find it
				let currentBeat = beatMap.get(item.linkName) || null;
				if (!currentBeat) {
					// Try to find by ID if linkName is actually an ID
					currentBeat = beatIdMap.get(item.linkName) || null;
				}
				
				if (currentBeat) {
					// Check if order_num or content needs to be updated
					const needsOrderUpdate = currentBeat.order_num !== beatOrderNum;
					const needsContentUpdate = intent !== currentBeat.intent || outcome !== currentBeat.outcome;
					
					if (needsOrderUpdate || needsContentUpdate) {
						currentBeat = await this.apiClient.updateBeat(currentBeat.id, {
							intent,
							outcome,
							order_num: beatOrderNum,
							// Keep scene_id as null or invalid (orphan) - don't update it
						});
					}
				}
			} else {
				// Create new orphan beat
				// The API requires scene_id, so we need to find or create an orphan scene for orphan beats
				// First, try to find an existing orphan scene specifically for orphan beats
				const allScenes = await this.apiClient.getScenesByStory(storyId);
				let orphanBeatScene = allScenes.find(s => !s.chapter_id && s.goal.startsWith("Orphan Beats Container"));
				
				if (!orphanBeatScene) {
					// Create a special orphan scene for orphan beats
					orphanBeatScene = await this.apiClient.createScene({
						story_id: storyId,
						chapter_id: null,
						order_num: 9999, // High number to keep it at the end
						goal: "Orphan Beats Container",
						time_ref: "",
					});
				}
				
				await this.apiClient.createBeat({
					scene_id: orphanBeatScene.id,
					order_num: beatOrderNum,
					type: "setup", // Default type
					intent,
					outcome,
				});
			}
			
			beatOrderNum++;
		}
	}

	// Process the "## Scenes & Beats" list and create/update/delete scenes and beats
	private async processSceneBeatList(
		list: { items: ParsedSceneBeatListItem[] },
		chapterId: string,
		storyId: string
	): Promise<void> {
		// Get existing scenes and beats for this chapter
		const existingScenes = await this.apiClient.getScenes(chapterId);
		const sceneMap = new Map<string, Scene>(); // Map by link name
		const sceneIdMap = new Map<string, Scene>(); // Map by ID
		for (const scene of existingScenes) {
			const fileName = this.fileManager.generateSceneFileName(scene);
			const linkName = fileName.replace(/\.md$/, "");
			sceneMap.set(linkName, scene);
			sceneIdMap.set(scene.id, scene);
		}

		// Get all beats for scenes in this chapter
		const beatMap = new Map<string, Beat>(); // Map by link name
		const beatIdMap = new Map<string, Beat>(); // Map by ID
		for (const scene of existingScenes) {
			const beats = await this.apiClient.getBeats(scene.id);
			for (const beat of beats) {
				const fileName = this.fileManager.generateBeatFileName(beat);
				const linkName = fileName.replace(/\.md$/, "");
				beatMap.set(linkName, beat);
				beatIdMap.set(beat.id, beat);
			}
		}

		let currentScene: Scene | null = null;
		let sceneOrderNum = 1; // Start from 1, will increment as we process
		const beatOrderNums = new Map<string, number>(); // Track order_num per scene

		// Process list items - order matters!
		for (const item of list.items) {
			if (item.type === "scene") {
				// Reset currentScene before processing
				currentScene = null;
				// Parse scene display text
				// Can be: "Scene N: goal - timeRef", "goal - timeRef", or just "goal"
				let goal: string;
				let timeRef: string = "";

				const sceneMatch = item.displayText.match(/Scene\s+\d+:\s*(.+)/);
				if (sceneMatch) {
					// Format: "Scene N: goal - timeRef" or "Scene N: goal"
					const sceneText = sceneMatch[1].trim();
					const parts = sceneText.split(/\s*-\s*/);
					goal = parts[0].trim();
					timeRef = parts.length > 1 ? parts.slice(1).join(" - ").trim() : "";
				} else {
					// Format: "goal - timeRef" or just "goal"
					const parts = item.displayText.split(/\s*-\s*/);
					goal = parts[0].trim();
					timeRef = parts.length > 1 ? parts.slice(1).join(" - ").trim() : "";
				}

				// Reset currentScene before processing
				currentScene = null;

				if (item.linkName) {
					// Scene exists - find it
					currentScene = sceneMap.get(item.linkName) || null;
					if (!currentScene) {
						// Try to find by ID if linkName is actually an ID
						currentScene = sceneIdMap.get(item.linkName) || null;
					}
					
					if (currentScene) {
						// Check if order_num needs to be updated
						const needsOrderUpdate = currentScene.order_num !== sceneOrderNum;
						const needsContentUpdate = goal !== currentScene.goal || timeRef !== currentScene.time_ref;
						
						if (needsOrderUpdate || needsContentUpdate) {
							currentScene = await this.apiClient.updateScene(currentScene.id, {
								goal,
								time_ref: timeRef,
								order_num: sceneOrderNum,
							});
						}
						
						// Initialize beat order for this scene (important for empty scenes)
						// Check if scene already has beats, if not, start from 1
						const existingBeats = await this.apiClient.getBeats(currentScene.id);
						if (existingBeats.length === 0) {
							beatOrderNums.set(currentScene.id, 1);
						} else {
							// Scene has beats, use max order_num + 1 for new beats
							const maxOrderNum = Math.max(...existingBeats.map(b => b.order_num));
							beatOrderNums.set(currentScene.id, maxOrderNum + 1);
						}
					}
				} else {
					// Create new scene
					currentScene = await this.apiClient.createScene({
						story_id: storyId,
						chapter_id: chapterId,
						order_num: sceneOrderNum,
						goal,
						time_ref: timeRef,
					});
					
					// Update maps
					const sceneFileName = this.fileManager.generateSceneFileName(currentScene);
					const sceneLinkName = sceneFileName.replace(/\.md$/, "");
					sceneMap.set(sceneLinkName, currentScene);
					sceneIdMap.set(currentScene.id, currentScene);
					
					// Initialize beat order for new scene (starts at 1)
					beatOrderNums.set(currentScene.id, 1);
				}
				
				sceneOrderNum++;
			} else if (item.type === "beat" && currentScene) {
				// Parse beat display text
				// Can be: "Beat N: intent -> outcome", "intent -> outcome", or just "intent"
				let intent: string;
				let outcome: string = "";

				const beatMatch = item.displayText.match(/Beat\s+\d+:\s*(.+)/);
				if (beatMatch) {
					// Format: "Beat N: intent -> outcome" or "Beat N: intent"
					const beatText = beatMatch[1].trim();
					const parts = beatText.split(/\s*->\s*/);
					intent = parts[0].trim();
					outcome = parts.length > 1 ? parts.slice(1).join(" -> ").trim() : "";
				} else {
					// Format: "intent -> outcome" or just "intent"
					const parts = item.displayText.split(/\s*->\s*/);
					intent = parts[0].trim();
					outcome = parts.length > 1 ? parts.slice(1).join(" -> ").trim() : "";
				}

				// Get current beat order_num for this scene
				const currentBeatOrderNum = beatOrderNums.get(currentScene.id) || 1;

				if (item.linkName) {
					// Beat exists - find it
					let currentBeat = beatMap.get(item.linkName) || null;
					if (!currentBeat) {
						// Try to find by ID if linkName is actually an ID
						currentBeat = beatIdMap.get(item.linkName) || null;
					}
					
					if (currentBeat) {
						// Check if order_num or content needs to be updated
						const needsOrderUpdate = currentBeat.order_num !== currentBeatOrderNum;
						const needsContentUpdate = intent !== currentBeat.intent || outcome !== currentBeat.outcome;
						const needsSceneUpdate = currentBeat.scene_id !== currentScene.id;
						
						if (needsOrderUpdate || needsContentUpdate || needsSceneUpdate) {
							currentBeat = await this.apiClient.updateBeat(currentBeat.id, {
								intent,
								outcome,
								order_num: currentBeatOrderNum,
								scene_id: currentScene.id,
							});
							
							// Update maps if scene changed
							if (needsSceneUpdate) {
								const beatFileName = this.fileManager.generateBeatFileName(currentBeat);
								const beatLinkName = beatFileName.replace(/\.md$/, "");
								beatMap.set(beatLinkName, currentBeat);
								beatIdMap.set(currentBeat.id, currentBeat);
							}
						}
					}
				} else {
					// Create new beat
					const newBeat = await this.apiClient.createBeat({
						scene_id: currentScene.id,
						order_num: currentBeatOrderNum,
						type: "setup", // Default type
						intent,
						outcome,
					});
					
					// Update maps
					const beatFileName = this.fileManager.generateBeatFileName(newBeat);
					const beatLinkName = beatFileName.replace(/\.md$/, "");
					beatMap.set(beatLinkName, newBeat);
					beatIdMap.set(newBeat.id, newBeat);
				}
				
				// Increment beat order for next beat in this scene
				beatOrderNums.set(currentScene.id, currentBeatOrderNum + 1);
			}
		}
	}

	// Find content block file by link name (searches recursively in all type subfolders)
	private async findContentBlockFileByLinkName(
		contentsRoot: string,
		linkName: string
	): Promise<string | null> {
		const root = this.fileManager.getVault().getAbstractFileByPath(contentsRoot);
		if (!(root instanceof TFolder)) return null;

		const target = `${linkName}.md`;

		// Recursive search in all subfolders
		const stack: TFolder[] = [root];
		while (stack.length) {
			const folder = stack.pop()!;
			for (const child of folder.children) {
				if (child instanceof TFolder) {
					stack.push(child);
				} else if (child instanceof TFile && child.name === target) {
					return child.path;
				}
			}
		}

		return null;
	}

	// Find prose block by content when file name doesn't match (searches recursively in all type subfolders)
	private async findContentBlockByContent(
		contentsFolderPath: string,
		content: string
	): Promise<ContentBlock | null> {
		try {
			const folder = this.fileManager.getVault().getAbstractFileByPath(contentsFolderPath);
			if (!(folder instanceof TFolder)) {
				return null;
			}

			const normalizedContent = content.trim();

			// Recursive search through all subfolders
			const stack: TFolder[] = [folder];
			while (stack.length) {
				const currentFolder = stack.pop()!;
				for (const child of currentFolder.children) {
					if (child instanceof TFolder) {
						stack.push(child);
					} else if (child instanceof TFile && child.extension === "md") {
						const contentBlock = await this.fileManager.readContentBlockFromFile(child.path);
						if (contentBlock && contentBlock.content.trim() === normalizedContent) {
							return contentBlock;
						}
					}
				}
			}
		} catch (err) {
			console.error("Error searching for prose block by content:", err);
		}

		return null;
	}

	// Find prose block by ID when we have remote ID but need local file (searches recursively in all type subfolders)
	private async findContentBlockById(
		contentsFolderPath: string,
		id: string
	): Promise<ContentBlock | null> {
		try {
			const folder = this.fileManager.getVault().getAbstractFileByPath(contentsFolderPath);
			if (!(folder instanceof TFolder)) {
				return null;
			}

			// Recursive search through all subfolders
			const stack: TFolder[] = [folder];
			while (stack.length) {
				const currentFolder = stack.pop()!;
				for (const child of currentFolder.children) {
					if (child instanceof TFolder) {
						stack.push(child);
					} else if (child instanceof TFile && child.extension === "md") {
						const contentBlock = await this.fileManager.readContentBlockFromFile(child.path);
						if (contentBlock && contentBlock.id === id) {
							return contentBlock;
						}
					}
				}
			}
		} catch (err) {
			console.error("Error searching for prose block by ID:", err);
		}

		return null;
	}

	// Get the correct file path for a content block (always uses type subfolders)
	private getContentBlockFilePath(storyFolderPath: string, contentBlock: ContentBlock): string {
		const fileName = this.fileManager.generateContentBlockFileName(contentBlock);
		const typeFolderPath = this.fileManager.getContentBlockFolderPath(storyFolderPath, contentBlock.type || "text");
		return `${typeFolderPath}/${fileName}`;
	}

	// Get content block file path from contents folder path (03-contents)
	private async getContentBlockFilePathFromContents(
		contentsFolderPath: string, 
		contentBlock: ContentBlock
	): Promise<string> {
		const typeFolder = this.fileManager.getContentTypeFolder(contentBlock.type || "text");
		const typeFolderPath = `${contentsFolderPath}/${typeFolder}`;
		// Ensure type subfolder exists
		await this.fileManager.ensureFolderExists(typeFolderPath);
		const fileName = this.fileManager.generateContentBlockFileName(contentBlock);
		return `${typeFolderPath}/${fileName}`;
	}

	// Resolve conflict using modal or auto-resolve based on mode
	private async resolveConflict(
		localContentBlock: ContentBlock,
		remoteContentBlock: ContentBlock
	): Promise<ConflictResolutionResult> {
		// In local mode, automatically resolve to "local wins"
		if (this.settings.mode === "local") {
			return { resolution: "local" };
		}

		// In remote mode, show modal for user to choose
		return new Promise((resolve) => {
			const modal = new ConflictModal(
				this.app,
				localContentBlock,
				remoteContentBlock,
				async (result) => {
					resolve(result);
				}
			);
			modal.open();
		});
	}

	// Update the Chapter section in chapter file (hierarchical structure)
	private async updateChapterProseSectionHierarchical(
		originalContent: string,
		updatedSections: string[],
		file: TFile,
		frontmatter: Record<string, string>
	): Promise<void> {
		// Extract frontmatter
		const frontmatterMatch = originalContent.match(/^---\n([\s\S]*?)\n---/);
		const frontmatterText = frontmatterMatch ? frontmatterMatch[0] : "";

		// Extract content after frontmatter
		const bodyStart = frontmatterMatch ? frontmatterMatch[0].length : 0;
		const bodyContent = originalContent.substring(bodyStart).trim();

		// Get chapter number and title from frontmatter
		const chapterNumber = frontmatter.number || "1";
		const chapterTitle = frontmatter.title || "Untitled";

		// Find Chapter section and replace it
		// Capture everything until the end since Scene and Beat are part of the chapter section
		// Accept with or without blank line after the chapter header
		// Use non-greedy match to stop at next ## Chapter if it exists
		const chapterHeaderPattern = `##\\s+Chapter\\s+${chapterNumber}:\\s+[^\\n]+`;
		const chapterSectionMatch = bodyContent.match(new RegExp(`([\\s\\S]*?${chapterHeaderPattern}\\s*\\n+)([\\s\\S]*?)(?=\\n##\\s+Chapter\\s+\\d+:|$)`, "i"));
		
		if (!chapterSectionMatch) {
			// No Chapter section found, add it
			const newChapterSection = `\n\n## Chapter ${chapterNumber}: ${chapterTitle}\n\n${updatedSections.join("\n\n")}\n\n`;
			const updatedContent = `${frontmatterText}\n${bodyContent}${newChapterSection}`;
			await this.fileManager.getVault().modify(file, updatedContent);
			return;
		}

		// Replace Chapter section content
		const beforeChapter = chapterSectionMatch[1];
		const afterChapter = bodyContent.substring(chapterSectionMatch.index! + chapterSectionMatch[0].length);
		const newChapterContent = updatedSections.join("\n\n");
		const updatedBody = `${beforeChapter}${newChapterContent}\n\n${afterChapter}`;
		const updatedContent = `${frontmatterText}\n${updatedBody}`;
		await this.fileManager.getVault().modify(file, updatedContent);
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

