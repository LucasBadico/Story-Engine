import type { OutlineEntry } from "../parsers/outlineParser";
import { OutlineParser } from "../parsers/outlineParser";

export class OutlineReconciler {
	private readonly outlineParser = new OutlineParser();

	reconcile(localContent: string | null, generatedContent: string): string {
		if (!localContent) {
			return generatedContent;
		}

		const localEntries = this.outlineParser.parse(localContent);
		const generatedEntries = this.outlineParser.parse(generatedContent);

		const localEntryMap = new Map(localEntries.map((entry) => [this.buildEntryKey(entry), entry.raw]));
		const generatedLines = generatedContent.split("\n");
		const entryQueueByRaw = this.buildEntryQueues(generatedEntries);

		const mergedLines: string[] = [];
		for (const line of generatedLines) {
			const queue = entryQueueByRaw.get(line);
			if (queue && queue.length > 0) {
				const entry = queue.shift()!;
				const key = this.buildEntryKey(entry);
				mergedLines.push(localEntryMap.get(key) ?? line);
			} else {
				mergedLines.push(line);
			}
		}

		const untracked = this.collectUntrackedSegments(localContent, new Set(localEntries.map((entry) => entry.raw)));
		if (untracked.length > 0) {
			mergedLines.push(
				"",
				"<!-- story-engine/untracked-start -->",
				"> [!warning] Texto preservado",
				"> Estas linhas não foram reconhecidas pelo sync. Revise e mova para a seção correta.",
				"",
				untracked.join("\n"),
				"",
				"<!-- story-engine/untracked-end -->",
				""
			);
		}

		return mergedLines.join("\n");
	}

	private buildEntryQueues(entries: OutlineEntry[]): Map<string, OutlineEntry[]> {
		const queues = new Map<string, OutlineEntry[]>();
		for (const entry of entries) {
			const bucket = queues.get(entry.raw) ?? [];
			bucket.push(entry);
			queues.set(entry.raw, bucket);
		}
		return queues;
	}

	private buildEntryKey(entry: OutlineEntry): string {
		if (entry.link) {
			return `${entry.type}:${entry.link}`;
		}
		if (entry.placeholderLabel) {
			return `${entry.type}:placeholder:${entry.placeholderLabel.toLowerCase()}:${entry.depth}`;
		}
		return `${entry.type}:${entry.title.toLowerCase()}:${entry.depth}`;
	}

	private collectUntrackedSegments(localContent: string, trackedLines: Set<string>): string[] {
		const untracked: string[] = [];
		let insideBlock = false;

		for (const line of localContent.split("\n")) {
			if (line.includes("<!-- story-engine/untracked-start -->")) {
				insideBlock = true;
				continue;
			}
			if (line.includes("<!-- story-engine/untracked-end -->")) {
				insideBlock = false;
				continue;
			}
			if (insideBlock) {
				continue;
			}

			if (!trackedLines.has(line) && line.trim().length > 0) {
				untracked.push(line);
			}
		}

		return untracked;
	}
}

