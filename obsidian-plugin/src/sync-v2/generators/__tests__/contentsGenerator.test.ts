import { describe, expect, it } from "vitest";
import type { ChapterWithContent, ContentBlock, Story } from "../../../types";
import { ContentsGenerator } from "../ContentsGenerator";

const story: Story = {
	id: "story-1",
	tenant_id: "tenant-1",
	title: "The Great Adventure",
	status: "draft",
	version_number: 1,
	root_story_id: "story-1",
	previous_story_id: null,
	world_id: "world-1",
	created_by_user_id: "user-1",
	created_at: "2024-01-01",
	updated_at: "2024-01-01",
};

const chapters: ChapterWithContent[] = [
	{
		chapter: {
			id: "ch-1",
			story_id: "story-1",
			number: 1,
			title: "Beginning",
			status: "draft",
			created_at: "2024-01-01",
			updated_at: "2024-01-01",
		},
		scenes: [
			{
				scene: {
					id: "sc-1",
					story_id: "story-1",
					chapter_id: "ch-1",
					order_num: 1,
					pov_character_id: null,
					location_id: null,
					time_ref: "",
					goal: "Meet hero",
					created_at: "2024-01-01",
					updated_at: "2024-01-01",
				},
				beats: [
					{
						id: "bt-1",
						scene_id: "sc-1",
						order_num: 1,
						type: "exposition",
						intent: "Introduction",
						outcome: "Hero introduced",
						created_at: "2024-01-01",
						updated_at: "2024-01-01",
					},
				],
			},
		],
	},
];

const textBlock = (overrides: Partial<ContentBlock> = {}): ContentBlock => ({
	id: overrides.id ?? "cb-1",
	chapter_id: overrides.chapter_id ?? "ch-1",
	order_num: overrides.order_num ?? 1,
	type: "text",
	kind: overrides.kind ?? "prose",
	content: overrides.content ?? "Sample paragraph",
	metadata: overrides.metadata ?? {},
	created_at: "2024-01-01",
	updated_at: "2024-01-01",
});

describe("ContentsGenerator", () => {
	it("generates fences for chapters/scenes/beats and placeholders", () => {
		const generator = new ContentsGenerator(() => "2025-01-01T00:00:00Z");
		const chapterBlocks = new Map<string, ContentBlock[]>([["ch-1", [textBlock()]]]);
		const sceneBlocks = new Map<string, ContentBlock[]>([["sc-1", [textBlock({ id: "cb-2" })]]]);
		const beatBlocks = new Map<string, ContentBlock[]>([["bt-1", [textBlock({ id: "cb-3" })]]]);

		const output = generator.generateStoryContents({
			story,
			chapters,
			chapterContentBlocks: chapterBlocks,
			sceneContentBlocks: sceneBlocks,
			beatContentBlocks: beatBlocks,
		});

		expect(output).toContain("type: story-contents");
		expect(output).toContain("<!--chapter-start");
		expect(output).toContain("<!--scene-start");
		expect(output).toContain("<!--beat-start");
		expect(output).toContain("Sample paragraph");
	});

	it("generates chapter contents", () => {
		const generator = new ContentsGenerator(() => "2025-01-01T00:00:00Z");
		const chapter: ChapterWithContent = {
			chapter: {
				id: "ch-1",
				story_id: "story-1",
				number: 1,
				title: "Beginning",
				status: "draft",
				created_at: "2024-01-01",
				updated_at: "2024-01-01",
			},
			scenes: [
				{
					scene: {
						id: "sc-1",
						story_id: "story-1",
						chapter_id: "ch-1",
						order_num: 1,
						pov_character_id: null,
						location_id: null,
						time_ref: "",
						goal: "Meet hero",
						created_at: "2024-01-01",
						updated_at: "2024-01-01",
					},
					beats: [
						{
							id: "bt-1",
							scene_id: "sc-1",
							order_num: 1,
							type: "exposition",
							intent: "Introduction",
							outcome: "Hero introduced",
							created_at: "2024-01-01",
							updated_at: "2024-01-01",
						},
					],
				},
			],
		};

		const chapterBlocks = new Map<string, ContentBlock[]>([["ch-1", [textBlock()]]]);
		const sceneBlocks = new Map<string, ContentBlock[]>([["sc-1", [textBlock({ id: "cb-2" })]]]);
		const beatBlocks = new Map<string, ContentBlock[]>([["bt-1", [textBlock({ id: "cb-3" })]]]);

		const output = generator.generateChapterContents(
			chapter,
			chapterBlocks,
			sceneBlocks,
			beatBlocks
		);

		expect(output).toContain("type: chapter-contents");
		expect(output).toContain("# Beginning - Contents");
		expect(output).toContain("Sample paragraph");
		expect(output).toContain("<!--scene-start");
		expect(output).toContain("<!--beat-start");
	});
});

