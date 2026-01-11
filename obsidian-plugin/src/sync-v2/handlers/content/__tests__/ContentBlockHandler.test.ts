import { describe, expect, it, vi, beforeEach } from "vitest";
import type { App } from "obsidian";
import type { ContentBlock, StoryEngineSettings } from "../../../../types";
import type { SyncContext } from "../../../types/sync";
import { ContentBlockHandler } from "../ContentBlockHandler";
import * as detectEntityMentionsModule from "../../../utils/detectEntityMentions";
import * as contentBlockHelpersModule from "../../../utils/contentBlockHelpers";

const contentBlock: ContentBlock = {
	id: "cb-1",
	chapter_id: "ch-1",
	order_num: 1,
	type: "text",
	kind: "prose",
	content: "Sample text",
	metadata: {},
	created_at: "2025-01-01T00:00:00Z",
	updated_at: "2025-01-01T00:00:00Z",
};

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
	const apiClient = {
		getContentBlock: vi.fn().mockResolvedValue(contentBlock),
		getChapter: vi.fn().mockResolvedValue({ id: "ch-1", story_id: "story-1" }),
		getStory: vi.fn().mockResolvedValue({ id: "story-1", title: "Test Story" }),
		getContentAnchors: vi.fn().mockResolvedValue([]),
		updateContentBlock: vi.fn().mockResolvedValue(contentBlock),
		deleteContentBlock: vi.fn(),
	};
	const fileManager = {
		getStoryFolderPath: vi.fn().mockReturnValue("StoryFolder"),
		ensureFolderExists: vi.fn(),
		writeContentBlockFile: vi.fn(),
		readFile: vi.fn().mockResolvedValue(`---
id: cb-1
chapter_id: ch-1
order_num: 1
type: text
kind: prose
---

Sample text
`),
	};

	const context: SyncContext = {
		app: {} as App,
		apiClient: apiClient as any,
		fileManager: fileManager as any,
		settings,
		timestamp: () => "2025-01-01T00:00:00Z",
		backupMode: "snapshots",
		emitWarning: vi.fn(),
	};

	return { context, apiClient, fileManager };
};

describe("ContentBlockHandler", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("writes content block file when pulling block", async () => {
		const { context, apiClient, fileManager } = createContext();
		const handler = new ContentBlockHandler();

		await handler.pull("cb-1", context);

		expect(apiClient.getContentBlock).toHaveBeenCalledWith("cb-1");
		expect(fileManager.writeContentBlockFile).toHaveBeenCalledWith(
			contentBlock,
			"StoryFolder/03-contents/00-texts/cb-0001-prose.md",
			"Test Story"
		);
	});

	describe("push", () => {
		it("updates content block when content changes", async () => {
			const { context, apiClient, fileManager } = createContext();
			fileManager.readFile = vi.fn().mockResolvedValue(`---
id: cb-1
chapter_id: ch-1
order_num: 1
type: text
kind: prose
---

Updated content
`);

			vi.spyOn(detectEntityMentionsModule, "detectEntityMentions").mockReturnValue([]);
			vi.spyOn(contentBlockHelpersModule, "resolveContentBlockHierarchy").mockResolvedValue(null);

			const handler = new ContentBlockHandler();
			await handler.push(contentBlock, context);

			expect(fileManager.readFile).toHaveBeenCalled();
			expect(apiClient.updateContentBlock).toHaveBeenCalledWith("cb-1", {
				content: "Updated content",
			});
		});

		it("updates order_num when changed", async () => {
			const { context, apiClient, fileManager } = createContext();
			fileManager.readFile = vi.fn().mockResolvedValue(`---
id: cb-1
chapter_id: ch-1
order_num: 5
type: text
kind: prose
---

Sample text
`);

			vi.spyOn(detectEntityMentionsModule, "detectEntityMentions").mockReturnValue([]);
			vi.spyOn(contentBlockHelpersModule, "resolveContentBlockHierarchy").mockResolvedValue(null);

			const handler = new ContentBlockHandler();
			await handler.push(contentBlock, context);

			expect(apiClient.updateContentBlock).toHaveBeenCalledWith("cb-1", {
				content: "Sample text",
				order_num: 5,
			});
		});

		it("updates chapter_id when changed", async () => {
			const { context, apiClient, fileManager } = createContext();
			fileManager.readFile = vi.fn().mockResolvedValue(`---
id: cb-1
chapter_id: ch-2
order_num: 1
type: text
kind: prose
---

Sample text
`);

			vi.spyOn(detectEntityMentionsModule, "detectEntityMentions").mockReturnValue([]);
			vi.spyOn(contentBlockHelpersModule, "resolveContentBlockHierarchy").mockResolvedValue(null);

			const handler = new ContentBlockHandler();
			await handler.push(contentBlock, context);

			expect(apiClient.updateContentBlock).toHaveBeenCalledWith("cb-1", {
				content: "Sample text",
				chapter_id: "ch-2",
			});
		});

		it("handles null chapter_id", async () => {
			const { context, apiClient, fileManager } = createContext();
			fileManager.readFile = vi.fn().mockResolvedValue(`---
id: cb-1
chapter_id: null
order_num: 1
type: text
kind: prose
---

Sample text
`);

			vi.spyOn(detectEntityMentionsModule, "detectEntityMentions").mockReturnValue([]);
			vi.spyOn(contentBlockHelpersModule, "resolveContentBlockHierarchy").mockResolvedValue(null);

			const handler = new ContentBlockHandler();
			const blockWithoutChapter = { ...contentBlock, chapter_id: "ch-1" };
			await handler.push(blockWithoutChapter, context);

			expect(apiClient.updateContentBlock).toHaveBeenCalledWith("cb-1", {
				content: "Sample text",
				chapter_id: null,
			});
		});

		it("does not update if content unchanged", async () => {
			const { context, apiClient, fileManager } = createContext();
			fileManager.readFile = vi.fn().mockResolvedValue(`---
id: cb-1
chapter_id: ch-1
order_num: 1
type: text
kind: prose
---

Sample text
`);

			vi.spyOn(detectEntityMentionsModule, "detectEntityMentions").mockReturnValue([]);
			vi.spyOn(contentBlockHelpersModule, "resolveContentBlockHierarchy").mockResolvedValue(null);

			const handler = new ContentBlockHandler();
			await handler.push(contentBlock, context);

			// Should not update if only content changed but content is the same
			expect(apiClient.updateContentBlock).not.toHaveBeenCalled();
		});

		it("uses custom ID field from settings", async () => {
			const { context, apiClient, fileManager } = createContext();
			const customSettings = { ...settings, frontmatterIdField: "story_engine_id" };
			const customContext = { ...context, settings: customSettings };

			fileManager.readFile = vi.fn().mockResolvedValue(`---
story_engine_id: cb-1
chapter_id: ch-1
order_num: 1
type: text
kind: prose
---

Updated content
`);

			vi.spyOn(detectEntityMentionsModule, "detectEntityMentions").mockReturnValue([]);
			vi.spyOn(contentBlockHelpersModule, "resolveContentBlockHierarchy").mockResolvedValue(null);

			const handler = new ContentBlockHandler();
			await handler.push(contentBlock, customContext);

			expect(apiClient.updateContentBlock).toHaveBeenCalledWith("cb-1", {
				content: "Updated content",
			});
		});

		it("emits warning when file not found", async () => {
			const { context, apiClient, fileManager } = createContext();
			fileManager.readFile = vi.fn().mockRejectedValue(new Error("File not found"));

			const handler = new ContentBlockHandler();
			await handler.push(contentBlock, context);

			expect(context.emitWarning).toHaveBeenCalledWith(
				expect.objectContaining({
					code: "content_block_file_not_found",
				})
			);
			expect(apiClient.updateContentBlock).not.toHaveBeenCalled();
		});

		it("emits warning when ID mismatch", async () => {
			const { context, apiClient, fileManager } = createContext();
			fileManager.readFile = vi.fn().mockResolvedValue(`---
id: cb-wrong
chapter_id: ch-1
order_num: 1
type: text
kind: prose
---

Sample text
`);

			const handler = new ContentBlockHandler();
			await handler.push(contentBlock, context);

			expect(context.emitWarning).toHaveBeenCalledWith(
				expect.objectContaining({
					code: "content_block_id_mismatch",
				})
			);
			expect(apiClient.updateContentBlock).not.toHaveBeenCalled();
		});

		it("processes entity mentions and creates citations", async () => {
			const { context, apiClient, fileManager } = createContext();
			fileManager.readFile = vi.fn().mockResolvedValue(`---
id: cb-1
chapter_id: ch-1
order_num: 1
type: text
kind: prose
---

Text mentioning [[worlds/eldoria/characters/aria-moon]] and [[worlds/eldoria/locations/crystal-cave]].
`);

			const mockHierarchy = {
				contentBlockId: "cb-1",
				chapter: { id: "ch-1", title: "Introduction", number: 1 },
				story: { id: "story-1", title: "Test Story" },
			};

			vi.spyOn(detectEntityMentionsModule, "detectEntityMentions").mockReturnValue([
				{
					linkText: "[[worlds/eldoria/characters/aria-moon]]",
					filenamePath: "worlds/eldoria/characters/aria-moon",
					format: "official",
				},
				{
					linkText: "[[worlds/eldoria/locations/crystal-cave]]",
					filenamePath: "worlds/eldoria/locations/crystal-cave",
					format: "official",
				},
			]);

			vi.spyOn(detectEntityMentionsModule, "resolveEntityMention").mockResolvedValueOnce({
				entityId: "char-123",
				entityType: "character",
				worldId: "world-1",
			}).mockResolvedValueOnce({
				entityId: "loc-456",
				entityType: "location",
				worldId: "world-1",
			});

			vi.spyOn(contentBlockHelpersModule, "resolveContentBlockHierarchy").mockResolvedValue(
				mockHierarchy as any
			);
			vi.spyOn(contentBlockHelpersModule, "buildHierarchyContext").mockReturnValue(
				"Chapter 1: Introduction"
			);
			vi.spyOn(contentBlockHelpersModule, "createCitationRelations").mockResolvedValue({
				created: 2,
				errors: [],
			});

			const handler = new ContentBlockHandler();
			await handler.push(contentBlock, context);

			expect(detectEntityMentionsModule.detectEntityMentions).toHaveBeenCalledWith(
				expect.stringContaining("aria-moon")
			);
			expect(contentBlockHelpersModule.resolveContentBlockHierarchy).toHaveBeenCalledWith(
				"cb-1",
				apiClient
			);
			expect(contentBlockHelpersModule.createCitationRelations).toHaveBeenCalledWith(
				[
					{ entityId: "char-123", entityType: "character", worldId: "world-1" },
					{ entityId: "loc-456", entityType: "location", worldId: "world-1" },
				],
				mockHierarchy,
				"cb-1",
				apiClient,
				"Chapter 1: Introduction"
			);
		});

		it("handles errors during entity mention processing gracefully", async () => {
			const { context, apiClient, fileManager } = createContext();
			fileManager.readFile = vi.fn().mockResolvedValue(`---
id: cb-1
chapter_id: ch-1
order_num: 1
type: text
kind: prose
---

Text with [[invalid-link]].
`);

			vi.spyOn(detectEntityMentionsModule, "detectEntityMentions").mockReturnValue([
				{
					linkText: "[[invalid-link]]",
					filenamePath: "invalid-link",
					format: "obsidian",
				},
			]);

			vi.spyOn(detectEntityMentionsModule, "resolveEntityMention").mockResolvedValue(null);
			vi.spyOn(contentBlockHelpersModule, "resolveContentBlockHierarchy").mockResolvedValue(null);

			const handler = new ContentBlockHandler();
			await handler.push(contentBlock, context);

			// Should not throw, but may emit warnings
			expect(apiClient.updateContentBlock).toHaveBeenCalled();
		});

		it("resolves story via chapter_id", async () => {
			const { context, apiClient, fileManager } = createContext();
			fileManager.readFile = vi.fn().mockResolvedValue(`---
id: cb-1
chapter_id: ch-1
order_num: 1
type: text
kind: prose
---

Updated content
`);

			apiClient.getChapter = vi.fn().mockResolvedValue({
				id: "ch-1",
				story_id: "story-1",
				number: 1,
				title: "Chapter 1",
				status: "draft",
				created_at: "2025-01-01T00:00:00Z",
				updated_at: "2025-01-01T00:00:00Z",
			});
			apiClient.getStory = vi.fn().mockResolvedValue({
				id: "story-1",
				title: "Test Story",
				world_id: "world-1",
				tenant_id: "tenant-1",
				status: "draft",
				version: 1,
				root_story_id: "story-1",
				previous_version_id: null,
				created_at: "2025-01-01T00:00:00Z",
				updated_at: "2025-01-01T00:00:00Z",
			});

			vi.spyOn(detectEntityMentionsModule, "detectEntityMentions").mockReturnValue([]);
			vi.spyOn(contentBlockHelpersModule, "resolveContentBlockHierarchy").mockResolvedValue(null);

			const handler = new ContentBlockHandler();
			await handler.push(contentBlock, context);

			expect(apiClient.getChapter).toHaveBeenCalledWith("ch-1");
			expect(apiClient.getStory).toHaveBeenCalledWith("story-1");
		});

		it("resolves story via ContentAnchors when chapter_id is null", async () => {
			const blockWithoutChapter = { ...contentBlock, chapter_id: null };
			const { context, apiClient, fileManager } = createContext();
			fileManager.readFile = vi.fn().mockResolvedValue(`---
id: cb-1
chapter_id: null
order_num: 1
type: text
kind: prose
---

Updated content
`);

			apiClient.getContentAnchors = vi.fn().mockResolvedValue([
				{ id: "anchor-1", content_block_id: "cb-1", entity_type: "scene", entity_id: "sc-1" },
			]);
			(apiClient as any).getScene = vi.fn().mockResolvedValue({
				id: "sc-1",
				story_id: "story-1",
				chapter_id: "ch-1",
				order_num: 1,
				goal: "Test Scene",
				time_ref: "Morning",
				pov_character_id: null,
				location_id: null,
				created_at: "2025-01-01T00:00:00Z",
				updated_at: "2025-01-01T00:00:00Z",
			});
			apiClient.getStory = vi.fn().mockResolvedValue({
				id: "story-1",
				title: "Test Story",
				world_id: "world-1",
				tenant_id: "tenant-1",
				status: "draft",
				version: 1,
				root_story_id: "story-1",
				previous_version_id: null,
				created_at: "2025-01-01T00:00:00Z",
				updated_at: "2025-01-01T00:00:00Z",
			});

			vi.spyOn(detectEntityMentionsModule, "detectEntityMentions").mockReturnValue([]);
			vi.spyOn(contentBlockHelpersModule, "resolveContentBlockHierarchy").mockResolvedValue(null);

			const handler = new ContentBlockHandler();
			await handler.push(blockWithoutChapter, context);

			expect(apiClient.getContentAnchors).toHaveBeenCalledWith("cb-1");
			expect((apiClient as any).getScene).toHaveBeenCalledWith("sc-1");
			expect(apiClient.getStory).toHaveBeenCalledWith("story-1");
		});
	});

	it("deletes content block", async () => {
		const { context, apiClient } = createContext();
		const handler = new ContentBlockHandler();

		await handler.delete("cb-1", context);

		expect(apiClient.deleteContentBlock).toHaveBeenCalledWith("cb-1");
	});
});

