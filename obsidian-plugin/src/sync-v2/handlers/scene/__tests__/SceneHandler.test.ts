import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import type { Scene, StoryEngineSettings } from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import { SceneHandler } from "../SceneHandler";

const scene: Scene = {
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
};

const settings: StoryEngineSettings = {
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

const createContext = () => {
	const apiClient = {
		getScene: vi.fn().mockResolvedValue(scene),
		getBeats: vi.fn().mockResolvedValue([]),
		getStory: vi.fn().mockResolvedValue({ id: "story-1", title: "Test Story" }),
		getContentBlocksByScene: vi.fn().mockResolvedValue([]),
		deleteScene: vi.fn(),
	};
	const fileManager = {
		getStoryFolderPath: vi.fn().mockReturnValue("StoryFolder"),
		ensureFolderExists: vi.fn(),
		writeSceneFile: vi.fn(),
	};

	const context: SyncContext = {
		app: {} as App,
		apiClient: apiClient as any,
		fileManager: fileManager as any,
		settings,
		timestamp: () => "2025-01-01T00:00:00Z",
		backupMode: "snapshots",
	};

	return { context, apiClient, fileManager };
};

describe("SceneHandler", () => {
	it("writes scene file when pulling scene", async () => {
		const { context, apiClient, fileManager } = createContext();
		const handler = new SceneHandler();

		await handler.pull("sc-1", context);

		expect(apiClient.getScene).toHaveBeenCalledWith("sc-1");
		expect(fileManager.writeSceneFile).toHaveBeenCalledWith(
			expect.objectContaining({ scene }),
			"StoryFolder/01-scenes/sc-0001-meet-hero.md",
			"Test Story",
			[],
			[]
		);
	});
});

