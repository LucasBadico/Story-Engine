/**
 * Helper utilities for working with frontmatter, especially for reading/writing IDs
 * with customizable field names to avoid conflicts with other plugins.
 */

/**
 * Get the ID field name from settings, defaulting to "id"
 */
export function getIdFieldName(idField?: string): string {
	return idField || "id";
}

/**
 * Read entity ID from frontmatter using the configured field name
 * @param frontmatter - The parsed frontmatter object
 * @param idField - Custom ID field name (from settings), defaults to "id"
 * @returns The entity ID as string, or undefined if not found
 */
export function getFrontmatterId(
	frontmatter: Record<string, unknown>,
	idField?: string
): string | undefined {
	const fieldName = getIdFieldName(idField);
	const value = frontmatter[fieldName];
	
	// Handle null/undefined/missing values first
	if (value === null || value === undefined) {
		return undefined;
	}
	
	// Handle string type and check if it's empty
	if (typeof value === "string") {
		// Empty string or whitespace-only string should return undefined
		const trimmed = value.trim();
		return trimmed === "" ? undefined : value;
	}
	
	// Convert to string if it's a number or other type
	// Note: empty string was already handled above as typeof "string"
	return String(value);
}

/**
 * Set entity ID in frontmatter object using the configured field name
 * @param frontmatter - The frontmatter object to update
 * @param id - The entity ID to set
 * @param idField - Custom ID field name (from settings), defaults to "id"
 */
export function setFrontmatterId(
	frontmatter: Record<string, unknown>,
	id: string,
	idField?: string
): void {
	const fieldName = getIdFieldName(idField);
	frontmatter[fieldName] = id;
}

/**
 * Generate frontmatter fields object with ID field using custom field name
 * This is a helper for FrontmatterGenerator to ensure the ID field uses the correct name
 * @param id - The entity ID
 * @param otherFields - Other fields to include
 * @param idField - Custom ID field name (from settings), defaults to "id"
 * @returns Object with fields including the ID field with correct name
 */
export function buildFrontmatterFields(
	id: string,
	otherFields: Record<string, string | number | null> = {},
	idField?: string
): Record<string, string | number | null> {
	const fieldName = getIdFieldName(idField);
	return {
		[fieldName]: id,
		...otherFields,
	};
}

