import type { App } from "obsidian";
import type { SyncContext, SyncWarning } from "../types/sync";
import type {
	Conflict,
	ConflictResolution,
	ConflictResolutionResult,
	ConflictResolutionStrategy,
	ConflictType,
} from "./types";
import { ConflictModal } from "./ConflictModal";

/**
 * ConflictResolver handles conflict detection and resolution for Sync V2
 * 
 * Supports multiple resolution strategies:
 * - "manual": User chooses via UI modal
 * - "local": Always use local version
 * - "remote": Always use remote version
 * - "last_write_wins": Use version with latest timestamp
 * - "merge": Attempt intelligent merge (text content only)
 */
export class ConflictResolver {
	constructor(
		private app: App,
		private context: SyncContext
	) {}

	/**
	 * Detect if there's a conflict between local and remote versions
	 * 
	 * @param entityType Type of entity (story, chapter, scene, etc.)
	 * @param entityId ID of the entity
	 * @param localData Local version (from file)
	 * @param remoteData Remote version (from API)
	 * @param localTimestamp Local modification timestamp (if available)
	 * @param remoteTimestamp Remote modification timestamp (from API)
	 * @returns Conflict object if conflict detected, null otherwise
	 */
	detectConflict(
		entityType: string,
		entityId: string,
		filePath: string,
		localData: unknown,
		remoteData: unknown,
		localTimestamp?: string,
		remoteTimestamp?: string
	): Conflict | null {
		// Check for simultaneous edit conflict
		if (this.isSimultaneousEdit(localData, remoteData, localTimestamp, remoteTimestamp)) {
			return {
				type: "simultaneous_edit",
				entityId,
				entityType,
				filePath,
				localTimestamp,
				remoteTimestamp,
				localData,
				remoteData,
				context: {
					message: "Entity was modified both locally and remotely",
				},
			};
		}

		// Additional conflict types can be detected here:
		// - Rename conflict (file path changed but entity exists with different path)
		// - Deletion conflict (entity deleted remotely but exists locally)
		// - File exists conflict (trying to create entity but file already exists)

		return null;
	}

	/**
	 * Resolve a conflict using the configured strategy
	 * 
	 * @param conflict Conflict to resolve
	 * @returns Resolution result
	 */
	async resolve(conflict: Conflict): Promise<ConflictResolutionResult> {
		const strategy = this.getResolutionStrategy();

		try {
			let resolution: ConflictResolution;

			switch (strategy) {
				case "local":
					resolution = this.resolveLocal(conflict);
					break;
				case "remote":
					resolution = this.resolveRemote(conflict);
					break;
				case "last_write_wins":
					resolution = this.resolveLastWriteWins(conflict);
					break;
				case "merge":
					resolution = await this.resolveMerge(conflict);
					break;
				case "manual":
				default:
					resolution = await this.resolveManual(conflict);
					break;
			}

			return {
				success: true,
				resolution,
			};
		} catch (error) {
			return {
				success: false,
				resolution: {
					strategy,
					autoResolved: false,
				},
				error: error instanceof Error ? error.message : String(error),
			};
		}
	}

	/**
	 * Get the resolution strategy from settings
	 */
	private getResolutionStrategy(): ConflictResolutionStrategy {
		const setting = this.context.settings.conflictResolution;

		// Map settings values to strategy
		switch (setting) {
			case "local":
				return "local";
			case "manual":
				return "manual";
			case "service":
				// "service" means use remote (service wins)
				return "remote";
			default:
				return "manual";
		}
	}

	/**
	 * Check if there's a simultaneous edit conflict
	 */
	private isSimultaneousEdit(
		localData: unknown,
		remoteData: unknown,
		localTimestamp?: string,
		remoteTimestamp?: string
	): boolean {
		// If timestamps are available, check if both were modified
		if (localTimestamp && remoteTimestamp) {
			// If remote timestamp is newer, but local data is different, it's a conflict
			const localTs = new Date(localTimestamp).getTime();
			const remoteTs = new Date(remoteTimestamp).getTime();

			// If remote is newer but local was also modified (different content), conflict
			if (remoteTs > localTs && !this.isDataEqual(localData, remoteData)) {
				return true;
			}
		}

		// If no timestamps, check if data is different (simpler heuristic)
		if (!localTimestamp && !remoteTimestamp) {
			return !this.isDataEqual(localData, remoteData);
		}

		return false;
	}

	/**
	 * Check if two data objects are equal (deep comparison for objects, simple for primitives)
	 */
	private isDataEqual(a: unknown, b: unknown): boolean {
		if (a === b) {
			return true;
		}

		if (typeof a !== typeof b) {
			return false;
		}

		if (typeof a === "object" && a !== null && b !== null) {
			const aObj = a as Record<string, unknown>;
			const bObj = b as Record<string, unknown>;

			const aKeys = Object.keys(aObj);
			const bKeys = Object.keys(bObj);

			if (aKeys.length !== bKeys.length) {
				return false;
			}

			for (const key of aKeys) {
				if (!bKeys.includes(key)) {
					return false;
				}
				if (!this.isDataEqual(aObj[key], bObj[key])) {
					return false;
				}
			}

			return true;
		}

		return false;
	}

	/**
	 * Resolve conflict by using local version
	 */
	private resolveLocal(conflict: Conflict): ConflictResolution {
		return {
			strategy: "local",
			resolvedData: conflict.localData,
			autoResolved: true,
		};
	}

	/**
	 * Resolve conflict by using remote version
	 */
	private resolveRemote(conflict: Conflict): ConflictResolution {
		return {
			strategy: "remote",
			resolvedData: conflict.remoteData,
			autoResolved: true,
		};
	}

	/**
	 * Resolve conflict using last-write-wins strategy (latest timestamp wins)
	 */
	private resolveLastWriteWins(conflict: Conflict): ConflictResolution {
		if (conflict.localTimestamp && conflict.remoteTimestamp) {
			const localTs = new Date(conflict.localTimestamp).getTime();
			const remoteTs = new Date(conflict.remoteTimestamp).getTime();

			if (remoteTs >= localTs) {
				return {
					strategy: "last_write_wins",
					resolvedData: conflict.remoteData,
					autoResolved: true,
				};
			}
		}

		// Default to local if timestamps are not available or local is newer
		return {
			strategy: "last_write_wins",
			resolvedData: conflict.localData,
			autoResolved: true,
		};
	}

	/**
	 * Resolve conflict using merge strategy (attempt intelligent merge)
	 * For now, falls back to manual resolution
	 */
	private async resolveMerge(conflict: Conflict): Promise<ConflictResolution> {
		// TODO: Implement intelligent merge for text content
		// For now, fall back to manual resolution
		return this.resolveManual(conflict);
	}

	/**
	 * Resolve conflict manually (show UI modal to user)
	 */
	private async resolveManual(conflict: Conflict): Promise<ConflictResolution> {
		const choice = await new ConflictModal(this.app, conflict).open();
		const resolvedData = choice === "remote" ? conflict.remoteData : conflict.localData;

		this.context.emitWarning?.({
			code: "conflict_detected",
			message: `Conflict detected for ${conflict.entityType} ${conflict.entityId}. Manual resolution selected (${choice}).`,
			filePath: conflict.filePath,
			severity: "warning",
		});

		return {
			strategy: "manual",
			resolvedData,
			autoResolved: false,
		};
	}
}

