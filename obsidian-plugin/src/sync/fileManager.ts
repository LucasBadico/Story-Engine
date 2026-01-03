import { Notice, TFile, TFolder, Vault } from "obsidian";
import { Story, Chapter, ChapterWithContent, StoryMetadata, Scene, Beat, SceneWithBeats, ContentBlock, ContentBlockReference } from "../types";

// Structure to organize content blocks by their associations
export interface ContentBlockOrganization {
	chapterOnly: ContentBlock[]; // Content blocks only associated with chapter
	byScene: Map<string, { scene: Scene; contentBlocks: ContentBlock[] }>; // Content blocks by scene ID
	byBeat: Map<string, { beat: Beat; contentBlocks: ContentBlock[] }>; // Content blocks by beat ID
}

export class FileManager {
	constructor(
		private vault: Vault,
		private baseFolder: string
	) {}

	// Expose vault for sync operations
	getVault(): Vault {
		return this.vault;
	}

	// Get the folder path for a specific story
	getStoryFolderPath(storyTitle: string): string {
		const sanitized = this.sanitizeFolderName(storyTitle);
		return `${this.baseFolder}/${sanitized}`;
	}

	// Sanitize folder/file names
	private sanitizeFolderName(name: string): string {
		return name
			.replace(/[<>:"/\\|?*]/g, "-")
			.replace(/\s+/g, " ")
			.trim();
	}

	// Generate frontmatter with Obsidian tags
	private generateFrontmatter(
		baseFields: Record<string, string | number | null>,
		extraFields?: Record<string, string | number | null>,
		options?: {
			entityType: "story" | "chapter" | "scene" | "beat" | "prose-block";
			storyName?: string;
			date?: string; // ISO date string (YYYY-MM-DD) or Date object
		}
	): string {
		const fields: Record<string, string | number | null> = { ...baseFields };

		// Add extra fields if provided
		if (extraFields) {
			Object.assign(fields, extraFields);
		}

		// Generate tags (without # prefix - Obsidian adds it automatically)
		const tags: string[] = [];
		if (options) {
			// Entity type tag
			tags.push(`story-engine/${options.entityType}`);

			// Story name tag (sanitized)
			if (options.storyName) {
				const sanitizedStoryName = this.sanitizeFolderName(options.storyName)
					.toLowerCase()
					.replace(/\s+/g, "-");
				tags.push(`story/${sanitizedStoryName}`);
			}

			// Date tag in format YYYY/MM/DD
			if (options.date) {
				const date = typeof options.date === "string" ? new Date(options.date) : options.date;
				if (!isNaN(date.getTime())) {
					const year = date.getFullYear();
					const month = String(date.getMonth() + 1).padStart(2, "0");
					const day = String(date.getDate()).padStart(2, "0");
					tags.push(`date/${year}/${month}/${day}`);
				}
			}
		}

		// Build frontmatter lines
		const lines: string[] = ["---"];

		// Add all fields
		for (const [key, value] of Object.entries(fields)) {
			if (value === null || value === undefined) {
				lines.push(`${key}: null`);
			} else if (typeof value === "string") {
				// Escape quotes and wrap in quotes if contains special chars
				const escaped = value.replace(/"/g, '\\"');
				if (value.includes(":") || value.includes("\n") || value.includes('"')) {
					lines.push(`${key}: "${escaped}"`);
				} else {
					lines.push(`${key}: ${escaped}`);
				}
			} else {
				lines.push(`${key}: ${value}`);
			}
		}

		// Add tags if any
		if (tags.length > 0) {
			lines.push(`tags:`);
			for (const tag of tags) {
				lines.push(`  - ${tag}`);
			}
		}

		lines.push("---", "");

		return lines.join("\n");
	}

	// Ensure folder exists
	async ensureFolderExists(path: string): Promise<void> {
		const folder = this.vault.getAbstractFileByPath(path);
		if (!folder) {
			await this.vault.createFolder(path);
		}
	}

	// Write story metadata (story.md)
	async writeStoryMetadata(
		story: Story, 
		folderPath: string, 
		chapters?: ChapterWithContent[], 
		orphanScenes?: SceneWithBeats[], 
		orphanBeats?: Beat[],
		chapterContentData?: Map<string, { contentBlocks: ContentBlock[], contentBlockRefs: ContentBlockReference[] }>
	): Promise<void> {
		await this.ensureFolderExists(folderPath);

		const baseFields = {
			id: story.id,
			title: story.title,
			status: story.status,
			version: story.version_number,
			root_story_id: story.root_story_id,
			previous_version_id: story.previous_story_id,
			created_at: story.created_at,
			updated_at: story.updated_at,
		};

		const frontmatter = this.generateFrontmatter(baseFields, undefined, {
			entityType: "story",
			storyName: story.title,
			date: story.created_at,
		});

		let content = `${frontmatter}\n# ${story.title}\n\nVersion: ${story.version_number}\nStatus: ${story.status}\n\n`;

		// Filter out temporary chapters used for prose storage
		const temporaryChapterTitles = ["Story Prose", "Scene-Level Prose", "Beat-Level Prose"];
		const filteredChapters = chapters?.filter(c => 
			!temporaryChapterTitles.includes(c.chapter.title) && c.chapter.number < 9000
		) || [];

		// Generate hierarchical list of chapters, scenes and beats if provided
		if (filteredChapters.length > 0) {
			content += `## Chapters, Scenes & Beats\n\n`;
			content += `> [!info] How to use this list\n`;
			content += `> - **Create new item**: Add a line with \`-\` followed by the text\n`;
			content += `>   - **Chapter**: No indentation (level 0)\n`;
			content += `>   - **Scene**: Use 1 tab indentation (inside a chapter)\n`;
			content += `>   - **Beat**: Use 2 tabs indentation (inside a scene)\n`;
			content += `> - **Reorder**: Move items up/down to change order\n`;
			content += `> - **Links**: Use \`[[link-name|text]]\` to create links to existing files\n`;
			content += `> - **Format**:\n`;
			content += `>   - **Chapter**:\n`;
			content += `>     - Complete: \`Chapter N: title\`\n`;
			content += `>     - Simplified: \`title\`\n`;
			content += `>   - **Scene**:\n`;
			content += `>     - Complete: \`Scene N: goal - timeRef\`\n`;
			content += `>     - Simplified:\n`;
			content += `>       - \`goal - timeRef\`\n`;
			content += `>       - \`goal\`\n`;
			content += `>   - **Beat**:\n`;
			content += `>     - Complete: \`Beat N: intent -> outcome\`\n`;
			content += `>     - Simplified:\n`;
			content += `>       - \`intent -> outcome\`\n`;
			content += `>       - \`intent\`\n`;
			content += `> - **Markers**: \`-\` indicates no associated prose, \`+\` indicates associated prose\n\n`;
			
			// Add chapters with their scenes and beats
			for (const chapterWithContent of filteredChapters) {
				const chapter = chapterWithContent.chapter;
				const chapterFileName = `Chapter-${chapter.number}.md`;
				const chapterLinkName = chapterFileName.replace(/\.md$/, "");
				
				// For now, we don't track prose at story level, so always use "-"
				// In the future, we could check if chapter has any prose blocks
				content += `- [[${chapterLinkName}|Chapter ${chapter.number}: ${chapter.title}]]\n`;
				
				// Add scenes for this chapter
				for (const { scene, beats } of chapterWithContent.scenes) {
					const sceneFileName = this.generateSceneFileName(scene);
					const sceneLinkName = sceneFileName.replace(/\.md$/, "");
					const sceneDisplayText = scene.time_ref 
						? `${scene.goal} - ${scene.time_ref}`
						: scene.goal;
					
					// For now, we don't track prose at story level, so always use "-"
					content += `\t- [[${sceneLinkName}|Scene ${scene.order_num}: ${sceneDisplayText}]]\n`;
					
					// Add beats for this scene
					for (const beat of beats) {
						const beatFileName = this.generateBeatFileName(beat);
						const beatLinkName = beatFileName.replace(/\.md$/, "");
						const beatDisplayText = beat.outcome
							? `${beat.intent} -> ${beat.outcome}`
							: beat.intent;
						
						// For now, we don't track prose at story level, so always use "-"
						content += `\t\t- [[${beatLinkName}|Beat ${beat.order_num}: ${beatDisplayText}]]\n`;
					}
				}
			}
			content += `\n`;
		}
		
		// Generate separate list for orphan scenes (scenes without chapter_id)
		if (orphanScenes && orphanScenes.length > 0) {
			content += `## Orphan Scenes\n\n`;
			content += `> [!info] Scenes without a chapter\n`;
			content += `> These scenes are not associated with any chapter. You can:\n`;
			content += `> - **Reorder**: Move items up/down to change order\n`;
			content += `> - **Links**: Use \`[[link-name|text]]\` to create links to existing files\n`;
			content += `> - **Format**:\n`;
			content += `>   - Complete: \`Scene N: goal - timeRef\`\n`;
			content += `>   - Simplified:\n`;
			content += `>     - \`goal - timeRef\`\n`;
			content += `>     - \`goal\`\n`;
			content += `> - **Markers**: \`-\` indicates no associated prose, \`+\` indicates associated prose\n\n`;
			
			for (const { scene, beats } of orphanScenes) {
				const sceneFileName = this.generateSceneFileName(scene);
				const sceneLinkName = sceneFileName.replace(/\.md$/, "");
				const sceneDisplayText = scene.time_ref 
					? `${scene.goal} - ${scene.time_ref}`
					: scene.goal;
				
				// For now, we don't track prose at story level, so always use "-"
				content += `- [[${sceneLinkName}|Scene ${scene.order_num}: ${sceneDisplayText}]]\n`;
				
				// Add beats for this scene
				for (const beat of beats) {
					const beatFileName = this.generateBeatFileName(beat);
					const beatLinkName = beatFileName.replace(/\.md$/, "");
					const beatDisplayText = beat.outcome
						? `${beat.intent} -> ${beat.outcome}`
						: beat.intent;
					
					// For now, we don't track prose at story level, so always use "-"
					content += `\t- [[${beatLinkName}|Beat ${beat.order_num}: ${beatDisplayText}]]\n`;
				}
			}
			content += `\n`;
		}
		
		// Generate separate list for orphan beats (beats without scene_id)
		if (orphanBeats && orphanBeats.length > 0) {
			content += `## Orphan Beats\n\n`;
			content += `> [!info] Beats without a scene\n`;
			content += `> These beats are not associated with any scene. You can:\n`;
			content += `> - **Reorder**: Move items up/down to change order\n`;
			content += `> - **Links**: Use \`[[link-name|text]]\` to create links to existing files\n`;
			content += `> - **Format**:\n`;
			content += `>   - Complete: \`Beat N: intent -> outcome\`\n`;
			content += `>   - Simplified:\n`;
			content += `>     - \`intent -> outcome\`\n`;
			content += `>     - \`intent\`\n`;
			content += `> - **Markers**: \`-\` indicates no associated prose, \`+\` indicates associated prose\n\n`;
			
			for (const beat of orphanBeats) {
				const beatFileName = this.generateBeatFileName(beat);
				const beatLinkName = beatFileName.replace(/\.md$/, "");
				const beatDisplayText = beat.outcome
					? `${beat.intent} -> ${beat.outcome}`
					: beat.intent;
				
				// For now, we don't track prose at story level, so always use "-"
				content += `- [[${beatLinkName}|Beat ${beat.order_num}: ${beatDisplayText}]]\n`;
			}
			content += `\n`;
		}

		// Generate hierarchical prose section for the story
		content += `# Story: ${story.title}\n\n`;

		// Add chapters with their scenes, beats, and prose blocks
		if (filteredChapters.length > 0) {
			for (const chapterWithContent of filteredChapters) {
				const chapter = chapterWithContent.chapter;
				const chapterFileName = `Chapter-${chapter.number}.md`;
				const chapterLinkName = chapterFileName.replace(/\.md$/, "");
				
				content += `## Chapter ${chapter.number}: [[${chapterLinkName}|${chapter.title}]]\n\n`;
				
				// Get prose data for this chapter
				const proseData = chapterProseData?.get(chapter.id);
				let organization: ReturnType<typeof this.organizeContentBlocks> | null = null;
				
				if (contentData) {
					organization = this.organizeContentBlocks(
						contentData.contentBlocks,
						contentData.contentBlockRefs,
						chapterWithContent.scenes
					);
					
					// Add chapter-only content blocks
					for (const contentBlock of organization.chapterOnly) {
						const fileName = this.generateContentBlockFileName(contentBlock);
						const linkName = fileName.replace(/\.md$/, "");
						content += `[[${linkName}|${contentBlock.content.substring(0, 50)}...]]\n\n`;
					}
				}
				
				// Add scenes for this chapter
				for (const { scene, beats } of chapterWithContent.scenes) {
					const sceneFileName = this.generateSceneFileName(scene);
					const sceneLinkName = sceneFileName.replace(/\.md$/, "");
					const sceneDisplayText = scene.time_ref 
						? `${scene.goal} - ${scene.time_ref}`
						: scene.goal;
					
					content += `### Scene: [[${sceneLinkName}|${sceneDisplayText}]]\n\n`;
					
					// Add scene content blocks
					if (organization) {
						const sceneContentBlocks = organization.byScene.get(scene.id)?.contentBlocks || [];
						for (const contentBlock of sceneContentBlocks) {
							const fileName = this.generateContentBlockFileName(contentBlock);
							const linkName = fileName.replace(/\.md$/, "");
							content += `[[${linkName}|${contentBlock.content.substring(0, 50)}...]]\n\n`;
						}
					}
					
					// Add beats for this scene
					for (const beat of beats) {
						const beatFileName = this.generateBeatFileName(beat);
						const beatLinkName = beatFileName.replace(/\.md$/, "");
						const beatDisplayText = beat.outcome
							? `${beat.intent} -> ${beat.outcome}`
							: beat.intent;
						
						content += `#### Beat: [[${beatLinkName}|${beatDisplayText}]]\n\n`;
						
						// Add beat content blocks
						if (organization) {
							const beatContentBlocks = organization.byBeat.get(beat.id)?.contentBlocks || [];
							for (const contentBlock of beatContentBlocks) {
								const fileName = this.generateContentBlockFileName(contentBlock);
								const linkName = fileName.replace(/\.md$/, "");
								content += `[[${linkName}|${contentBlock.content.substring(0, 50)}...]]\n\n`;
							}
						}
					}
				}
			}
		}

		// Add orphan scenes in the prose section
		if (orphanScenes && orphanScenes.length > 0) {
			for (const { scene, beats } of orphanScenes) {
				const sceneFileName = this.generateSceneFileName(scene);
				const sceneLinkName = sceneFileName.replace(/\.md$/, "");
				const sceneDisplayText = scene.time_ref 
					? `${scene.goal} - ${scene.time_ref}`
					: scene.goal;
				
				content += `### Scene: [[${sceneLinkName}|${sceneDisplayText}]]\n\n`;
				
				// Add beats for this scene
				for (const beat of beats) {
					const beatFileName = this.generateBeatFileName(beat);
					const beatLinkName = beatFileName.replace(/\.md$/, "");
					const beatDisplayText = beat.outcome
						? `${beat.intent} -> ${beat.outcome}`
						: beat.intent;
					
					content += `#### Beat: [[${beatLinkName}|${beatDisplayText}]]\n\n`;
				}
			}
		}

		const filePath = `${folderPath}/story.md`;

		const file = this.vault.getAbstractFileByPath(filePath);
		if (file instanceof TFile) {
			await this.vault.modify(file, content);
		} else {
			await this.vault.create(filePath, content);
		}
	}

	// Write chapter file
	async writeChapterFile(
		chapterWithContent: ChapterWithContent,
		filePath: string,
		storyName?: string,
		contentBlocks?: ContentBlock[],
		contentBlockRefs?: ContentBlockReference[],
		orphanScenes?: SceneWithBeats[] // Include orphan scenes for easy association
	): Promise<void> {
		const { chapter, scenes } = chapterWithContent;

		const baseFields = {
			id: chapter.id,
			story_id: chapter.story_id,
			number: chapter.number,
			title: chapter.title,
			status: chapter.status,
			created_at: chapter.created_at,
			updated_at: chapter.updated_at,
		};

		const frontmatter = this.generateFrontmatter(baseFields, undefined, {
			entityType: "chapter",
			storyName: storyName,
			date: chapter.created_at,
		});

		let content = `${frontmatter}\n# ${chapter.title}\n\n`;

		// Organize prose blocks by their associations
		const organization = this.organizeContentBlocks(
			contentBlocks || [],
			contentBlockRefs || [],
			scenes
		);

		// Generate list of scenes and beats with prose indicators
		content += `## Scenes & Beats\n\n`;
		content += `> [!info] How to use this list\n`;
		content += `> - **Create new item**: Add a line with \`-\` followed by the text\n`;
		content += `>   - **Scene**: No indentation (level 0)\n`;
		content += `>   - **Beat**: Use 1 tab indentation (inside a scene)\n`;
		content += `> - **Reorder**: Move items up/down to change order\n`;
		content += `> - **Links**: Use \`[[link-name|text]]\` to create links to existing files\n`;
		content += `> - **Format**:\n`;
		content += `>   - **Scene**:\n`;
		content += `>     - Complete: \`Scene N: goal - timeRef\`\n`;
		content += `>     - Simplified:\n`;
		content += `>       - \`goal - timeRef\`\n`;
		content += `>       - \`goal\`\n`;
		content += `>   - **Beat**:\n`;
		content += `>     - Complete: \`Beat N: intent -> outcome\`\n`;
		content += `>     - Simplified:\n`;
		content += `>       - \`intent -> outcome\`\n`;
		content += `>       - \`intent\`\n`;
		content += `> - **Markers**: \`-\` indicates no associated prose, \`+\` indicates associated prose\n`;
		content += `> - **Orphan scenes** (without chapter) are shown below for easy association\n\n`;
		
		// Add scenes for this chapter
		for (const { scene, beats } of scenes) {
			const sceneFileName = this.generateSceneFileName(scene);
			const sceneLinkName = sceneFileName.replace(/\.md$/, "");
			const sceneDisplayText = scene.time_ref 
				? `${scene.goal} - ${scene.time_ref}`
				: scene.goal;
			
			// Check if scene has prose blocks (not associated with beats)
			const sceneContentBlocks = organization.byScene.get(scene.id)?.contentBlocks || [];
			const hasSceneContent = sceneContentBlocks.length > 0;
			const sceneMarker = hasSceneProse ? "+" : "-";
			
			content += `${sceneMarker} [[${sceneLinkName}|Scene ${scene.order_num}: ${sceneDisplayText}]]\n`;
			
			// Add beats for this scene
			for (const beat of beats) {
				const beatFileName = this.generateBeatFileName(beat);
				const beatLinkName = beatFileName.replace(/\.md$/, "");
				const beatDisplayText = beat.outcome
					? `${beat.intent} -> ${beat.outcome}`
					: beat.intent;
				
				// Check if beat has prose blocks
				const beatContentBlocks = organization.byBeat.get(beat.id)?.contentBlocks || [];
				const hasBeatContent = beatContentBlocks.length > 0;
				const beatMarker = hasBeatProse ? "+" : "-";
				
				content += `\t${beatMarker} [[${beatLinkName}|Beat ${beat.order_num}: ${beatDisplayText}]]\n`;
			}
		}
		
		// Add orphan scenes (scenes without chapter_id) for easy association
		if (orphanScenes && orphanScenes.length > 0) {
			content += `\n`;
			content += `> [!info] Orphan Scenes (not yet associated with this chapter)\n`;
			content += `> You can associate these scenes with this chapter by moving them here.\n\n`;
			
			for (const { scene, beats } of orphanScenes) {
				const sceneFileName = this.generateSceneFileName(scene);
				const sceneLinkName = sceneFileName.replace(/\.md$/, "");
				const sceneDisplayText = scene.time_ref 
					? `${scene.goal} - ${scene.time_ref}`
					: scene.goal;
				
				// For now, we don't track prose at chapter level, so always use "-"
				content += `- [[${sceneLinkName}|Scene ${scene.order_num}: ${sceneDisplayText}]]\n`;
				
				// Add beats for this orphan scene
				for (const beat of beats) {
					const beatFileName = this.generateBeatFileName(beat);
					const beatLinkName = beatFileName.replace(/\.md$/, "");
					const beatDisplayText = beat.outcome
						? `${beat.intent} -> ${beat.outcome}`
						: beat.intent;
					
					// For now, we don't track prose at chapter level, so always use "-"
					content += `\t- [[${beatLinkName}|Beat ${beat.order_num}: ${beatDisplayText}]]\n`;
				}
			}
		}
		
		content += `\n`;

		// Generate hierarchical prose section with chapter header
		content += `## Chapter ${chapter.number}: ${chapter.title}\n\n`;

		// Add prose blocks that are only associated with chapter
		for (const contentBlock of organization.chapterOnly) {
			const fileName = this.generateContentBlockFileName(contentBlock);
			const linkName = fileName.replace(/\.md$/, "");
			content += `[[${linkName}|${contentBlock.content.substring(0, 50)}...]]\n\n`;
		}

		// Add scenes with their prose blocks and beats
		for (const { scene, beats } of scenes) {
			const sceneFileName = this.generateSceneFileName(scene);
			const sceneLinkName = sceneFileName.replace(/\.md$/, "");
			
			// Format scene header: ## Scene: [[link|goal - timeRef]] or ## Scene: [[link|goal]]
			const sceneDisplayText = scene.time_ref 
				? `${scene.goal} - ${scene.time_ref}`
				: scene.goal;
			content += `## Scene: [[${sceneLinkName}|${sceneDisplayText}]]\n\n`;

			// Get prose blocks for this scene (not associated with any beat)
			const sceneContentBlocks = organization.byScene.get(scene.id)?.contentBlocks || [];
			for (const contentBlock of sceneContentBlocks) {
				const fileName = this.generateContentBlockFileName(contentBlock);
				const linkName = fileName.replace(/\.md$/, "");
				content += `[[${linkName}|${contentBlock.content.substring(0, 50)}...]]\n\n`;
			}

			// Add beats with their prose blocks
			for (const beat of beats) {
				const beatFileName = this.generateBeatFileName(beat);
				const beatLinkName = beatFileName.replace(/\.md$/, "");
				
				// Format beat header: ### Beat: [[link|intent -> outcome]] or ### Beat: [[link|intent]]
				const beatDisplayText = beat.outcome
					? `${beat.intent} -> ${beat.outcome}`
					: beat.intent;
				content += `### Beat: [[${beatLinkName}|${beatDisplayText}]]\n\n`;

				// Get prose blocks for this beat
				const beatContentBlocks = organization.byBeat.get(beat.id)?.contentBlocks || [];
				for (const contentBlock of beatContentBlocks) {
					const fileName = this.generateContentBlockFileName(contentBlock);
					const linkName = fileName.replace(/\.md$/, "");
					content += `[[${linkName}|${contentBlock.content.substring(0, 50)}...]]\n\n`;
				}
			}
		}

		const file = this.vault.getAbstractFileByPath(filePath);
		if (file instanceof TFile) {
			await this.vault.modify(file, content);
		} else {
			await this.vault.create(filePath, content);
		}
	}

	// Read story metadata
	async readStoryMetadata(folderPath: string): Promise<StoryMetadata> {
		const filePath = `${folderPath}/story.md`;
		const file = this.vault.getAbstractFileByPath(filePath);

		if (!(file instanceof TFile)) {
			throw new Error(`Story metadata file not found: ${filePath}`);
		}

		const content = await this.vault.read(file);
		const frontmatter = this.parseFrontmatter(content);

		return {
			frontmatter: {
				id: frontmatter.id,
				title: frontmatter.title,
				status: frontmatter.status,
				version: parseInt(frontmatter.version),
				root_story_id: frontmatter.root_story_id,
				previous_version_id: frontmatter.previous_version_id || null,
				created_at: frontmatter.created_at,
				updated_at: frontmatter.updated_at,
			},
			content: content.split("---").slice(2).join("---").trim(),
		};
	}

	// Parse YAML frontmatter
	parseFrontmatter(content: string): Record<string, string> {
		const match = content.match(/^---\n([\s\S]*?)\n---/);
		if (!match) {
			return {};
		}

		const frontmatterText = match[1];
		const result: Record<string, string> = {};

		for (const line of frontmatterText.split("\n")) {
			const colonIndex = line.indexOf(":");
			if (colonIndex > 0) {
				const key = line.slice(0, colonIndex).trim();
				const value = line
					.slice(colonIndex + 1)
					.trim()
					.replace(/^["']|["']$/g, "");
				result[key] = value;
			}
		}

		return result;
	}

	// Copy story folder to versions folder
	async createVersionSnapshot(
		storyFolderPath: string,
		versionNumber: number
	): Promise<void> {
		const versionsPath = `${storyFolderPath}/versions`;
		await this.ensureFolderExists(versionsPath);

		const versionFolderPath = `${versionsPath}/v${versionNumber}`;

		// Check if version already exists
		const existingVersion = this.vault.getAbstractFileByPath(versionFolderPath);
		if (existingVersion) {
			console.log(`Version v${versionNumber} already exists, skipping snapshot`);
			return;
		}

		await this.ensureFolderExists(versionFolderPath);

		// Copy all files from story folder to version folder (except versions folder)
		const storyFolder = this.vault.getAbstractFileByPath(storyFolderPath);
		if (!(storyFolder instanceof TFolder)) {
			throw new Error(`Story folder not found: ${storyFolderPath}`);
		}

		await this.copyFolderContents(storyFolder, versionFolderPath, "versions");

		console.log(`Created version snapshot: v${versionNumber}`);
	}

	// Recursively copy folder contents
	private async copyFolderContents(
		sourceFolder: TFolder,
		destPath: string,
		excludeFolderName?: string
	): Promise<void> {
		for (const child of sourceFolder.children) {
			if (child instanceof TFile) {
				const relativePath = child.path.replace(sourceFolder.path + "/", "");
				const destFilePath = `${destPath}/${relativePath}`;
				const content = await this.vault.read(child);
				await this.vault.create(destFilePath, content);
			} else if (child instanceof TFolder) {
				// Skip excluded folder (e.g., "versions")
				if (excludeFolderName && child.name === excludeFolderName) {
					continue;
				}
				const relativePath = child.path.replace(sourceFolder.path + "/", "");
				const destFolderPath = `${destPath}/${relativePath}`;
				await this.ensureFolderExists(destFolderPath);
				await this.copyFolderContents(child, destFolderPath, excludeFolderName);
			}
		}
	}

	// Write scene file
	async writeSceneFile(
		sceneWithBeats: SceneWithBeats,
		filePath: string,
		storyName?: string,
		contentBlocks?: ContentBlock[],
		orphanBeats?: Beat[] // Include orphan beats for easy association
	): Promise<void> {
		const { scene, beats } = sceneWithBeats;

		const baseFields: Record<string, string | number | null> = {
			id: scene.id,
			story_id: scene.story_id,
			chapter_id: scene.chapter_id ?? null,
			order_num: scene.order_num,
			time_ref: scene.time_ref || "",
			goal: scene.goal || "",
			created_at: scene.created_at,
			updated_at: scene.updated_at,
		};

		// Add optional fields
		const extraFields: Record<string, string | number | null> = {};
		if (scene.pov_character_id) {
			extraFields.pov_character_id = scene.pov_character_id;
		}
		if (scene.location_id) {
			extraFields.location_id = scene.location_id;
		}

		const frontmatter = this.generateFrontmatter(baseFields, extraFields, {
			entityType: "scene",
			storyName: storyName,
			date: scene.created_at,
		});

		let content = `${frontmatter}\n# Scene ${scene.order_num}\n\n`;
		
		if (scene.goal) {
			content += `**Goal:** ${scene.goal}\n\n`;
		}
		if (scene.time_ref) {
			content += `**Time:** ${scene.time_ref}\n\n`;
		}

		// Generate list of beats with prose indicators
		if (beats.length > 0) {
			content += `## Beats\n\n`;
			content += `> [!info] How to use this list\n`;
			content += `> - **Create new item**: Add a line with \`-\` followed by the text\n`;
			content += `>   - **Beat**: No indentation (level 0)\n`;
			content += `> - **Reorder**: Move items up/down to change order\n`;
			content += `> - **Links**: Use \`[[link-name|text]]\` to create links to existing files\n`;
			content += `> - **Format**:\n`;
			content += `>   - **Beat**:\n`;
			content += `>     - Complete: \`Beat N: intent -> outcome\`\n`;
			content += `>     - Simplified:\n`;
			content += `>       - \`intent -> outcome\`\n`;
			content += `>       - \`intent\`\n`;
			content += `> - **Markers**: \`-\` indicates no associated prose, \`+\` indicates associated prose\n`;
			content += `> - **Orphan beats** (without scene) are shown below for easy association\n\n`;
			
			// Check which beats have prose blocks
			const beatsWithProse = new Set<string>();
			if (contentBlocks) {
				for (const contentBlock of contentBlocks) {
					// Check if prose block is referenced by any beat
					// We'll need to check prose block references, but for now we'll use a simple check
					// This will be updated when we process prose block references
				}
			}
			
			for (const beat of beats.sort((a, b) => a.order_num - b.order_num)) {
				const beatFileName = this.generateBeatFileName(beat);
				const beatLinkName = beatFileName.replace(/\.md$/, "");
				const beatDisplayText = beat.outcome
					? `${beat.intent} -> ${beat.outcome}`
					: beat.intent;
				
				// For now, we don't track prose at scene level, so always use "-"
				// In the future, we could check if beat has any prose blocks
				const hasBeatProse = beatsWithProse.has(beat.id);
				const beatMarker = hasBeatProse ? "+" : "-";
				
				content += `${beatMarker} [[${beatLinkName}|Beat ${beat.order_num}: ${beatDisplayText}]]\n`;
			}
			
			// Add orphan beats (beats without scene_id) for easy association
			if (orphanBeats && orphanBeats.length > 0) {
				content += `\n`;
				content += `> [!info] Orphan Beats (not yet associated with this scene)\n`;
				content += `> You can associate these beats with this scene by moving them here.\n\n`;
				
				for (const beat of orphanBeats) {
					const beatFileName = this.generateBeatFileName(beat);
					const beatLinkName = beatFileName.replace(/\.md$/, "");
					const beatDisplayText = beat.outcome
						? `${beat.intent} -> ${beat.outcome}`
						: beat.intent;
					
					// For now, we don't track prose at scene level, so always use "-"
					content += `- [[${beatLinkName}|Beat ${beat.order_num}: ${beatDisplayText}]]\n`;
				}
			}
			
			content += `\n`;
		}

		// Generate hierarchical prose section for the scene
		const sceneFileName = this.generateSceneFileName(scene);
		const sceneLinkName = sceneFileName.replace(/\.md$/, "");
		const sceneDisplayText = scene.time_ref 
			? `${scene.goal} - ${scene.time_ref}`
			: scene.goal;
		
		content += `## Scene: [[${sceneLinkName}|${sceneDisplayText}]]\n\n`;

		// Add prose blocks at scene level (not associated with any beat)
		if (contentBlocks && contentBlocks.length > 0) {
			// Sort by order_num
			const sortedContentBlocks = [...contentBlocks].sort((a, b) => (a.order_num || 0) - (b.order_num || 0));
			for (const contentBlock of sortedContentBlocks) {
				const fileName = this.generateContentBlockFileName(contentBlock);
				const linkName = fileName.replace(/\.md$/, "");
				content += `[[${linkName}|${contentBlock.content.substring(0, 50)}...]]\n\n`;
			}
		}

		// Add beats with their hierarchical structure
		for (const beat of beats.sort((a, b) => a.order_num - b.order_num)) {
			const beatFileName = this.generateBeatFileName(beat);
			const beatLinkName = beatFileName.replace(/\.md$/, "");
			const beatDisplayText = beat.outcome
				? `${beat.intent} -> ${beat.outcome}`
				: beat.intent;
			
			content += `### Beat: [[${beatLinkName}|${beatDisplayText}]]\n\n`;
		}

		const file = this.vault.getAbstractFileByPath(filePath);
		if (file instanceof TFile) {
			await this.vault.modify(file, content);
		} else {
			await this.vault.create(filePath, content);
		}
	}

	// Write beat file
	async writeBeatFile(
		beat: Beat,
		filePath: string,
		storyName?: string,
		contentBlocks?: ContentBlock[]
	): Promise<void> {
		const baseFields = {
			id: beat.id,
			scene_id: beat.scene_id,
			order_num: beat.order_num,
			type: beat.type,
			intent: beat.intent || "",
			outcome: beat.outcome || "",
			created_at: beat.created_at,
			updated_at: beat.updated_at,
		};

		const frontmatter = this.generateFrontmatter(baseFields, undefined, {
			entityType: "beat",
			storyName: storyName,
			date: beat.created_at,
		});

		let content = `${frontmatter}\n# Beat ${beat.order_num} - ${beat.type}\n\n`;
		
		if (beat.intent) {
			content += `**Intent:** ${beat.intent}\n\n`;
		}
		if (beat.outcome) {
			content += `**Outcome:** ${beat.outcome}\n\n`;
		}

		// Generate hierarchical prose section for the beat
		const beatFileName = this.generateBeatFileName(beat);
		const beatLinkName = beatFileName.replace(/\.md$/, "");
		const beatDisplayText = beat.outcome
			? `${beat.intent} -> ${beat.outcome}`
			: beat.intent;
		
		content += `## Beat: [[${beatLinkName}|${beatDisplayText}]]\n\n`;

		// Add content blocks for this beat
		if (contentBlocks && contentBlocks.length > 0) {
			const sortedContentBlocks = [...contentBlocks].sort((a, b) => (a.order_num || 0) - (b.order_num || 0));
			for (const contentBlock of sortedContentBlocks) {
				const fileName = this.generateContentBlockFileName(contentBlock);
				const linkName = fileName.replace(/\.md$/, "");
				content += `[[${linkName}|${contentBlock.content.substring(0, 50)}...]]\n\n`;
			}
		}

		const file = this.vault.getAbstractFileByPath(filePath);
		if (file instanceof TFile) {
			await this.vault.modify(file, content);
		} else {
			await this.vault.create(filePath, content);
		}
	}

	// List all chapter files in a story folder
	async listChapterFiles(storyFolderPath: string): Promise<string[]> {
		const chaptersPath = `${storyFolderPath}/chapters`;
		const folder = this.vault.getAbstractFileByPath(chaptersPath);

		if (!(folder instanceof TFolder)) {
			return [];
		}

		const chapterFiles: string[] = [];
		for (const child of folder.children) {
			if (child instanceof TFile && child.extension === "md") {
				chapterFiles.push(child.path);
			}
		}

		return chapterFiles.sort();
	}

	// List all scene files in a chapter folder
	async listSceneFiles(chapterFolderPath: string): Promise<string[]> {
		const scenesPath = `${chapterFolderPath}/scenes`;
		const folder = this.vault.getAbstractFileByPath(scenesPath);

		if (!(folder instanceof TFolder)) {
			return [];
		}

		const sceneFiles: string[] = [];
		for (const child of folder.children) {
			if (child instanceof TFile && child.extension === "md") {
				sceneFiles.push(child.path);
			}
		}

		return sceneFiles.sort();
	}

	// List all scene files in a story folder
	async listStorySceneFiles(storyFolderPath: string): Promise<string[]> {
		const scenesPath = `${storyFolderPath}/scenes`;
		const folder = this.vault.getAbstractFileByPath(scenesPath);

		if (!(folder instanceof TFolder)) {
			return [];
		}

		const sceneFiles: string[] = [];
		for (const child of folder.children) {
			if (child instanceof TFile && child.extension === "md") {
				sceneFiles.push(child.path);
			}
		}

		return sceneFiles.sort();
	}

	// List all beat files in a story folder
	async listStoryBeatFiles(storyFolderPath: string): Promise<string[]> {
		const beatsPath = `${storyFolderPath}/beats`;
		const folder = this.vault.getAbstractFileByPath(beatsPath);

		if (!(folder instanceof TFolder)) {
			return [];
		}

		const beatFiles: string[] = [];
		for (const child of folder.children) {
			if (child instanceof TFile && child.extension === "md") {
				beatFiles.push(child.path);
			}
		}

		return beatFiles.sort();
	}

	// List all beat files in a scene folder
	async listBeatFiles(sceneFolderPath: string): Promise<string[]> {
		const beatsPath = `${sceneFolderPath}/beats`;
		const folder = this.vault.getAbstractFileByPath(beatsPath);

		if (!(folder instanceof TFolder)) {
			return [];
		}

		const beatFiles: string[] = [];
		for (const child of folder.children) {
			if (child instanceof TFile && child.extension === "md") {
				beatFiles.push(child.path);
			}
		}

		return beatFiles.sort();
	}

	// Generate filename for content block based on date and content preview
	generateContentBlockFileName(contentBlock: ContentBlock): string {
		// Parse created_at date
		const date = new Date(contentBlock.created_at);
		const year = date.getFullYear();
		const month = String(date.getMonth() + 1).padStart(2, "0");
		const day = String(date.getDate()).padStart(2, "0");
		const hours = String(date.getHours()).padStart(2, "0");
		const minutes = String(date.getMinutes()).padStart(2, "0");
		
		// Format: 2024-01-15T14-30
		const dateStr = `${year}-${month}-${day}T${hours}-${minutes}`;

		// Get first 30 characters of content and sanitize
		const contentPreview = contentBlock.content
			.substring(0, 30)
			.trim()
			.replace(/[<>:"/\\|?*\n\r\t]/g, "-")
			.replace(/\s+/g, "-")
			.replace(/-+/g, "-")
			.replace(/^-|-$/g, "")
			.toLowerCase();

		// If content is empty, use a default
		const textPart = contentPreview || "content-block";

		return `${dateStr}_${textPart}.md`;
	}

	// Organize prose blocks by their associations (chapter, scene, beat)
	private organizeContentBlocks(
		proseBlocks: ProseBlock[],
		proseBlockRefs: ProseBlockReference[],
		scenes: SceneWithBeats[]
	): ProseBlockOrganization {
		const organization: ProseBlockOrganization = {
			chapterOnly: [],
			byScene: new Map(),
			byBeat: new Map(),
		};

		// Create maps for quick lookup
		const proseBlockRefsByProseBlock = new Map<string, ProseBlockReference[]>();
		for (const ref of proseBlockRefs) {
			if (!proseBlockRefsByProseBlock.has(ref.prose_block_id)) {
				proseBlockRefsByProseBlock.set(ref.prose_block_id, []);
			}
			proseBlockRefsByProseBlock.get(ref.prose_block_id)!.push(ref);
		}

		// Create scene and beat maps for quick lookup
		const sceneMap = new Map<string, Scene>();
		const beatMap = new Map<string, Beat>();
		for (const { scene, beats } of scenes) {
			sceneMap.set(scene.id, scene);
			for (const beat of beats) {
				beatMap.set(beat.id, beat);
			}
		}

		// Sort prose blocks by order_num
		const sortedProseBlocks = [...proseBlocks].sort((a, b) => a.order_num - b.order_num);

		for (const proseBlock of sortedProseBlocks) {
			const refs = proseBlockRefsByProseBlock.get(proseBlock.id) || [];
			
			// Find scene and beat references
			const sceneRef = refs.find(r => r.entity_type === "scene");
			const beatRef = refs.find(r => r.entity_type === "beat");

			if (beatRef && beatMap.has(beatRef.entity_id)) {
				// Associated with a beat (and implicitly with its scene)
				const beat = beatMap.get(beatRef.entity_id)!;
				if (!organization.byBeat.has(beat.id)) {
					organization.byBeat.set(beat.id, { beat, proseBlocks: [] });
				}
				organization.byBeat.get(beat.id)!.proseBlocks.push(proseBlock);
			} else if (sceneRef && sceneMap.has(sceneRef.entity_id)) {
				// Associated with a scene but not a beat
				const scene = sceneMap.get(sceneRef.entity_id)!;
				if (!organization.byScene.has(scene.id)) {
					organization.byScene.set(scene.id, { scene, proseBlocks: [] });
				}
				organization.byScene.get(scene.id)!.proseBlocks.push(proseBlock);
			} else {
				// Only associated with chapter
				organization.chapterOnly.push(proseBlock);
			}
		}

		return organization;
	}

	// Generate filename for scene based on date and goal
	generateSceneFileName(scene: Scene): string {
		// Parse created_at date
		const date = new Date(scene.created_at);
		const year = date.getFullYear();
		const month = String(date.getMonth() + 1).padStart(2, "0");
		const day = String(date.getDate()).padStart(2, "0");
		const hours = String(date.getHours()).padStart(2, "0");
		const minutes = String(date.getMinutes()).padStart(2, "0");
		
		// Format: 2024-01-15T14-30
		const dateStr = `${year}-${month}-${day}T${hours}-${minutes}`;

		// Sanitize goal for filename
		const goalSanitized = (scene.goal || "scene")
			.trim()
			.replace(/[<>:"/\\|?*\n\r\t]/g, "-")
			.replace(/\s+/g, "-")
			.replace(/-+/g, "-")
			.replace(/^-|-$/g, "")
			.toLowerCase();

		return `${dateStr}_${goalSanitized}.md`;
	}

	// Generate filename for beat based on date and intent
	generateBeatFileName(beat: Beat): string {
		// Parse created_at date
		const date = new Date(beat.created_at);
		const year = date.getFullYear();
		const month = String(date.getMonth() + 1).padStart(2, "0");
		const day = String(date.getDate()).padStart(2, "0");
		const hours = String(date.getHours()).padStart(2, "0");
		const minutes = String(date.getMinutes()).padStart(2, "0");
		
		// Format: 2024-01-15T14-30
		const dateStr = `${year}-${month}-${day}T${hours}-${minutes}`;

		// Sanitize intent for filename
		const intentSanitized = (beat.intent || "beat")
			.trim()
			.replace(/[<>:"/\\|?*\n\r\t]/g, "-")
			.replace(/\s+/g, "-")
			.replace(/-+/g, "-")
			.replace(/^-|-$/g, "")
			.toLowerCase();

		return `${dateStr}_${intentSanitized}.md`;
	}

	// Write prose block file
	async writeContentBlockFile(
		contentBlock: ContentBlock,
		filePath: string,
		storyName?: string
	): Promise<void> {
		const baseFields = {
			id: proseBlock.id,
			chapter_id: proseBlock.chapter_id,
			order_num: proseBlock.order_num,
			kind: proseBlock.kind,
			word_count: proseBlock.word_count,
			created_at: proseBlock.created_at,
			updated_at: proseBlock.updated_at,
		};

		const frontmatter = this.generateFrontmatter(baseFields, undefined, {
			entityType: "prose-block",
			storyName: storyName,
			date: proseBlock.created_at,
		});

		const content = `${frontmatter}${proseBlock.content}`;

		const file = this.vault.getAbstractFileByPath(filePath);
		if (file instanceof TFile) {
			await this.vault.modify(file, content);
		} else {
			await this.vault.create(filePath, content);
		}
	}

	// Read prose block from file
	async readContentBlockFromFile(filePath: string): Promise<ContentBlock | null> {
		const file = this.vault.getAbstractFileByPath(filePath);
		if (!(file instanceof TFile)) {
			return null;
		}

		try {
			const content = await this.vault.read(file);
			const frontmatter = this.parseFrontmatter(content);

			// Extract content after frontmatter
			const contentMatch = content.match(/^---\n[\s\S]*?\n---\n([\s\S]*)$/);
			const proseContent = contentMatch ? contentMatch[1].trim() : "";

			if (!frontmatter.id) {
				return null;
			}

			return {
				id: frontmatter.id,
				chapter_id: frontmatter.chapter_id || "",
				order_num: parseInt(frontmatter.order_num || "0", 10),
				kind: frontmatter.kind || "final",
				content: proseContent,
				word_count: parseInt(frontmatter.word_count || "0", 10),
				created_at: frontmatter.created_at || "",
				updated_at: frontmatter.updated_at || "",
			};
		} catch (err) {
			console.error(`Failed to read prose block from ${filePath}:`, err);
			return null;
		}
	}
}

