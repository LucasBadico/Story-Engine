import { ProseBlock } from "../types";

export interface ParsedParagraph {
	content: string; // Texto sem o link
	linkName: string | null; // Nome do arquivo (se tiver link)
	originalOrder: number; // Posição no chapter (0-indexed)
}

export interface ParsedScene {
	linkName: string | null; // Nome do arquivo (se tiver link)
	goal: string; // Texto antes do "-" ou texto completo se não tiver "-"
	timeRef: string; // Texto depois do "-" ou vazio
	originalOrder: number; // Posição no chapter (0-indexed)
}

export interface ParsedBeat {
	linkName: string | null; // Nome do arquivo (se tiver link)
	intent: string; // Texto antes do "->" ou texto completo se não tiver "->"
	outcome: string; // Texto depois do "->" ou vazio
	originalOrder: number; // Posição relativa à scene (0-indexed)
}

export interface ParsedSection {
	type: "prose" | "scene" | "beat";
	prose?: ParsedParagraph;
	scene?: ParsedScene;
	beat?: ParsedBeat;
	originalOrder: number; // Posição absoluta no chapter (0-indexed)
}

export interface HierarchicalProse {
	sections: ParsedSection[];
}

export interface ProseBlockComparison {
	paragraph: ParsedParagraph;
	localProseBlock: ProseBlock | null; // Do arquivo .md
	remoteProseBlock: ProseBlock | null; // Da API
	status: "new" | "unchanged" | "local_modified" | "remote_modified" | "conflict";
}

/**
 * Parse the Prose section from a chapter markdown content
 * Extracts hierarchical structure: scenes (##), beats (###), and prose blocks
 */
export function parseHierarchicalProse(chapterContent: string): HierarchicalProse {
	const sections: ParsedSection[] = [];

	// Extract content after frontmatter
	const frontmatterMatch = chapterContent.match(/^---\n([\s\S]*?)\n---/);
	const contentStart = frontmatterMatch ? frontmatterMatch[0].length : 0;
	const bodyContent = chapterContent.substring(contentStart).trim();

	// Find the "## Prose" section
	// Capture everything until the end of the document since Scene and Beat are part of Prose section
	const proseSectionMatch = bodyContent.match(/##\s+Prose\s*\n\n([\s\S]*)$/);
	if (!proseSectionMatch) {
		return { sections: [] };
	}

	const proseContent = proseSectionMatch[1].trim();
	const lines = proseContent.split(/\n/);

	let currentScene: ParsedScene | null = null;
	let currentBeat: ParsedBeat | null = null;
	let order = 0;

	for (let i = 0; i < lines.length; i++) {
		const line = lines[i].trim();

		// Empty line - skip
		if (!line) {
			continue;
		}

		// Check for scene header FIRST: ## Scene: [[link|text]] or ## Scene: text
		const sceneMatch = line.match(/^##\s+Scene:\s*(.+)$/);
		if (sceneMatch) {
			const sceneText = sceneMatch[1].trim();
			const parsedScene = parseSceneHeader(sceneText);
			parsedScene.originalOrder = order++;
			currentScene = parsedScene;
			currentBeat = null; // Reset beat when new scene starts
			sections.push({
				type: "scene",
				scene: parsedScene,
				originalOrder: parsedScene.originalOrder,
			});
			continue;
		}

		// Check for beat header: ### Beat: [[link|text]] or ### Beat: text
		const beatMatch = line.match(/^###\s+Beat:\s*(.+)$/);
		if (beatMatch) {
			const beatText = beatMatch[1].trim();
			const parsedBeat = parseBeatHeader(beatText);
			parsedBeat.originalOrder = order++;
			currentBeat = parsedBeat;
			sections.push({
				type: "beat",
				beat: parsedBeat,
				originalOrder: parsedBeat.originalOrder,
			});
			continue;
		}

		// Skip other markdown headers (e.g., # Title, ## Prose, etc.)
		if (line.startsWith('#')) {
			continue;
		}

		// Check for prose block: [[link|content]] or plain text
		// Match can have optional whitespace around the link
		const proseMatch = line.match(/^\s*\[\[([^\|]+)\|([^\]]+)\]\]\s*$/);
		if (proseMatch) {
			const linkName = proseMatch[1].trim();
			const content = proseMatch[2].trim();
			const paragraph: ParsedParagraph = {
				content,
				linkName,
				originalOrder: order++,
			};
			sections.push({
				type: "prose",
				prose: paragraph,
				originalOrder: paragraph.originalOrder,
			});
			continue;
		}

		// Plain text paragraph (new prose block without link)
		// Only if it's not a markdown header and not empty
		if (line.length > 0 && !line.startsWith('#')) {
			const paragraph: ParsedParagraph = {
				content: line,
				linkName: null,
				originalOrder: order++,
			};
			sections.push({
				type: "prose",
				prose: paragraph,
				originalOrder: paragraph.originalOrder,
			});
		}
	}

	return { sections };
}

/**
 * Parse a scene header: ## [[link|goal - timeRef]] or ## goal - timeRef or ## goal
 */
function parseSceneHeader(text: string): ParsedScene {
	// Check if it's a link format: [[link|text]]
	const linkMatch = text.match(/^\[\[([^\|]+)\|([^\]]+)\]\]$/);
	if (linkMatch) {
		const linkName = linkMatch[1].trim();
		const displayText = linkMatch[2].trim();
		const { goal, timeRef } = parseSceneText(displayText);
		return {
			linkName,
			goal,
			timeRef,
			originalOrder: 0, // Will be set by caller
		};
	}

	// Check if it's just a link: [[link]]
	const simpleLinkMatch = text.match(/^\[\[([^\]]+)\]\]$/);
	if (simpleLinkMatch) {
		return {
			linkName: simpleLinkMatch[1].trim(),
			goal: "",
			timeRef: "",
			originalOrder: 0,
		};
	}

	// Plain text - parse goal and timeRef
	const { goal, timeRef } = parseSceneText(text);
	return {
		linkName: null,
		goal,
		timeRef,
		originalOrder: 0,
	};
}

/**
 * Parse scene text to extract goal and timeRef
 * Format: "goal - timeRef" or just "goal"
 */
function parseSceneText(text: string): { goal: string; timeRef: string } {
	const parts = text.split(/\s*-\s*/);
	if (parts.length >= 2) {
		return {
			goal: parts[0].trim(),
			timeRef: parts.slice(1).join(" - ").trim(), // Join in case there are multiple "-"
		};
	}
	return {
		goal: text.trim(),
		timeRef: "",
	};
}

/**
 * Parse a beat header: ### [[link|intent -> outcome]] or ### intent -> outcome or ### intent
 */
function parseBeatHeader(text: string): ParsedBeat {
	// Check if it's a link format: [[link|text]]
	const linkMatch = text.match(/^\[\[([^\|]+)\|([^\]]+)\]\]$/);
	if (linkMatch) {
		const linkName = linkMatch[1].trim();
		const displayText = linkMatch[2].trim();
		const { intent, outcome } = parseBeatText(displayText);
		return {
			linkName,
			intent,
			outcome,
			originalOrder: 0, // Will be set by caller
		};
	}

	// Check if it's just a link: [[link]]
	const simpleLinkMatch = text.match(/^\[\[([^\]]+)\]\]$/);
	if (simpleLinkMatch) {
		return {
			linkName: simpleLinkMatch[1].trim(),
			intent: "",
			outcome: "",
			originalOrder: 0,
		};
	}

	// Plain text - parse intent and outcome
	const { intent, outcome } = parseBeatText(text);
	return {
		linkName: null,
		intent,
		outcome,
		originalOrder: 0,
	};
}

/**
 * Parse beat text to extract intent and outcome
 * Format: "intent -> outcome" or just "intent"
 */
function parseBeatText(text: string): { intent: string; outcome: string } {
	const parts = text.split(/\s*->\s*/);
	if (parts.length >= 2) {
		return {
			intent: parts[0].trim(),
			outcome: parts.slice(1).join(" -> ").trim(), // Join in case there are multiple "->"
		};
	}
	return {
		intent: text.trim(),
		outcome: "",
	};
}

/**
 * Legacy function - kept for backward compatibility
 * Parse the Prose section from a chapter markdown content
 * Extracts paragraphs and identifies which ones have links to prose blocks
 */
export function parseChapterProse(chapterContent: string): ParsedParagraph[] {
	const hierarchical = parseHierarchicalProse(chapterContent);
	const paragraphs: ParsedParagraph[] = [];

	for (const section of hierarchical.sections) {
		if (section.type === "prose" && section.prose) {
			paragraphs.push(section.prose);
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
