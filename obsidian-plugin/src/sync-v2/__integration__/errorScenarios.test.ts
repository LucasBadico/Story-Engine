import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import type { StoryEngineSettings, StoryWithHierarchy } from "../../types";
import type { SyncContext } from "../types/sync";
import { SyncOrchestrator } from "../core/SyncOrchestrator";
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

const storyWithHierarchy: StoryWithHierarchy = {
	story: {
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
	},
	chapters: [
		{
			chapter: {
				id: "ch-1",
				story_id: "story-1",
				number: 1,
				title: "Chapter 1",
				status: "draft",
				created_at: "2025-01-01T00:00:00Z",
				updated_at: "2025-01-01T00:00:00Z",
			},
			scenes: [],
		},
	],
};

describe("Sync V2 integration - error scenarios", () => {
	it("returns sync_v2_push_failed when contents file is missing", async () => {
		const apiClient = {
			getStoryWithHierarchy: vi.fn().mockResolvedValue(storyWithHierarchy),
		};

		const fileManager = new MockFileManager("StoryFolder");
		await fileManager.writeStoryMetadata(
			storyWithHierarchy.story,
			fileManager.getStoryFolderPath("Test Story"),
			storyWithHierarchy.chapters.map((c) => c.chapter)
		);

		const context: SyncContext = {
			app: {} as App,
			apiClient: apiClient as any,
			fileManager: fileManager as any,
			settings,
			timestamp: () => "2025-01-10T00:00:00Z",
			backupMode: "off",
		};

		const orchestrator = new SyncOrchestrator(context);
		const result = await orchestrator.run({
			type: "push_story",
			payload: { folderPath: fileManager.getStoryFolderPath("Test Story") },
		});

		expect(result.success).toBe(false);
		expect(result.errors?.[0]?.code).toBe("sync_v2_push_failed");
	});

	it("returns push_content_update error when update fails", async () => {
		const localContents = readFileSync(
			resolve(__dirname, "fixtures/story/story.contents.local.md"),
			"utf8"
		);
		const remoteContents = readFileSync(
			resolve(__dirname, "fixtures/story/story.contents.remote.md"),
			"utf8"
		);

		const apiClient = {
			getStoryWithHierarchy: vi.fn().mockResolvedValue(storyWithHierarchy),
			updateContentBlock: vi.fn().mockRejectedValue(new Error("update failed")),
			createEntityRelation: vi.fn(),
		};

		const fileManager = new MockFileManager("StoryFolder");
		const storyFolder = fileManager.getStoryFolderPath("Test Story");
		await fileManager.writeStoryMetadata(
			storyWithHierarchy.story,
			storyFolder,
			storyWithHierarchy.chapters.map((c) => c.chapter)
		);
		fileManager.setFile(`${storyFolder}/story.contents.md`, localContents);

		const context: SyncContext = {
			app: {} as App,
			apiClient: apiClient as any,
			fileManager: fileManager as any,
			settings,
			timestamp: () => "2025-01-10T00:00:00Z",
			backupMode: "off",
		};

		const orchestrator = new SyncOrchestrator(context);
		(orchestrator as any).contentsGenerator = {
			generateStoryContents: () => remoteContents,
		};

		const result = await orchestrator.run({
			type: "push_story",
			payload: { folderPath: storyFolder },
		});

		expect(result.success).toBe(false);
		expect(result.errors?.[0]?.code).toBe("push_content_update");
	});
});
