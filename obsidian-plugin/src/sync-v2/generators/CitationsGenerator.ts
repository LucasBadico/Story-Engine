import type { CitationsGeneratorInput } from "../types/generators";
import { getIdFieldName } from "../utils/frontmatterHelpers";

export class CitationsGenerator {
	constructor(private readonly now: () => string = () => new Date().toISOString()) {}

	generate(input: CitationsGeneratorInput): string {
		const { entity, citations } = input;
		const lines: string[] = [];

		// Use configured ID field name, default to "id"
		const idField = getIdFieldName(input.options?.idField);

		lines.push(
			"---",
			`${idField}: ${entity.id}`,
			`type: ${entity.type}-citations`,
			`synced_at: ${input.options?.syncedAt ?? this.now()}`,
			"---",
			"",
			`# ${entity.name} - Citations`,
			""
		);

		lines.push(
			"> [!warning] ⚠️ Arquivo auto-gerado - NÃO EDITE",
			"> Este arquivo é atualizado automaticamente durante o sync.",
			">",
			"> **Para adicionar citações**: Atualize os arquivos `.relations.md` relevantes.",
			""
		);

		const grouped = this.groupByStory(citations);
		for (const [storyId, entries] of grouped.entries()) {
			const storyLabel = entries[0].storyTitle;
			const storyPath = entries[0].storyPath ?? storyId;
			lines.push(`## [[${storyPath}|${storyLabel}]]`, "");

			const byRelation = this.groupByRelationType(entries);
			for (const [relationType, relationEntries] of byRelation.entries()) {
				lines.push(`### ${this.titleCase(relationType)} (\`relation_type: ${relationType}\`)`);
				relationEntries.forEach((entry) => {
					const context = entry.chapterTitle ? ` (Chapter: ${entry.chapterTitle})` : "";
					const sourcePath = entry.sourcePath ?? entry.sourceId;
					lines.push(
						`- [[${sourcePath}|${entry.sourceTitle}]]${context}${
							entry.summary ? `\n  - *"${entry.summary}"*` : ""
						}`
					);
				});
				lines.push("");
			}
		}

		lines.push("---", "", "## Summary", "", "| Story | relation_type | Count |", "|-------|---------------|-------|");

		for (const [storyId, entries] of grouped.entries()) {
			const byRelation = this.groupByRelationType(entries);
			for (const [relationType, relationEntries] of byRelation.entries()) {
				lines.push(`| ${entries[0].storyTitle} | ${relationType} | ${relationEntries.length} |`);
			}
		}

		const total = citations.length;
		lines.push(`| **Total** | | **${total}** |`);

		return lines.join("\n").trimEnd() + "\n";
	}

	private groupByStory(citations: CitationsGeneratorInput["citations"]) {
		const map = new Map<string, typeof citations>();
		citations.forEach((citation) => {
			if (!map.has(citation.storyId)) {
				map.set(citation.storyId, []);
			}
			map.get(citation.storyId)!.push(citation);
		});
		return map;
	}

	private groupByRelationType(citations: CitationsGeneratorInput["citations"]) {
		const map = new Map<string, typeof citations>();
		citations.forEach((citation) => {
			if (!map.has(citation.relationType)) {
				map.set(citation.relationType, []);
			}
			map.get(citation.relationType)!.push(citation);
		});
		return map;
	}

	private titleCase(value: string): string {
		return value
			.split(/[_\s]/g)
			.map((part) => part.charAt(0).toUpperCase() + part.slice(1))
			.join(" ");
	}
}

