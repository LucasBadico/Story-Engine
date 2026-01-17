import type { Lore, World } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { FrontmatterGenerator } from "../../generators/FrontmatterGenerator";
import { slugify } from "../../utils/slugify";
import { parseWorldEntityFile } from "../../parsers/worldEntityParser";

export class LoreHandler {
	readonly entityType = "lore";
	private readonly frontmatterGenerator = new FrontmatterGenerator();

	async pull(id: string, context: SyncContext): Promise<Lore> {
		const lore = await context.apiClient.getLore(id);
		const world = await context.apiClient.getWorld(lore.world_id);
		const folderPath = context.fileManager.getWorldFolderPath(world.name);
		const loreFolder = `${folderPath}/lore`;
		await context.fileManager.ensureFolderExists(loreFolder);
		const filePath = `${loreFolder}/${slugify(lore.name)}.md`;
		await context.fileManager.writeFile(filePath, this.renderLore(lore, world, context));
		return lore;
	}

	async push(entity: Lore, context: SyncContext): Promise<void> {
		const world = await context.apiClient.getWorld(entity.world_id);
		const folderPath = context.fileManager.getWorldFolderPath(world.name);
		const filePath = `${folderPath}/lore/${slugify(entity.name)}.md`;

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

		await context.apiClient.updateLore(entity.id, {
			name: parsed.name,
			description,
		});
	}

	async delete(id: string, context: SyncContext): Promise<void> {
		await context.apiClient.deleteLore(id);
	}

	private renderLore(lore: Lore, world: World, context: SyncContext): string {
		const baseFields = {
			id: lore.id,
			world_id: lore.world_id,
			category: lore.category ?? null,
			parent_id: lore.parent_id ?? null,
			hierarchy_level: lore.hierarchy_level,
			created_at: lore.created_at,
			updated_at: lore.updated_at,
		};

		const frontmatter = this.frontmatterGenerator.generate(baseFields, undefined, {
			entityType: "lore",
			worldName: world.name,
			date: lore.created_at,
			idField: context.settings.frontmatterIdField,
		});

		return [
			frontmatter,
			"",
			`# ${lore.name}`,
			"",
			"## Description",
			lore.description || "_No description yet._",
			"",
			"## Rules",
			lore.rules || "_Rules not documented._",
			"",
		].join("\n");
	}
}

