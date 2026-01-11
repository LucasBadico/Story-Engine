import { buildFrontmatterFields, getIdFieldName } from "../utils/frontmatterHelpers";

export type EntityType =
	| "story"
	| "chapter"
	| "scene"
	| "beat"
	| "content-block"
	| "world"
	| "character"
	| "location"
	| "faction"
	| "artifact"
	| "event"
	| "lore"
	| "archetype"
	| "trait";

export interface FrontmatterOptions {
	entityType: EntityType;
	storyName?: string;
	worldName?: string;
	date?: string; // ISO date string (YYYY-MM-DD) or Date object
	/** Custom ID field name (from settings), defaults to "id" */
	idField?: string;
}

export class FrontmatterGenerator {
	generate(
		baseFields: Record<string, string | number | null>,
		extraFields?: Record<string, string | number | null>,
		options?: FrontmatterOptions
	): string {
		// Extract ID field name from options if provided
		const idField = options?.idField;
		const customIdField = getIdFieldName(idField);
		
		// If baseFields contains an 'id' field, we need to rename it to the custom field name
		let fields: Record<string, string | number | null> = { ...baseFields };
		
		// If using custom ID field and different from "id", move/rename the id field
		if (customIdField !== "id" && "id" in fields) {
			// Move 'id' to custom field name
			const idValue = fields.id;
			delete fields.id;
			fields = {
				[customIdField]: idValue,
				...fields,
			};
		}

		// Add extra fields if provided
		if (extraFields) {
			// Ensure we don't add 'id' back if using custom field
			const sanitizedExtraFields = { ...extraFields };
			if (customIdField !== "id" && "id" in sanitizedExtraFields) {
				delete sanitizedExtraFields.id;
			}
			Object.assign(fields, sanitizedExtraFields);
		}
		
		// Final cleanup: remove any leftover 'id' field if we're using custom field
		if (customIdField !== "id" && "id" in fields) {
			delete fields.id;
		}

		// Generate tags (without # prefix - Obsidian adds it automatically)
		const tags: string[] = [];
		if (options) {
			// Entity type tag
			tags.push(`story-engine/${options.entityType}`);

			// Story name tag (sanitized)
			if (options.storyName) {
				const sanitizedStoryName = this.sanitizeForTag(options.storyName);
				tags.push(`story/${sanitizedStoryName}`);
			}

			// World name tag (sanitized)
			if (options.worldName) {
				const sanitizedWorldName = this.sanitizeForTag(options.worldName);
				tags.push(`world/${sanitizedWorldName}`);
			}

			// Date tag in format YYYY/MM/DD
			if (options.date) {
				const date = typeof options.date === "string" ? new Date(options.date) : options.date;
				if (!isNaN(date.getTime())) {
					const year = date.getFullYear();
					const month = String(date.getMonth() + 1).padStart(2, "0");
					const day = String(date.getDate()).padStart(2, "0");
					tags.push(`date/${year}/${month}/${day}`);
				}
			}
		}

		// Build frontmatter lines
		const lines: string[] = ["---"];

		// Add all fields
		for (const [key, value] of Object.entries(fields)) {
			if (value === null || value === undefined) {
				lines.push(`${key}: null`);
			} else if (typeof value === "string") {
				// Escape quotes and wrap in quotes if contains special chars
				const escaped = value.replace(/"/g, '\\"');
				if (value.includes(":") || value.includes("\n") || value.includes('"')) {
					lines.push(`${key}: "${escaped}"`);
				} else {
					lines.push(`${key}: ${escaped}`);
				}
			} else {
				lines.push(`${key}: ${value}`);
			}
		}

		// Add tags if any
		if (tags.length > 0) {
			lines.push(`tags:`);
			for (const tag of tags) {
				lines.push(`  - ${tag}`);
			}
		}

		lines.push("---");
		return lines.join("\n");
	}

	private sanitizeForTag(value: string): string {
		return value
			.toLowerCase()
			.normalize("NFKD")
			.replace(/[^\w\s-]/g, "")
			.trim()
			.replace(/\s+/g, "-")
			.replace(/-+/g, "-")
			.replace(/^-|-$/g, "");
	}
}

