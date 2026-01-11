import { DiffEngine } from "./DiffEngine";
import type { DiffResult } from "./types";
import type { SyncWarning } from "../types/sync";

export interface ReconcileResult {
	mergedContent: string;
	diff: DiffResult;
	warnings: SyncWarning[];
}

export class ContentsReconciler {
	constructor(private readonly diffEngine = new DiffEngine()) {}

	reconcile(localContent: string | null, generatedContent: string): ReconcileResult {
		if (!localContent) {
			return {
				mergedContent: generatedContent,
				diff: { operations: [], untrackedSegments: [] },
				warnings: [],
			};
		}

		const diff = this.diffEngine.diffContents(localContent, generatedContent);
		let mergedContent = generatedContent;
		const warnings: SyncWarning[] = [];

		if (diff.untrackedSegments.length > 0) {
			const untrackedBlock = [
				"",
				"<!-- story-engine/untracked-start -->",
				"> [!warning] Texto preservado",
				"> Estas linhas não foram reconhecidas pelo sync. Revise e mova para a seção correta.",
				"",
				diff.untrackedSegments.join("\n\n"),
				"",
				"<!-- story-engine/untracked-end -->",
				"",
			].join("\n");
			mergedContent = `${generatedContent.trimEnd()}\n${untrackedBlock}`;
			warnings.push({
				code: "contents_untracked_segments",
				message: `Preservamos ${diff.untrackedSegments.length} trechos não reconhecidos em story.contents.`,
				details: diff.untrackedSegments,
				severity: "warning",
			});
		}

		return {
			mergedContent,
			diff,
			warnings,
		};
	}
}

