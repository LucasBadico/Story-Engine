import { describe, expect, it, vi } from "vitest";
import type { SyncContext } from "../../types/sync";
import { ContentCitationService } from "../ContentCitationService";

vi.mock("../../utils/detectEntityMentions", () => ({
	detectEntityMentions: vi.fn().mockReturnValue([
		{
			linkText: "[[john-doe]]",
			filenamePath: "characters/john-doe",
			displayLabel: "John Doe",
			format: "official",
		},
	]),
	resolveEntityMention: vi.fn().mockResolvedValue({
		entityId: "char-1",
		entityType: "character",
		worldId: "world-1",
	}),
}));

describe("ContentCitationService", () => {
	it("creates a relation for each resolved mention", async () => {
		const context = {
			apiClient: {
				createEntityRelation: vi.fn(),
			},
			app: {
				vault: {} as any,
				metadataCache: {} as any,
			},
			settings: {
				frontmatterIdField: "id",
				mode: "remote",
				tenantId: "tenant-1",
			},
		} as unknown as SyncContext;

		const service = new ContentCitationService(context);
		await service.syncCitations("cb-1", "Referencing [[john-doe]] in story", "world-1");

		expect(context.apiClient.createEntityRelation).toHaveBeenCalledWith(
			expect.objectContaining({
				source_type: "content_block",
				source_id: "cb-1",
				target_type: "character",
				target_id: "char-1",
				relation_type: "citation",
				world_id: "world-1",
			})
		);
	});
});
