import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import type { Beat, StoryEngineSettings } from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import { BeatHandler } from "../BeatHandler";

const beat: Beat = {
	id: "bt-1",
	scene_id: "sc-1",
	order_num: 1,
	type: "exposition",
	intent: "Introduce hero",
	outcome: "Hero introduced",
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
		getBeat: vi.fn().mockResolvedValue(beat),
		getScene: vi.fn().mockResolvedValue({
			id: "sc-1",
			story_id: "story-1",
			chapter_id: "ch-1",
			order_num: 1,
		}),
		getChapter: vi.fn().mockResolvedValue({ id: "ch-1", number: 1 }),
		getStory: vi.fn().mockResolvedValue({ id: "story-1", title: "Test Story" }),
		getContentBlocksByBeat: vi.fn().mockResolvedValue([]),
		listRelationsByTarget: vi.fn().mockResolvedValue({ data: [], pagination: { has_more: false } }),
		deleteBeat: vi.fn(),
	};
	const fileManager = {
		getStoryFolderPath: vi.fn().mockReturnValue("StoryFolder"),
		ensureFolderExists: vi.fn(),
		writeBeatFile: vi.fn(),
		writeFile: vi.fn(),
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

describe("BeatHandler", () => {
	it("writes beat file when pulling beat", async () => {
		const { context, apiClient, fileManager } = createContext();
		const handler = new BeatHandler();

		await handler.pull("bt-1", context);

		expect(apiClient.getBeat).toHaveBeenCalledWith("bt-1");
		expect(fileManager.writeBeatFile).toHaveBeenCalledWith(
			beat,
			"StoryFolder/02-beats/bt-0001-0001-0001-introduce-hero.md",
			"Test Story",
			[],
			{
				linkMode: "full_path",
				storyFolderPath: "StoryFolder",
				chapterOrder: 1,
				sceneOrder: 1,
			}
		);
	});
});

