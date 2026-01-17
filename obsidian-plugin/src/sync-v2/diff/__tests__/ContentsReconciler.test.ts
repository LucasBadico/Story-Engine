import { describe, expect, it } from "vitest";
import { ContentsReconciler } from "../ContentsReconciler";

describe("ContentsReconciler", () => {
	it("returns generated content when no local file", () => {
		const reconciler = new ContentsReconciler();
		const result = reconciler.reconcile(null, "generated");
		expect(result.mergedContent).toBe("generated");
		expect(result.diff.operations).toEqual([]);
		expect(result.warnings).toEqual([]);
	});

	it("preserves untracked segments from local file", () => {
		const reconciler = new ContentsReconciler();
		const local = "Writer note";
		const generated = `
<!--content-start:0001:first:cb-1-->
Hello world
<!--content-end:0001:first:cb-1-->
`;
		const result = reconciler.reconcile(local, generated);
		expect(result.mergedContent).toContain("Writer note");
		expect(result.mergedContent).toContain("story-engine/untracked-start");
		expect(result.warnings).toHaveLength(1);
		expect(result.warnings?.[0].code).toBe("contents_untracked_segments");
	});
});

