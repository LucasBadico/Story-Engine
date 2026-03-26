import type { Trait } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { FrontmatterGenerator } from "../../generators/FrontmatterGenerator";
import { slugify } from "../../utils/slugify";
import { parseWorldEntityFile } from "../../parsers/worldEntityParser";

export class TraitHandler {
	readonly entityType = "trait";
	private readonly frontmatterGenerator = new FrontmatterGenerator();

	async pull(id: string, context: SyncContext): Promise<Trait> {
		const trait = await context.apiClient.getTrait(id);
		const worldsRoot = context.fileManager.getWorldsRootPath();
		const charactersFolder = `${worldsRoot}/characters`;
		const folderPath = `${charactersFolder}/_traits`;
		await context.fileManager.ensureFolderExists(worldsRoot);
		await context.fileManager.ensureFolderExists(charactersFolder);
		await context.fileManager.ensureFolderExists(folderPath);
		const filePath = `${folderPath}/${slugify(trait.name)}.md`;
		await context.fileManager.writeFile(filePath, this.renderTrait(trait, context));
		return trait;
	}

	async push(entity: Trait, context: SyncContext): Promise<void> {
		const worldsRoot = context.fileManager.getWorldsRootPath();
		const folderPath = `${worldsRoot}/characters/_traits`;
		const filePath = `${folderPath}/${slugify(entity.name)}.md`;

		let localContent: string;
		try {
			localContent = await context.fileManager.readFile(filePath);
		} catch {
			return;
		}

		const parsed = parseWorldEntityFile(localContent);
		const description = parsed.description ?? undefined;

		if (parsed.name === entity.name && (description ?? "") === (entity.description ?? "")) {
			return;
		}

		await context.apiClient.updateTrait(entity.id, {
			name: parsed.name,
			description,
		});
	}

	async delete(id: string, context: SyncContext): Promise<void> {
		await context.apiClient.deleteTrait(id);
	}

	private renderTrait(trait: Trait, context: SyncContext): string {
		const baseFields = {
			id: trait.id,
			tenant_id: trait.tenant_id,
			category: trait.category,
			created_at: trait.created_at,
			updated_at: trait.updated_at,
		};

		const frontmatter = this.frontmatterGenerator.generate(baseFields, undefined, {
			entityType: "trait",
			date: trait.created_at,
			idField: context.settings.frontmatterIdField,
		});

		return [
			frontmatter,
			"",
			`# ${trait.name}`,
			"",
			"## Description",
			trait.description || "_No description yet._",
			"",
		].join("\n");
	}
}

