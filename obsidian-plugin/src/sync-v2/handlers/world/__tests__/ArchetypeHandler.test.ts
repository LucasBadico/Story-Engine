import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import type { Archetype } from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import { ArchetypeHandler } from "../ArchetypeHandler";

	const archetype: Archetype = {
		id: "arch-1",
		tenant_id: "tenant-1",
		name: "Warrior",
		description: "Front-line fighter",
		created_at: "2025-01-01T00:00:00Z",
		updated_at: "2025-01-01T00:00:00Z",
	};

const createContext = () => {
	const apiClient = {
		getArchetype: vi.fn().mockResolvedValue(archetype),
		deleteArchetype: vi.fn(),
	};
	const fileManager = {
		getWorldsRootPath: vi.fn().mockReturnValue("StoryFolder/worlds"),
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

describe("ArchetypeHandler", () => {
	it("writes archetype file", async () => {
		const { context, fileManager } = createContext();
		const handler = new ArchetypeHandler();

		await handler.pull("arch-1", context);

		expect(fileManager.writeFile).toHaveBeenCalledWith(
			"StoryFolder/worlds/characters/_archetypes/warrior.md",
			expect.stringContaining("# Warrior")
		);
	});
});

