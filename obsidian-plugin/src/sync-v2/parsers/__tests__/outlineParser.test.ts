import { describe, expect, it } from "vitest";
import { OutlineParser } from "../outlineParser";

const SAMPLE_OUTLINE = `
- [[ch-01-the-beginning|Chapter 1: The Beginning]] +
\t- [[sc-01-meet-hero|Scene 1: Meet the hero - Morning]] +
\t\t- [[bt-01-intro|Beat 1: Introduction]] +
\t\t- _New beat: intent here..._
`;

describe("OutlineParser", () => {
	const parser = new OutlineParser();

	it("parses entries with depth, status, and placeholders", () => {
		const entries = parser.parse(SAMPLE_OUTLINE);
		expect(entries).toHaveLength(4);

		const chapter = entries[0];
		expect(chapter.type).toBe("chapter");
		expect(chapter.status).toBe("has_content");
		expect(chapter.title).toBe("Chapter 1 The Beginning");
		expect(chapter.order).toBe(1);

		const beatPlaceholder = entries[3];
		expect(beatPlaceholder.type).toBe("beat");
		expect(beatPlaceholder.status).toBe("placeholder");
		expect(beatPlaceholder.placeholderLabel).toContain("intent here");
	});

	it("formats entries back into outline lines", () => {
		const entries = parser.parse(SAMPLE_OUTLINE);
		const line = parser.formatEntry(entries[1]);
		expect(line).toContain("[[sc-01-meet-hero|Scene 1 Meet the hero - Morning]]");
		expect(line.trim().endsWith("+")).toBeTruthy();
	});
});

