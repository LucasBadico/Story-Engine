import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import type { App } from "obsidian";
import type { StoryEngineSettings, StoryWithHierarchy } from "../../../../../types";
import type { SyncContext } from "../../../../types/sync";
import { StoryRelationsService } from "../StoryRelationsService";

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
	chapters: [],
};

const createContext = () => {
	const apiClient = {
		listRelationsByTarget: vi.fn().mockResolvedValue({ data: [], pagination: { has_more: false } }),
		getCharacter: vi.fn().mockResolvedValue({ id: "char-1", name: "Test Character" }),
		getLocation: vi.fn().mockResolvedValue({ id: "loc-1", name: "Test Location" }),
		getFaction: vi.fn().mockResolvedValue({ id: "faction-1", name: "Test Faction" }),
		getArtifact: vi.fn().mockResolvedValue({ id: "art-1", name: "Test Artifact" }),
		getEvent: vi.fn().mockResolvedValue({ id: "evt-1", name: "Test Event" }),
		getLore: vi.fn().mockResolvedValue({ id: "lore-1", name: "Test Lore" }),
		getWorld: vi.fn().mockResolvedValue({ id: "world-1", name: "Test World" }),
	};
	const fileManager = {
		writeFile: vi.fn(),
		readFile: vi.fn(),
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

describe("StoryRelationsService", () => {
	let consoleWarnSpy: ReturnType<typeof vi.spyOn>;

	beforeEach(() => {
		consoleWarnSpy = vi.spyOn(console as any, "warn").mockImplementation(() => {});
	});

	afterEach(() => {
		consoleWarnSpy?.mockRestore();
	});
	it("generates relations file from API data", async () => {
		const { context, apiClient, fileManager } = createContext();
		const relationsGenerator = {
			generate: vi.fn().mockReturnValue("generated relations"),
		};
		const relationsPushHandler = {
			pushRelations: vi.fn(),
		};
		const service = new StoryRelationsService(
			relationsGenerator as any,
			relationsPushHandler as any
		);

		apiClient.listRelationsByTarget.mockResolvedValue({
			data: [
				{
					source_type: "character",
					source_id: "char-1",
					relation_type: "pov",
					context: "Protagonist",
				},
			],
			pagination: { has_more: false },
		});

		await service.generateRelationsFile(mockStory, "StoryFolder", context);

		expect(apiClient.listRelationsByTarget).toHaveBeenCalledWith({
			targetType: "story",
			targetId: "story-1",
		});
		expect(relationsGenerator.generate).toHaveBeenCalled();
		expect(fileManager.writeFile).toHaveBeenCalledWith(
			"StoryFolder/story.relations.md",
			"generated relations"
		);
	});

	it("writes placeholder relations file when API fails", async () => {
		const { context, apiClient, fileManager } = createContext();
		const relationsGenerator = { generate: vi.fn() };
		const service = new StoryRelationsService(relationsGenerator as any, {} as any);

		apiClient.listRelationsByTarget.mockRejectedValue(new Error("boom"));

		await service.generateRelationsFile(mockStory, "StoryFolder", context);

		expect(fileManager.writeFile).toHaveBeenCalledWith(
			"StoryFolder/story.relations.md",
			expect.stringContaining("_Relations will be populated")
		);
	});

	it("pushes relations and emits warnings from handler", async () => {
		const { context, fileManager, emitWarning } = createContext();
		const relationsPushHandler = {
			pushRelations: vi.fn().mockResolvedValue({
				created: 0,
				updated: 0,
				deleted: 0,
				warnings: ["warn"],
			}),
		};
		const service = new StoryRelationsService({} as any, relationsPushHandler as any);
		fileManager.readFile.mockResolvedValue("relations content");

		await service.pushRelations(mockStory, "StoryFolder", context);

		expect(relationsPushHandler.pushRelations).toHaveBeenCalledWith(
			"StoryFolder/story.relations.md",
			"story",
			"story-1",
			context,
			undefined
		);
		expect(emitWarning).toHaveBeenCalledWith(
			expect.objectContaining({ code: "relations_push_warning" })
		);
	});

	it("emits warning when push fails", async () => {
		const { context, fileManager, emitWarning } = createContext();
		const relationsPushHandler = {
			pushRelations: vi.fn().mockRejectedValue(new Error("failed")),
		};
		const service = new StoryRelationsService({} as any, relationsPushHandler as any);
		fileManager.readFile.mockResolvedValue("relations content");

		await service.pushRelations(mockStory, "StoryFolder", context);

		expect(emitWarning).toHaveBeenCalledWith(
			expect.objectContaining({ code: "relations_push_error" })
		);
	});
});
