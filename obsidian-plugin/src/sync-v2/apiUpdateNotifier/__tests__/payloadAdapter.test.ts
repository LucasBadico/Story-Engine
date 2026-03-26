import { describe, expect, it } from "vitest";
import type {
	SyncEntityPayload,
	ChapterEntityPayload,
	SceneEntityPayload,
	ContentEntityPayload,
} from "../../../sync/entitySyncTypes";
import type { Story, Chapter, Scene, Beat, ContentBlock } from "../../../types";
import { adaptPayloadToEvent } from "../payloadAdapter";

describe("payloadAdapter", () => {
	describe("adaptPayloadToEvent", () => {
		it("should convert chapter payload to chapter.updated event", () => {
			const story: Story = {
				id: "story-1",
				tenant_id: "tenant-1",
				title: "Test Story",
				status: "draft",
				version_number: 1,
				root_story_id: "story-1",
				previous_story_id: null,
				world_id: null,
				created_by_user_id: "user-1",
				created_at: "2025-01-01T00:00:00Z",
				updated_at: "2025-01-01T00:00:00Z",
			};

			const chapter: Chapter = {
				id: "chapter-1",
				story_id: "story-1",
				title: "Test Chapter",
				number: 1,
				status: "draft",
				created_at: "2025-01-01T00:00:00Z",
				updated_at: "2025-01-01T00:00:00Z",
			};

			const payload: ChapterEntityPayload = {
				type: "chapter",
				story,
				chapter,
				scenes: [],
				contentBlocks: [],
				contentBlockRefs: [],
			};

			const event = adaptPayloadToEvent(payload);

			expect(event.type).toBe("chapter.updated");
			expect(event.entityId).toBe("chapter-1");
			expect(event.entityType).toBe("chapter");
			expect(event.storyId).toBe("story-1");
			expect(event.timestamp).toBeDefined();
			expect(new Date(event.timestamp).getTime()).toBeGreaterThan(0);
		});

		it("should convert scene payload to scene.updated event", () => {
			const story: Story = {
				id: "story-1",
				tenant_id: "tenant-1",
				title: "Test Story",
				status: "draft",
				version_number: 1,
				root_story_id: "story-1",
				previous_story_id: null,
				world_id: null,
				created_by_user_id: "user-1",
				created_at: "2025-01-01T00:00:00Z",
				updated_at: "2025-01-01T00:00:00Z",
			};

			const scene: Scene = {
				id: "scene-1",
				story_id: "story-1",
				chapter_id: "chapter-1",
				order_num: 1,
				goal: "Test Scene",
				time_ref: "2025-01-01T00:00:00Z",
				created_at: "2025-01-01T00:00:00Z",
				updated_at: "2025-01-01T00:00:00Z",
			};

			const payload: SceneEntityPayload = {
				type: "scene",
				story,
				scene,
				beats: [],
			};

			const event = adaptPayloadToEvent(payload);

			expect(event.type).toBe("scene.updated");
			expect(event.entityId).toBe("scene-1");
			expect(event.entityType).toBe("scene");
			expect(event.storyId).toBe("story-1");
			expect(event.timestamp).toBeDefined();
			expect(new Date(event.timestamp).getTime()).toBeGreaterThan(0);
		});

		it("should convert content payload to content_block.updated event", () => {
			const story: Story = {
				id: "story-1",
				tenant_id: "tenant-1",
				title: "Test Story",
				status: "draft",
				version_number: 1,
				root_story_id: "story-1",
				previous_story_id: null,
				world_id: null,
				created_by_user_id: "user-1",
				created_at: "2025-01-01T00:00:00Z",
				updated_at: "2025-01-01T00:00:00Z",
			};

			const contentBlock: ContentBlock = {
				id: "block-1",
				chapter_id: "chapter-1",
				type: "text",
				kind: "paragraph",
				content: "Test content",
				metadata: {},
				created_at: "2025-01-01T00:00:00Z",
				updated_at: "2025-01-01T00:00:00Z",
			};

			const payload: ContentEntityPayload = {
				type: "content",
				story,
				contentBlock,
			};

			const event = adaptPayloadToEvent(payload);

			expect(event.type).toBe("content_block.updated");
			expect(event.entityId).toBe("block-1");
			expect(event.entityType).toBe("content_block");
			expect(event.storyId).toBe("story-1");
			expect(event.timestamp).toBeDefined();
			expect(new Date(event.timestamp).getTime()).toBeGreaterThan(0);
		});

		it("should throw error for unsupported payload type", () => {
			const invalidPayload = {
				type: "invalid",
			} as unknown as SyncEntityPayload;

			expect(() => adaptPayloadToEvent(invalidPayload)).toThrow(
				"Unsupported payload type"
			);
		});
	});
});

