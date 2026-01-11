import type { App, TFile } from "obsidian";
import type { SyncContext } from "../types/sync";
import { getFrontmatterId } from "./frontmatterHelpers";

/**
 * Result of detecting an entity mention from a link
 */
export interface DetectedEntityMention {
	/** The full link text (e.g., `[[worlds/eldoria/characters/aria-moon]]` or `[[aria-moon]]`) */
	linkText: string;
	/** The filename path from the link (e.g., `worlds/eldoria/characters/aria-moon` or `aria-moon`) */
	filenamePath: string;
	/** The display label if present (e.g., `Aria Moon` from `[[aria-moon|Aria Moon]]`) */
	displayLabel?: string;
	/** The detected format: 'official' (with full path) or 'obsidian' (without path) */
	format: "official" | "obsidian";
	/** Resolved entity ID (if successfully resolved) */
	entityId?: string;
	/** Resolved entity type (if successfully resolved) */
	entityType?: string;
	/** World name if extracted from path (official format only) */
	worldName?: string;
}

/**
 * Parse markdown content to detect entity mentions in link format
 * Supports two formats:
 * 1. Official format (written by sync): `[[filename path]]` or `[[filename path|label]]`
 *    Example: `[[worlds/eldoria/characters/aria-moon]]` or `[[worlds/eldoria/characters/aria-moon|Aria Moon]]`
 * 2. Obsidian format: `[[filename without path]]` or `[[filename without path|label]]`
 *    Example: `[[aria-moon]]` or `[[aria-moon|Aria]]`
 *
 * @param content - The markdown content to parse
 * @returns Array of detected entity mentions (links)
 */
export function detectEntityMentions(content: string): DetectedEntityMention[] {
	const mentions: DetectedEntityMention[] = [];
	
	// Regex to match Obsidian links: [[link|label]] or [[link]]
	// This matches both formats
	const linkRegex = /\[\[([^\]]+)\]\]/g;
	let match: RegExpExecArray | null;
	
	while ((match = linkRegex.exec(content)) !== null) {
		const fullLink = match[0]; // e.g., `[[worlds/eldoria/characters/aria-moon|Aria Moon]]`
		const linkContent = match[1]; // e.g., `worlds/eldoria/characters/aria-moon|Aria Moon`
		
		// Split link content into filename path and display label
		const [filenamePath, displayLabel] = linkContent.split("|").map((s) => s.trim());
		
		// Determine format based on whether the path contains `/` (indicating a full path)
		const format: "official" | "obsidian" = filenamePath.includes("/") ? "official" : "obsidian";
		
		mentions.push({
			linkText: fullLink,
			filenamePath,
			displayLabel: displayLabel || undefined,
			format,
		});
	}
	
	return mentions;
}

/**
 * Resolve a detected entity mention to its entity ID and type
 * For official format: resolves using Obsidian vault by full path (e.g., `worlds/eldoria/characters/aria-moon.md`)
 * For Obsidian format: resolves using Obsidian vault by filename (e.g., `aria-moon.md`)
 * Then reads the file to infer entity type and ID from frontmatter
 *
 * @param mention - The detected entity mention
 * @param context - Sync context with app, apiClient, and fileManager
 * @returns Resolved entity ID and type, or null if resolution fails
 */
export async function resolveEntityMention(
	mention: DetectedEntityMention,
	context: SyncContext
): Promise<{ entityId: string; entityType: string; worldId?: string } | null> {
	try {
		const vault = context.app.vault;
		const metadataCache = context.app.metadataCache;
		
		let file: TFile | null = null;
		
		if (mention.format === "official") {
			// Official format: resolve by full path
			// Example: worlds/eldoria/characters/aria-moon -> worlds/eldoria/characters/aria-moon.md
			const filePath = `${mention.filenamePath}.md`;
			const abstractFile = vault.getAbstractFileByPath(filePath);
			// Check if it's a file by verifying it has file-like properties
			// instanceof doesn't work with mocks, so we check for properties instead
			// TFolder has 'children' property, TFile doesn't
			// If it has a path and is not a folder (no children property), it's a file
			if (abstractFile && "path" in abstractFile && !("children" in abstractFile)) {
				file = abstractFile as TFile;
			}
		} else {
			// Obsidian format: resolve by filename using Obsidian's link resolution
			// Try multiple strategies:
			
			// Strategy 1: Try exact filename match (basename without extension)
			const markdownFiles = vault.getMarkdownFiles();
			file = markdownFiles.find((f) => {
				return f.basename === mention.filenamePath;
			}) || null;
			
			// Strategy 2: Try using metadataCache to resolve link (handles aliases, titles, etc.)
			if (!file) {
				const resolvedFile = metadataCache.getFirstLinkpathDest(mention.filenamePath, "");
				if (resolvedFile && "path" in resolvedFile && !("children" in resolvedFile)) {
					file = resolvedFile as TFile;
				}
			}
			
			// Strategy 3: Try filename with .md extension in any location
			if (!file) {
				file = markdownFiles.find((f) => {
					return f.name === `${mention.filenamePath}.md`;
				}) || null;
			}
		}
		
		if (!file) {
			return null; // Could not resolve link
		}
		
		// Read the file to get frontmatter and infer entity type
		const fileContent = await vault.read(file);
		const frontmatter = parseFrontmatter(fileContent);
		
		// Infer entity type from file path and frontmatter
		const entityType = inferEntityTypeFromFile(file.path, frontmatter);
		if (!entityType) {
			return null; // Could not infer entity type
		}
		
		// Get entity ID from frontmatter using configured field name
		const idField = context.settings.frontmatterIdField;
		const entityId = getFrontmatterId(frontmatter, idField);
		if (!entityId) {
			return null; // No entity ID in frontmatter
		}
		
		// Get world_id if present (for world entities)
		// Convert null to undefined for optional fields
		const worldId = frontmatter.world_id as string | null | undefined;
		
		return {
			entityId,
			entityType,
			worldId: worldId === null ? undefined : worldId,
		};
	} catch (error) {
		console.warn("[Sync V2] Failed to resolve entity mention", {
			mention,
			error,
		});
		return null;
	}
}

/**
 * Parse frontmatter from markdown content
 * Returns a record with frontmatter fields
 * Handles simple key-value pairs, arrays (tags), null values, and basic types
 */
export function parseFrontmatter(content: string): Record<string, unknown> {
	const match = content.match(/^---\n([\s\S]*?)\n---/);
	if (!match) {
		return {};
	}

	const frontmatterText = match[1];
	const result: Record<string, unknown> = {};
	const lines = frontmatterText.split("\n");
	
	let i = 0;
	while (i < lines.length) {
		const line = lines[i].trim();
		
		// Skip empty lines and comments
		if (!line || line.startsWith("#")) {
			i++;
			continue;
		}
		
		// Check if this is a key-value pair
		const colonIndex = line.indexOf(":");
		if (colonIndex > 0) {
			const key = line.slice(0, colonIndex).trim();
			let value = line.slice(colonIndex + 1).trim();
			
			// Check if this is an array (tags: or tags: [])
			if (key === "tags" && (value === "" || value === "[]")) {
				// Tags array - collect list items
				const tags: string[] = [];
				i++; // Move to next line
				
				// Collect all list items (lines starting with `-`)
				while (i < lines.length) {
					const nextLine = lines[i].trim();
					if (nextLine.startsWith("-")) {
						const tagMatch = nextLine.match(/^\s*-\s*(.+)/);
						if (tagMatch) {
							let tagValue = tagMatch[1].trim().replace(/^["']|["']$/g, "");
							tags.push(tagValue);
						}
						i++;
					} else if (nextLine === "" || nextLine.startsWith("#")) {
						// Empty line or comment, continue to next
						i++;
					} else {
						// End of tags array
						break;
					}
				}
				
				result[key] = tags;
				continue;
			}
			
			// Handle regular key-value pairs
			// Remove quotes if present
			if ((value.startsWith('"') && value.endsWith('"')) || (value.startsWith("'") && value.endsWith("'"))) {
				value = value.slice(1, -1);
			}
			
			// Handle null values
			if (value === "null" || value === "") {
				result[key] = null;
			} else if (value === "true" || value === "false") {
				result[key] = value === "true";
			} else if (/^-?\d+$/.test(value)) {
				result[key] = parseInt(value, 10);
			} else if (/^-?\d+\.\d+$/.test(value)) {
				result[key] = parseFloat(value);
			} else {
				result[key] = value;
			}
		} else if (line.startsWith("-")) {
			// List item outside of tags - skip or handle as needed
			// For now, skip standalone list items
		}
		
		i++;
	}

	return result;
}

/**
 * Infer entity type from file path and frontmatter
 * Priority: explicit entity_type field > frontmatter tags > file path patterns
 */
function inferEntityTypeFromFile(
	filePath: string,
	frontmatter: Record<string, unknown>
): string | null {
	// Priority 1: Check frontmatter for explicit entity type field (highest priority)
	if (frontmatter.entity_type && typeof frontmatter.entity_type === "string") {
		return frontmatter.entity_type;
	}
	
	// Priority 2: Check frontmatter tags for story-engine/{entityType} tag
	if (frontmatter.tags && Array.isArray(frontmatter.tags)) {
		for (const tag of frontmatter.tags) {
			if (typeof tag === "string" && tag.startsWith("story-engine/")) {
				const entityType = tag.replace("story-engine/", "");
				return entityType;
			}
		}
	}
	
	// Priority 3: Check file path patterns (lowest priority, used as fallback)
	const pathPatterns: Record<string, string> = {
		"/_archetypes/": "archetype", // More specific patterns first
		"/_traits/": "trait",
		"/worlds/": "world",
		"/characters/": "character",
		"/locations/": "location",
		"/factions/": "faction",
		"/artifacts/": "artifact",
		"/events/": "event",
		"/lore/": "lore",
		"/chapters/": "chapter",
		"/scenes/": "scene",
		"/beats/": "beat",
		"/contents/": "content_block",
	};
	
	for (const [pattern, entityType] of Object.entries(pathPatterns)) {
		if (filePath.includes(pattern)) {
			return entityType;
		}
	}
	
	return null;
}

