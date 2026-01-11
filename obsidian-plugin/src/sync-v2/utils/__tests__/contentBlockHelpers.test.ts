import { describe, expect, it, vi } from "vitest";
import type { StoryEngineClient } from "../../../api/client";
import {
	resolveContentBlockHierarchy,
	buildHierarchyContext,
	createCitationRelations,
	type ContentBlockHierarchy,
} from "../contentBlockHelpers";

describe("contentBlockHelpers", () => {
	describe("resolveContentBlockHierarchy", () => {
		it("resolves hierarchy at beat level (most specific)", async () => {
			const mockApiClient = {
				getContentAnchors: vi.fn().mockResolvedValue([
					{ id: "anchor-1", content_block_id: "cb-1", entity_type: "beat", entity_id: "bt-1" },
				]),
				getBeat: vi.fn().mockResolvedValue({
					id: "bt-1",
					scene_id: "sc-1",
					order_num: 3,
					intent: "Confrontation",
					type: "dialogue",
					outcome: "",
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
				getScene: vi.fn().mockResolvedValue({
					id: "sc-1",
					story_id: "story-1",
					chapter_id: "ch-1",
					order_num: 2,
					goal: "The Meeting",
					time_ref: "Morning",
					pov_character_id: null,
					location_id: null,
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
				getChapter: vi.fn().mockResolvedValue({
					id: "ch-1",
					story_id: "story-1",
					number: 1,
					title: "Introduction",
					status: "draft",
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
				getStory: vi.fn().mockResolvedValue({
					id: "story-1",
					title: "Test Story",
					world_id: "world-1",
					tenant_id: "tenant-1",
					status: "draft",
					version: 1,
					root_story_id: "story-1",
					previous_version_id: null,
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
			} as unknown as StoryEngineClient;

			const result = await resolveContentBlockHierarchy("cb-1", mockApiClient);

			expect(result).not.toBeNull();
			expect(result?.contentBlockId).toBe("cb-1");
			expect(result?.beat?.id).toBe("bt-1");
			expect(result?.beat?.intent).toBe("Confrontation");
			expect(result?.beat?.order_num).toBe(3);
			expect(result?.scene?.id).toBe("sc-1");
			expect(result?.scene?.goal).toBe("The Meeting");
			expect(result?.chapter?.id).toBe("ch-1");
			expect(result?.chapter?.title).toBe("Introduction");
			expect(result?.story.id).toBe("story-1");
			expect(result?.worldId).toBe("world-1");
		});

		it("resolves hierarchy at scene level", async () => {
			const mockApiClient = {
				getContentAnchors: vi.fn().mockResolvedValue([
					{ id: "anchor-1", content_block_id: "cb-1", entity_type: "scene", entity_id: "sc-1" },
				]),
				getScene: vi.fn().mockResolvedValue({
					id: "sc-1",
					story_id: "story-1",
					chapter_id: "ch-1",
					order_num: 2,
					goal: "The Meeting",
					time_ref: "Morning",
					pov_character_id: null,
					location_id: null,
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
				getChapter: vi.fn().mockResolvedValue({
					id: "ch-1",
					story_id: "story-1",
					number: 1,
					title: "Introduction",
					status: "draft",
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
				getStory: vi.fn().mockResolvedValue({
					id: "story-1",
					title: "Test Story",
					world_id: "world-1",
					tenant_id: "tenant-1",
					status: "draft",
					version: 1,
					root_story_id: "story-1",
					previous_version_id: null,
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
			} as unknown as StoryEngineClient;

			const result = await resolveContentBlockHierarchy("cb-1", mockApiClient);

			expect(result).not.toBeNull();
			expect(result?.beat).toBeUndefined();
			expect(result?.scene?.id).toBe("sc-1");
			expect(result?.chapter?.id).toBe("ch-1");
			expect(result?.story.id).toBe("story-1");
		});

		it("resolves hierarchy at chapter level", async () => {
			const mockApiClient = {
				getContentAnchors: vi.fn().mockResolvedValue([
					{ id: "anchor-1", content_block_id: "cb-1", entity_type: "chapter", entity_id: "ch-1" },
				]),
				getChapter: vi.fn().mockResolvedValue({
					id: "ch-1",
					story_id: "story-1",
					number: 1,
					title: "Introduction",
					status: "draft",
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
				getStory: vi.fn().mockResolvedValue({
					id: "story-1",
					title: "Test Story",
					world_id: "world-1",
					tenant_id: "tenant-1",
					status: "draft",
					version: 1,
					root_story_id: "story-1",
					previous_version_id: null,
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
			} as unknown as StoryEngineClient;

			const result = await resolveContentBlockHierarchy("cb-1", mockApiClient);

			expect(result).not.toBeNull();
			expect(result?.beat).toBeUndefined();
			expect(result?.scene).toBeUndefined();
			expect(result?.chapter?.id).toBe("ch-1");
			expect(result?.story.id).toBe("story-1");
		});

		it("falls back to chapter_id when no anchors exist", async () => {
			const mockApiClient = {
				getContentAnchors: vi.fn().mockResolvedValue([]),
				getContentBlock: vi.fn().mockResolvedValue({
					id: "cb-1",
					chapter_id: "ch-1",
					order_num: 1,
					type: "text",
					kind: "prose",
					content: "Test",
					metadata: {},
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
				getChapter: vi.fn().mockResolvedValue({
					id: "ch-1",
					story_id: "story-1",
					number: 1,
					title: "Introduction",
					status: "draft",
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
				getStory: vi.fn().mockResolvedValue({
					id: "story-1",
					title: "Test Story",
					world_id: "world-1",
					tenant_id: "tenant-1",
					status: "draft",
					version: 1,
					root_story_id: "story-1",
					previous_version_id: null,
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
			} as unknown as StoryEngineClient;

			const result = await resolveContentBlockHierarchy("cb-1", mockApiClient);

			expect(result).not.toBeNull();
			expect(result?.chapter?.id).toBe("ch-1");
			expect(result?.story.id).toBe("story-1");
		});

		it("returns null when no hierarchy can be resolved", async () => {
			const mockApiClient = {
				getContentAnchors: vi.fn().mockResolvedValue([]),
				getContentBlock: vi.fn().mockResolvedValue({
					id: "cb-1",
					chapter_id: null,
					order_num: 1,
					type: "text",
					kind: "prose",
					content: "Test",
					metadata: {},
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
			} as unknown as StoryEngineClient;

			const result = await resolveContentBlockHierarchy("cb-1", mockApiClient);

			expect(result).toBeNull();
		});

		it("returns null when API call fails", async () => {
			const mockApiClient = {
				getContentAnchors: vi.fn().mockRejectedValue(new Error("API error")),
			} as unknown as StoryEngineClient;

			const result = await resolveContentBlockHierarchy("cb-1", mockApiClient);

			expect(result).toBeNull();
		});
	});

	describe("buildHierarchyContext", () => {
		it("builds context string with all levels", () => {
			const hierarchy: ContentBlockHierarchy = {
				contentBlockId: "cb-1",
				beat: { id: "bt-1", intent: "Confrontation", order_num: 3 },
				scene: { id: "sc-1", goal: "The Meeting", order_num: 2 },
				chapter: { id: "ch-1", title: "Introduction", number: 1 },
				story: { id: "story-1", title: "Test Story" },
			};

			const result = buildHierarchyContext(hierarchy);

			expect(result).toBe("Chapter 1: Introduction > Scene 2: The Meeting > Beat 3: Confrontation");
		});

		it("builds context string with scene and chapter only", () => {
			const hierarchy: ContentBlockHierarchy = {
				contentBlockId: "cb-1",
				scene: { id: "sc-1", goal: "The Meeting", order_num: 2 },
				chapter: { id: "ch-1", title: "Introduction", number: 1 },
				story: { id: "story-1", title: "Test Story" },
			};

			const result = buildHierarchyContext(hierarchy);

			expect(result).toBe("Chapter 1: Introduction > Scene 2: The Meeting");
		});

		it("builds context string with chapter only", () => {
			const hierarchy: ContentBlockHierarchy = {
				contentBlockId: "cb-1",
				chapter: { id: "ch-1", title: "Introduction", number: 1 },
				story: { id: "story-1", title: "Test Story" },
			};

			const result = buildHierarchyContext(hierarchy);

			expect(result).toBe("Chapter 1: Introduction");
		});

		it("uses story title when no hierarchy levels", () => {
			const hierarchy: ContentBlockHierarchy = {
				contentBlockId: "cb-1",
				story: { id: "story-1", title: "Test Story" },
			};

			const result = buildHierarchyContext(hierarchy);

			expect(result).toBe("Test Story");
		});

		it("handles missing titles with fallbacks", () => {
			const hierarchy: ContentBlockHierarchy = {
				contentBlockId: "cb-1",
				beat: { id: "bt-1", order_num: 0 }, // no intent, order_num 0
				scene: { id: "sc-1", order_num: 0 }, // no goal, order_num 0
				chapter: { id: "ch-1", title: "Introduction", number: 0 }, // number 0
				story: { id: "story-1", title: "Test Story" },
			};

			const result = buildHierarchyContext(hierarchy);

			expect(result).toBe("Chapter ?: Introduction > Scene ?: Untitled Scene > Beat ?: Untitled Beat");
		});
	});

	describe("createCitationRelations", () => {
		it("creates citation relations at beat level", async () => {
			const hierarchy: ContentBlockHierarchy = {
				contentBlockId: "cb-1",
				beat: { id: "bt-1", intent: "Confrontation", order_num: 3 },
				scene: { id: "sc-1", goal: "The Meeting", order_num: 2 },
				chapter: { id: "ch-1", title: "Introduction", number: 1 },
				story: { id: "story-1", title: "Test Story" },
				worldId: "world-1",
			};

			const mentions = [
				{ entityId: "char-123", entityType: "character", worldId: "world-1" },
				{ entityId: "loc-456", entityType: "location", worldId: "world-1" },
			];

			const mockApiClient = {
				listRelationsBySource: vi.fn().mockResolvedValue({
					data: [],
				}),
				getCharacter: vi.fn().mockResolvedValue({ id: "char-123", name: "Aria" }),
				getLocation: vi.fn().mockResolvedValue({ id: "loc-456", name: "Crystal Cave" }),
				createRelation: vi.fn().mockResolvedValue({
					id: "rel-1",
					tenant_id: "tenant-1",
					source_type: "beat",
					source_id: "bt-1",
					target_type: "character",
					target_id: "char-123",
					relation_type: "citation",
					direction: "source",
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
			} as unknown as StoryEngineClient;

			const result = await createCitationRelations(
				mentions,
				hierarchy,
				"cb-1",
				mockApiClient,
				"Chapter 1: Introduction > Scene 2: The Meeting > Beat 3: Confrontation"
			);

			expect(result.created).toBe(2);
			expect(result.errors).toHaveLength(0);
			expect(mockApiClient.listRelationsBySource).toHaveBeenCalledWith({
				sourceType: "beat",
				sourceId: "bt-1",
			});
			expect(mockApiClient.createRelation).toHaveBeenCalledTimes(2);
			expect(mockApiClient.createRelation).toHaveBeenCalledWith({
				sourceType: "beat",
				sourceId: "bt-1",
				targetType: "character",
				targetId: "char-123",
				relationType: "citation",
				context: expect.stringContaining("Chapter 1: Introduction > Scene 2: The Meeting > Beat 3: Confrontation"),
			});
		});

		it("creates citation relations at scene level when no beat", async () => {
			const hierarchy: ContentBlockHierarchy = {
				contentBlockId: "cb-1",
				scene: { id: "sc-1", goal: "The Meeting", order_num: 2 },
				chapter: { id: "ch-1", title: "Introduction", number: 1 },
				story: { id: "story-1", title: "Test Story" },
			};

			const mentions = [{ entityId: "char-123", entityType: "character" }];

			const mockApiClient = {
				listRelationsBySource: vi.fn().mockResolvedValue({ data: [] }),
				getCharacter: vi.fn().mockResolvedValue({ id: "char-123", name: "Aria" }),
				createRelation: vi.fn().mockResolvedValue({
					id: "rel-1",
					tenant_id: "tenant-1",
					source_type: "scene",
					source_id: "sc-1",
					target_type: "character",
					target_id: "char-123",
					relation_type: "citation",
					direction: "source",
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
			} as unknown as StoryEngineClient;

			const result = await createCitationRelations(
				mentions,
				hierarchy,
				"cb-1",
				mockApiClient,
				"Chapter 1: Introduction > Scene 2: The Meeting"
			);

			expect(result.created).toBe(1);
			expect(mockApiClient.createRelation).toHaveBeenCalledWith({
				sourceType: "scene",
				sourceId: "sc-1",
				targetType: "character",
				targetId: "char-123",
				relationType: "citation",
				context: expect.stringContaining("Chapter 1: Introduction > Scene 2: The Meeting"),
			});
		});

		it("skips duplicate citations", async () => {
			const hierarchy: ContentBlockHierarchy = {
				contentBlockId: "cb-1",
				beat: { id: "bt-1", intent: "Confrontation", order_num: 3 },
				story: { id: "story-1", title: "Test Story" },
			};

			const mentions = [{ entityId: "char-123", entityType: "character" }];

			const mockApiClient = {
				listRelationsBySource: vi.fn().mockResolvedValue({
					data: [
						{
							id: "rel-existing",
							tenant_id: "tenant-1",
							source_type: "beat",
							source_id: "bt-1",
							target_type: "character",
							target_id: "char-123",
							relation_type: "citation",
							direction: "source",
							created_at: "2025-01-01T00:00:00Z",
							updated_at: "2025-01-01T00:00:00Z",
						},
					],
				}),
				getCharacter: vi.fn().mockResolvedValue({ id: "char-123", name: "Aria" }),
				createRelation: vi.fn(),
			} as unknown as StoryEngineClient;

			const result = await createCitationRelations(
				mentions,
				hierarchy,
				"cb-1",
				mockApiClient,
				"Beat 3: Confrontation"
			);

			expect(result.created).toBe(0);
			expect(result.errors).toHaveLength(0);
			expect(mockApiClient.createRelation).not.toHaveBeenCalled();
		});

		it("validates target entities exist before creating", async () => {
			const hierarchy: ContentBlockHierarchy = {
				contentBlockId: "cb-1",
				chapter: { id: "ch-1", title: "Introduction", number: 1 },
				story: { id: "story-1", title: "Test Story" },
			};

			const mentions = [
				{ entityId: "char-123", entityType: "character" },
				{ entityId: "char-invalid", entityType: "character" },
			];

			const mockApiClient = {
				listRelationsBySource: vi.fn().mockResolvedValue({ data: [] }),
				getCharacter: vi
					.fn()
					.mockResolvedValueOnce({ id: "char-123", name: "Aria" })
					.mockRejectedValueOnce(new Error("Character not found")),
				createRelation: vi.fn().mockResolvedValue({
					id: "rel-1",
					tenant_id: "tenant-1",
					source_type: "chapter",
					source_id: "ch-1",
					target_type: "character",
					target_id: "char-123",
					relation_type: "citation",
					direction: "source",
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
			} as unknown as StoryEngineClient;

			const result = await createCitationRelations(
				mentions,
				hierarchy,
				"cb-1",
				mockApiClient,
				"Chapter 1: Introduction"
			);

			expect(result.created).toBe(1);
			expect(result.errors.length).toBeGreaterThan(0);
			expect(result.errors[0]).toContain("char-invalid");
			expect(mockApiClient.createRelation).toHaveBeenCalledTimes(1);
		});

		it("handles unsupported entity types", async () => {
			const hierarchy: ContentBlockHierarchy = {
				contentBlockId: "cb-1",
				chapter: { id: "ch-1", title: "Introduction", number: 1 },
				story: { id: "story-1", title: "Test Story" },
			};

			const mentions = [{ entityId: "unknown-123", entityType: "unsupported_type" }];

			const mockApiClient = {
				listRelationsBySource: vi.fn().mockResolvedValue({ data: [] }),
				createRelation: vi.fn(),
			} as unknown as StoryEngineClient;

			const result = await createCitationRelations(
				mentions,
				hierarchy,
				"cb-1",
				mockApiClient,
				"Chapter 1: Introduction"
			);

			expect(result.created).toBe(0);
			expect(result.errors).toHaveLength(1);
			expect(result.errors[0]).toContain("Unsupported entity type");
			expect(mockApiClient.createRelation).not.toHaveBeenCalled();
		});

		it("supports all world entity types", async () => {
			const hierarchy: ContentBlockHierarchy = {
				contentBlockId: "cb-1",
				chapter: { id: "ch-1", title: "Introduction", number: 1 },
				story: { id: "story-1", title: "Test Story" },
			};

			const mentions = [
				{ entityId: "char-1", entityType: "character" },
				{ entityId: "loc-1", entityType: "location" },
				{ entityId: "fac-1", entityType: "faction" },
				{ entityId: "art-1", entityType: "artifact" },
				{ entityId: "evt-1", entityType: "event" },
				{ entityId: "lor-1", entityType: "lore" },
			];

			const mockApiClient = {
				listRelationsBySource: vi.fn().mockResolvedValue({ data: [] }),
				getCharacter: vi.fn().mockResolvedValue({ id: "char-1" }),
				getLocation: vi.fn().mockResolvedValue({ id: "loc-1" }),
				getFaction: vi.fn().mockResolvedValue({ id: "fac-1" }),
				getArtifact: vi.fn().mockResolvedValue({ id: "art-1" }),
				getEvent: vi.fn().mockResolvedValue({ id: "evt-1" }),
				getLore: vi.fn().mockResolvedValue({ id: "lor-1" }),
				createRelation: vi.fn().mockResolvedValue({
					id: "rel-1",
					tenant_id: "tenant-1",
					source_type: "chapter",
					source_id: "ch-1",
					target_type: "character",
					target_id: "char-1",
					relation_type: "citation",
					direction: "source",
					created_at: "2025-01-01T00:00:00Z",
					updated_at: "2025-01-01T00:00:00Z",
				}),
			} as unknown as StoryEngineClient;

			const result = await createCitationRelations(
				mentions,
				hierarchy,
				"cb-1",
				mockApiClient,
				"Chapter 1: Introduction"
			);

			expect(result.created).toBe(6);
			expect(result.errors).toHaveLength(0);
			expect(mockApiClient.createRelation).toHaveBeenCalledTimes(6);
		});
	});
});

