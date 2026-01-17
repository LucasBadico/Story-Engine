import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import type { StoryEngineSettings, StoryWithHierarchy } from "../../../../../types";
import type { SyncContext } from "../../../../types/sync";
import { StoryFileService } from "../StoryFileService";

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

const mockStory: StoryWithHierarchy = {
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
				title: "Chapter One",
				status: "draft",
				created_at: "2025-01-01T00:00:00Z",
				updated_at: "2025-01-01T00:00:00Z",
			},
			scenes: [],
		},
		{
			chapter: {
				id: "ch-2",
				story_id: "story-1",
				number: 2,
				title: "Chapter Two",
				status: "draft",
				created_at: "2025-01-01T00:00:00Z",
				updated_at: "2025-01-01T00:00:00Z",
			},
			scenes: [],
		},
	],
};

const createContext = () => {
	const fileManager = {
		readFile: vi.fn(),
		writeFile: vi.fn(),
		ensureFolderExists: vi.fn(),
		writeChapterFile: vi.fn(),
		writeSceneFile: vi.fn(),
		writeBeatFile: vi.fn(),
		writeContentBlockFile: vi.fn(),
	};
	const context: SyncContext = {
		app: {} as App,
		apiClient: {
			getContentBlocksByScene: vi.fn().mockResolvedValue([]),
			getContentBlocksByBeat: vi.fn().mockResolvedValue([]),
			listRelationsByTarget: vi.fn().mockResolvedValue({ data: [], pagination: { has_more: false } }),
		} as any,
		fileManager: fileManager as any,
		settings: baseSettings,
		timestamp: () => "2025-01-10T00:00:00Z",
		backupMode: "snapshots",
	};

	return { context, fileManager };
};

describe("StoryFileService", () => {
	it("reads file silently when it exists", async () => {
		const { context, fileManager } = createContext();
		const service = new StoryFileService();
		fileManager.readFile.mockResolvedValue("content");

		const result = await service.readFileSilently(context, "StoryFolder/story.md");

		expect(result).toBe("content");
	});

	it("returns null when read fails", async () => {
		const { context, fileManager } = createContext();
		const service = new StoryFileService();
		fileManager.readFile.mockRejectedValue(new Error("missing"));

		const result = await service.readFileSilently(context, "StoryFolder/story.md");

		expect(result).toBeNull();
	});

	it("writes chapter files for each chapter", async () => {
		const { context, fileManager } = createContext();
		const service = new StoryFileService();

		await service.writeIndividualEntityFiles(mockStory, "StoryFolder", context);

		expect(fileManager.ensureFolderExists).toHaveBeenCalledWith("StoryFolder/00-chapters");
		expect(fileManager.writeChapterFile).toHaveBeenCalledTimes(2);
	});

	it("creates chapter folder for each file", async () => {
		const { context, fileManager } = createContext();
		const service = new StoryFileService();

		await service.writeIndividualEntityFiles(mockStory, "StoryFolder", context);

		expect(fileManager.ensureFolderExists).toHaveBeenCalledWith("StoryFolder/00-chapters");
	});
});
