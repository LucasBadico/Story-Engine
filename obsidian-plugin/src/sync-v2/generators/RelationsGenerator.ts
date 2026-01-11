import type { RelationsGeneratorInput } from "../types/generators";
import { getIdFieldName } from "../utils/frontmatterHelpers";

const TARGET_LABELS: Record<string, string> = {
	character: "Main Characters",
	location: "Key Locations",
	faction: "Referenced Factions",
	event: "Timeline Events",
	artifact: "Artifacts",
	lore: "Lore References",
};

export class RelationsGenerator {
	constructor(private readonly now: () => string = () => new Date().toISOString()) {}

	generate(input: RelationsGeneratorInput): string {
		const { entity, relations } = input;
		const lines: string[] = [];

		// Use configured ID field name, default to "id"
		const idField = getIdFieldName(input.options?.idField);
		
		lines.push(
			"---",
			`${idField}: ${entity.id}`,
			`type: ${entity.type}-relations`,
			`synced_at: ${input.options?.syncedAt ?? this.now()}`
		);
		if (entity.worldId) {
			lines.push(`world_id: ${entity.worldId}`);
		}
		lines.push("---", "", `# ${entity.name} - Relations`, "");

		if (input.options?.showHelpBox !== false) {
			lines.push(
				"> [!tip] Como editar relações",
				"> - **Adicionar**: Edite a linha `_Add new..._` da seção",
				"> - **Remover**: Delete a linha da relação",
				"> - **Formato**: `[[entity|Name]] - description`",
				""
			);
		}

		if (entity.worldId && entity.worldName) {
			lines.push("## World");
			lines.push(`[[${entity.worldId}|${entity.worldName}]]`, "");
		}

		const grouped = this.groupByTarget(relations);
		for (const [targetType, items] of grouped.entries()) {
			lines.push(`## ${TARGET_LABELS[targetType] ?? targetType}`);

			if (items.length === 0) {
				lines.push(this.placeholderLine(targetType));
				lines.push("");
				continue;
			}

			items
				.sort((a, b) => a.targetName.localeCompare(b.targetName))
				.forEach((entry) => {
					const description = entry.summary ?? entry.contextLabel ?? "";
					const desc = description ? ` - ${description}` : "";
					lines.push(`- [[${entry.targetId}|${entry.targetName}]]${desc}`);
				});

			lines.push(this.placeholderLine(targetType), "");
		}

		return lines.join("\n").trimEnd() + "\n";
	}

	private groupByTarget(relations: RelationsGeneratorInput["relations"]) {
		const map = new Map<string, typeof relations>();
		relations.forEach((relation) => {
			const key = relation.targetType;
			if (!map.has(key)) {
				map.set(key, []);
			}
			map.get(key)!.push(relation);
		});

		// Ensure sections exist even if empty
		Object.keys(TARGET_LABELS).forEach((type) => {
			if (!map.has(type)) {
				map.set(type, []);
			}
		});

		return map;
	}

	private placeholderLine(targetType: string): string {
		const label = TARGET_LABELS[targetType] ?? targetType;
		const noun = label.replace(/s$/, "").toLowerCase();
		return `- _Add new ${noun}: [[file|Name]] - description_`;
	}
}

