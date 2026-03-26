import type { World, WorldEvent } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { FrontmatterGenerator } from "../../generators/FrontmatterGenerator";
import { slugify } from "../../utils/slugify";
import { parseWorldEntityFile } from "../../parsers/worldEntityParser";

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

	async push(entity: WorldEvent, context: SyncContext): Promise<void> {
		const world = await context.apiClient.getWorld(entity.world_id);
		const folderPath = context.fileManager.getWorldFolderPath(world.name);
		const filePath = `${folderPath}/events/${slugify(entity.name)}.md`;

		let localContent: string;
		try {
			localContent = await context.fileManager.readFile(filePath);
		} catch {
			return;
		}

		const parsed = parseWorldEntityFile(localContent);
		const description = parsed.description ?? null;

		if (parsed.name === entity.name && description === (entity.description ?? null)) {
			return;
		}

		await context.apiClient.updateEvent(entity.id, {
			name: parsed.name,
			description,
		});
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

