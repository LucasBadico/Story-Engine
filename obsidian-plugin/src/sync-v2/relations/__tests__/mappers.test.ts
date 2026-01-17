import { describe, expect, it } from "vitest";
import { mapRelationsToGeneratorInput, mapCitationsToGeneratorInput } from "../mappers";
import type { EntityRelation } from "../../types/relations";

const RELATION: EntityRelation = {
	id: "rel-1",
	tenant_id: "tenant-1",
	source_type: "story",
	source_id: "story-1",
	target_type: "character",
	target_id: "char-1",
	relation_type: "pov",
	context: "Scenes 1,2",
	created_at: "2025-01-01T00:00:00Z",
	updated_at: "2025-01-01T00:00:00Z",
	direction: "source",
};

describe("relations mappers", () => {
	it("maps relations to generator input using resolver", () => {
		const input = mapRelationsToGeneratorInput({
			entity: { id: "story-1", name: "The Great Adventure", type: "story", worldId: "world-1", worldName: "Eldoria" },
			relations: [RELATION],
			resolveTarget: () => ({
				targetId: "char-1",
				targetName: "John Smith",
				contextLabel: "Scene 1",
			}),
			options: { syncedAt: "2025-01-02T00:00:00Z" },
		});

		expect(input.entity.worldName).toBe("Eldoria");
		expect(input.relations).toHaveLength(1);
		expect(input.relations[0]).toMatchObject({
			targetType: "character",
			targetName: "John Smith",
			contextLabel: "Scene 1",
		});
	});

	it("maps relations to citations input using source resolver", () => {
		const citationRelation: EntityRelation = {
			...RELATION,
			source_type: "scene",
			source_id: "sc-1",
			target_type: "character",
			target_id: "char-1",
			direction: "target",
		};

		const input = mapCitationsToGeneratorInput({
			entity: { id: "char-1", name: "John Smith", type: "character" },
			relations: [citationRelation],
			resolveSource: () => ({
				storyId: "story-1",
				storyTitle: "The Great Adventure",
				sourceTitle: "Scene 1: Arrival",
				sourceType: "scene",
				chapterTitle: "Chapter 1",
			}),
		});

		expect(input.citations).toHaveLength(1);
		expect(input.citations[0]).toMatchObject({
			storyTitle: "The Great Adventure",
			sourceType: "scene",
			sourceTitle: "Scene 1: Arrival",
			chapterTitle: "Chapter 1",
		});
	});
});

