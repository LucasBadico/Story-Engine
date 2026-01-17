import type { Faction, World } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { FrontmatterGenerator } from "../../generators/FrontmatterGenerator";
import { slugify } from "../../utils/slugify";
import { parseWorldEntityFile } from "../../parsers/worldEntityParser";

export class FactionHandler {
	readonly entityType = "faction";
	private readonly frontmatterGenerator = new FrontmatterGenerator();

	async pull(id: string, context: SyncContext): Promise<Faction> {
		const faction = await context.apiClient.getFaction(id);
		const world = await context.apiClient.getWorld(faction.world_id);
		const folderPath = context.fileManager.getWorldFolderPath(world.name);
		const factionsFolder = `${folderPath}/factions`;
		await context.fileManager.ensureFolderExists(factionsFolder);
		const filePath = `${factionsFolder}/${slugify(faction.name)}.md`;
		await context.fileManager.writeFile(filePath, this.renderFaction(faction, world, context));
		return faction;
	}

	async push(entity: Faction, context: SyncContext): Promise<void> {
		const world = await context.apiClient.getWorld(entity.world_id);
		const folderPath = context.fileManager.getWorldFolderPath(world.name);
		const filePath = `${folderPath}/factions/${slugify(entity.name)}.md`;

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

		await context.apiClient.updateFaction(entity.id, {
			name: parsed.name,
			description,
		});
	}

	async delete(id: string, context: SyncContext): Promise<void> {
		await context.apiClient.deleteFaction(id);
	}

	private renderFaction(faction: Faction, world: World, context: SyncContext): string {
		const baseFields = {
			id: faction.id,
			world_id: faction.world_id,
			type: faction.type ?? null,
			hierarchy_level: faction.hierarchy_level,
			parent_id: faction.parent_id ?? null,
			created_at: faction.created_at,
			updated_at: faction.updated_at,
		};

		const frontmatter = this.frontmatterGenerator.generate(baseFields, undefined, {
			entityType: "faction",
			worldName: world.name,
			date: faction.created_at,
			idField: context.settings.frontmatterIdField,
		});

		return [
			frontmatter,
			"",
			`# ${faction.name}`,
			"",
			"## Description",
			faction.description || "_No description yet._",
			"",
			"## Beliefs & Structure",
			faction.beliefs || "_Beliefs pending._",
			"",
			faction.structure || "_Structure pending._",
			"",
		].join("\n");
	}
}

