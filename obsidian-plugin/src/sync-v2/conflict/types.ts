/**
 * Types for conflict detection and resolution in Sync V2
 */

export type ConflictType = "simultaneous_edit" | "rename" | "deletion" | "file_exists";

export type ConflictResolutionStrategy = "manual" | "local" | "remote" | "last_write_wins" | "merge";

export interface Conflict {
	/**
	 * Type of conflict detected
	 */
	type: ConflictType;

	/**
	 * Entity ID involved in the conflict
	 */
	entityId: string;

	/**
	 * Entity type (story, chapter, scene, beat, content_block, character, etc.)
	 */
	entityType: string;

	/**
	 * File path of the conflicted entity
	 */
	filePath: string;

	/**
	 * Local version timestamp (if available)
	 */
	localTimestamp?: string;

	/**
	 * Remote version timestamp (if available)
	 */
	remoteTimestamp?: string;

	/**
	 * Local content/metadata (entity data or file content)
	 */
	localData: unknown;

	/**
	 * Remote content/metadata (entity data from API)
	 */
	remoteData: unknown;

	/**
	 * Additional context about the conflict
	 */
	context?: {
		message?: string;
		details?: Record<string, unknown>;
	};
}

export interface ConflictResolution {
	/**
	 * Selected resolution strategy
	 */
	strategy: ConflictResolutionStrategy;

	/**
	 * Resolved data (merged content, local data, or remote data)
	 */
	resolvedData?: unknown;

	/**
	 * Whether the conflict was automatically resolved
	 */
	autoResolved: boolean;
}

export interface ConflictResolutionResult {
	/**
	 * Whether the conflict was resolved successfully
	 */
	success: boolean;

	/**
	 * Resolution details
	 */
	resolution: ConflictResolution;

	/**
	 * Error message if resolution failed
	 */
	error?: string;
}

