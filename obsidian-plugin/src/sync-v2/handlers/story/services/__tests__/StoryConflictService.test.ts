import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import type { StoryEngineSettings, StoryWithHierarchy } from "../../../../../types";
import type { SyncContext } from "../../../../types/sync";
import type { Conflict, ConflictResolutionResult } from "../../../../conflict/types";
import { StoryConflictService } from "../StoryConflictService";

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

const mockStory: StoryWithHierarchy["story"] = {
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
	updated_at: "2025-01-02T00:00:00Z",
};

const createContext = () => {
	const fileManager = {
		readFile: vi.fn(),
	};
	const emitWarning = vi.fn();
	const context: SyncContext = {
		app: {} as App,
		apiClient: {} as any,
		fileManager: fileManager as any,
		settings: baseSettings,
		timestamp: () => "2025-01-10T00:00:00Z",
		backupMode: "snapshots",
		emitWarning,
	};

	return { context, fileManager, emitWarning };
};

describe("StoryConflictService", () => {
	it("does nothing when file does not exist", async () => {
		const { context, fileManager } = createContext();
		fileManager.readFile.mockRejectedValue(new Error("missing"));
		const resolverFactory = vi.fn();
		const service = new StoryConflictService(resolverFactory as any);

		await service.checkConflicts("StoryFolder/story.md", mockStory, context);

		expect(resolverFactory).not.toHaveBeenCalled();
	});

	it("does nothing when no timestamp is present", async () => {
		const { context, fileManager } = createContext();
		fileManager.readFile.mockResolvedValue("---\nid: story-1\n---\n\n# Story");
		const resolverFactory = vi.fn();
		const service = new StoryConflictService(resolverFactory as any);

		await service.checkConflicts("StoryFolder/story.md", mockStory, context);

		expect(resolverFactory).not.toHaveBeenCalled();
	});

	it("emits warning when manual resolution is required", async () => {
		const { context, fileManager, emitWarning } = createContext();
		fileManager.readFile.mockResolvedValue(
			"---\nupdated_at: 2025-01-01T00:00:00Z\n---\n\n# Story"
		);
		const conflict: Conflict = {
			type: "simultaneous_edit",
			entityId: "story-1",
			entityType: "story",
			filePath: "StoryFolder/story.md",
			localData: {},
			remoteData: {},
		};
		const resolver = {
			detectConflict: vi.fn().mockReturnValue(conflict),
			resolve: vi.fn().mockResolvedValue({
				success: true,
				resolution: { strategy: "manual", autoResolved: false },
			} as ConflictResolutionResult),
		};
		const service = new StoryConflictService(() => resolver as any);

		await service.checkConflicts("StoryFolder/story.md", mockStory, context);

		expect(resolver.detectConflict).toHaveBeenCalled();
		expect(resolver.resolve).toHaveBeenCalledWith(conflict);
		expect(emitWarning).toHaveBeenCalledWith(
			expect.objectContaining({ code: "conflict_requires_manual_resolution" })
		);
	});

	it("emits warning when resolution fails", async () => {
		const { context, fileManager, emitWarning } = createContext();
		fileManager.readFile.mockResolvedValue(
			"---\nupdated_at: 2025-01-01T00:00:00Z\n---\n\n# Story"
		);
		const conflict: Conflict = {
			type: "simultaneous_edit",
			entityId: "story-1",
			entityType: "story",
			filePath: "StoryFolder/story.md",
			localData: {},
			remoteData: {},
		};
		const resolver = {
			detectConflict: vi.fn().mockReturnValue(conflict),
			resolve: vi.fn().mockResolvedValue({
				success: false,
				resolution: { strategy: "manual", autoResolved: false },
				error: "failed",
			} as ConflictResolutionResult),
		};
		const service = new StoryConflictService(() => resolver as any);

		await service.checkConflicts("StoryFolder/story.md", mockStory, context);

		expect(emitWarning).toHaveBeenCalledWith(
			expect.objectContaining({ code: "conflict_resolution_failed" })
		);
	});

	it("does nothing when no conflict is detected", async () => {
		const { context, fileManager, emitWarning } = createContext();
		fileManager.readFile.mockResolvedValue(
			"---\nupdated_at: 2025-01-01T00:00:00Z\n---\n\n# Story"
		);
		const resolver = {
			detectConflict: vi.fn().mockReturnValue(null),
			resolve: vi.fn(),
		};
		const service = new StoryConflictService(() => resolver as any);

		await service.checkConflicts("StoryFolder/story.md", mockStory, context);

		expect(resolver.resolve).not.toHaveBeenCalled();
		expect(emitWarning).not.toHaveBeenCalled();
	});
});
