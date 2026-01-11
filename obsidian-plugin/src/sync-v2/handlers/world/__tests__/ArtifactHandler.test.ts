import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import type { Artifact, World } from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import { ArtifactHandler } from "../ArtifactHandler";

const artifact: Artifact = {
	id: "art-1",
	tenant_id: "tenant-1",
	world_id: "world-1",
	name: "Sunblade",
	description: "Legendary sword",
	rarity: "mythic",
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
		getArtifact: vi.fn().mockResolvedValue(artifact),
		getWorld: vi.fn().mockResolvedValue(world),
		deleteArtifact: vi.fn(),
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

describe("ArtifactHandler", () => {
	it("writes artifact file", async () => {
		const { context, fileManager } = createContext();
		const handler = new ArtifactHandler();

		await handler.pull("art-1", context);

		expect(fileManager.writeFile).toHaveBeenCalledWith(
			"StoryFolder/worlds/eldoria/artifacts/sunblade.md",
			expect.stringContaining("# Sunblade")
		);
	});
});

