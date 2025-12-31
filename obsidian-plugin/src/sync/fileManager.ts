import { Notice, TFile, TFolder, Vault } from "obsidian";
import { Story, Chapter, ChapterWithContent, StoryMetadata } from "../types";

export class FileManager {
	constructor(
		private vault: Vault,
		private baseFolder: string
	) {}

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

		const frontmatter = [
			"---",
			`id: ${story.id}`,
			`title: "${story.title}"`,
			`status: ${story.status}`,
			`version: ${story.version_number}`,
			`root_story_id: ${story.root_story_id}`,
			story.previous_story_id
				? `previous_version_id: ${story.previous_story_id}`
				: `previous_version_id: null`,
			`created_at: ${story.created_at}`,
			`updated_at: ${story.updated_at}`,
			"---",
			"",
		].join("\n");

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
		filePath: string
	): Promise<void> {
		const { chapter, scenes } = chapterWithContent;

		const frontmatter = [
			"---",
			`id: ${chapter.id}`,
			`story_id: ${chapter.story_id}`,
			`number: ${chapter.number}`,
			`title: "${chapter.title}"`,
			`status: ${chapter.status}`,
			`created_at: ${chapter.created_at}`,
			`updated_at: ${chapter.updated_at}`,
			"---",
			"",
		].join("\n");

		let content = `${frontmatter}\n# ${chapter.title}\n\n`;

		// Add scenes
		for (const { scene, beats } of scenes) {
			content += `## Scene\n\n${scene.content}\n\n`;

			if (beats.length > 0) {
				content += `### Beats\n\n`;
				for (const beat of beats) {
					content += `- ${beat.content}\n`;
				}
				content += `\n`;
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
	private parseFrontmatter(content: string): Record<string, string> {
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
}

