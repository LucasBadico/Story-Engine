import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import type { Trait } from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import { TraitHandler } from "../TraitHandler";

const trait: Trait = {
	id: "trait-1",
	tenant_id: "tenant-1",
	name: "Bravery",
	category: "virtue",
	description: "Fearless",
	created_at: "2025-01-01T00:00:00Z",
	updated_at: "2025-01-01T00:00:00Z",
};

const createContext = () => {
	const apiClient = {
		getTrait: vi.fn().mockResolvedValue(trait),
		deleteTrait: vi.fn(),
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

describe("TraitHandler", () => {
	it("writes trait file", async () => {
		const { context, fileManager } = createContext();
		const handler = new TraitHandler();

		await handler.pull("trait-1", context);

		expect(fileManager.writeFile).toHaveBeenCalledWith(
			"StoryFolder/worlds/characters/_traits/bravery.md",
			expect.stringContaining("# Bravery")
		);
	});
});

