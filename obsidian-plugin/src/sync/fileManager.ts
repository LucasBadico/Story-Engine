import { Notice, TFile, TFolder, Vault } from "obsidian";
import { Story, Chapter, ChapterWithContent, StoryMetadata, Scene, Beat, SceneWithBeats, ProseBlock, ProseBlockReference } from "../types";

// Structure to organize prose blocks by their associations
export interface ProseBlockOrganization {
	chapterOnly: ProseBlock[]; // Prose blocks only associated with chapter
	byScene: Map<string, { scene: Scene; proseBlocks: ProseBlock[] }>; // Prose blocks by scene ID
	byBeat: Map<string, { beat: Beat; proseBlocks: ProseBlock[] }>; // Prose blocks by beat ID
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
	async writeStoryMetadata(story: Story, folderPath: string): Promise<void> {
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

		const content = `${frontmatter}\n# ${story.title}\n\nVersion: ${story.version_number}\nStatus: ${story.status}\n`;
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
		proseBlocks?: ProseBlock[],
		proseBlockRefs?: ProseBlockReference[]
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
		const organization = this.organizeProseBlocks(
			proseBlocks || [],
			proseBlockRefs || [],
			scenes
		);

		// Generate hierarchical prose section
		content += `## Prose\n\n`;

		// Add prose blocks that are only associated with chapter
		for (const proseBlock of organization.chapterOnly) {
			const fileName = this.generateProseBlockFileName(proseBlock);
			const linkName = fileName.replace(/\.md$/, "");
			content += `[[${linkName}|${proseBlock.content}]]\n\n`;
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
			const sceneProseBlocks = organization.byScene.get(scene.id)?.proseBlocks || [];
			for (const proseBlock of sceneProseBlocks) {
				const fileName = this.generateProseBlockFileName(proseBlock);
				const linkName = fileName.replace(/\.md$/, "");
				content += `[[${linkName}|${proseBlock.content}]]\n\n`;
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
				const beatProseBlocks = organization.byBeat.get(beat.id)?.proseBlocks || [];
				for (const proseBlock of beatProseBlocks) {
					const fileName = this.generateProseBlockFileName(proseBlock);
					const linkName = fileName.replace(/\.md$/, "");
					content += `[[${linkName}|${proseBlock.content}]]\n\n`;
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
		proseBlocks?: ProseBlock[]
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

		// Add prose blocks section with content and links (if referenced)
		if (proseBlocks && proseBlocks.length > 0) {
			content += `## Prose Relacionado\n\n`;
			// Sort by order_num
			const sortedProseBlocks = [...proseBlocks].sort((a, b) => a.order_num - b.order_num);
			for (const proseBlock of sortedProseBlocks) {
				const fileName = this.generateProseBlockFileName(proseBlock);
				// Remove .md extension for link
				const linkName = fileName.replace(/\.md$/, "");
				// Include content as link text
				content += `[[${linkName}|${proseBlock.content}]]\n\n`;
			}
		}

		// Add beats if any
		if (beats.length > 0) {
			content += `## Beats\n\n`;
			for (const beat of beats) {
				content += `### Beat ${beat.order_num} - ${beat.type}\n\n`;
				if (beat.intent) {
					content += `**Intent:** ${beat.intent}\n\n`;
				}
				if (beat.outcome) {
					content += `**Outcome:** ${beat.outcome}\n\n`;
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

	// Write beat file
	async writeBeatFile(
		beat: Beat,
		filePath: string,
		storyName?: string
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

	// Generate filename for prose block based on date and content preview
	generateProseBlockFileName(proseBlock: ProseBlock): string {
		// Parse created_at date
		const date = new Date(proseBlock.created_at);
		const year = date.getFullYear();
		const month = String(date.getMonth() + 1).padStart(2, "0");
		const day = String(date.getDate()).padStart(2, "0");
		const hours = String(date.getHours()).padStart(2, "0");
		const minutes = String(date.getMinutes()).padStart(2, "0");
		
		// Format: 2024-01-15T14-30
		const dateStr = `${year}-${month}-${day}T${hours}-${minutes}`;

		// Get first 30 characters of content and sanitize
		const contentPreview = proseBlock.content
			.substring(0, 30)
			.trim()
			.replace(/[<>:"/\\|?*\n\r\t]/g, "-")
			.replace(/\s+/g, "-")
			.replace(/-+/g, "-")
			.replace(/^-|-$/g, "")
			.toLowerCase();

		// If content is empty, use a default
		const textPart = contentPreview || "prose-block";

		return `${dateStr}_${textPart}.md`;
	}

	// Organize prose blocks by their associations (chapter, scene, beat)
	private organizeProseBlocks(
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
	async writeProseBlockFile(
		proseBlock: ProseBlock,
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
	async readProseBlockFromFile(filePath: string): Promise<ProseBlock | null> {
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

