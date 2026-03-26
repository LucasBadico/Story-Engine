import { describe, expect, it, vi } from "vitest";
import type { StoryEngineClient } from "../../../api/client";
import { PushExecutor } from "../PushExecutor";
import type { PushAction } from "../PushPlanner";
import type { ContentCitationService } from "../ContentCitationService";

const createApiClient = () => ({
	updateChapter: vi.fn(),
	updateScene: vi.fn(),
	moveScene: vi.fn(),
	updateBeat: vi.fn(),
	moveBeat: vi.fn(),
	updateContentBlock: vi.fn(),
});

describe("PushExecutor", () => {
	it("applies reorder and move actions", async () => {
		const apiClient = createApiClient();
		const citationService = { syncCitations: vi.fn() } as unknown as ContentCitationService;
		const executor = new PushExecutor(apiClient as unknown as StoryEngineClient, citationService);

		const actions: PushAction[] = [
			{ type: "chapter_reorder", chapterId: "ch-1", newOrder: 2 },
			{ type: "scene_reorder", sceneId: "sc-1", newOrder: 3 },
			{ type: "scene_move", sceneId: "sc-2", toChapterId: "ch-2" },
			{ type: "beat_reorder", beatId: "bt-1", newOrder: 4 },
			{ type: "beat_move", beatId: "bt-2", toSceneId: "sc-3" },
			{ type: "content_update", contentBlockId: "cb-1", newContent: "Updated content" },
		];

		const summary = await executor.execute(actions);

		expect(summary.applied).toBe(actions.length);
		expect(summary.errors).toHaveLength(0);
		expect(apiClient.updateChapter).toHaveBeenCalledWith("ch-1", { number: 2 });
		expect(apiClient.updateScene).toHaveBeenCalledWith("sc-1", { order_num: 3 });
		expect(apiClient.moveScene).toHaveBeenCalledWith("sc-2", "ch-2");
		expect(apiClient.updateBeat).toHaveBeenCalledWith("bt-1", { order_num: 4 });
		expect(apiClient.moveBeat).toHaveBeenCalledWith("bt-2", "sc-3");
		expect(apiClient.updateContentBlock).toHaveBeenCalledWith("cb-1", {
			content: "Updated content",
		});
		expect(citationService.syncCitations).toHaveBeenCalledWith("cb-1", "Updated content", undefined);
	});

	it("collects errors per action", async () => {
		const apiClient = createApiClient();
		const citationService = { syncCitations: vi.fn() } as unknown as ContentCitationService;
		const executor = new PushExecutor(apiClient as unknown as StoryEngineClient, citationService);
		apiClient.updateScene.mockRejectedValueOnce(new Error("boom"));

		const actions: PushAction[] = [{ type: "scene_reorder", sceneId: "sc-err", newOrder: 1 }];
		const summary = await executor.execute(actions);

		expect(summary.applied).toBe(0);
		expect(summary.errors).toHaveLength(1);
		expect(summary.errors[0]).toMatchObject({
			code: "push_scene_reorder",
		});
	});
});

