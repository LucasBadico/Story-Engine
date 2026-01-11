import { describe, expect, it } from "vitest";
import { CitationsGenerator } from "../CitationsGenerator";
import type { CitationsGeneratorInput } from "../../types/generators";

const input: CitationsGeneratorInput = {
	entity: {
		id: "character-1",
		name: "John Smith",
		type: "character",
	},
	citations: [
		{
			storyId: "story-1",
			storyTitle: "The Great Adventure",
			relationType: "pov",
			sourceType: "scene",
			sourceId: "sc-1",
			sourceTitle: "Scene 1: Introduction",
			chapterTitle: "Chapter 1",
			summary: "First POV scene",
		},
	],
};

describe("CitationsGenerator", () => {
	it("groups citations by story and relation type", () => {
		const generator = new CitationsGenerator(() => "2025-01-01T00:00:00Z");
		const output = generator.generate(input);

		expect(output).toContain("type: character-citations");
		expect(output).toContain("## [[story-1|The Great Adventure]]");
		expect(output).toContain("### Pov (`relation_type: pov`)");
		expect(output).toContain("Scene 1: Introduction");
		expect(output).toContain("## Summary");
		expect(output).toContain("| The Great Adventure | pov | 1 |");
	});
});

