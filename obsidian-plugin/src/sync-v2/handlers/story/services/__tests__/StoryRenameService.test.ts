import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import type { App } from "obsidian";
import type { ContentBlock, StoryEngineSettings, StoryWithHierarchy } from "../../../../../types";
import type { SyncContext } from "../../../../types/sync";
import type { DiffOperation } from "../../../../diff/types";
import { StoryRenameService } from "../StoryRenameService";
import { PathResolver } from "../../../../fileRenamer/PathResolver";

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
					beats: [
						{
							id: "bt-1",
							scene_id: "sc-1",
							order_num: 1,
							type: "exposition",
							intent: "Introduce hero",
							outcome: "Hero introduced",
							created_at: "2025-01-01T00:00:00Z",
							updated_at: "2025-01-01T00:00:00Z",
						},
					],
				},
			],
		},
	],
};

const contentBlock: ContentBlock = {
	id: "cb-1",
	chapter_id: "ch-1",
	order_num: 1,
	type: "text",
	kind: "paragraph",
	content: "Hello",
	metadata: { title: "Intro" },
	created_at: "2025-01-01T00:00:00Z",
	updated_at: "2025-01-01T00:00:00Z",
};

const createContext = () => {
	const apiClient = {
		getContentBlock: vi.fn().mockResolvedValue(contentBlock),
	};
	const context: SyncContext = {
		app: {} as App,
		apiClient: apiClient as any,
		fileManager: {} as any,
		settings: baseSettings,
		timestamp: () => "2025-01-10T00:00:00Z",
		backupMode: "snapshots",
	};

	return { context, apiClient };
};

describe("StoryRenameService", () => {
	let consoleWarnSpy: ReturnType<typeof vi.spyOn>;

	beforeEach(() => {
		consoleWarnSpy = vi.spyOn(console as any, "warn").mockImplementation(() => {});
	});

	afterEach(() => {
		consoleWarnSpy?.mockRestore();
	});
	it("renames chapter files when reordered", async () => {
		const { context } = createContext();
		const renamer = { rename: vi.fn().mockResolvedValue(undefined) };
		const service = new StoryRenameService(() => renamer as any);
		const operations: DiffOperation[] = [
			{
				kind: "reordered",
				fenceId: "ch-1",
				fenceType: "chapter",
				metadata: { oldOrder: 1, newOrder: 2 },
			},
		];
		const pathResolver = new PathResolver("StoryFolder");
		const oldPath = pathResolver.getChapterPath(mockStory.chapters[0].chapter, { order: 1 });
		const newPath = pathResolver.getChapterPath(mockStory.chapters[0].chapter, { order: 2 });

		await service.handleReorders(operations, mockStory, "StoryFolder", context);

		expect(renamer.rename).toHaveBeenCalledWith({ oldPath, newPath });
	});

	it("renames scene files when reordered", async () => {
		const { context } = createContext();
		const renamer = { rename: vi.fn().mockResolvedValue(undefined) };
		const service = new StoryRenameService(() => renamer as any);
		const operations: DiffOperation[] = [
			{
				kind: "reordered",
				fenceId: "sc-1",
				fenceType: "scene",
				metadata: { oldOrder: 1, newOrder: 2 },
			},
		];
		const pathResolver = new PathResolver("StoryFolder");
		const scene = mockStory.chapters[0].scenes[0].scene;
		const oldPath = pathResolver.getScenePath(scene, { order: 1, chapterOrder: 1 });
		const newPath = pathResolver.getScenePath(scene, { order: 2, chapterOrder: 1 });

		await service.handleReorders(operations, mockStory, "StoryFolder", context);

		expect(renamer.rename).toHaveBeenCalledWith({ oldPath, newPath });
	});

	it("renames beat files when reordered", async () => {
		const { context } = createContext();
		const renamer = { rename: vi.fn().mockResolvedValue(undefined) };
		const service = new StoryRenameService(() => renamer as any);
		const operations: DiffOperation[] = [
			{
				kind: "reordered",
				fenceId: "bt-1",
				fenceType: "beat",
				metadata: { oldOrder: 1, newOrder: 2 },
			},
		];
		const pathResolver = new PathResolver("StoryFolder");
		const beat = mockStory.chapters[0].scenes[0].beats[0];
		const oldPath = pathResolver.getBeatPath(beat, { order: 1, chapterOrder: 1, sceneOrder: 1 });
		const newPath = pathResolver.getBeatPath(beat, { order: 2, chapterOrder: 1, sceneOrder: 1 });

		await service.handleReorders(operations, mockStory, "StoryFolder", context);

		expect(renamer.rename).toHaveBeenCalledWith({ oldPath, newPath });
	});

	it("renames content block files when reordered", async () => {
		const { context } = createContext();
		const renamer = { rename: vi.fn().mockResolvedValue(undefined) };
		const service = new StoryRenameService(() => renamer as any);
		const operations: DiffOperation[] = [
			{
				kind: "reordered",
				fenceId: "cb-1",
				fenceType: "content",
				metadata: { oldOrder: 1, newOrder: 2 },
			},
		];
		const pathResolver = new PathResolver("StoryFolder");
		const oldPath = pathResolver.getContentBlockPath(contentBlock, { order: 1 });
		const newPath = pathResolver.getContentBlockPath(contentBlock, { order: 2 });

		await service.handleReorders(operations, mockStory, "StoryFolder", context);

		expect(renamer.rename).toHaveBeenCalledWith({ oldPath, newPath });
	});

	it("ignores non-reordered operations", async () => {
		const { context } = createContext();
		const renamer = { rename: vi.fn().mockResolvedValue(undefined) };
		const service = new StoryRenameService(() => renamer as any);
		const operations: DiffOperation[] = [
			{ kind: "updated", fenceId: "ch-1", fenceType: "chapter" },
		];

		await service.handleReorders(operations, mockStory, "StoryFolder", context);

		expect(renamer.rename).not.toHaveBeenCalled();
	});

	it("does not throw when rename fails", async () => {
		const { context } = createContext();
		const renamer = { rename: vi.fn().mockRejectedValue(new Error("fail")) };
		const service = new StoryRenameService(() => renamer as any);
		const operations: DiffOperation[] = [
			{
				kind: "reordered",
				fenceId: "ch-1",
				fenceType: "chapter",
				metadata: { oldOrder: 1, newOrder: 2 },
			},
		];

		await expect(
			service.handleReorders(operations, mockStory, "StoryFolder", context)
		).resolves.toBeUndefined();
	});
});
