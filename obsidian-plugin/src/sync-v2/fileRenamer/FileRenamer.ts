import type { SyncContext } from "../types/sync";

export interface RenameResult {
	oldPath: string;
	newPath: string;
	updatedReferences: number;
}

export interface ReferenceReplacement {
	pattern: RegExp;
	replacement: string;
}

export interface ReferenceUpdateRequest {
	filePath: string;
	replacements: ReferenceReplacement[];
}

export interface RenameRequest {
	oldPath: string;
	newPath: string;
	references?: ReferenceUpdateRequest[];
}

export class FileRenamer {
	constructor(private readonly context: SyncContext) {}

	async rename(request: RenameRequest): Promise<RenameResult> {
		await this.context.fileManager.renameFile(request.oldPath, request.newPath);
		const updatedReferences = await this.updateReferences(request.references ?? []);
		return {
			oldPath: request.oldPath,
			newPath: request.newPath,
			updatedReferences,
		};
	}

	private async updateReferences(updates: ReferenceUpdateRequest[]): Promise<number> {
		let updatedFiles = 0;
		for (const update of updates) {
			const original = await this.context.fileManager.readFile(update.filePath);
			let modified = original;
			for (const replacement of update.replacements) {
				modified = modified.replace(replacement.pattern, replacement.replacement);
			}
			if (modified !== original) {
				await this.context.fileManager.writeFile(update.filePath, modified);
				updatedFiles += 1;
			}
		}
		return updatedFiles;
	}
}

