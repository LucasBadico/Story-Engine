import type { App } from "obsidian";
import type { StoryWithHierarchy } from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import { ConflictResolver } from "../../../conflict/ConflictResolver";
import { parseFrontmatter } from "../../../utils/detectEntityMentions";

export interface ConflictCheckOptions {
	localContentPath?: string;
	remoteContent?: string;
	entityType?: string;
}

/**
 * Service responsible for detecting and resolving conflicts during story sync.
 * Extracted from StoryHandler for better separation of concerns.
 */
export class StoryConflictService {
	constructor(
		private readonly resolverFactory: (app: App, context: SyncContext) => ConflictResolver = 
			(app, ctx) => new ConflictResolver(app, ctx)
	) {}

	/**
	 * Check for conflicts between local and remote versions of a file.
	 * Emits warnings through context.emitWarning when conflicts are detected.
	 */
	async checkConflicts(
		path: string,
		story: StoryWithHierarchy["story"],
		context: SyncContext,
		opts?: ConflictCheckOptions
	): Promise<void> {
		const local = opts?.localContentPath
			? await this.readFileSilently(context, opts.localContentPath)
			: await this.readFileSilently(context, path);

		if (!local) {
			// No existing file, no conflict possible
			return;
		}

		const parsed = parseFrontmatter(local);
		const localTimestamp =
			(parsed.updated_at as string | undefined) ?? (parsed.synced_at as string | undefined);

		if (!localTimestamp) {
			// No timestamp in file, can't detect conflict
			return;
		}

		const remoteTimestamp = story.updated_at;
		const entityType = opts?.entityType ?? "story";
		const resolver = this.resolverFactory(context.app, context);

		const conflict = resolver.detectConflict(
			entityType,
			story.id,
			path,
			{ updated_at: localTimestamp },
			{ updated_at: remoteTimestamp },
			localTimestamp,
			remoteTimestamp
		);

		if (!conflict) {
			return;
		}

		const resolution = await resolver.resolve(conflict);

		if (!resolution.success) {
			context.emitWarning?.({
				code: "conflict_resolution_failed",
				message: `Failed to resolve conflict for ${entityType.replace("-", " ")}: ${resolution.error}`,
				filePath: path,
				severity: "warning",
			});
			return;
		}

		if (!resolution.resolution.autoResolved && resolution.resolution.strategy === "manual") {
			context.emitWarning?.({
				code: "conflict_requires_manual_resolution",
				message: `Conflict detected for ${entityType.replace("-", " ")}. Manual resolution required.`,
				filePath: path,
				severity: "warning",
				details: conflict,
			});
		}
	}

	private async readFileSilently(context: SyncContext, path: string): Promise<string | null> {
		try {
			return await context.fileManager.readFile(path);
		} catch {
			return null;
		}
	}
}
