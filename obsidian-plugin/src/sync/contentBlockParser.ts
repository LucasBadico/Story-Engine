import { ContentBlock } from "../types";

export interface ParsedParagraph {
	content: string; // Text without the link
	linkName: string | null; // File name (if has link)
	originalOrder: number; // Position in chapter (0-indexed)
}

export interface ParsedScene {
	linkName: string | null; // File name (if has link)
	goal: string; // Text before "-" or full text if no "-"
	timeRef: string; // Text after "-" or empty
	originalOrder: number; // Position in chapter (0-indexed)
}

export interface ParsedBeat {
	linkName: string | null; // File name (if has link)
	intent: string; // Text before "->" or full text if no "->"
	outcome: string; // Text after "->" or empty
	originalOrder: number; // Position relative to scene (0-indexed)
}

export interface ParsedSection {
	type: "prose" | "scene" | "beat";
	prose?: ParsedParagraph;
	scene?: ParsedScene;
	beat?: ParsedBeat;
	originalOrder: number; // Absolute position in chapter (0-indexed)
}

export interface HierarchicalProse {
	sections: ParsedSection[];
}

export interface ParsedSceneBeatListItem {
	type: "scene" | "beat";
	linkName: string | null;
	displayText: string;
	hasProse: boolean;
	indentLevel: number; // 0 for scene, 1 for beat
	originalOrder: number;
}

export interface ParsedSceneBeatList {
	items: ParsedSceneBeatListItem[];
}

export interface ParsedChapterListItem {
	type: "chapter" | "scene" | "beat";
	linkName: string | null;
	displayText: string;
	hasProse: boolean;
	indentLevel: number; // 0 for chapter, 1 for scene, 2 for beat
	originalOrder: number;
}

export interface ParsedChapterList {
	items: ParsedChapterListItem[];
}

export interface ParsedBeatListItem {
	linkName: string | null;
	displayText: string;
	hasProse: boolean;
	originalOrder: number;
}

export interface ParsedBeatList {
	items: ParsedBeatListItem[];
}

export interface ContentBlockComparison {
	paragraph: ParsedParagraph;
	localContentBlock: ContentBlock | null; // From .md file
	remoteContentBlock: ContentBlock | null; // From API
	status: "new" | "unchanged" | "local_modified" | "remote_modified" | "conflict";
}

/**
 * Parse the Prose section from a chapter markdown content
 * Extracts hierarchical structure: scenes (##), beats (###), and content blocks
 */
export function parseHierarchicalProse(chapterContent: string): HierarchicalProse {
	const sections: ParsedSection[] = [];

	// Extract content after frontmatter
	const frontmatterMatch = chapterContent.match(/^---\n([\s\S]*?)\n---/);
	const contentStart = frontmatterMatch ? frontmatterMatch[0].length : 0;
	const bodyContent = chapterContent.substring(contentStart).trim();

	// Find the "## Chapter N: {title}" section
	// Capture everything until the end of the document since Scene and Beat are part of the chapter section
	// Accept with or without blank line after the chapter header
	// Use non-greedy match to stop at next ## Chapter if it exists (to avoid capturing duplicate sections)
	const chapterSectionMatch = bodyContent.match(/##\s+Chapter\s+\d+:\s+[^\n]+\s*\n+([\s\S]*?)(?=\n##\s+Chapter\s+\d+:|$)/);
	if (!chapterSectionMatch) {
		// Try without requiring newline (content on same line as chapter header)
		const chapterSectionMatchSameLine = bodyContent.match(/##\s+Chapter\s+\d+:\s+[^\n]+\s+([^\n]+)/);
		if (chapterSectionMatchSameLine) {
			const proseContent = chapterSectionMatchSameLine[1].trim();
			// Process single line content
			if (proseContent.length > 0 && !proseContent.startsWith('#')) {
				const paragraph: ParsedParagraph = {
					content: proseContent,
					linkName: null,
					originalOrder: 0,
				};
				sections.push({
					type: "prose",
					prose: paragraph,
					originalOrder: 0,
				});
			}
			return { sections };
		}
		return { sections: [] };
	}

		const proseContent = chapterSectionMatch[1].trim();
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

		// Skip other markdown headers (e.g., # Title, ## Capítulo, etc.)
		// But don't skip Scene and Beat headers which are part of the chapter content
		if (line.startsWith('#') && !line.match(/^##\s+Scene:/) && !line.match(/^###\s+Beat:/)) {
			continue;
		}

		// Check for content block: [[link|content]] or plain text
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

		// Plain text paragraph (new content block without link)
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
 * Extracts paragraphs and identifies which ones have links to content blocks
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
 * Compare local and remote content blocks to determine status
 */
export function compareContentBlocks(
	paragraph: ParsedParagraph,
	localContentBlock: ContentBlock | null,
	remoteContentBlock: ContentBlock | null
): ContentBlockComparison["status"] {
	const paragraphContent = paragraph.content.trim();

	// Paragraph without link - check if remote exists with same content
	if (!paragraph.linkName) {
		if (remoteContentBlock && remoteContentBlock.content.trim() === paragraphContent) {
			// Remote exists with same content - treat as unchanged if local matches, otherwise need to create local file
			if (localContentBlock && localContentBlock.id === remoteContentBlock.id) {
				return "unchanged";
			}
			// Remote exists but no local file - need to create local file (treat as unchanged for content, but need file creation)
			return "unchanged";
		}
		// No remote match - this is a new content block
		return "new";
	}

	// No local file found (shouldn't happen if link exists, but handle gracefully)
	if (!localContentBlock) {
		// If remote exists with same content, treat as remote_modified
		if (remoteContentBlock && remoteContentBlock.content.trim() === paragraphContent) {
			return "remote_modified";
		}
		return "new";
	}

	// No remote content block found (deleted on server?)
	if (!remoteContentBlock) {
		return "local_modified";
	}

	const localContent = localContentBlock.content.trim();
	const remoteContent = remoteContentBlock.content.trim();

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

/**
 * Parse the "## Scenes & Beats" list from chapter content
 * Format:
 * ## Scenes & Beats
 * 
 * + [[link|Scene N: text]]
 *   - [[link|Beat N: text]]
 *   + [[link|Beat N: text]]
 * - [[link|Scene N: text]]
 */
export function parseSceneBeatList(chapterContent: string): ParsedSceneBeatList {
	const items: ParsedSceneBeatListItem[] = [];

	// Extract content after frontmatter
	const frontmatterMatch = chapterContent.match(/^---\n([\s\S]*?)\n---/);
	const contentStart = frontmatterMatch ? frontmatterMatch[0].length : 0;
	const bodyContent = chapterContent.substring(contentStart).trim();

	// Find the "## Scenes & Beats" section
	const listSectionMatch = bodyContent.match(/##\s+Scenes\s+&\s+Beats\s*\n+([\s\S]*?)(?=\n##|$)/);
	if (!listSectionMatch) {
		return { items: [] };
	}

	const listContent = listSectionMatch[1].trim();
	const lines = listContent.split(/\n/);

	let order = 0;
	for (const line of lines) {
		const trimmedLine = line.trim();
		if (!trimmedLine) {
			continue;
		}

		// Check if it's a scene or beat item
		// Format: [+/-] [[link|text]] or [+/-] text
		const itemMatch = trimmedLine.match(/^([+-])\s+(.+)$/);
		if (!itemMatch) {
			continue;
		}

		const hasProse = itemMatch[1] === "+";
		const itemText = itemMatch[2].trim();
		// Count tabs at the beginning of the line (more reliable than spaces)
		const tabMatch = line.match(/^(\t*)/);
		const indentLevel = tabMatch ? tabMatch[1].length : 0;
		const isBeat = indentLevel > 0;

		// Parse link if present
		const linkMatch = itemText.match(/^\[\[([^\|]+)\|([^\]]+)\]\]$/);
		let linkName: string | null = null;
		let displayText: string;

		if (linkMatch) {
			linkName = linkMatch[1].trim();
			displayText = linkMatch[2].trim();
		} else {
			displayText = itemText;
		}

		items.push({
			type: isBeat ? "beat" : "scene",
			linkName,
			displayText,
			hasProse,
			indentLevel: isBeat ? 1 : 0,
			originalOrder: order++,
		});
	}

	return { items };
}

/**
 * Parse the "## Chapters, Scenes & Beats" list from story content
 * Format:
 * ## Chapters, Scenes & Beats
 * 
 * - [[link|Chapter N: title]]
 *   - [[link|Scene N: goal - timeRef]]
 *     - [[link|Beat N: intent -> outcome]]
 *     - [[link|Beat N: intent]]
 *   - [[link|Scene N: goal]]
 */
export function parseChapterList(storyContent: string): ParsedChapterList {
	const items: ParsedChapterListItem[] = [];

	// Extract content after frontmatter
	const frontmatterMatch = storyContent.match(/^---\n([\s\S]*?)\n---/);
	const contentStart = frontmatterMatch ? frontmatterMatch[0].length : 0;
	const bodyContent = storyContent.substring(contentStart).trim();

	// Find the "## Chapters, Scenes & Beats" section (or "## Chapters" for backward compatibility)
	const listSectionMatch = bodyContent.match(/##\s+Chapters(?:,\s*Scenes\s*&\s*Beats)?\s*\n+([\s\S]*?)(?=\n##|$)/);
	if (!listSectionMatch) {
		return { items: [] };
	}

	const listContent = listSectionMatch[1].trim();
	const lines = listContent.split(/\n/);

	let order = 0;
	for (const line of lines) {
		const trimmedLine = line.trim();
		if (!trimmedLine) {
			continue;
		}

		// Check if it's a chapter, scene or beat item
		// Format: [+/-] [[link|text]] or [+/-] text
		const itemMatch = trimmedLine.match(/^([+-])\s+(.+)$/);
		if (!itemMatch) {
			continue;
		}

		const hasProse = itemMatch[1] === "+";
		const itemText = itemMatch[2].trim();
		// Count tabs at the beginning of the line (more reliable than spaces)
		const tabMatch = line.match(/^(\t*)/);
		const indentLevel = tabMatch ? tabMatch[1].length : 0;
		
		// Determine type based on indent level (tabs)
		// 0 tabs = chapter, 1 tab = scene, 2 tabs = beat
		let type: "chapter" | "scene" | "beat";
		if (indentLevel === 0) {
			type = "chapter";
		} else if (indentLevel === 1) {
			type = "scene";
		} else if (indentLevel === 2) {
			type = "beat";
		} else {
			// Skip items with unexpected indent level
			continue;
		}

		// Parse link if present
		const linkMatch = itemText.match(/^\[\[([^\|]+)\|([^\]]+)\]\]$/);
		let linkName: string | null = null;
		let displayText: string;

		if (linkMatch) {
			linkName = linkMatch[1].trim();
			displayText = linkMatch[2].trim();
		} else {
			displayText = itemText;
		}

		items.push({
			type,
			linkName,
			displayText,
			hasProse,
			indentLevel: indentLevel / 2, // Normalize to 0, 1, 2
			originalOrder: order++,
		});
	}

	return { items };
}

/**
 * Parse the "## Beats" list from scene content
 * Format:
 * ## Beats
 * 
 * - [[link|Beat N: intent -> outcome]]
 * + [[link|Beat N: intent -> outcome]]
 */
export function parseBeatList(sceneContent: string): ParsedBeatList {
	const items: ParsedBeatListItem[] = [];

	// Extract content after frontmatter
	const frontmatterMatch = sceneContent.match(/^---\n([\s\S]*?)\n---/);
	const contentStart = frontmatterMatch ? frontmatterMatch[0].length : 0;
	const bodyContent = sceneContent.substring(contentStart).trim();

	// Find the "## Beats" section
	const listSectionMatch = bodyContent.match(/##\s+Beats\s*\n+([\s\S]*?)(?=\n##|$)/);
	if (!listSectionMatch) {
		return { items: [] };
	}

	const listContent = listSectionMatch[1].trim();
	const lines = listContent.split(/\n/);

	let order = 0;
	for (const line of lines) {
		const trimmedLine = line.trim();
		if (!trimmedLine) {
			continue;
		}

		// Skip the info callout block
		if (trimmedLine.startsWith(">")) {
			continue;
		}

		// Check if it's a beat item
		// Format: [+/-] [[link|text]] or [+/-] text
		const itemMatch = trimmedLine.match(/^([+-])\s+(.+)$/);
		if (!itemMatch) {
			continue;
		}

		const hasProse = itemMatch[1] === "+";
		const itemText = itemMatch[2].trim();

		// Parse link if present
		const linkMatch = itemText.match(/^\[\[([^\|]+)\|([^\]]+)\]\]$/);
		let linkName: string | null = null;
		let displayText: string;

		if (linkMatch) {
			linkName = linkMatch[1].trim();
			displayText = linkMatch[2].trim();
		} else {
			displayText = itemText;
		}

		items.push({
			linkName,
			displayText,
			hasProse,
			originalOrder: order++,
		});
	}

	return { items };
}

/**
 * Parse the "## Orphan Scenes" list from story content
 * Format:
 * ## Orphan Scenes
 * 
 * - [[link|Scene N: goal - timeRef]]
 * + [[link|Scene N: goal - timeRef]]
 *   - [[link|Beat N: intent -> outcome]]
 */
export function parseOrphanScenesList(storyContent: string): ParsedSceneBeatList {
	const items: ParsedSceneBeatListItem[] = [];

	// Extract content after frontmatter
	const frontmatterMatch = storyContent.match(/^---\n([\s\S]*?)\n---/);
	const contentStart = frontmatterMatch ? frontmatterMatch[0].length : 0;
	const bodyContent = storyContent.substring(contentStart).trim();

	// Find the "## Orphan Scenes" section
	const listSectionMatch = bodyContent.match(/##\s+Orphan\s+Scenes\s*\n+([\s\S]*?)(?=\n##|$)/);
	if (!listSectionMatch) {
		return { items: [] };
	}

	const listContent = listSectionMatch[1].trim();
	const lines = listContent.split(/\n/);

	let order = 0;
	for (const line of lines) {
		const trimmedLine = line.trim();
		if (!trimmedLine) {
			continue;
		}

		// Skip the info callout block
		if (trimmedLine.startsWith(">")) {
			continue;
		}

		// Check if it's a scene or beat item
		// Format: [+/-] [[link|text]] or [+/-] text
		const itemMatch = trimmedLine.match(/^([+-])\s+(.+)$/);
		if (!itemMatch) {
			continue;
		}

		const hasProse = itemMatch[1] === "+";
		const itemText = itemMatch[2].trim();
		// Count tabs at the beginning of the line (more reliable than spaces)
		const tabMatch = line.match(/^(\t*)/);
		const indentLevel = tabMatch ? tabMatch[1].length : 0;
		const isBeat = indentLevel > 0;

		// Parse link if present
		const linkMatch = itemText.match(/^\[\[([^\|]+)\|([^\]]+)\]\]$/);
		let linkName: string | null = null;
		let displayText: string;

		if (linkMatch) {
			linkName = linkMatch[1].trim();
			displayText = linkMatch[2].trim();
		} else {
			displayText = itemText;
		}

		items.push({
			type: isBeat ? "beat" : "scene",
			linkName,
			displayText,
			hasProse,
			indentLevel: isBeat ? 1 : 0,
			originalOrder: order++,
		});
	}

	return { items };
}

/**
 * Parse the "## Orphan Beats" list from story content
 * Format:
 * ## Orphan Beats
 * 
 * - [[link|Beat N: intent -> outcome]]
 * + [[link|Beat N: intent -> outcome]]
 */
export function parseOrphanBeatsList(storyContent: string): ParsedBeatList {
	const items: ParsedBeatListItem[] = [];

	// Extract content after frontmatter
	const frontmatterMatch = storyContent.match(/^---\n([\s\S]*?)\n---/);
	const contentStart = frontmatterMatch ? frontmatterMatch[0].length : 0;
	const bodyContent = storyContent.substring(contentStart).trim();

	// Find the "## Orphan Beats" section
	const listSectionMatch = bodyContent.match(/##\s+Orphan\s+Beats\s*\n+([\s\S]*?)(?=\n##|$)/);
	if (!listSectionMatch) {
		return { items: [] };
	}

	const listContent = listSectionMatch[1].trim();
	const lines = listContent.split(/\n/);

	let order = 0;
	for (const line of lines) {
		const trimmedLine = line.trim();
		if (!trimmedLine) {
			continue;
		}

		// Skip the info callout block
		if (trimmedLine.startsWith(">")) {
			continue;
		}

		// Check if it's a beat item
		// Format: [+/-] [[link|text]] or [+/-] text
		const itemMatch = trimmedLine.match(/^([+-])\s+(.+)$/);
		if (!itemMatch) {
			continue;
		}

		const hasProse = itemMatch[1] === "+";
		const itemText = itemMatch[2].trim();

		// Parse link if present
		const linkMatch = itemText.match(/^\[\[([^\|]+)\|([^\]]+)\]\]$/);
		let linkName: string | null = null;
		let displayText: string;

		if (linkMatch) {
			linkName = linkMatch[1].trim();
			displayText = linkMatch[2].trim();
		} else {
			displayText = itemText;
		}

		items.push({
			linkName,
			displayText,
			hasProse,
			originalOrder: order++,
		});
	}

	return { items };
}

/**
 * Parse content blocks from story content using hierarchical headers
 * Format:
 * # Story: title
 * prose at story level
 * 
 * ## Chapter: title
 * prose at chapter level (no scene)
 * 
 * ### Scene: title
 * prose at scene level
 * 
 * #### Beat: title
 * prose at beat level
 */
export function parseStoryProse(storyContent: string): HierarchicalProse {
	const sections: ParsedSection[] = [];

	// Extract content after frontmatter
	const frontmatterMatch = storyContent.match(/^---\n([\s\S]*?)\n---/);
	const contentStart = frontmatterMatch ? frontmatterMatch[0].length : 0;
	const bodyContent = storyContent.substring(contentStart).trim();

	// Find the hierarchical prose section - starts with "# Story:" or "## Chapter N:" or "## Chapter:"
	// This is AFTER the list sections (## Chapters, Scenes & Beats, ## Orphan Scenes, etc.)
	const storyHeaderMatch = bodyContent.match(/^(#\s+Story:\s*.+)$/m);
	const chapterHeaderMatch = bodyContent.match(/^(##\s+Chapter\s*\d*:\s*.+)$/m);
	
	let proseStartIndex = -1;
	if (storyHeaderMatch) {
		proseStartIndex = bodyContent.indexOf(storyHeaderMatch[0]);
	} else if (chapterHeaderMatch) {
		proseStartIndex = bodyContent.indexOf(chapterHeaderMatch[0]);
	}
	
	if (proseStartIndex === -1) {
		return { sections: [] };
	}
	
	const proseContent = bodyContent.substring(proseStartIndex);
	const lines = proseContent.split(/\n/);

	let order = 0;
	let currentChapter: { linkName: string | null; title: string } | null = null;
	let currentScene: ParsedScene | null = null;
	let currentBeat: ParsedBeat | null = null;

	for (const line of lines) {
		const trimmedLine = line.trim();

		// Skip empty lines
		if (!trimmedLine) {
			continue;
		}

		// Skip info callouts
		if (trimmedLine.startsWith(">")) {
			continue;
		}

		// Check for Story header: # Story: [[link|title]] or # Story: title
		const storyMatch = trimmedLine.match(/^#\s+Story:\s*(.+)$/i);
		if (storyMatch) {
			// Story header - just skip, we're already in story context
			continue;
		}

		// Check for Chapter header: ## Chapter N: [[link|title]] or ## Chapter: title
		const chapterMatch = trimmedLine.match(/^##\s+Chapter\s*\d*:\s*(.+)$/i);
		if (chapterMatch) {
			const chapterText = chapterMatch[1].trim();
			const linkMatch = chapterText.match(/^\[\[([^\|]+)\|([^\]]+)\]\]$/);
			if (linkMatch) {
				currentChapter = {
					linkName: linkMatch[1].trim(),
					title: linkMatch[2].trim(),
				};
			} else {
				currentChapter = {
					linkName: null,
					title: chapterText,
				};
			}
			currentScene = null;
			currentBeat = null;
			continue;
		}

		// Check for Scene header: ### Scene: [[link|goal - timeRef]] or ### Scene: goal
		const sceneMatch = trimmedLine.match(/^###\s+Scene:\s*(.+)$/i);
		if (sceneMatch) {
			const sceneText = sceneMatch[1].trim();
			currentScene = parseSceneHeaderText(sceneText);
			currentScene.originalOrder = order++;
			sections.push({
				type: "scene",
				scene: currentScene,
				originalOrder: currentScene.originalOrder,
			});
			currentBeat = null;
			continue;
		}

		// Check for Beat header: #### Beat: [[link|intent -> outcome]] or #### Beat: intent
		const beatMatch = trimmedLine.match(/^####\s+Beat:\s*(.+)$/i);
		if (beatMatch) {
			const beatText = beatMatch[1].trim();
			currentBeat = parseBeatHeaderText(beatText);
			currentBeat.originalOrder = order++;
			sections.push({
				type: "beat",
				beat: currentBeat,
				originalOrder: currentBeat.originalOrder,
			});
			continue;
		}

		// Skip other headers that don't match our pattern
		if (trimmedLine.startsWith("#")) {
			continue;
		}

		// Skip list items (these are chapter/scene/beat lists)
		if (trimmedLine.match(/^[+-]\s+/)) {
			continue;
		}

		// Check for content block: [[link|content]] or plain text
		const proseMatch = trimmedLine.match(/^\s*\[\[([^\|]+)\|([^\]]+)\]\]\s*$/);
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

		// Plain text paragraph (new content block without link)
		if (trimmedLine.length > 0) {
			const paragraph: ParsedParagraph = {
				content: trimmedLine,
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
 * Helper to parse scene header text
 */
function parseSceneHeaderText(text: string): ParsedScene {
	const linkMatch = text.match(/^\[\[([^\|]+)\|([^\]]+)\]\]$/);
	if (linkMatch) {
		const linkName = linkMatch[1].trim();
		const displayText = linkMatch[2].trim();
		const parts = displayText.split(/\s*-\s*/);
		return {
			linkName,
			goal: parts[0].trim(),
			timeRef: parts.length > 1 ? parts.slice(1).join(" - ").trim() : "",
			originalOrder: 0,
		};
	}

	const simpleLinkMatch = text.match(/^\[\[([^\]]+)\]\]$/);
	if (simpleLinkMatch) {
		return {
			linkName: simpleLinkMatch[1].trim(),
			goal: "",
			timeRef: "",
			originalOrder: 0,
		};
	}

	const parts = text.split(/\s*-\s*/);
	return {
		linkName: null,
		goal: parts[0].trim(),
		timeRef: parts.length > 1 ? parts.slice(1).join(" - ").trim() : "",
		originalOrder: 0,
	};
}

/**
 * Helper to parse beat header text
 */
function parseBeatHeaderText(text: string): ParsedBeat {
	const linkMatch = text.match(/^\[\[([^\|]+)\|([^\]]+)\]\]$/);
	if (linkMatch) {
		const linkName = linkMatch[1].trim();
		const displayText = linkMatch[2].trim();
		const parts = displayText.split(/\s*->\s*/);
		return {
			linkName,
			intent: parts[0].trim(),
			outcome: parts.length > 1 ? parts.slice(1).join(" -> ").trim() : "",
			originalOrder: 0,
		};
	}

	const simpleLinkMatch = text.match(/^\[\[([^\]]+)\]\]$/);
	if (simpleLinkMatch) {
		return {
			linkName: simpleLinkMatch[1].trim(),
			intent: "",
			outcome: "",
			originalOrder: 0,
		};
	}

	const parts = text.split(/\s*->\s*/);
	return {
		linkName: null,
		intent: parts[0].trim(),
		outcome: parts.length > 1 ? parts.slice(1).join(" -> ").trim() : "",
		originalOrder: 0,
	};
}

/**
 * Parse content blocks from scene content using hierarchical headers
 * Format (no chapters in scenes):
 * 
 * ## Beats (list - skipped)
 * - [[beat-1|Beat 1: intent]]
 * 
 * ### Scene: title (or just start with prose)
 * prose at scene level
 * 
 * #### Beat: title (or ### Beat:)
 * prose at beat level
 */
export function parseSceneProse(sceneContent: string): HierarchicalProse {
	const sections: ParsedSection[] = [];

	// Extract content after frontmatter
	const frontmatterMatch = sceneContent.match(/^---\n([\s\S]*?)\n---/);
	const contentStart = frontmatterMatch ? frontmatterMatch[0].length : 0;
	const bodyContent = sceneContent.substring(contentStart).trim();

	// Find the hierarchical prose section
	// It starts with "### Scene:" or "#### Beat:" or "### Beat:" or just prose text after ## Beats list
	const sceneHeaderMatch = bodyContent.match(/^(###\s+Scene:\s*.+)$/m);
	const beatHeaderMatch = bodyContent.match(/^(#{3,4}\s+Beat:\s*.+)$/m);
	
	let proseStartIndex = -1;
	
	if (sceneHeaderMatch) {
		proseStartIndex = bodyContent.indexOf(sceneHeaderMatch[0]);
	} else if (beatHeaderMatch) {
		proseStartIndex = bodyContent.indexOf(beatHeaderMatch[0]);
	} else {
		// No explicit headers - look for prose after ## Beats list
		const beatsListEnd = bodyContent.match(/##\s+Beats\s*\n+[\s\S]*?(?=\n###|\n####|\n[^#\-\+>\s]|$)/);
		if (beatsListEnd) {
			proseStartIndex = bodyContent.indexOf(beatsListEnd[0]) + beatsListEnd[0].length;
		} else {
			// No ## Beats section, start from beginning (skip any ## headers that are lists)
			const firstNonListContent = bodyContent.match(/(?:^|\n)([^#\-\+>\s\n].+)/);
			if (firstNonListContent) {
				proseStartIndex = bodyContent.indexOf(firstNonListContent[1]);
			}
		}
	}
	
	if (proseStartIndex === -1) {
		return { sections: [] };
	}
	
	const proseContent = bodyContent.substring(proseStartIndex);
	const lines = proseContent.split(/\n/);

	let order = 0;
	let currentBeat: ParsedBeat | null = null;

	for (const line of lines) {
		const trimmedLine = line.trim();

		// Skip empty lines
		if (!trimmedLine) {
			continue;
		}

		// Skip info callouts
		if (trimmedLine.startsWith(">")) {
			continue;
		}

		// Check for Scene header: ### Scene: [[link|goal]] or ### Scene: goal
		const sceneMatch = trimmedLine.match(/^###\s+Scene:\s*(.+)$/i);
		if (sceneMatch) {
			const sceneText = sceneMatch[1].trim();
			const parsedScene = parseSceneHeaderText(sceneText);
			parsedScene.originalOrder = order++;
			sections.push({
				type: "scene",
				scene: parsedScene,
				originalOrder: parsedScene.originalOrder,
			});
			currentBeat = null;
			continue;
		}

		// Check for Beat header: ### Beat: or #### Beat:
		const beatMatch = trimmedLine.match(/^#{3,4}\s+Beat:\s*(.+)$/i);
		if (beatMatch) {
			const beatText = beatMatch[1].trim();
			currentBeat = parseBeatHeaderText(beatText);
			currentBeat.originalOrder = order++;
			sections.push({
				type: "beat",
				beat: currentBeat,
				originalOrder: currentBeat.originalOrder,
			});
			continue;
		}

		// Skip other headers (like ## Beats list header)
		if (trimmedLine.startsWith("#")) {
			continue;
		}

		// Skip list items (these are beat lists)
		if (trimmedLine.match(/^[+-]\s+/)) {
			continue;
		}

		// Check for content block: [[link|content]] or plain text
		const proseMatch = trimmedLine.match(/^\s*\[\[([^\|]+)\|([^\]]+)\]\]\s*$/);
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

		// Plain text paragraph (new content block without link)
		if (trimmedLine.length > 0) {
			const paragraph: ParsedParagraph = {
				content: trimmedLine,
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
 * Parse content blocks from beat content using hierarchical headers
 * Format:
 * 
 * ## Beat: [[link|intent -> outcome]]
 * prose at beat level
 */
export function parseBeatProse(beatContent: string): HierarchicalProse {
	const sections: ParsedSection[] = [];

	// Extract content after frontmatter
	const frontmatterMatch = beatContent.match(/^---\n([\s\S]*?)\n---/);
	const contentStart = frontmatterMatch ? frontmatterMatch[0].length : 0;
	const bodyContent = beatContent.substring(contentStart).trim();

	// Find the hierarchical prose section - starts with "## Beat:"
	const beatHeaderMatch = bodyContent.match(/^(##\s+Beat:\s*.+)$/m);
	
	let proseStartIndex = -1;
	if (beatHeaderMatch) {
		proseStartIndex = bodyContent.indexOf(beatHeaderMatch[0]);
	}
	
	if (proseStartIndex === -1) {
		return { sections: [] };
	}
	
	const proseContent = bodyContent.substring(proseStartIndex);
	const lines = proseContent.split(/\n/);

	let order = 0;

	for (const line of lines) {
		const trimmedLine = line.trim();

		// Skip empty lines
		if (!trimmedLine) {
			continue;
		}

		// Skip info callouts
		if (trimmedLine.startsWith(">")) {
			continue;
		}

		// Check for Beat header: ## Beat: [[link|text]] or ## Beat: text
		const beatMatch = trimmedLine.match(/^##\s+Beat:\s*(.+)$/i);
		if (beatMatch) {
			// Beat header - just skip, we're already in beat context
			continue;
		}

		// Skip other headers
		if (trimmedLine.startsWith("#")) {
			continue;
		}

		// Skip list items
		if (trimmedLine.match(/^[+-]\s+/)) {
			continue;
		}

		// Check for content block: [[link|content]] or plain text
		const proseMatch = trimmedLine.match(/^\s*\[\[([^\|]+)\|([^\]]+)\]\]\s*$/);
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

		// Plain text paragraph (new content block without link)
		if (trimmedLine.length > 0) {
			const paragraph: ParsedParagraph = {
				content: trimmedLine,
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

