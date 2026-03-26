export type OutlineEntryType = "chapter" | "scene" | "beat";

export type OutlineEntryStatus = "has_content" | "empty" | "placeholder";

export interface OutlineEntry {
	type: OutlineEntryType;
	title: string;
	link?: string;
	raw: string;
	status: OutlineEntryStatus;
	depth: number;
	order: number;
	placeholderLabel?: string;
}

const PLACEHOLDER_PATTERN = /^_(.+)_$/;
const LINK_PATTERN = /\[\[([^[\]|]+)(?:\|([^[\]]+))?\]\]/;

export class OutlineParser {
	parse(markdown: string): OutlineEntry[] {
		const entries: OutlineEntry[] = [];
		const counters: Record<OutlineEntryType, number> = {
			chapter: 0,
			scene: 0,
			beat: 0,
		};

		for (const rawLine of markdown.split("\n")) {
			const line = rawLine.replace(/\r$/, "");
			if (!line.trim() || line.trimStart().startsWith(">")) {
				continue;
			}

			const entry = this.parseLine(line, counters);
			if (entry) {
				entries.push(entry);
			}
		}

		return entries;
	}

	sanitizeTitle(title: string): string {
		return title
			.normalize("NFKD")
			.replace(/[^\w\s-]/g, "")
			.replace(/\s+/g, " ")
			.trim();
	}

	formatEntry(entry: OutlineEntry): string {
		const indent = "\t".repeat(entry.depth);
		const marker = entry.status === "has_content" ? "+" : "-";
		const display =
			entry.status === "placeholder" && entry.placeholderLabel
				? `_${entry.placeholderLabel}_`
				: entry.link
				? `[[${entry.link}|${entry.title}]]`
				: entry.title;
		return `${indent}- ${display} ${marker}`.trimEnd();
	}

	private parseLine(
		line: string,
		counters: Record<OutlineEntryType, number>
	): OutlineEntry | null {
		const match = line.match(/^([\t ]*)([-+])\s+(.+)$/);
		if (!match) {
			return null;
		}

		const [, indent, , restRaw] = match;
		const depth = this.computeDepth(indent);
		const type = this.depthToType(depth);
		if (!type) {
			return null;
		}

		let rest = restRaw.trim();
		let status: OutlineEntryStatus = "empty";

		const trailingStatus = rest.match(/([+-])\s*$/);
		if (trailingStatus) {
			status = trailingStatus[1] === "+" ? "has_content" : "empty";
			rest = rest.slice(0, trailingStatus.index).trim();
		}

		const placeholder = rest.match(PLACEHOLDER_PATTERN);
		let placeholderLabel: string | undefined;
		if (placeholder) {
			status = "placeholder";
			placeholderLabel = placeholder[1];
		}

		const linkMatch = rest.match(LINK_PATTERN);
		let link: string | undefined;
		let title = rest;

		if (linkMatch) {
			link = linkMatch[1].trim();
			title = (linkMatch[2] ?? linkMatch[1]).trim();
		} else if (placeholderLabel) {
			title = placeholderLabel.replace(/^_+|_+$/g, "").trim();
		}

		counters[type] += 1;

		return {
			type,
			title: this.sanitizeTitle(title),
			link,
			raw: line,
			status,
			depth,
			order: counters[type],
			placeholderLabel,
		};
	}

	private computeDepth(indent: string): number {
		const tabs = (indent.match(/\t/g) || []).length;
		const spaces = indent.replace(/\t/g, "").length;
		return tabs + Math.floor(spaces / 2);
	}

	private depthToType(depth: number): OutlineEntryType | null {
		if (depth <= 0) return "chapter";
		if (depth === 1) return "scene";
		if (depth >= 2) return "beat";
		return null;
	}
}

