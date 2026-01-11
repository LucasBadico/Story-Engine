export interface RelationEntry {
	link?: string;
	displayText: string;
	description?: string;
	placeholder: boolean;
	raw: string;
}

export interface ParsedRelationsSection {
	name: string;
	entries: RelationEntry[];
}

export interface RelationsFile {
	frontmatter: Record<string, string>;
	sections: ParsedRelationsSection[];
}

const FRONTMATTER_REGEX = /^---\n([\s\S]*?)\n---/;
const HEADING_REGEX = /^##\s+(.+)$/;
const ENTRY_REGEX = /^-\s+(.*)$/;
const LINK_REGEX = /\[\[([^[\]|]+)(?:\|([^[\]]+))?\]\]/;

export class RelationsParser {
	parse(content: string): RelationsFile {
		const frontmatter = this.parseFrontmatter(content);
		const body = content.replace(FRONTMATTER_REGEX, "").trim();
		const sections: ParsedRelationsSection[] = [];

		let currentSection: ParsedRelationsSection | null = null;

		for (const rawLine of body.split("\n")) {
			const line = rawLine.trimEnd();
			if (!line) continue;

			const headingMatch = line.match(HEADING_REGEX);
			if (headingMatch) {
				currentSection = {
					name: headingMatch[1].trim(),
					entries: [],
				};
				sections.push(currentSection);
				continue;
			}

			if (!currentSection) {
				continue;
			}

			const entryMatch = line.match(ENTRY_REGEX);
			if (!entryMatch) {
				continue;
			}

			const entry = entryMatch[1].trim();
			const placeholder = entry.startsWith("_") && entry.endsWith("_");
			const linkMatch = entry.match(LINK_REGEX);
			let link: string | undefined;
			let displayText = entry;
			let description: string | undefined;

			if (linkMatch) {
				link = linkMatch[1].trim();
				displayText = (linkMatch[2] ?? linkMatch[1]).trim();
				const remainder = entry.slice(linkMatch[0].length).trim();
				if (remainder.startsWith("-")) {
					description = remainder.slice(1).trim();
				}
			} else if (entry.includes("-")) {
				const [name, desc] = entry.split("-").map((part) => part.trim());
				displayText = name;
				description = desc;
			}

			currentSection.entries.push({
				link,
				displayText,
				description,
				placeholder,
				raw: line,
			});
		}

		return { frontmatter, sections };
	}

	formatEntry(entry: RelationEntry): string {
		const prefix = entry.placeholder ? "_" : "- ";
		const suffix = entry.placeholder ? "_" : "";
		const link = entry.link ? `[[${entry.link}|${entry.displayText}]]` : entry.displayText;
		const description = entry.description ? ` - ${entry.description}` : "";
		return `${prefix}${link}${description}${suffix}`;
	}

	private parseFrontmatter(content: string): Record<string, string> {
		const match = content.match(FRONTMATTER_REGEX);
		if (!match) {
			return {};
		}

		const lines = match[1].split("\n");
		const data: Record<string, string> = {};
		for (const line of lines) {
			const colon = line.indexOf(":");
			if (colon === -1) continue;
			const key = line.slice(0, colon).trim();
			const value = line
				.slice(colon + 1)
				.trim()
				.replace(/^["']|["']$/g, "");
			data[key] = value;
		}
		return data;
	}
}

