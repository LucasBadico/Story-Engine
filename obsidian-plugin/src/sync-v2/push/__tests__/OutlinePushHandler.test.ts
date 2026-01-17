import { describe, expect, it, vi } from "vitest";
import { OutlinePushHandler } from "../OutlinePushHandler";

describe("OutlinePushHandler", () => {
	it("detects chapter reorder", async () => {
		// Mock parser that returns pre-defined entries
		const mockParser = {
			parse: vi.fn().mockReturnValue([
				{
					type: "chapter",
					link: "StoryFolder/00-chapters/ch-0002-chapter-two.md",
					order: 1,
					title: "Chapter 2",
					status: "has_content",
					depth: 0,
					raw: "",
				},
				{
					type: "chapter",
					link: "StoryFolder/00-chapters/ch-0001-chapter-one.md",
					order: 2,
					title: "Chapter 1",
					status: "has_content",
					depth: 0,
					raw: "",
				},
			]),
		};

		const handler = new OutlinePushHandler(mockParser as any);

		const chapterFiles: Record<string, string> = {
			"ch-0002-chapter-two.md": "---\nid: ch-02\n---\n# Chapter 2",
			"ch-0001-chapter-one.md": "---\nid: ch-01\n---\n# Chapter 1",
		};

		const context = {
			apiClient: {
				getStory: vi.fn().mockResolvedValue({ id: "story-1", title: "My Story" }),
				getStoryWithHierarchy: vi.fn().mockResolvedValue({
					story: { id: "story-1", title: "My Story" },
					chapters: [
						{ chapter: { id: "ch-01", number: 1, title: "Chapter 1" }, scenes: [] },
						{ chapter: { id: "ch-02", number: 2, title: "Chapter 2" }, scenes: [] },
					],
				}),
			},
			fileManager: {
				readFile: vi.fn().mockImplementation((path: string) => {
					if (path.endsWith("story.outline.md")) {
						return Promise.resolve("outline content");
					}
					for (const [filename, content] of Object.entries(chapterFiles)) {
						if (path.endsWith(filename)) {
							return Promise.resolve(content);
						}
					}
					return Promise.reject(new Error(`File not found: ${path}`));
				}),
				getStoryFolderPath: vi.fn().mockReturnValue("StoryFolder"),
			},
			settings: { frontmatterIdField: "id" },
		} as any;

		const result = await handler.analyzeOutline("StoryFolder/story.outline.md", "story-1", context);

		// ch-02: remote number=2, local order=1 -> reorder
		expect(result.actions.length).toBeGreaterThanOrEqual(1);
		const ch02Action = result.actions.find(a => a.chapterId === "ch-02");
		expect(ch02Action).toMatchObject({
			type: "chapter_reorder",
			chapterId: "ch-02",
			oldOrder: 2,
			newOrder: 1,
		});
	});
});
