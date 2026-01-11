import { describe, expect, it } from "vitest";
import { DiffEngine } from "../DiffEngine";

const BASE_CONTENT = `
<!--content-start:0001:first:cb-1-->
Hello world
<!--content-end:0001:first:cb-1-->
<!--content-start:0002:second:cb-2-->
Second block
<!--content-end:0002:second:cb-2-->
`;

describe("DiffEngine", () => {
	const engine = new DiffEngine();

	it("detects new fences between versions", () => {
		const local = BASE_CONTENT;
		const remote = `
<!--content-start:0001:first:cb-1-->
Hello world!
<!--content-end:0001:first:cb-1-->
<!--content-start:0002:second:cb-2-->
Second block
<!--content-end:0002:second:cb-2-->
<!--content-start:0003:third:cb-3-->
New item
<!--content-end:0003:third:cb-3-->
`;

		const diff = engine.diffContents(local, remote);
		expect(
			diff.operations.some((op) => op.kind === "created" && op.fenceId === "cb-3")
		).toBe(true);
	});

	it("captures updates to existing fences", () => {
		const local = BASE_CONTENT;
		const remote = `
<!--content-start:0001:first:cb-1-->
Hello world!!!
<!--content-end:0001:first:cb-1-->
<!--content-start:0002:second:cb-2-->
Second block
<!--content-end:0002:second:cb-2-->
`;
		const diff = engine.diffContents(local, remote);
		expect(diff.operations.some((op) => op.kind === "updated" && op.fenceId === "cb-1")).toBe(
			true
		);
	});

	it("detects reorder operations", () => {
		const local = BASE_CONTENT;
		const remote = `
<!--content-start:0001:second:cb-2-->
Second block
<!--content-end:0001:second:cb-2-->
<!--content-start:0002:first:cb-1-->
Hello world
<!--content-end:0002:first:cb-1-->
`;
		const diff = engine.diffContents(local, remote);
		expect(
			diff.operations.some((op) => op.kind === "reordered" && op.fenceId === "cb-1")
		).toBe(true);
	});

	it("detects moves between parent fences", () => {
		const local = `
<!--scene-start:0001:scene-a:sc-1-->
Scene A
<!--content-start:0001:first:cb-1-->
Hello world
<!--content-end:0001:first:cb-1-->
<!--scene-end:0001:scene-a:sc-1-->
<!--scene-start:0002:scene-b:sc-2-->
Scene B
<!--scene-end:0002:scene-b:sc-2-->
`;
		const remote = `
<!--scene-start:0001:scene-a:sc-1-->
Scene A
<!--scene-end:0001:scene-a:sc-1-->
<!--scene-start:0002:scene-b:sc-2-->
Scene B
<!--content-start:0001:first:cb-1-->
Hello world
<!--content-end:0001:first:cb-1-->
<!--scene-end:0002:scene-b:sc-2-->
`;
		const diff = engine.diffContents(local, remote);
		expect(diff.operations.some((op) => op.kind === "moved" && op.fenceId === "cb-1")).toBe(
			true
		);
	});

	it("handles unknown segments without deleting them", () => {
		const local = `Editor note\n${BASE_CONTENT}`;
		const diff = engine.diffContents(local, BASE_CONTENT);
		expect(diff.operations.length).toBe(0);
		expect(diff.untrackedSegments).toContain("Editor note");
	});

	it("detects deletions", () => {
		const local = BASE_CONTENT;
		const remote = `
<!--content-start:0001:first:cb-1-->
Hello world
<!--content-end:0001:first:cb-1-->
`;
		const diff = engine.diffContents(local, remote);
		expect(diff.operations.some((op) => op.kind === "deleted" && op.fenceId === "cb-2")).toBe(
			true
		);
	});

	it("detects placeholder conversions", () => {
		const local = `
<!--content-start:0000:new-content:placeholder-->
_Write your content here..._
<!--content-end:0000:new-content:placeholder-->
`;
		const remote = `
<!--content-start:0001:first:cb-1-->
Hello world
<!--content-end:0001:first:cb-1-->
`;
		const diff = engine.diffContents(local, remote);
		expect(diff.operations.some((op) => op.kind === "created" && op.fenceId === "cb-1")).toBe(
			true
		);
	});
});

