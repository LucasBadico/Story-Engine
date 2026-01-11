import type { SyncOrchestrator } from "../core/SyncOrchestrator";
import type { SyncContext } from "../types/sync";
import { ApiUpdateNotifierV2, type ApiUpdateEvent } from "./ApiUpdateNotifierV2";
import { Notice } from "obsidian";

/**
 * ApiUpdateNotifierHandler - Handles API update notifications for Sync V2
 * 
 * This handler:
 * - Subscribes to ApiUpdateNotifierV2 events
 * - Processes events and pulls updated entities using SyncOrchestrator
 * - Handles conflicts when local and remote change simultaneously
 * - Notifies users about remote changes
 */
export class ApiUpdateNotifierHandler {
	private unsubscribe?: () => void;
	private isProcessing = false;
	private processingQueue: ApiUpdateEvent[] = [];

	constructor(
		private notifier: ApiUpdateNotifierV2,
		private orchestrator: SyncOrchestrator,
		private context: SyncContext
	) {}

	/**
	 * Start listening to API update events
	 */
	start(): void {
		if (this.unsubscribe) {
			// Already subscribed
			return;
		}

		this.unsubscribe = this.notifier.subscribe(async (event) => {
			await this.handleEvent(event);
		});
	}

	/**
	 * Stop listening to API update events
	 */
	stop(): void {
		if (this.unsubscribe) {
			this.unsubscribe();
			this.unsubscribe = undefined;
		}
		this.processingQueue = [];
		this.isProcessing = false;
	}

	/**
	 * Handle an API update event
	 */
	private async handleEvent(event: ApiUpdateEvent): Promise<void> {
		// Add to queue
		this.processingQueue.push(event);

		// Process queue if not already processing
		if (!this.isProcessing) {
			await this.processQueue();
		}
	}

	/**
	 * Process queued events
	 */
	private async processQueue(): Promise<void> {
		if (this.processingQueue.length === 0) {
			this.isProcessing = false;
			return;
		}

		this.isProcessing = true;

		// Process events one by one
		while (this.processingQueue.length > 0) {
			const event = this.processingQueue.shift();
			if (!event) {
				break;
			}

			try {
				await this.processEvent(event);
			} catch (err) {
				console.error(`Failed to process API update event ${event.type} for ${event.entityType}:${event.entityId}`, err);
				this.context.emitWarning?.({
					code: "api_update_failed",
					message: `Failed to sync ${event.entityType} ${event.entityId} from API update`,
					severity: "warning",
					details: { event, error: err },
				});
			}
		}

		this.isProcessing = false;
	}

	/**
	 * Process a single API update event
	 */
	private async processEvent(event: ApiUpdateEvent): Promise<void> {
		const [entityType, action] = event.type.split(".");

		// Handle deletions separately
		if (action === "deleted") {
			await this.handleDeletion(event);
			return;
		}

		// Map event type to sync operation
		const operation = this.mapEventToOperation(event);
		if (!operation) {
			// Event type not supported
			return;
		}

		// Execute pull operation
		const result = await this.orchestrator.run(operation);

		// Handle warnings
		if (result.warnings && result.warnings.length > 0) {
			for (const warning of result.warnings) {
				this.context.emitWarning?.(warning);
			}
		}

		// Notify user about successful update
		if (result.success) {
			const entityName = this.getEntityDisplayName(event);
			new Notice(`${entityName} synced from API`);
		}
	}

	/**
	 * Handle deletion events
	 * 
	 * Note: Full deletion support requires:
	 * 1. A way to map entity IDs to file paths (e.g., EntityRegistry with file paths)
	 * 2. Deletion of the entity file and potentially related files
	 * 3. Cleanup of parent entity files (e.g., if a chapter is deleted, update story.md)
	 * 
	 * For now, we just log the deletion and emit a warning.
	 */
	private async handleDeletion(event: ApiUpdateEvent): Promise<void> {
		console.log(`Entity deleted: ${event.entityType}:${event.entityId}`);
		
		// TODO: Implement full deletion support
		// This would require:
		// 1. Finding the file(s) corresponding to the entity
		// 2. Deleting the file(s) using context.app.vault.delete()
		// 3. Updating parent files if needed (e.g., story.md, chapter files)
		
		this.context.emitWarning?.({
			code: "entity_deleted",
			message: `${event.entityType} ${event.entityId} was deleted on the server. Local file deletion is not yet implemented.`,
			severity: "warning",
			details: { event },
		});
	}

	/**
	 * Map API update event to sync operation
	 */
	private mapEventToOperation(event: ApiUpdateEvent): import("../types/sync").SyncOperation | null {
		const [entityType] = event.type.split(".");

		// Deletions are handled separately in handleDeletion()
		// This method only handles update/create operations

		switch (entityType) {
			case "story":
				return { type: "pull_story", payload: { storyId: event.entityId } };
			case "chapter":
				return { type: "pull_chapter", payload: { chapterId: event.entityId } };
			case "scene":
				return { type: "pull_scene", payload: { sceneId: event.entityId } };
			case "beat":
				return { type: "pull_beat", payload: { beatId: event.entityId } };
			case "content_block":
				return { type: "pull_content_block", payload: { contentBlockId: event.entityId } };
			case "character":
				return { type: "pull_character", payload: { entityId: event.entityId } };
			case "location":
				return { type: "pull_location", payload: { entityId: event.entityId } };
			case "faction":
				return { type: "pull_faction", payload: { entityId: event.entityId } };
			case "artifact":
				return { type: "pull_artifact", payload: { entityId: event.entityId } };
			case "event":
				return { type: "pull_event", payload: { entityId: event.entityId } };
			case "lore":
				return { type: "pull_lore", payload: { entityId: event.entityId } };
			case "archetype":
				return { type: "pull_archetype", payload: { entityId: event.entityId } };
			case "trait":
				return { type: "pull_trait", payload: { entityId: event.entityId } };
			case "relation":
				// Relations are handled differently - we need to pull the source entity
				// TODO: Handle relations separately
				return null;
			default:
				return null;
		}
	}

	/**
	 * Get display name for entity
	 */
	private getEntityDisplayName(event: ApiUpdateEvent): string {
		const entityTypeName = event.entityType.replace("_", " ");
		return `${entityTypeName} ${event.entityId}`;
	}

	dispose(): void {
		this.stop();
	}
}

