import { describe, expect, it } from "vitest";
import { RelationsParser } from "../relationsParser";

const parser = new RelationsParser();

const SAMPLE_RELATIONS = `---
id: story-1
type: story-relations
---

## Main Characters
- [[john-smith|John Smith]] - Protagonist
- _Add new character: [[file|Name]] - role_

## Locations
- [[crystal-mountains|Crystal Mountains]] - Opening scene
`;

describe("RelationsParser", () => {
	it("parses sections and entries including placeholders", () => {
		const parsed = parser.parse(SAMPLE_RELATIONS);
		expect(parsed.frontmatter.id).toBe("story-1");
		expect(parsed.sections).toHaveLength(2);

		const placeholder = parsed.sections[0].entries[1];
		expect(placeholder.placeholder).toBe(true);
		expect(placeholder.raw).toContain("_Add new character");
	});
});

