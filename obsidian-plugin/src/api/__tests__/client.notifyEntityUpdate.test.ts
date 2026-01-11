import { describe, expect, it, vi, beforeEach } from "vitest";
import { StoryEngineClient } from "../client";
import { apiUpdateNotifier } from "../../sync/apiUpdateNotifier";
import { apiUpdateNotifierV2 } from "../../sync-v2/apiUpdateNotifier/apiUpdateNotifier";
import type {
	SyncEntityPayload,
	ChapterEntityPayload,
	SceneEntityPayload,
	ContentEntityPayload,
} from "../../sync/entitySyncTypes";
import type { Story, Chapter, Scene, ContentBlock } from "../../types";

// Mock the notifiers
vi.mock("../../sync/apiUpdateNotifier", () => ({
	apiUpdateNotifier: {
		notify: vi.fn(),
	},
}));

vi.mock("../../sync-v2/apiUpdateNotifier/apiUpdateNotifier", () => ({
	apiUpdateNotifierV2: {
		notify: vi.fn(),
	},
}));

describe("StoryEngineClient.notifyEntityUpdate", () => {
	let client: StoryEngineClient;

	beforeEach(() => {
		vi.clearAllMocks();
		client = new StoryEngineClient("http://localhost:8080", "test-key", "test-tenant");
	});

	describe("V1 notification", () => {
		it("should notify V1 notifier for chapter payload", async () => {
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

			await client.notifyEntityUpdate(payload);

			expect(apiUpdateNotifier.notify).toHaveBeenCalledWith(payload);
		});

		it("should notify V1 notifier for scene payload", async () => {
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

			await client.notifyEntityUpdate(payload);

			expect(apiUpdateNotifier.notify).toHaveBeenCalledWith(payload);
		});

		it("should notify V1 notifier for content payload", async () => {
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

			await client.notifyEntityUpdate(payload);

			expect(apiUpdateNotifier.notify).toHaveBeenCalledWith(payload);
		});
	});

	describe("V2 notification", () => {
		it("should notify V2 notifier for chapter payload", async () => {
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

			await client.notifyEntityUpdate(payload);

			expect(apiUpdateNotifierV2.notify).toHaveBeenCalledWith(
				expect.objectContaining({
					type: "chapter.updated",
					entityId: "chapter-1",
					entityType: "chapter",
					storyId: "story-1",
				})
			);
		});

		it("should notify V2 notifier for scene payload", async () => {
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

			await client.notifyEntityUpdate(payload);

			expect(apiUpdateNotifierV2.notify).toHaveBeenCalledWith(
				expect.objectContaining({
					type: "scene.updated",
					entityId: "scene-1",
					entityType: "scene",
					storyId: "story-1",
				})
			);
		});

		it("should notify V2 notifier for content payload", async () => {
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

			await client.notifyEntityUpdate(payload);

			expect(apiUpdateNotifierV2.notify).toHaveBeenCalledWith(
				expect.objectContaining({
					type: "content_block.updated",
					entityId: "block-1",
					entityType: "content_block",
					storyId: "story-1",
				})
			);
		});
	});

	describe("Error handling", () => {
		it("should still notify V1 even if V2 notification fails", async () => {
			const consoleWarnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
			(apiUpdateNotifierV2.notify as any).mockRejectedValue(new Error("V2 error"));

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

			// Should not throw
			await expect(client.notifyEntityUpdate(payload)).resolves.not.toThrow();

			// V1 should still be called
			expect(apiUpdateNotifier.notify).toHaveBeenCalledWith(payload);

			// V2 should be attempted but fail gracefully
			expect(apiUpdateNotifierV2.notify).toHaveBeenCalled();
			expect(consoleWarnSpy).toHaveBeenCalledWith(
				"Failed to notify V2 notifier",
				expect.any(Error)
			);

			consoleWarnSpy.mockRestore();
		});
	});
});

