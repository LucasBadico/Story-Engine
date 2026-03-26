import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import type { StoryEngineSettings } from "../../../types";
import type { EntityRelation } from "../../types/relations";
import type { SyncContext } from "../../types/sync";
import { RelationsPushHandler } from "../RelationsPushHandler";

const baseSettings: StoryEngineSettings = {
	apiUrl: "",
	llmGatewayUrl: "",
	apiKey: "",
	tenantId: "",
	tenantName: "",
	syncFolderPath: "",
	autoVersionSnapshots: true,
	conflictResolution: "service",
	mode: "local",
	syncVersion: "v2",
	showHelpBox: true,
	autoSyncOnApiUpdates: true,
	autoPushOnFileBlur: true,
	backupMode: "snapshots",
	backupRetentionDays: 7,
};

const createContext = (relationsFileContent: string, existingRelations: EntityRelation[] = []) => {
	const apiClient = {
		listRelationsByTarget: vi.fn().mockResolvedValue({ data: existingRelations, pagination: { has_more: false } }),
		listRelationsByWorld: vi.fn().mockResolvedValue({ data: existingRelations, pagination: { has_more: false } }),
  createRelation: vi.fn().mockResolvedValue({ relation: { ...existingRelations[0], id: "rel-new" } }),
		updateRelation: vi.fn().mockResolvedValue({ relation: existingRelations[0] }),
		deleteRelation: vi.fn().mockResolvedValue(undefined),
		getCharacter: vi.fn().mockResolvedValue({ id: "char-1", name: "Test Character" }),
		getLocation: vi.fn().mockResolvedValue({ id: "loc-1", name: "Test Location" }),
		getStory: vi.fn().mockResolvedValue({ id: "story-1", title: "Test Story" }),
		getWorld: vi.fn().mockResolvedValue({ id: "world-1", name: "Test World" }),
		getFaction: vi.fn().mockResolvedValue({ id: "faction-1", name: "Test Faction" }),
		getArtifact: vi.fn().mockResolvedValue({ id: "art-1", name: "Test Artifact" }),
		getEvent: vi.fn().mockResolvedValue({ id: "evt-1", name: "Test Event" }),
		getLore: vi.fn().mockResolvedValue({ id: "lore-1", name: "Test Lore" }),
	};
	const fileManager = {
		readFile: vi.fn().mockResolvedValue(relationsFileContent),
	};

	const emitWarning = vi.fn();
	const context: SyncContext = {
		app: {} as App,
		apiClient: apiClient as any,
		fileManager: fileManager as any,
		settings: baseSettings,
		timestamp: () => "2025-01-10T00:00:00Z",
		backupMode: "snapshots",
		emitWarning,
	};

	return { context, apiClient, fileManager, emitWarning };
};

describe("RelationsPushHandler", () => {
	it("creates new relation when adding entry to story relations file", async () => {
		const relationsFile = `---
id: story-1
type: story-relations
synced_at: 2025-01-01T00:00:00Z
---

# Test Story - Relations

## Main Characters
- [[char-1|John Smith]] - Main character
- _Add new main character: [[file|Name]] - description_
`;

		const existingRelations: EntityRelation[] = [];
		const { context, apiClient } = createContext(relationsFile, existingRelations);

		const handler = new RelationsPushHandler();
		const result = await handler.pushRelations("story.relations.md", "story", "story-1", context);

		expect(apiClient.listRelationsByTarget).toHaveBeenCalledWith({
			targetType: "story",
			targetId: "story-1",
		});
		expect(apiClient.createRelation).toHaveBeenCalledWith({
			sourceType: "character",
			sourceId: "char-1",
			targetType: "story",
			targetId: "story-1",
			relationType: "pov",
			context: "Main character",
		});
		expect(result.created).toBe(1);
		expect(result.updated).toBe(0);
		expect(result.deleted).toBe(0);
	});

	it("updates existing relation when description changes", async () => {
		const relationsFile = `---
id: story-1
type: story-relations
synced_at: 2025-01-01T00:00:00Z
---

# Test Story - Relations

## Main Characters
- [[char-1|John Smith]] - Updated description
- _Add new main character: [[file|Name]] - description_
`;

		const existingRelations: EntityRelation[] = [
			{
				id: "rel-1",
				tenant_id: "tenant-1",
				source_type: "character",
				source_id: "char-1",
				target_type: "story",
				target_id: "story-1",
				relation_type: "pov",
				context: "Old description",
				created_at: "2025-01-01T00:00:00Z",
				updated_at: "2025-01-01T00:00:00Z",
				direction: "source" as const,
			},
		];
		const { context, apiClient } = createContext(relationsFile, existingRelations);

		const handler = new RelationsPushHandler();
		const result = await handler.pushRelations("story.relations.md", "story", "story-1", context);

		expect(apiClient.updateRelation).toHaveBeenCalledWith({
			id: "rel-1",
			context: "Updated description",
		});
		expect(result.created).toBe(0);
		expect(result.updated).toBe(1);
		expect(result.deleted).toBe(0);
	});

	it("deletes relation when entry is removed from file", async () => {
		const relationsFile = `---
id: story-1
type: story-relations
synced_at: 2025-01-01T00:00:00Z
---

# Test Story - Relations

## Main Characters
- _Add new main character: [[file|Name]] - description_
`;

		const existingRelations: EntityRelation[] = [
			{
				id: "rel-1",
				tenant_id: "tenant-1",
				source_type: "character",
				source_id: "char-1",
				target_type: "story",
				target_id: "story-1",
				relation_type: "pov",
				context: "Description",
				created_at: "2025-01-01T00:00:00Z",
				updated_at: "2025-01-01T00:00:00Z",
				direction: "source" as const,
			},
		];
		const { context, apiClient } = createContext(relationsFile, existingRelations);

		const handler = new RelationsPushHandler();
		const result = await handler.pushRelations("story.relations.md", "story", "story-1", context);

		expect(apiClient.deleteRelation).toHaveBeenCalledWith("rel-1");
		expect(result.created).toBe(0);
		expect(result.updated).toBe(0);
		expect(result.deleted).toBe(1);
	});

	it("validates that source entity exists before creating relation", async () => {
		const relationsFile = `---
id: story-1
type: story-relations
synced_at: 2025-01-01T00:00:00Z
---

# Test Story - Relations

## Main Characters
- [[char-invalid|Invalid Character]] - Description
- _Add new main character: [[file|Name]] - description_
`;

		const existingRelations: EntityRelation[] = [];
		const { context, apiClient, emitWarning } = createContext(relationsFile, existingRelations);
		apiClient.getCharacter = vi.fn().mockRejectedValue(new Error("Character not found"));

		const handler = new RelationsPushHandler();
		const result = await handler.pushRelations("story.relations.md", "story", "story-1", context);

		expect(apiClient.createRelation).not.toHaveBeenCalled();
		expect(result.created).toBe(0);
		expect(result.warnings.length).toBeGreaterThan(0);
		expect(result.warnings.some((w) => w.includes("Could not resolve entity ID"))).toBe(true);
	});

	it("validates that target entity exists before creating relation", async () => {
		const relationsFile = `---
id: story-1
type: story-relations
synced_at: 2025-01-01T00:00:00Z
---

# Test Story - Relations

## Main Characters
- [[char-1|John Smith]] - Main character
- _Add new main character: [[file|Name]] - description_
`;

		const existingRelations: EntityRelation[] = [];
		const { context, apiClient } = createContext(relationsFile, existingRelations);
		apiClient.getStory = vi.fn().mockRejectedValue(new Error("Story not found"));

		const handler = new RelationsPushHandler();
		const result = await handler.pushRelations("story.relations.md", "story", "story-1", context);

		expect(apiClient.createRelation).not.toHaveBeenCalled();
		expect(result.created).toBe(0);
		expect(result.warnings.length).toBeGreaterThan(0);
		expect(result.warnings.some((w) => w.includes("does not exist"))).toBe(true);
	});

	it("skips placeholders when processing relations", async () => {
		const relationsFile = `---
id: story-1
type: story-relations
synced_at: 2025-01-01T00:00:00Z
---

# Test Story - Relations

## Main Characters
- [[char-1|John Smith]] - Main character
- _Add new main character: [[file|Name]] - description_
`;

		const existingRelations: EntityRelation[] = [];
		const { context, apiClient } = createContext(relationsFile, existingRelations);

		const handler = new RelationsPushHandler();
		const result = await handler.pushRelations("story.relations.md", "story", "story-1", context);

		// Should only create one relation (the non-placeholder one)
		expect(apiClient.createRelation).toHaveBeenCalledTimes(1);
		expect(result.created).toBe(1);
	});
});

