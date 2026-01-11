import { describe, expect, it } from "vitest";
import { RelationsGenerator } from "../RelationsGenerator";
import type { RelationsGeneratorInput } from "../../types/generators";

const input: RelationsGeneratorInput = {
	entity: {
		id: "story-1",
		name: "The Great Adventure",
		type: "story",
		worldId: "world-1",
		worldName: "Eldoria",
	},
	relations: [
		{
			targetType: "character",
			targetId: "john-smith",
			targetName: "John Smith",
			relationType: "pov",
			summary: "Scenes 1,2",
		},
		{
			targetType: "location",
			targetId: "crystal-mountains",
			targetName: "Crystal Mountains",
			relationType: "setting",
			summary: "Opening",
		},
	],
};

describe("RelationsGenerator", () => {
	it("groups relations per target type and adds placeholders", () => {
		const generator = new RelationsGenerator(() => "2025-01-01T00:00:00Z");
		const output = generator.generate(input);

		expect(output).toContain("type: story-relations");
		expect(output).toContain("## World");
		expect(output).toContain("## Main Characters");
		expect(output).toContain("[[john-smith|John Smith]] - Scenes 1,2");
		expect(output).toContain("_Add new main character: [[file|Name]] - description_");
	});
});

