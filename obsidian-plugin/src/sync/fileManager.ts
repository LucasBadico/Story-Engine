import { Vault, TFile, TFolder } from "obsidian";
import {
	Story,
	Chapter,
	Scene,
	Beat,
	ChapterWithContent,
	Frontmatter,
} from "../types";

export class FileManager {
	constructor(
		private vault: Vault,
		private basePath: string = "Stories"
	) {}

	getVault(): Vault {
		return this.vault;
	}

	// Parse frontmatter from markdown content
	parseFrontmatter(content: string): { frontmatter: Frontmatter; body: string } {
		const frontmatterRegex = /^---\s*\n([\s\S]*?)\n---\s*\n([\s\S]*)$/;
		const match = content.match(frontmatterRegex);

		if (!match) {
			return {
				frontmatter: {
					id: "",
					type: "story",
				},
				body: content,
			};
		}

		const frontmatterText = match[1];
		const body = match[2];

		// Parse YAML-like frontmatter
		const frontmatter: Frontmatter = {
			id: "",
			type: "story",
		};

		const lines = frontmatterText.split("\n");
		for (const line of lines) {
			const colonIndex = line.indexOf(":");
			if (colonIndex === -1) continue;

			const key = line.substring(0, colonIndex).trim();
			const value = line.substring(colonIndex + 1).trim().replace(/^["']|["']$/g, "");

			switch (key) {
				case "id":
					frontmatter.id = value;
					break;
				case "type":
					if (
						value === "story" ||
						value === "chapter" ||
						value === "scene" ||
						value === "beat"
					) {
						frontmatter.type = value;
					}
					break;
				case "story_id":
					frontmatter.story_id = value;
					break;
				case "chapter_id":
					frontmatter.chapter_id = value;
					break;
				case "scene_id":
					frontmatter.scene_id = value;
					break;
				case "number":
					// Store number in frontmatter for chapters
					(frontmatter as any).number = parseInt(value, 10);
					break;
				case "version":
					frontmatter.version = parseInt(value, 10);
					break;
				case "synced_at":
					frontmatter.synced_at = value;
					break;
			}
		}

		return { frontmatter, body };
	}

	// Serialize frontmatter and body to markdown
	serializeFrontmatter(frontmatter: Frontmatter, body: string): string {
		const lines: string[] = ["---"];

		lines.push(`id: ${frontmatter.id}`);
		lines.push(`type: ${frontmatter.type}`);

		if (frontmatter.story_id) {
			lines.push(`story_id: ${frontmatter.story_id}`);
		}
		if (frontmatter.chapter_id) {
			lines.push(`chapter_id: ${frontmatter.chapter_id}`);
		}
		if (frontmatter.scene_id) {
			lines.push(`scene_id: ${frontmatter.scene_id}`);
		}
		if ((frontmatter as any).number !== undefined) {
			lines.push(`number: ${(frontmatter as any).number}`);
		}
		if (frontmatter.version !== undefined) {
			lines.push(`version: ${frontmatter.version}`);
		}
		if (frontmatter.synced_at) {
			lines.push(`synced_at: ${frontmatter.synced_at}`);
		}

		lines.push("---");
		lines.push("");

		return lines.join("\n") + body;
	}

	// Sanitize filename for filesystem
	sanitizeFilename(name: string): string {
		// Remove or replace invalid characters
		return name
			.replace(/[<>:"/\\|?*]/g, "-")
			.replace(/\s+/g, " ")
			.trim();
	}

	// Get story folder path
	getStoryFolderPath(storyTitle: string): string {
		const sanitizedTitle = this.sanitizeFilename(storyTitle);
		return `${this.basePath}/${sanitizedTitle}`;
	}

	// Ensure folder exists
	async ensureFolderExists(path: string): Promise<void> {
		const folders = path.split("/");
		let currentPath = "";

		for (const folder of folders) {
			if (folder === "") continue;
			currentPath = currentPath ? `${currentPath}/${folder}` : folder;

			const folderExists = await this.vault.adapter.exists(currentPath);
			if (!folderExists) {
				await this.vault.createFolder(currentPath);
			}
		}
	}

	// Write story metadata file
	async writeStoryMetadata(
		story: Story,
		folderPath: string
	): Promise<void> {
		await this.ensureFolderExists(folderPath);

		const frontmatter: Frontmatter = {
			id: story.id,
			type: "story",
			version: story.version_number,
			synced_at: new Date().toISOString(),
		};

		const body = `# ${story.title}

**Status**: ${story.status}
**Version**: ${story.version_number}
**Created**: ${new Date(story.created_at).toLocaleString()}
**Updated**: ${new Date(story.updated_at).toLocaleString()}

## Synopsis
(User can add content here)
`;

		const content = this.serializeFrontmatter(frontmatter, body);
		const filePath = `${folderPath}/metadata.md`;

		const existingFile = this.vault.getAbstractFileByPath(filePath);
		if (existingFile instanceof TFile) {
			await this.vault.modify(existingFile, content);
		} else {
			await this.vault.create(filePath, content);
		}
	}

	// Read story metadata file
	async readStoryMetadata(folderPath: string): Promise<{
		frontmatter: Frontmatter;
		content: string;
	}> {
		const filePath = `${folderPath}/metadata.md`;
		const file = this.vault.getAbstractFileByPath(filePath);

		if (!(file instanceof TFile)) {
			throw new Error(`Metadata file not found: ${filePath}`);
		}

		const content = await this.vault.read(file);
		const parsed = this.parseFrontmatter(content);
		return { frontmatter: parsed.frontmatter, content: parsed.body };
	}

	// Write chapter file with scenes and beats
	async writeChapterFile(
		chapter: ChapterWithContent,
		filePath: string
	): Promise<void> {
		const frontmatter: Frontmatter & { number?: number } = {
			id: chapter.chapter.id,
			type: "chapter",
			story_id: chapter.chapter.story_id,
			number: chapter.chapter.number,
			synced_at: new Date().toISOString(),
		};

		let body = `# Chapter ${chapter.chapter.number}: ${chapter.chapter.title}\n\n`;

		if (chapter.scenes.length === 0) {
			body += "(No scenes yet)\n";
		} else {
			body += "## Scenes\n\n";

			for (const sceneWithBeats of chapter.scenes) {
				const scene = sceneWithBeats.scene;
				body += `### Scene ${scene.order_num}`;
				if (scene.goal) {
					body += `: ${scene.goal}`;
				}
				body += `\n`;
				body += `**ID**: ${scene.id}\n`;
				body += `**Order**: ${scene.order_num}\n`;
				if (scene.time_ref) {
					body += `**Time**: ${scene.time_ref}\n`;
				}
				if (scene.goal) {
					body += `**Goal**: ${scene.goal}\n`;
				}
				body += `\n`;

				if (sceneWithBeats.beats.length > 0) {
					body += `#### Beats\n\n`;

					for (const beat of sceneWithBeats.beats) {
						body += `**Beat ${beat.order_num}** (${beat.type})\n`;
						if (beat.intent) {
							body += `- **Intent**: ${beat.intent}\n`;
						}
						if (beat.outcome) {
							body += `- **Outcome**: ${beat.outcome}\n`;
						}
						body += `\n`;
					}
				}

				body += `---\n\n`;
			}
		}

		const content = this.serializeFrontmatter(frontmatter, body);

		const existingFile = this.vault.getAbstractFileByPath(filePath);
		if (existingFile instanceof TFile) {
			await this.vault.modify(existingFile, content);
		} else {
			await this.vault.create(filePath, content);
		}
	}

	// Read chapter file
	async readChapterFile(filePath: string): Promise<{
		frontmatter: Frontmatter;
		content: string;
	}> {
		const file = this.vault.getAbstractFileByPath(filePath);

		if (!(file instanceof TFile)) {
			throw new Error(`Chapter file not found: ${filePath}`);
		}

		const content = await this.vault.read(file);
		const parsed = this.parseFrontmatter(content);
		return { frontmatter: parsed.frontmatter, content: parsed.body };
	}

	// Update frontmatter in a file
	async updateFrontmatter(
		filePath: string,
		updates: Partial<Frontmatter>
	): Promise<void> {
		const file = this.vault.getAbstractFileByPath(filePath);

		if (!(file instanceof TFile)) {
			throw new Error(`File not found: ${filePath}`);
		}

		const content = await this.vault.read(file);
		const { frontmatter, body } = this.parseFrontmatter(content);

		const updatedFrontmatter: Frontmatter = {
			...frontmatter,
			...updates,
		};

		const newContent = this.serializeFrontmatter(updatedFrontmatter, body);
		await this.vault.modify(file, newContent);
	}
}

