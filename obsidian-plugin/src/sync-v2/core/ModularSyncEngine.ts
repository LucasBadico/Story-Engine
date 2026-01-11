import { Notice } from "obsidian";
import type { SyncEngine } from "../../sync/engine";
import type { SyncEntityTarget } from "../../sync/entitySyncTypes";
import { SyncOrchestrator } from "./SyncOrchestrator";
import type { SyncContext, SyncOperationType, SyncResult } from "../types/sync";
import { ApiUpdateNotifierHandler } from "../apiUpdateNotifier/ApiUpdateNotifierHandler";
import { apiUpdateNotifierV2 } from "../apiUpdateNotifier/apiUpdateNotifier";

export class ModularSyncEngine implements SyncEngine {
	private readonly orchestrator: SyncOrchestrator;
	private readonly context: SyncContext;
	private apiUpdateNotifierHandler?: ApiUpdateNotifierHandler;

	constructor(context: SyncContext) {
		this.context = context;
		this.orchestrator = new SyncOrchestrator(context);
		this.initializeApiUpdateNotifier();
	}

	getOrchestrator(): SyncOrchestrator {
		return this.orchestrator;
	}

	getContext(): SyncContext {
		return this.context;
	}

	/**
	 * Initialize API update notifier handler
	 */
	private initializeApiUpdateNotifier(): void {
		// Only enable if autoSyncOnApiUpdates is enabled
		if (!this.context.settings.autoSyncOnApiUpdates) {
			return;
		}

		this.apiUpdateNotifierHandler = new ApiUpdateNotifierHandler(
			apiUpdateNotifierV2,
			this.orchestrator,
			this.context
		);
		this.apiUpdateNotifierHandler.start();
	}

	dispose(): void {
		if (this.apiUpdateNotifierHandler) {
			this.apiUpdateNotifierHandler.dispose();
			this.apiUpdateNotifierHandler = undefined;
		}
		this.orchestrator.dispose();
	}

	async pullStory(storyId: string, target?: SyncEntityTarget): Promise<void> {
		const result = await this.orchestrator.run({
			type: "pull_story",
			payload: { storyId, target },
		});
		this.handleResult("pull_story", result);
	}

	async pullAllStories(): Promise<void> {
		const result = await this.orchestrator.run({
			type: "pull_all_stories",
			payload: {},
		});
		this.handleResult("pull_all_stories", result);
	}

	async pushStory(folderPath: string, target?: SyncEntityTarget): Promise<void> {
		const result = await this.orchestrator.run({
			type: "push_story",
			payload: { folderPath, target },
		});
		this.handleResult("push_story", result);
	}

	private handleResult(operation: SyncOperationType, result: SyncResult): void {
		if (result.success) {
			if (result.warnings?.length) {
				const summary =
					result.warnings.length === 1
						? result.warnings[0].message
						: `${result.warnings.length} avisos detectados; revise o log.`;
				console.warn(`[Sync V2] warnings for ${operation}`, result.warnings);
				new Notice(`Sync V2 (${operation}) aviso: ${summary}`, 6000);
			}
			if (result.message) {
				new Notice(result.message, 3000);
			}
			return;
		}

		const firstError = result.errors?.[0];
		const message =
			firstError?.message ??
			result.message ??
			`Sync V2 operation "${operation}" failed without error details.`;

		new Notice(`Sync V2 (${operation}): ${message}`, 8000);
		if (result.warnings?.length) {
			console.warn(`[Sync V2] warnings for ${operation}`, result.warnings);
		}
		if (result.errors?.length) {
			console.warn("[Sync V2] operation failed", {
				operation,
				errors: result.errors,
			});
		}
	}
}

