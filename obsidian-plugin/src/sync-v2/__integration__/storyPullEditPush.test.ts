import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { SyncOrchestrator } from "../core/SyncOrchestrator";
import type { StoryEngineSettings, StoryWithHierarchy } from "../../types";
import type { SyncContext } from "../types/sync";
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
			scenes: [
				{
					scene: {
						id: "sc-1",
						story_id: "story-1",
						chapter_id: "ch-1",
						order_num: 1,
						pov_character_id: null,
						location_id: null,
						time_ref: "Morning",
						goal: "Meet hero",
						created_at: "2025-01-01T00:00:00Z",
						updated_at: "2025-01-01T00:00:00Z",
					},
					beats: [],
				},
			],
		},
	],
};

describe("Sync V2 integration - story pull/edit/push", () => {
	it("pushes content updates detected in story.contents.md", async () => {
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
			updateContentBlock: vi.fn().mockResolvedValue({}),
			createEntityRelation: vi.fn(),
			listRelationsByTarget: vi.fn().mockResolvedValue({ data: [], pagination: { has_more: false } }),
		};

		const fileManager = new MockFileManager("StoryFolder");
		const context: SyncContext = {
			app: {} as App,
			apiClient: apiClient as any,
			fileManager: fileManager as any,
			settings,
			timestamp: () => "2025-01-10T00:00:00Z",
			backupMode: "off",
			emitWarning: vi.fn(),
		};

		const orchestrator = new SyncOrchestrator(context);
		(orchestrator as any).contentsGenerator = {
			generateStoryContents: () => remoteContents,
		};

		await orchestrator.run({
			type: "pull_story",
			payload: { storyId: "story-1" },
		});

		const storyFolder = fileManager.getStoryFolderPath("Test Story");
		fileManager.setFile(`${storyFolder}/story.contents.md`, localContents);

		await orchestrator.run({
			type: "push_story",
			payload: { folderPath: storyFolder },
		});

		expect(apiClient.updateContentBlock).toHaveBeenCalledWith("cb-1", {
			content: "Modified content by user",
		});
	});
});
