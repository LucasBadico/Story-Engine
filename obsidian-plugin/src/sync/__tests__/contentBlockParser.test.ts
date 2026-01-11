import { describe, expect, it } from "vitest";
import {
	compareContentBlocks,
	parseBeatHeader,
	parseBeatText,
	parseHierarchicalProse,
	parseSceneHeader,
	parseSceneText,
	ParsedParagraph,
} from "../contentBlockParser";
import { ContentBlock } from "../../types";

const buildContentBlock = (overrides: Partial<ContentBlock> = {}): ContentBlock => ({
	id: overrides.id ?? "cb-001",
	chapter_id: overrides.chapter_id ?? "ch-001",
	order_num: overrides.order_num ?? 1,
	type: overrides.type ?? "text",
	kind: overrides.kind ?? "prose",
	content: overrides.content ?? "Sample paragraph",
	metadata: overrides.metadata ?? {},
	created_at: overrides.created_at ?? "2025-01-01T00:00:00.000Z",
	updated_at: overrides.updated_at ?? "2025-01-01T00:00:00.000Z",
});

const buildParagraph = (content: string, linkName: string | null = null): ParsedParagraph => ({
	content,
	linkName,
	originalOrder: 0,
});

describe("scene and beat parsing helpers", () => {
	it("parses scene headers with wiki links and goal/time extraction", () => {
		const parsed = parseSceneHeader("[[sc-01-meet|Meet the hero - Morning]]");
		expect(parsed).toEqual({
			linkName: "sc-01-meet",
			goal: "Meet the hero",
			timeRef: "Morning",
			originalOrder: 0,
		});

		const fallback = parseSceneText("Investigate ruins");
		expect(fallback).toEqual({ goal: "Investigate ruins", timeRef: "" });
	});

	it("parses beat headers with plain text fallback and arrows", () => {
		const parsed = parseBeatHeader("[[bt-01|Interrogation -> Gains clue]]");
		expect(parsed).toEqual({
			linkName: "bt-01",
			intent: "Interrogation",
			outcome: "Gains clue",
			originalOrder: 0,
		});

		const fallback = parseBeatText("Search perimeter -> Finds tracks -> Raises stakes");
		expect(fallback).toEqual({
			intent: "Search perimeter",
			outcome: "Finds tracks -> Raises stakes",
		});
	});
});

describe("parseHierarchicalProse", () => {
	it("extracts scenes, beats, and prose blocks with preserved order", () => {
		const content = `---
id: ch-001
---

## Chapter 1: The Beginning

## Scene: [[sc-01|Meet the hero - Morning]]
[[cb-001|John arrives in town]]
Plain paragraph without link

### Beat: [[bt-01|Greeting -> Accepts call]]
[[cb-002|Dialogue with the mentor]]
`;

		const result = parseHierarchicalProse(content);
		expect(result.sections.map((section) => section.type)).toEqual(["scene", "prose", "prose", "beat", "prose"]);

		const scene = result.sections[0].scene;
		expect(scene).toMatchObject({
			linkName: "sc-01",
			goal: "Meet the hero",
			timeRef: "Morning",
		});

		const beat = result.sections[3].beat;
		expect(beat).toMatchObject({
			linkName: "bt-01",
			intent: "Greeting",
			outcome: "Accepts call",
		});

		const linkParagraph = result.sections[1].prose;
		expect(linkParagraph).toMatchObject({
			content: "John arrives in town",
			linkName: "cb-001",
		});

		const plaintextParagraph = result.sections[2].prose;
		expect(plaintextParagraph).toMatchObject({
			content: "Plain paragraph without link",
			linkName: null,
		});
	});
});

describe("compareContentBlocks", () => {
	it("marks paragraphs without link as new when nothing matches remotely", () => {
		const paragraph = buildParagraph("Fresh prose with no link");
		expect(compareContentBlocks(paragraph, null, null)).toBe("new");
	});

	it("treats matching local and remote blocks as unchanged", () => {
		const paragraph = buildParagraph("Shared prose", "cb-123");
		const local = buildContentBlock({ id: "cb-123", content: "Shared prose" });
		const remote = buildContentBlock({ id: "cb-123", content: "Shared prose" });

		expect(compareContentBlocks(paragraph, local, remote)).toBe("unchanged");
	});

	it("detects remote modifications when local matches paragraph content", () => {
		const paragraph = buildParagraph("Canonical prose", "cb-456");
		const local = buildContentBlock({ id: "cb-456", content: "Canonical prose" });
		const remote = buildContentBlock({ id: "cb-456", content: "Edited on server" });

		expect(compareContentBlocks(paragraph, local, remote)).toBe("remote_modified");
	});

	it("detects local modifications when paragraph diverges from synced state", () => {
		const paragraph = buildParagraph("Author edit", "cb-789");
		const local = buildContentBlock({ id: "cb-789", content: "Synced prose" });
		const remote = buildContentBlock({ id: "cb-789", content: "Synced prose" });

		expect(compareContentBlocks(paragraph, local, remote)).toBe("local_modified");
	});

	it("flags conflicts when paragraph, local, and remote all differ", () => {
		const paragraph = buildParagraph("Writer version", "cb-999");
		const local = buildContentBlock({ id: "cb-999", content: "Local draft" });
		const remote = buildContentBlock({ id: "cb-999", content: "Remote edit" });

		expect(compareContentBlocks(paragraph, local, remote)).toBe("conflict");
	});

	it("handles missing local block but matching remote content as remote_modified", () => {
		const paragraph = buildParagraph("API generated prose", "cb-222");
		const remote = buildContentBlock({ id: "cb-222", content: "API generated prose" });

		expect(compareContentBlocks(paragraph, null, remote)).toBe("remote_modified");
	});
});

