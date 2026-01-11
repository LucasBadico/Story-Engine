import type { SyncEntityPayload } from "../../sync/entitySyncTypes";
import type { ApiUpdateEvent } from "./ApiUpdateNotifierV2";

/**
 * Adapter to convert V1 SyncEntityPayload to V2 ApiUpdateEvent
 * 
 * This allows the StoryEngineClient (which uses V1 payloads) to
 * notify the V2 ApiUpdateNotifierV2 system.
 */
export function adaptPayloadToEvent(payload: SyncEntityPayload): ApiUpdateEvent {
	const timestamp = new Date().toISOString();

	switch (payload.type) {
		case "chapter":
			return {
				type: "chapter.updated",
				entityId: payload.chapter.id,
				entityType: "chapter",
				storyId: payload.story.id,
				timestamp,
			};

		case "scene":
			return {
				type: "scene.updated",
				entityId: payload.scene.id,
				entityType: "scene",
				storyId: payload.story.id,
				timestamp,
			};

		case "content":
			return {
				type: "content_block.updated",
				entityId: payload.contentBlock.id,
				entityType: "content_block",
				storyId: payload.story.id,
				timestamp,
			};

		default:
			// TypeScript exhaustiveness check
			const _exhaustive: never = payload;
			throw new Error(`Unsupported payload type: ${(_exhaustive as any).type}`);
	}
}

