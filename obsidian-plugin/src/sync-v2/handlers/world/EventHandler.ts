import type { World, WorldEvent } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { FrontmatterGenerator } from "../../generators/FrontmatterGenerator";
import { slugify } from "../../utils/slugify";

export class EventHandler {
	readonly entityType = "event";
	private readonly frontmatterGenerator = new FrontmatterGenerator();

	async pull(id: string, context: SyncContext): Promise<WorldEvent> {
		const event = await context.apiClient.getEvent(id);
		const world = await context.apiClient.getWorld(event.world_id);
		const folderPath = context.fileManager.getWorldFolderPath(world.name);
		const eventsFolder = `${folderPath}/events`;
		await context.fileManager.ensureFolderExists(eventsFolder);
		const filePath = `${eventsFolder}/${slugify(event.name)}.md`;
		await context.fileManager.writeFile(filePath, this.renderEvent(event, world, context));
		return event;
	}

	async push(_entity: WorldEvent, _context: SyncContext): Promise<void> {
		// TODO: implement push logic
	}

	async delete(id: string, context: SyncContext): Promise<void> {
		await context.apiClient.deleteEvent(id);
	}

	private renderEvent(event: WorldEvent, world: World, context: SyncContext): string {
		const baseFields = {
			id: event.id,
			world_id: event.world_id,
			type: event.type ?? null,
			importance: event.importance,
			timeline: event.timeline ?? null,
			parent_id: event.parent_id ?? null,
			created_at: event.created_at,
			updated_at: event.updated_at,
		};

		const frontmatter = this.frontmatterGenerator.generate(baseFields, undefined, {
			entityType: "event",
			worldName: world.name,
			date: event.created_at,
			idField: context.settings.frontmatterIdField,
		});

		return [
			frontmatter,
			"",
			`# ${event.name}`,
			"",
			"## Description",
			event.description || "_No description yet._",
			"",
		].join("\n");
	}
}

