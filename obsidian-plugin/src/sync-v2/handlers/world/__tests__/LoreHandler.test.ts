import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import type { Lore, World } from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import { LoreHandler } from "../LoreHandler";

const lore: Lore = {
	id: "lore-1",
	tenant_id: "tenant-1",
	world_id: "world-1",
	name: "Ancient Magic",
	category: "magic",
	description: "",
	rules: "",
	limitations: "",
	requirements: "",
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
		getLore: vi.fn().mockResolvedValue(lore),
		getWorld: vi.fn().mockResolvedValue(world),
		deleteLore: vi.fn(),
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

describe("LoreHandler", () => {
	it("writes lore file", async () => {
		const { context, fileManager } = createContext();
		const handler = new LoreHandler();

		await handler.pull("lore-1", context);

		expect(fileManager.writeFile).toHaveBeenCalledWith(
			"StoryFolder/worlds/eldoria/lore/ancient-magic.md",
			expect.stringContaining("# Ancient Magic")
		);
	});
});

