import type { Archetype } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { FrontmatterGenerator } from "../../generators/FrontmatterGenerator";
import { slugify } from "../../utils/slugify";
import { parseWorldEntityFile } from "../../parsers/worldEntityParser";

export class ArchetypeHandler {
	readonly entityType = "archetype";
	private readonly frontmatterGenerator = new FrontmatterGenerator();

	async pull(id: string, context: SyncContext): Promise<Archetype> {
		const archetype = await context.apiClient.getArchetype(id);
		const worldsRoot = context.fileManager.getWorldsRootPath();
		const charactersFolder = `${worldsRoot}/characters`;
		const folderPath = `${charactersFolder}/_archetypes`;
		await context.fileManager.ensureFolderExists(worldsRoot);
		await context.fileManager.ensureFolderExists(charactersFolder);
		await context.fileManager.ensureFolderExists(folderPath);
		const filePath = `${folderPath}/${slugify(archetype.name)}.md`;
		await context.fileManager.writeFile(filePath, this.renderArchetype(archetype, context));
		return archetype;
	}

	async push(entity: Archetype, context: SyncContext): Promise<void> {
		const worldsRoot = context.fileManager.getWorldsRootPath();
		const archetypeFolder = `${worldsRoot}/characters/_archetypes`;
		const filePath = `${archetypeFolder}/${slugify(entity.name)}.md`;

		let localContent: string;
		try {
			localContent = await context.fileManager.readFile(filePath);
		} catch {
			return;
		}

		const parsed = parseWorldEntityFile(localContent);
		const description = parsed.description ?? undefined;

		if (parsed.name === entity.name && description === (entity.description ?? undefined)) {
			return;
		}

		await context.apiClient.updateArchetype(entity.id, {
			name: parsed.name,
			description,
		});
	}

	async delete(id: string, context: SyncContext): Promise<void> {
		await context.apiClient.deleteArchetype(id);
	}

	private renderArchetype(archetype: Archetype, context: SyncContext): string {
		const baseFields = {
			id: archetype.id,
			tenant_id: archetype.tenant_id,
			created_at: archetype.created_at,
			updated_at: archetype.updated_at,
		};

		const frontmatter = this.frontmatterGenerator.generate(baseFields, undefined, {
			entityType: "archetype",
			date: archetype.created_at,
			idField: context.settings.frontmatterIdField,
		});

		return [
			frontmatter,
			"",
			`# ${archetype.name}`,
			"",
			"## Description",
			archetype.description || "_No description yet._",
			"",
		].join("\n");
	}
}

