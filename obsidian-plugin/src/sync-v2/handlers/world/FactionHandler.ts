import type { Faction, World } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { FrontmatterGenerator } from "../../generators/FrontmatterGenerator";
import { slugify } from "../../utils/slugify";

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

	async push(_entity: Faction, _context: SyncContext): Promise<void> {
		// TODO: implement push logic
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

