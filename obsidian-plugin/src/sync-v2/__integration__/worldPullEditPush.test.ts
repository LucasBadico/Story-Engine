import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import type { Character, StoryEngineSettings, World } from "../../types";
import type { SyncContext } from "../types/sync";
import { SyncOrchestrator } from "../core/SyncOrchestrator";
import { CharacterHandler } from "../handlers/world/CharacterHandler";
import { MockFileManager } from "./helpers/mockFileManager";

const settings: StoryEngineSettings = {
	apiUrl: "",
	llmGatewayUrl: "",
	apiKey: "",
	tenantId: "",
	tenantName: "",
	syncFolderPath: "StoryFolder",
	autoVersionSnapshots: true,
	conflictResolution: "local",
	mode: "local",
	syncVersion: "v2",
	showHelpBox: true,
	autoSyncOnApiUpdates: true,
	autoPushOnFileBlur: true,
	backupMode: "off",
	backupRetentionDays: 7,
};

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

describe("Sync V2 integration - world pull/edit/push", () => {
	it("updates character when local description changes", async () => {
		const updatedCharacterFile = readFileSync(
			resolve(__dirname, "fixtures/world/character.md"),
			"utf8"
		);

		const apiClient = {
			getWorld: vi.fn().mockResolvedValue(world),
			getCharacters: vi.fn().mockResolvedValue([character]),
			getLocations: vi.fn().mockResolvedValue([]),
			getFactions: vi.fn().mockResolvedValue([]),
			getArtifacts: vi.fn().mockResolvedValue([]),
			getEvents: vi.fn().mockResolvedValue([]),
			getLores: vi.fn().mockResolvedValue([]),
			getCharacter: vi.fn().mockResolvedValue(character),
			updateCharacter: vi.fn().mockResolvedValue({}),
			listRelationsByWorld: vi.fn().mockResolvedValue({ data: [], pagination: { has_more: false } }),
		};

		const fileManager = new MockFileManager("StoryFolder");
		const context: SyncContext = {
			app: {} as App,
			apiClient: apiClient as any,
			fileManager: fileManager as any,
			settings,
			timestamp: () => "2025-01-10T00:00:00Z",
			backupMode: "off",
		};

		const orchestrator = new SyncOrchestrator(context);

		await orchestrator.run({ type: "pull_world", payload: { worldId: "world-1" } });
		await orchestrator.run({ type: "pull_character", payload: { entityId: "char-1" } });

		const characterPath = `${fileManager.getWorldFolderPath(world.name)}/characters/aria-moon.md`;
		fileManager.setFile(characterPath, updatedCharacterFile);

		const handler = new CharacterHandler();
		await handler.push(character, context);

		expect(apiClient.updateCharacter).toHaveBeenCalledWith("char-1", {
			name: "Aria Moon",
			description: "Updated description",
		});
	});
});
