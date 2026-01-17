import { describe, expect, it } from "vitest";
import type { StoryWithHierarchy, ChapterWithContent } from "../../../types";
import { OutlineGenerator } from "../OutlineGenerator";

const story: StoryWithHierarchy = {
	story: {
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
	},
	chapters: [
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
						time_ref: "Morning",
						goal: "Meet the hero",
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
							outcome: "Reader meets hero",
							created_at: "2024-01-01",
							updated_at: "2024-01-01",
						},
					],
				},
			],
		},
	],
};

describe("OutlineGenerator", () => {
	it("generates hierarchy with placeholders", () => {
		const generator = new OutlineGenerator(() => "2025-01-01T00:00:00Z");
		const output = generator.generateStoryOutline(story, {
			storyFolderPath: "Stories/The Great Adventure",
		});

		expect(output).toContain("type: story-outline");
		expect(output).toContain("Chapter 1: Beginning");
		expect(output).toContain(
			"[[Stories/The Great Adventure/00-chapters/ch-0001-beginning.md|Chapter 1: Beginning]]"
		);
		expect(output).toContain("_New scene: goal - time_");
		expect(output).toContain("_New beat: intent here..._");
	});

	it("generates chapter outline", () => {
		const generator = new OutlineGenerator(() => "2025-01-01T00:00:00Z");
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
						time_ref: "Morning",
						goal: "Meet the hero",
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
							outcome: "Reader meets hero",
							created_at: "2024-01-01",
							updated_at: "2024-01-01",
						},
					],
				},
			],
		};

		const output = generator.generateChapterOutline(chapter, {
			storyFolderPath: "Stories/The Great Adventure",
		});

		expect(output).toContain("type: chapter-outline");
		expect(output).toContain("# Beginning");
		expect(output).toContain("## Hierarchy");
		expect(output).toContain("Scene 1: Meet the hero - Morning");
		expect(output).toContain("Beat 1: Introduction");
		expect(output).toContain("_New scene: goal - time_");
		expect(output).toContain("_New beat: intent here..._");
	});
});

