import { parseFrontmatter } from "../utils/detectEntityMentions";

export interface ParsedWorldEntity {
	id: string;
	name: string;
	description: string | null;
	frontmatter: Record<string, unknown>;
}

export function parseWorldEntityFile(content: string): ParsedWorldEntity {
	const frontmatter = parseFrontmatter(content);

	const nameMatch = content.match(/^#\s+(.+)$/m);
	const name = nameMatch ? nameMatch[1].trim() : "";

	const descriptionMatch = content.match(/##\s+Description\n([\s\S]*?)(?=\n##|$)/);
	let description: string | null = null;
	if (descriptionMatch) {
		const desc = descriptionMatch[1].trim();
		if (desc && desc !== "_No description yet._") {
			description = desc;
		}
	}

	return {
		id: (frontmatter.id as string) ?? "",
		name,
		description,
		frontmatter,
	};
}
