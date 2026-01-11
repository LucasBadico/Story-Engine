import { describe, expect, it, vi } from "vitest";
import { StoryEngineClient } from "../client";
import type { EntityRelation } from "../../sync-v2/types/relations";

const RELATION_FIXTURE: EntityRelation = {
	id: "rel-1",
	tenant_id: "tenant-1",
	source_type: "scene",
	source_id: "sc-1",
	target_type: "character",
	target_id: "ch-1",
	relation_type: "appearance",
	context: "Scene 1",
	created_at: "2025-01-01T00:00:00Z",
	updated_at: "2025-01-01T00:00:00Z",
	direction: "source",
};

describe("StoryEngineClient relations endpoints", () => {
	it("builds query for listRelationsBySource", async () => {
		const client = new StoryEngineClient("https://api.storyengine.dev", "token");
		const spy = vi.spyOn(client as any, "request").mockResolvedValue({
			data: [RELATION_FIXTURE],
			pagination: { has_more: true, next_cursor: "abc" },
		});

		const response = await client.listRelationsBySource({
			sourceType: "scene",
			sourceId: "sc-1",
			relationType: "character",
			cursor: "abc",
			limit: 50,
			orderBy: "created_at",
			orderDir: "desc",
			excludeMirrors: true,
		});

		expect(spy).toHaveBeenCalledWith(
			"GET",
			"/api/v1/relations/source?source_type=scene&source_id=sc-1&relation_type=character&cursor=abc&limit=50&order_by=created_at&order_dir=desc&exclude_mirrors=true"
		);
		expect(response.data).toHaveLength(1);
		expect(response.pagination?.has_more).toBe(true);
		expect(response.pagination?.next_cursor).toBe("abc");
	});

	it("builds query for listRelationsByTarget with required params only", async () => {
		const client = new StoryEngineClient("https://api.storyengine.dev", "token");
		const spy = vi.spyOn(client as any, "request").mockResolvedValue({
			data: [],
			pagination: { has_more: false },
		});

		await client.listRelationsByTarget({
			targetType: "character",
			targetId: "ch-1",
		});

		expect(spy).toHaveBeenCalledWith(
			"GET",
			"/api/v1/relations/target?target_type=character&target_id=ch-1"
		);
	});

	it("omits query params when listing relations by world with defaults", async () => {
		const client = new StoryEngineClient("https://api.storyengine.dev", "token");
		const spy = vi.spyOn(client as any, "request").mockResolvedValue({
			data: [],
			pagination: { has_more: false },
		});

		await client.listRelationsByWorld({
			worldId: "world-1",
		});

		expect(spy).toHaveBeenCalledWith("GET", "/api/v1/worlds/world-1/relations");
	});
});

