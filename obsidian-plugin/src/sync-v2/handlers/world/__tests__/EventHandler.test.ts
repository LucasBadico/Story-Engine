import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import type { World, WorldEvent } from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import { EventHandler } from "../EventHandler";

const worldEvent: WorldEvent = {
	id: "evt-1",
	tenant_id: "tenant-1",
	world_id: "world-1",
	name: "Cataclysm",
	type: "disaster",
	description: "",
	importance: 5,
	parent_id: null,
	timeline: "Age of Dawn",
	is_epoch: false,
	timeline_position: 1,
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
		getEvent: vi.fn().mockResolvedValue(worldEvent),
		getWorld: vi.fn().mockResolvedValue(world),
		deleteEvent: vi.fn(),
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

describe("EventHandler", () => {
	it("writes event file", async () => {
		const { context, fileManager } = createContext();
		const handler = new EventHandler();

		await handler.pull("evt-1", context);

		expect(fileManager.writeFile).toHaveBeenCalledWith(
			"StoryFolder/worlds/eldoria/events/cataclysm.md",
			expect.stringContaining("# Cataclysm")
		);
	});
});

