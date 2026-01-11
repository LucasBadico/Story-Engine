import type { SyncOrchestrator } from "../core/SyncOrchestrator";
import type { SyncContext } from "../types/sync";

export type ApiUpdateEventType = 
	| "story.created"
	| "story.updated"
	| "story.deleted"
	| "chapter.created"
	| "chapter.updated"
	| "chapter.deleted"
	| "scene.created"
	| "scene.updated"
	| "scene.deleted"
	| "beat.created"
	| "beat.updated"
	| "beat.deleted"
	| "content_block.created"
	| "content_block.updated"
	| "content_block.deleted"
	| "relation.created"
	| "relation.updated"
	| "relation.deleted"
	| "character.created"
	| "character.updated"
	| "character.deleted"
	| "location.created"
	| "location.updated"
	| "location.deleted"
	| "faction.created"
	| "faction.updated"
	| "faction.deleted"
	| "artifact.created"
	| "artifact.updated"
	| "artifact.deleted"
	| "event.created"
	| "event.updated"
	| "event.deleted"
	| "lore.created"
	| "lore.updated"
	| "lore.deleted";

export interface ApiUpdateEvent {
	type: ApiUpdateEventType;
	entityId: string;
	entityType: string;
	storyId?: string; // For story entities
	worldId?: string; // For world entities
	timestamp: string;
}

type EventSubscriber = (event: ApiUpdateEvent) => void | Promise<void>;

interface QueuedEvent {
	event: ApiUpdateEvent;
	timestamp: number;
}

/**
 * ApiUpdateNotifierV2 - Handles API update notifications for Sync V2
 * 
 * Features:
 * - Pub/sub system for API update events
 * - Debouncing to avoid flooding
 * - Integration with SyncOrchestrator to pull updated entities
 */
export class ApiUpdateNotifierV2 {
	private subscribers = new Set<EventSubscriber>();
	private eventQueue: QueuedEvent[] = [];
	private processingTimeoutId: number | null = null;
	private readonly DEBOUNCE_MS = 1000; // 1 second debounce
	private readonly MAX_QUEUE_SIZE = 100;

	/**
	 * Subscribe to API update events
	 * @param subscriber Callback function to handle events
	 * @returns Unsubscribe function
	 */
	subscribe(subscriber: EventSubscriber): () => void {
		this.subscribers.add(subscriber);
		return () => {
			this.subscribers.delete(subscriber);
		};
	}

	/**
	 * Notify subscribers of an API update event
	 * Events are debounced to avoid flooding
	 */
	async notify(event: ApiUpdateEvent): Promise<void> {
		// Add to queue
		this.eventQueue.push({
			event,
			timestamp: Date.now(),
		});

		// Limit queue size
		if (this.eventQueue.length > this.MAX_QUEUE_SIZE) {
			this.eventQueue.shift(); // Remove oldest
		}

		// Debounce processing
		if (this.processingTimeoutId !== null) {
			const clearFn = typeof window !== "undefined" ? window.clearTimeout : globalThis.clearTimeout;
			clearFn(this.processingTimeoutId);
		}

		const setTimeoutFn = typeof window !== "undefined" ? window.setTimeout : globalThis.setTimeout;
		this.processingTimeoutId = setTimeoutFn(() => {
			this.processQueue();
		}, this.DEBOUNCE_MS) as unknown as number;
	}

	/**
	 * Process queued events and notify subscribers
	 */
	private async processQueue(): Promise<void> {
		if (this.eventQueue.length === 0) {
			return;
		}

		// Get all events from queue
		const events = this.eventQueue.splice(0);

		// Deduplicate events (keep only the latest for each entity)
		const latestEvents = new Map<string, ApiUpdateEvent>();
		for (const queuedEvent of events) {
			const key = `${queuedEvent.event.entityType}:${queuedEvent.event.entityId}`;
			const existing = latestEvents.get(key);
			if (!existing || queuedEvent.timestamp > (latestEvents.get(key)?.timestamp ? Date.parse(existing.timestamp) : 0)) {
				latestEvents.set(key, queuedEvent.event);
			}
		}

		// Notify subscribers
		for (const event of latestEvents.values()) {
			for (const subscriber of this.subscribers) {
				try {
					await subscriber(event);
				} catch (err) {
					console.error("ApiUpdateNotifier subscriber failed", err);
				}
			}
		}

		this.processingTimeoutId = null;
	}

	/**
	 * Clear all queued events
	 */
	clear(): void {
		this.eventQueue = [];
		if (this.processingTimeoutId !== null) {
			const clearFn = typeof window !== "undefined" ? window.clearTimeout : globalThis.clearTimeout;
			clearFn(this.processingTimeoutId);
			this.processingTimeoutId = null;
		}
	}

	dispose(): void {
		this.clear();
		this.subscribers.clear();
	}
}

