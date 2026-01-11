import { describe, expect, it } from "vitest";
import { ContentsParser, PLACEHOLDER_DEFAULTS } from "../contentsParser";

const parser = new ContentsParser();

const SAMPLE_CONTENT = `
<!--chapter-start:0001:the-beginning:ch-uuid-1-->
## Chapter 1
<!--scene-start:0001:meet-hero:sc-uuid-1-->
### Scene 1
<!--content-start:0001:opening:cb-uuid-1-->
The hero arrives.
<!--content-end:0001:opening:cb-uuid-1-->
<!--scene-end:0001:meet-hero:sc-uuid-1-->
<!--chapter-end:0001:the-beginning:ch-uuid-1-->
`;

describe("ContentsParser", () => {
	it("parses hierarchy and returns nested fences", () => {
		const hierarchy = parser.parseHierarchy(SAMPLE_CONTENT);
		expect(hierarchy.chapters).toHaveLength(1);
		expect(hierarchy.chapters[0].children[0].children[0].innerText).toContain("hero arrives");
	});

	it("detects changes between contents", () => {
		const updated = SAMPLE_CONTENT.replace("The hero arrives.", "The hero arrives. Changed.");
		const changes = parser.detectChanges(SAMPLE_CONTENT, updated);
		expect(changes.some((c) => c.changeType === "updated")).toBeTruthy();
	});

	it("replaces placeholders and generates new fences", () => {
		const placeholder = parser.generatePlaceholder("scene");
		const withPlaceholder = `${SAMPLE_CONTENT}\n${placeholder}`;
		const replaced = parser.replacePlaceholder(withPlaceholder, withPlaceholder.indexOf("placeholder"), {
			type: "scene",
			order: 2,
			name: "new-scene",
			id: "sc-uuid-2",
			content: PLACEHOLDER_DEFAULTS.scene,
		});
		expect(replaced).toContain("sc-uuid-2");
	});
});

