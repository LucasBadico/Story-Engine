import { ProseBlock } from "../types";

export interface ParsedParagraph {
	content: string; // Texto sem o link
	linkName: string | null; // Nome do arquivo (se tiver link)
	originalOrder: number; // Posição no chapter (0-indexed)
}

export interface ProseBlockComparison {
	paragraph: ParsedParagraph;
	localProseBlock: ProseBlock | null; // Do arquivo .md
	remoteProseBlock: ProseBlock | null; // Da API
	status: "new" | "unchanged" | "local_modified" | "remote_modified" | "conflict";
}

/**
 * Parse the Prose section from a chapter markdown content
 * Extracts paragraphs and identifies which ones have links to prose blocks
 */
export function parseChapterProse(chapterContent: string): ParsedParagraph[] {
	const paragraphs: ParsedParagraph[] = [];

	// Extract content after frontmatter
	const frontmatterMatch = chapterContent.match(/^---\n([\s\S]*?)\n---/);
	const contentStart = frontmatterMatch ? frontmatterMatch[0].length : 0;
	const bodyContent = chapterContent.substring(contentStart).trim();

	// Find the "## Prose" section
	const proseSectionMatch = bodyContent.match(/##\s+Prose\s*\n\n([\s\S]*?)(?=\n##|\n*$)/);
	if (!proseSectionMatch) {
		return paragraphs;
	}

	const proseContent = proseSectionMatch[1].trim();

	// Split by double newlines to get paragraphs
	const rawParagraphs = proseContent.split(/\n\n+/).filter((p) => p.trim().length > 0);

	let order = 0;
	for (const rawPara of rawParagraphs) {
		// Check if paragraph is a link in format [[name|content]]
		// The entire paragraph should be just the link
		const linkMatch = rawPara.trim().match(/^\[\[([^\|]+)\|([^\]]+)\]\]$/);
		
		if (linkMatch) {
			// Has link - extract link name and content
			const linkName = linkMatch[1].trim();
			const content = linkMatch[2].trim();
			
			paragraphs.push({
				content,
				linkName,
				originalOrder: order++,
			});
		} else {
			// No link - new prose block (just plain text)
			paragraphs.push({
				content: rawPara.trim(),
				linkName: null,
				originalOrder: order++,
			});
		}
	}

	return paragraphs;
}

/**
 * Compare local and remote prose blocks to determine status
 */
export function compareProseBlocks(
	paragraph: ParsedParagraph,
	localProseBlock: ProseBlock | null,
	remoteProseBlock: ProseBlock | null
): ProseBlockComparison["status"] {
	// New paragraph without link
	if (!paragraph.linkName) {
		return "new";
	}

	// No local file found (shouldn't happen if link exists, but handle gracefully)
	if (!localProseBlock) {
		return "new";
	}

	// No remote prose block found (deleted on server?)
	if (!remoteProseBlock) {
		return "local_modified";
	}

	const localContent = localProseBlock.content.trim();
	const remoteContent = remoteProseBlock.content.trim();
	const paragraphContent = paragraph.content.trim();

	// Both local and remote match paragraph
	if (localContent === paragraphContent && remoteContent === paragraphContent) {
		return "unchanged";
	}

	// Only paragraph changed (local edit)
	if (paragraphContent !== localContent && paragraphContent !== remoteContent && localContent === remoteContent) {
		return "local_modified";
	}

	// Only remote changed (server edit)
	if (localContent === paragraphContent && remoteContent !== paragraphContent) {
		return "remote_modified";
	}

	// Both changed (conflict)
	if (paragraphContent !== localContent && paragraphContent !== remoteContent && localContent !== remoteContent) {
		return "conflict";
	}

	// Paragraph matches remote but not local (local file outdated)
	if (paragraphContent === remoteContent && localContent !== remoteContent) {
		return "remote_modified";
	}

	// Default to conflict if unclear
	return "conflict";
}

