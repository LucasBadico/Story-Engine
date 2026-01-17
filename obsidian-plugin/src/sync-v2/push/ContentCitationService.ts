import type { SyncContext } from "../types/sync";
import type { CreateEntityRelationInput } from "../../types";
import { detectEntityMentions, resolveEntityMention } from "../utils/detectEntityMentions";

export class ContentCitationService {
	constructor(private readonly context: SyncContext) {}

	async syncCitations(contentBlockId: string, content: string, worldId?: string): Promise<void> {
		const mentions = detectEntityMentions(content);
		const created = new Set<string>();

		for (const mention of mentions) {
			const resolved = await resolveEntityMention(mention, this.context);
			if (!resolved) {
				this.context.emitWarning?.({
					code: "citation_resolution_failed",
					message: `Não consegui resolver a menção ${mention.linkText}`,
					details: { mention },
					severity: "info",
				});
				continue;
			}

			const resolvedWorldId = worldId ?? resolved.worldId;
			if (!resolvedWorldId) {
				this.context.emitWarning?.({
					code: "citation_world_missing",
					message: `Não foi possível determinar o world_id para ${mention.linkText}`,
					details: { mention },
					severity: "warning",
				});
				continue;
			}

			const targetKey = `${resolved.entityType}:${resolved.entityId}`;
			if (created.has(targetKey)) {
				continue;
			}
			created.add(targetKey);

			const payload: CreateEntityRelationInput = {
				world_id: resolvedWorldId,
				source_type: "content_block",
				source_id: contentBlockId,
				target_type: resolved.entityType,
				target_id: resolved.entityId,
				relation_type: "citation",
				summary: mention.displayLabel ? `Referenciado como ${mention.displayLabel}` : undefined,
				create_mirror: true,
			};

			try {
				await this.context.apiClient.createEntityRelation(payload);
			} catch (error) {
				this.context.emitWarning?.({
					code: "citation_push_failed",
					message: `Não foi possível criar a citation para ${mention.linkText}`,
					details: error,
					severity: "warning",
				});
			}
		}
	}
}
