import { describe, expect, it, vi } from "vitest";
import type { App } from "obsidian";
import type { SyncContext } from "../../../sync-v2/types/sync";
import type { StoryEngineSettings } from "../../../types";
import { FileRenamer } from "../FileRenamer";

const settings: StoryEngineSettings = {
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

const createContext = () => {
	const files = new Map<string, string>();
	const fileManager = {
		renameFile: vi.fn(async (oldPath: string, newPath: string) => {
			if (!files.has(oldPath)) {
				throw new Error(`File not found: ${oldPath}`);
			}
			const content = files.get(oldPath)!;
			files.delete(oldPath);
			files.set(newPath, content);
		}),
		readFile: vi.fn(async (path: string) => {
			if (!files.has(path)) {
				throw new Error(`File not found: ${path}`);
			}
			return files.get(path)!;
		}),
		writeFile: vi.fn(async (path: string, content: string) => {
			files.set(path, content);
		}),
		getVault: vi.fn(),
	};

	const context: SyncContext = {
		app: {} as App,
		apiClient: {} as any,
		fileManager: fileManager as any,
		settings,
		timestamp: () => "2025-01-01T00:00:00Z",
		backupMode: "snapshots",
	};

	return { context, files, fileManager };
};

describe("FileRenamer", () => {
	it("renames file and updates references", async () => {
		const { context, files, fileManager } = createContext();
		files.set("StoryFolder/01-scenes/sc-01-old.md", "scene content");
		files.set("StoryFolder/story.outline.md", "Link [[sc-01-old|Scene]] end");

		const renamer = new FileRenamer(context);
		const result = await renamer.rename({
			oldPath: "StoryFolder/01-scenes/sc-01-old.md",
			newPath: "StoryFolder/01-scenes/sc-02-new.md",
			references: [
				{
					filePath: "StoryFolder/story.outline.md",
					replacements: [
						{
							pattern: /\[\[sc-01-old\|/g,
							replacement: "[[sc-02-new|",
						},
					],
				},
			],
		});

		expect(result).toEqual({
			oldPath: "StoryFolder/01-scenes/sc-01-old.md",
			newPath: "StoryFolder/01-scenes/sc-02-new.md",
			updatedReferences: 1,
		});
		expect(files.get("StoryFolder/01-scenes/sc-02-new.md")).toBe("scene content");
		expect(files.get("StoryFolder/story.outline.md")).toBe("Link [[sc-02-new|Scene]] end");
		expect(fileManager.renameFile).toHaveBeenCalledTimes(1);
	});

	it("returns zero updated references when no replacements occur", async () => {
		const { context, files } = createContext();
		files.set("folder/old.md", "content");
		files.set("folder/file.md", "no matching link");

		const renamer = new FileRenamer(context);
		const result = await renamer.rename({
			oldPath: "folder/old.md",
			newPath: "folder/new.md",
			references: [
				{
					filePath: "folder/file.md",
					replacements: [
						{
							pattern: /\[\[missing\|/g,
							replacement: "[[new|",
						},
					],
				},
			],
		});

		expect(result.updatedReferences).toBe(0);
		expect(files.get("folder/file.md")).toBe("no matching link");
	});
});

