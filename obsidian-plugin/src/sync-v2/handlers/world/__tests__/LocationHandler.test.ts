import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import type { Location, World } from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import { LocationHandler } from "../LocationHandler";

const location: Location = {
	id: "loc-1",
	tenant_id: "tenant-1",
	world_id: "world-1",
	name: "Crystal Cave",
	type: "dungeon",
	description: "A shiny cave",
	parent_id: null,
	hierarchy_level: 1,
	created_at: "2025-01-01T00:00:00Z",
	updated_at: "2025-01-01T00:00:00Z",
};

const world: World = {
	id: "world-1",
	tenant_id: "tenant-1",
	name: "Eldoria",
	description: "",
	genre: "Fantasy",
	is_implicit: false,
	rpg_system_id: null,
	time_config: null,
	created_at: "2025-01-01T00:00:00Z",
	updated_at: "2025-01-01T00:00:00Z",
};

const createContext = () => {
	const apiClient = {
		getLocation: vi.fn().mockResolvedValue(location),
		getWorld: vi.fn().mockResolvedValue(world),
		deleteLocation: vi.fn(),
	};
	const fileManager = {
		getWorldFolderPath: vi.fn().mockReturnValue("StoryFolder/worlds/eldoria"),
		ensureFolderExists: vi.fn(),
		writeFile: vi.fn(),
	};

	const context: SyncContext = {
		app: {} as App,
		apiClient: apiClient as any,
		fileManager: fileManager as any,
		settings: {} as any,
		timestamp: () => "2025-01-01T00:00:00Z",
		backupMode: "snapshots",
	};

	return { context, apiClient, fileManager };
};

describe("LocationHandler", () => {
	it("writes location file", async () => {
		const { context, fileManager } = createContext();
		const handler = new LocationHandler();

		await handler.pull("loc-1", context);

		expect(fileManager.writeFile).toHaveBeenCalledWith(
			"StoryFolder/worlds/eldoria/locations/crystal-cave.md",
			expect.stringContaining("# Crystal Cave")
		);
	});
});

