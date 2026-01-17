import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import type {
	Artifact,
	Character,
	Faction,
	Lore,
	Location,
	World,
	WorldEvent,
} from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import { WorldHandler } from "../WorldHandler";

const world: World = {
	id: "world-1",
	tenant_id: "tenant-1",
	name: "Eldoria",
	description: "Magic realm",
	genre: "Fantasy",
	is_implicit: false,
	rpg_system_id: null,
	time_config: null,
	created_at: "2025-01-01T00:00:00Z",
	updated_at: "2025-01-01T00:00:00Z",
};

const createContext = () => {
	const apiClient = {
		getWorld: vi.fn().mockResolvedValue(world),
		getCharacters: vi.fn().mockResolvedValue([] as Character[]),
		getLocations: vi.fn().mockResolvedValue([] as Location[]),
		getFactions: vi.fn().mockResolvedValue([] as Faction[]),
		getArtifacts: vi.fn().mockResolvedValue([] as Artifact[]),
		getEvents: vi.fn().mockResolvedValue([] as WorldEvent[]),
		getLores: vi.fn().mockResolvedValue([] as Lore[]),
		listRelationsByWorld: vi.fn().mockResolvedValue({ data: [], pagination: { has_more: false } }),
	};
	const fileManager = {
		getWorldFolderPath: vi.fn().mockReturnValue("StoryFolder/worlds/eldoria"),
		ensureFolderExists: vi.fn(),
		writeWorldMetadata: vi.fn(),
		writeFile: vi.fn(),
	};

	const context: SyncContext = {
		app: {} as App,
		apiClient: apiClient as any,
		fileManager: fileManager as any,
		settings: {} as any,
		timestamp: () => "2025-01-10T00:00:00Z",
		backupMode: "snapshots",
	};

	return { context, apiClient, fileManager };
};

describe("WorldHandler", () => {
	it("writes world files", async () => {
		const { context, fileManager } = createContext();
		const handler = new WorldHandler(() => "2025-01-10T00:00:00Z");

		await handler.pull("world-1", context);

		expect(fileManager.getWorldFolderPath).toHaveBeenCalledWith("Eldoria");
		expect(fileManager.writeWorldMetadata).toHaveBeenCalledWith(
			world,
			"StoryFolder/worlds/eldoria"
		);
		expect(fileManager.writeFile).toHaveBeenCalledWith(
			expect.stringContaining("world.outline.md"),
			expect.stringContaining("# Eldoria - Outline")
		);
	});
});

