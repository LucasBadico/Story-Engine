import { ContentsParser, type FenceChange } from "../parsers/contentsParser";
import type { DiffOperation, DiffResult } from "./types";

const FENCE_PATTERN =
	/<!--(chapter|scene|beat|content)-(start|end):(\d{4}):([a-z0-9-]+):([a-zA-Z0-9-]+)-->/gi;

export class DiffEngine {
	private readonly contentsParser = new ContentsParser();

	diffContents(local: string, generated: string): DiffResult {
		const changes = this.contentsParser.detectChanges(local, generated);
		return {
			operations: changes.map((change) => this.toDiffOperation(change)),
			untrackedSegments: this.findUntrackedSegments(local),
		};
	}

	private toDiffOperation(change: FenceChange): DiffOperation {
		let kind: DiffOperation["kind"] = "updated";
		if (change.changeType === "created") kind = "created";
		else if (change.changeType === "deleted") kind = "deleted";
		else if (change.changeType === "moved") kind = "moved";
		else if (change.changeType === "reordered") kind = "reordered";

		return {
			kind,
			fenceId: change.id,
			fenceType: change.type,
			metadata: {
				oldOrder: change.oldOrder,
				newOrder: change.newOrder,
				oldParentId: change.oldParentId,
				newParentId: change.newParentId,
			},
		};
	}

	private findUntrackedSegments(content: string): string[] {
		const trimmedContent = this.stripFrontmatter(content);
		const ranges = this.getFenceRanges(trimmedContent);
		const segments: string[] = [];
		let cursor = 0;

		for (const range of ranges) {
			if (range.start > cursor) {
				const snippet = trimmedContent.slice(cursor, range.start).trim();
				if (snippet.length > 0) {
					segments.push(snippet);
				}
			}
			cursor = Math.max(cursor, range.end);
		}

		if (cursor < trimmedContent.length) {
			const snippet = trimmedContent.slice(cursor).trim();
			if (snippet.length > 0) {
				segments.push(snippet);
			}
		}

		return segments;
	}

	private getFenceRanges(content: string): Array<{ start: number; end: number }> {
		const stack: Array<{ id: string; start: number }> = [];
		const ranges: Array<{ start: number; end: number }> = [];

		for (const match of content.matchAll(FENCE_PATTERN)) {
			const action = match[2];
			const id = match[5];
			const index = match.index ?? 0;

			if (action === "start") {
				stack.push({ id, start: index });
			} else {
				const startEntryIndex = stack.findLastIndex((entry) => entry.id === id);
				if (startEntryIndex === -1) {
					continue;
				}
				const startEntry = stack[startEntryIndex];
				stack.splice(startEntryIndex, 1);
				const end = index + match[0].length;
				ranges.push({ start: startEntry.start, end });
			}
		}

		return ranges.sort((a, b) => a.start - b.start);
	}

	private stripFrontmatter(content: string): string {
		const match = content.match(/^---\n[\s\S]*?\n---\n?/);
		return match ? content.slice(match[0].length) : content;
	}
}

