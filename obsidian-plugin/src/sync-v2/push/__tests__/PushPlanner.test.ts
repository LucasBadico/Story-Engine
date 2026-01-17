import { describe, expect, it } from "vitest";
import { PushPlanner } from "../PushPlanner";

const planner = new PushPlanner();

const baseChapter = (id: string, body: string, order: number, name: string): string =>
	[
		`<!--chapter-start:${order.toString().padStart(4, "0")}:${name}:${id}-->`,
		body,
		`<!--chapter-end:${order.toString().padStart(4, "0")}:${name}:${id}-->`,
	].join("\n");

const sceneBlock = (id: string, sceneOrder: number, goal: string): string =>
	[
		`<!--scene-start:${sceneOrder.toString().padStart(4, "0")}:${goal}:${id}-->`,
		`Scene ${sceneOrder}`,
		`<!--scene-end:${sceneOrder.toString().padStart(4, "0")}:${goal}:${id}-->`,
	].join("\n");

const beatBlock = (id: string, order: number, intent: string): string =>
	[
		`<!--beat-start:${order.toString().padStart(4, "0")}:${intent}:${id}-->`,
		`Beat ${order}`,
		`<!--beat-end:${order.toString().padStart(4, "0")}:${intent}:${id}-->`,
	].join("\n");

describe("PushPlanner", () => {
	it("detects scene reorders", () => {
		const remote = baseChapter(
			"ch-1",
			[
				sceneBlock("sc-1", 1, "scene-a"),
				sceneBlock("sc-2", 2, "scene-b"),
			].join("\n"),
			1,
			"chapter-one"
		);

		const local = baseChapter(
			"ch-1",
			[
				sceneBlock("sc-2", 1, "scene-b"),
				sceneBlock("sc-1", 2, "scene-a"),
			].join("\n"),
			1,
			"chapter-one"
		);

		const plan = planner.buildPlan(remote, local);
		const reorder = plan.actions.filter((action) => action.type === "scene_reorder");

		expect(reorder).toHaveLength(2);
		expect(reorder).toEqual(
			expect.arrayContaining([
				expect.objectContaining({ sceneId: "sc-1", newOrder: 2 }),
				expect.objectContaining({ sceneId: "sc-2", newOrder: 1 }),
			])
		);
		expect(plan.unsupportedOperations).toHaveLength(0);
	});

	it("detects scene move between chapters", () => {
		const remote = [
			baseChapter("ch-1", sceneBlock("sc-1", 1, "scene-a"), 1, "chapter-one"),
			baseChapter("ch-2", sceneBlock("sc-2", 1, "scene-b"), 2, "chapter-two"),
		].join("\n");

		const local = [
			baseChapter("ch-1", "", 1, "chapter-one"),
			baseChapter(
				"ch-2",
				[sceneBlock("sc-2", 1, "scene-b"), sceneBlock("sc-1", 2, "scene-a")].join("\n"),
				2,
				"chapter-two"
			),
		].join("\n");

		const plan = planner.buildPlan(remote, local);
		const move = plan.actions.find((action) => action.type === "scene_move");

		expect(move).toBeDefined();
		expect(move).toMatchObject({
			type: "scene_move",
			sceneId: "sc-1",
			toChapterId: "ch-2",
		});
	});

	it("detects beat move and reorder", () => {
		const remote = baseChapter(
			"ch-1",
				sceneBlock("sc-1", 1, "scene-a").replace(
				"Scene 1",
				[
					beatBlock("bt-1", 1, "beat-a"),
					beatBlock("bt-2", 2, "beat-b"),
				].join("\n")
			),
			1,
			"chapter-one"
		);

		const local = baseChapter(
			"ch-1",
				sceneBlock("sc-1", 1, "scene-a").replace(
				"Scene 1",
				[
					beatBlock("bt-2", 1, "beat-b"),
					beatBlock("bt-1", 2, "beat-a"),
				].join("\n")
			),
			1,
			"chapter-one"
		);

		const plan = planner.buildPlan(remote, local);
		expect(plan.actions.some((action) => action.type === "beat_reorder")).toBe(true);
	});

	it("collects unsupported operations", () => {
		const remote = "";
		const local = sceneBlock("new-id", 1, "scene-new");

		const plan = planner.buildPlan(remote, local);
		expect(plan.actions).toHaveLength(0);
		expect(plan.unsupportedOperations.length).toBeGreaterThan(0);
		expect(plan.warnings).toEqual(
			expect.arrayContaining([
				expect.objectContaining({ code: "push_unsupported_operations" }),
			])
		);
	});

	it("adds warning when untracked segments exist", () => {
		const remote = `
<!--scene-start:0001:scene-a:sc-1-->
Scene A
<!--scene-end:0001:scene-a:sc-1-->
`;
		const local = `Writer note\n${remote}`;
		const plan = planner.buildPlan(remote, local);
		expect(plan.warnings).toEqual(
			expect.arrayContaining([
				expect.objectContaining({ code: "push_untracked_segments" }),
			])
		);
	});

	it("detects content text updates", () => {
		const remote = `
<!--content-start:0001:intro:cb-1-->
Original text here
<!--content-end:0001:intro:cb-1-->
`;
		const local = `
<!--content-start:0001:intro:cb-1-->
Updated paragraph from user
<!--content-end:0001:intro:cb-1-->
`;

		const plan = planner.buildPlan(remote, local);
		expect(plan.actions).toContainEqual({
			type: "content_update",
			contentBlockId: "cb-1",
			newContent: "Updated paragraph from user",
		});
	});
});

