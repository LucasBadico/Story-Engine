import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import type { StoryEngineSettings } from "../../types";
import type { SyncContext } from "../types/sync";
import { RelationsPushHandler } from "../push/RelationsPushHandler";
import { MockFileManager } from "./helpers/mockFileManager";

const settings: StoryEngineSettings = {
	apiUrl: "",
	llmGatewayUrl: "",
	apiKey: "",
	tenantId: "",
	tenantName: "",
	syncFolderPath: "StoryFolder",
	autoVersionSnapshots: true,
	conflictResolution: "local",
	mode: "local",
	syncVersion: "v2",
	showHelpBox: true,
	autoSyncOnApiUpdates: true,
	autoPushOnFileBlur: true,
	backupMode: "off",
	backupRetentionDays: 7,
};

describe("Sync V2 integration - relations push", () => {
	it("creates relations from story.relations.md", async () => {
		const relationsContent = readFileSync(
			resolve(__dirname, "fixtures/story/story.relations.md"),
			"utf8"
		);

		const apiClient = {
			listRelationsByTarget: vi.fn().mockResolvedValue({ data: [], pagination: { has_more: false } }),
			getCharacter: vi.fn().mockResolvedValue({ id: "char-1", name: "John Smith" }),
			getLocation: vi.fn().mockResolvedValue({ id: "loc-1", name: "Location" }),
			getFaction: vi.fn().mockResolvedValue({ id: "faction-1", name: "Faction" }),
			getArtifact: vi.fn().mockResolvedValue({ id: "art-1", name: "Artifact" }),
			getEvent: vi.fn().mockResolvedValue({ id: "evt-1", name: "Event" }),
			getLore: vi.fn().mockResolvedValue({ id: "lore-1", name: "Lore" }),
			getStory: vi.fn().mockResolvedValue({ id: "story-1", title: "Test Story" }),
			createRelation: vi.fn().mockResolvedValue({ relation: { id: "rel-1" } }),
		};

		const fileManager = new MockFileManager("StoryFolder");
		fileManager.setFile("StoryFolder/Test Story/story.relations.md", relationsContent);

		const context: SyncContext = {
			app: {} as App,
			apiClient: apiClient as any,
			fileManager: fileManager as any,
			settings,
			timestamp: () => "2025-01-10T00:00:00Z",
			backupMode: "off",
		};

		const handler = new RelationsPushHandler();
		const result = await handler.pushRelations(
			"StoryFolder/Test Story/story.relations.md",
			"story",
			"story-1",
			context
		);

		expect(apiClient.createRelation).toHaveBeenCalledWith({
			sourceType: "character",
			sourceId: "char-1",
			targetType: "story",
			targetId: "story-1",
			relationType: "pov",
			context: "Protagonist",
		});
		expect(result.created).toBe(1);
	});
});
