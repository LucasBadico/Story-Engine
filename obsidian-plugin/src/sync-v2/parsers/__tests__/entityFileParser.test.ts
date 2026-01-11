import { describe, expect, it } from "vitest";
import { EntityFileParser } from "../entityFileParser";

const parser = new EntityFileParser();

const SAMPLE_ENTITY = `---
id: sc-01
title: Scene 1
goal: Meet the hero
---

# Scene 1: Meet the hero

## Notes
Some notes here.

## Beats
- [[bt-01|Introduction]]
`;

describe("EntityFileParser", () => {
	it("parses frontmatter and headings", () => {
		const parsed = parser.parse(SAMPLE_ENTITY);
		expect(parsed.frontmatter.id).toBe("sc-01");
		expect(parsed.headings).toHaveLength(3);
		expect(parser.getSectionContent(parsed, "Notes")).toContain("Some notes");
	});

	it("updates frontmatter values", () => {
		const updated = parser.updateFrontmatter(SAMPLE_ENTITY, { status: "draft" });
		const parsed = parser.parse(updated);
		expect(parsed.frontmatter.status).toBe("draft");
	});
});

