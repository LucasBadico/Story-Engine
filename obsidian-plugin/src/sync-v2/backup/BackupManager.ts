import type { App, TFile } from "obsidian";
import type { SyncContext } from "../types/sync";

export interface BackupResult {
	success: boolean;
	backupPath: string;
	filesCopied: string[];
	errors: string[];
}

export class BackupManager {
	constructor(private readonly app: App) {}

	async createBackup(
		filePaths: string[],
		operationType: "pull" | "push"
	): Promise<BackupResult> {
		const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
		const backupFolderPath = `.story-engine/backups/${timestamp}-${operationType}`;
		const filesCopied: string[] = [];
		const errors: string[] = [];

		try {
			await this.ensureFolderExists(backupFolderPath);
		} catch (error) {
			return {
				success: false,
				backupPath: backupFolderPath,
				filesCopied,
				errors: [`Failed to create backup folder: ${error}`],
			};
		}

		for (const filePath of filePaths) {
			try {
				const abstract = this.app.vault.getAbstractFileByPath(filePath);
				if (!abstract || !("path" in abstract) || "children" in abstract) {
					continue;
				}
				const content = await this.app.vault.read(abstract as TFile);
				const backupFilePath = `${backupFolderPath}/${filePath}`;
				const backupFolder = backupFilePath.substring(0, backupFilePath.lastIndexOf("/"));
				await this.ensureFolderExists(backupFolder);
				await this.app.vault.create(backupFilePath, content);
				filesCopied.push(filePath);
			} catch (error) {
				errors.push(`Failed to backup ${filePath}: ${error}`);
			}
		}

		const manifest = {
			timestamp,
			operationType,
			filesCopied,
			errors,
		};

		try {
			await this.app.vault.create(
				`${backupFolderPath}/manifest.json`,
				JSON.stringify(manifest, null, 2)
			);
		} catch (error) {
			errors.push(`Failed to write manifest: ${error}`);
		}

		return {
			success: errors.length === 0,
			backupPath: backupFolderPath,
			filesCopied,
			errors,
		};
	}

	getStoryFilesForBackup(storyFolderPath: string): string[] {
		return [
			`${storyFolderPath}/story.md`,
			`${storyFolderPath}/story.outline.md`,
			`${storyFolderPath}/story.contents.md`,
			`${storyFolderPath}/story.relations.md`,
		];
	}

	getWorldFilesForBackup(worldFolderPath: string): string[] {
		return [
			`${worldFolderPath}/world.md`,
			`${worldFolderPath}/world.outline.md`,
			`${worldFolderPath}/world.relations.md`,
		];
	}

	private async ensureFolderExists(path: string): Promise<void> {
		const parts = path.split("/");
		let current = "";
		for (const part of parts) {
			current = current ? `${current}/${part}` : part;
			const existing = this.app.vault.getAbstractFileByPath(current);
			if (!existing) {
				await this.app.vault.createFolder(current);
			}
		}
	}
}
