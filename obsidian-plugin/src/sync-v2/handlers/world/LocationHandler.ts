import type { Location, World } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { FrontmatterGenerator } from "../../generators/FrontmatterGenerator";
import { slugify } from "../../utils/slugify";

export class LocationHandler {
	readonly entityType = "location";
	private readonly frontmatterGenerator = new FrontmatterGenerator();

	async pull(id: string, context: SyncContext): Promise<Location> {
		const location = await context.apiClient.getLocation(id);
		const world = await context.apiClient.getWorld(location.world_id);
		const folderPath = context.fileManager.getWorldFolderPath(world.name);
		const locationsFolder = `${folderPath}/locations`;
		await context.fileManager.ensureFolderExists(locationsFolder);
		const filePath = `${locationsFolder}/${slugify(location.name)}.md`;
		await context.fileManager.writeFile(filePath, this.renderLocation(location, world, context));
		return location;
	}

	async push(_entity: Location, _context: SyncContext): Promise<void> {
		// TODO: implement push logic
	}

	async delete(id: string, context: SyncContext): Promise<void> {
		await context.apiClient.deleteLocation(id);
	}

	private renderLocation(location: Location, world: World, context: SyncContext): string {
		const baseFields = {
			id: location.id,
			world_id: location.world_id,
			type: location.type,
			hierarchy_level: location.hierarchy_level,
			parent_id: location.parent_id ?? null,
			created_at: location.created_at,
			updated_at: location.updated_at,
		};

		const frontmatter = this.frontmatterGenerator.generate(baseFields, undefined, {
			entityType: "location",
			worldName: world.name,
			date: location.created_at,
			idField: context.settings.frontmatterIdField,
		});

		return [
			frontmatter,
			"",
			`# ${location.name}`,
			"",
			"## Description",
			location.description || "_No description yet._",
			"",
			"## Notes",
			location.type ? `- Type: ${location.type}` : "- Type: —",
			"",
		].join("\n");
	}
}

