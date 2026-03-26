import { describe, expect, it, vi, beforeEach } from "vitest";
import type { App, TFile, Vault } from "obsidian";
import { BackupManager } from "../BackupManager";

describe("BackupManager", () => {
	let mockApp: App;
	let mockVault: Vault;
	let backupManager: BackupManager;

	beforeEach(() => {
		mockVault = {
			getAbstractFileByPath: vi.fn(),
			read: vi.fn(),
			create: vi.fn(),
			createFolder: vi.fn(),
		} as unknown as Vault;

		mockApp = {
			vault: mockVault,
		} as unknown as App;

		backupManager = new BackupManager(mockApp);
	});

	it("creates backup folder and copies files", async () => {
		const mockFile = {
			path: "stories/MyStory/story.md",
		} as TFile;

		vi.mocked(mockVault.getAbstractFileByPath).mockReturnValue(mockFile);
		vi.mocked(mockVault.read).mockResolvedValue("# My Story\n\nContent here");
		vi.mocked(mockVault.create).mockResolvedValue({} as TFile);
		vi.mocked(mockVault.createFolder).mockResolvedValue({} as any);

		const result = await backupManager.createBackup(["stories/MyStory/story.md"], "pull");

		expect(result.success).toBe(true);
		expect(result.filesCopied).toContain("stories/MyStory/story.md");
		expect(mockVault.create).toHaveBeenCalled();
	});

	it("creates manifest file", async () => {
		const mockFile = {
			path: "stories/MyStory/story.md",
		} as TFile;

		vi.mocked(mockVault.getAbstractFileByPath).mockReturnValue(mockFile);
		vi.mocked(mockVault.read).mockResolvedValue("content");
		vi.mocked(mockVault.create).mockResolvedValue({} as TFile);
		vi.mocked(mockVault.createFolder).mockResolvedValue({} as any);

		await backupManager.createBackup(["stories/MyStory/story.md"], "pull");

		const createCalls = vi.mocked(mockVault.create).mock.calls;
		const manifestCall = createCalls.find((call) => call[0].includes("manifest.json"));
		expect(manifestCall).toBeDefined();
	});

	it("handles missing files gracefully", async () => {
		vi.mocked(mockVault.getAbstractFileByPath).mockReturnValue(null);
		vi.mocked(mockVault.create).mockResolvedValue({} as TFile);
		vi.mocked(mockVault.createFolder).mockResolvedValue({} as any);

		const result = await backupManager.createBackup(["non-existent-file.md"], "pull");

		expect(result.success).toBe(true);
		expect(result.filesCopied).toHaveLength(0);
	});
});
