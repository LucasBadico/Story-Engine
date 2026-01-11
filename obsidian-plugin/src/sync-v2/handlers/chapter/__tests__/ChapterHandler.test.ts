import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import type { ChapterWithContent, Scene, SceneWithBeats, StoryEngineSettings } from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import { ChapterHandler } from "../ChapterHandler";

const mockChapter: ChapterWithContent["chapter"] = {
	id: "ch-1",
	story_id: "story-1",
	number: 1,
	title: "Chapter 1",
	status: "draft",
	created_at: "2025-01-01T00:00:00Z",
	updated_at: "2025-01-01T00:00:00Z",
};

const mockScene = (id: string): Scene => ({
	id,
	story_id: "story-1",
	chapter_id: "ch-1",
	order_num: 1,
	pov_character_id: null,
	location_id: null,
	time_ref: "Morning",
	goal: "Meet hero",
	created_at: "2025-01-01T00:00:00Z",
	updated_at: "2025-01-01T00:00:00Z",
});

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
		getChapter: vi.fn().mockResolvedValue(mockChapter),
		getScenes: vi.fn().mockResolvedValue([mockScene("sc-1")]),
		getBeats: vi.fn().mockResolvedValue([]),
		getStory: vi.fn().mockResolvedValue({
			id: "story-1",
			title: "Test Story",
		}),
		deleteChapter: vi.fn(),
	};
	const fileManager = {
		getStoryFolderPath: vi.fn().mockReturnValue("StoryFolder"),
		ensureFolderExists: vi.fn(),
		writeChapterFile: vi.fn(),
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

describe("ChapterHandler", () => {
	it("writes chapter file when pulling chapter", async () => {
		const { context, apiClient, fileManager } = createContext();
		const handler = new ChapterHandler();

		await handler.pull("ch-1", context);

		expect(apiClient.getChapter).toHaveBeenCalledWith("ch-1");
		expect(apiClient.getScenes).toHaveBeenCalledWith("ch-1");
		expect(fileManager.writeChapterFile).toHaveBeenCalledWith(
			expect.any(Object),
			"StoryFolder/00-chapters/ch-0001-chapter-1.md",
			"Test Story"
		);
	});
});

