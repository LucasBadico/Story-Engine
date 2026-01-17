import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import type { Character, World } from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import { CharacterHandler } from "../CharacterHandler";

const character: Character = {
	id: "char-1",
	tenant_id: "tenant-1",
	world_id: "world-1",
	class_level: 3,
	name: "Aria Moon",
	description: "Hero",
	created_at: "2025-01-01T00:00:00Z",
	updated_at: "2025-01-01T00:00:00Z",
	archetype_id: null,
	current_class_id: null,
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
		getCharacter: vi.fn().mockResolvedValue(character),
		getWorld: vi.fn().mockResolvedValue(world),
		deleteCharacter: vi.fn(),
		updateCharacter: vi.fn(),
	};
	const fileManager = {
		getWorldFolderPath: vi.fn().mockReturnValue("StoryFolder/worlds/eldoria"),
		ensureFolderExists: vi.fn(),
		writeFile: vi.fn(),
		readFile: vi.fn(),
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

describe("CharacterHandler", () => {
	it("writes character file", async () => {
		const { context, fileManager } = createContext();
		const handler = new CharacterHandler();

		await handler.pull("char-1", context);

		expect(fileManager.ensureFolderExists).toHaveBeenCalledWith(
			"StoryFolder/worlds/eldoria/characters"
		);
		expect(fileManager.writeFile).toHaveBeenCalledWith(
			"StoryFolder/worlds/eldoria/characters/aria-moon.md",
			expect.stringContaining("# Aria Moon")
		);
	});

	it("updates character when description changes", async () => {
		const { context, apiClient, fileManager } = createContext();
		const handler = new CharacterHandler();

		const localContent = `---
id: char-1
world_id: world-1
---

# Aria Moon

## Description
Updated description

## Metadata
- Tenant: tenant-1
`;
		vi.mocked(fileManager.readFile).mockResolvedValue(localContent);

		await handler.push(character, context);

		expect(apiClient.updateCharacter).toHaveBeenCalledWith("char-1", {
			name: "Aria Moon",
			description: "Updated description",
		});
	});

	it("does not update when nothing changes", async () => {
		const { context, apiClient, fileManager } = createContext();
		const handler = new CharacterHandler();

		const localContent = `---
id: char-1
world_id: world-1
---

# Aria Moon

## Description
Hero

## Metadata
- Tenant: tenant-1
`;
		vi.mocked(fileManager.readFile).mockResolvedValue(localContent);

		await handler.push(character, context);

		expect(apiClient.updateCharacter).not.toHaveBeenCalled();
	});
});

