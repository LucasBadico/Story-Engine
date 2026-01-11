export interface EntityFrontmatter {
	[key: string]: string;
}

export interface EntityHeading {
	level: number;
	title: string;
	content: string;
}

export interface ParsedEntityFile {
	frontmatter: EntityFrontmatter;
	body: string;
	headings: EntityHeading[];
}

const FRONTMATTER_REGEX = /^---\n([\s\S]*?)\n---/;
const HEADING_REGEX = /^(#{1,6})\s+(.+)$/;

export class EntityFileParser {
	parse(content: string): ParsedEntityFile {
		const frontmatter = this.parseFrontmatter(content);
		const body = content.replace(FRONTMATTER_REGEX, "").trimStart();
		const headings = this.parseHeadings(body);
		return { frontmatter, body, headings };
	}

	updateFrontmatter(content: string, updates: Record<string, string>): string {
		const frontmatter = this.parseFrontmatter(content);
		const merged = { ...frontmatter, ...updates };
		const serialized = this.serializeFrontmatter(merged);

		if (content.match(FRONTMATTER_REGEX)) {
			return content.replace(FRONTMATTER_REGEX, serialized);
		}

		return `${serialized}\n${content}`;
	}

	getSectionContent(parsed: ParsedEntityFile, title: string): string | undefined {
		const heading = parsed.headings.find(
			(h) => h.title.toLowerCase() === title.toLowerCase()
		);
		return heading?.content?.trim();
	}

	private parseFrontmatter(content: string): EntityFrontmatter {
		const match = content.match(FRONTMATTER_REGEX);
		if (!match) {
			return {};
		}

		const lines = match[1].split("\n");
		const data: EntityFrontmatter = {};
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

	private serializeFrontmatter(frontmatter: EntityFrontmatter): string {
		const lines = Object.entries(frontmatter).map(([key, value]) => `${key}: ${value}`);
		return `---\n${lines.join("\n")}\n---`;
	}

	private parseHeadings(body: string): EntityHeading[] {
		const lines = body.split("\n");
		const headings: EntityHeading[] = [];
		let current: EntityHeading | null = null;

		for (const line of lines) {
			const match = line.match(HEADING_REGEX);
			if (match) {
				if (current) {
					headings.push(current);
				}
				current = {
					level: match[1].length,
					title: match[2].trim(),
					content: "",
				};
			} else if (current) {
				current.content += `${line}\n`;
			}
		}

		if (current) {
			headings.push(current);
		}

		return headings.map((heading) => ({
			...heading,
			content: heading.content.trim(),
		}));
	}
}

