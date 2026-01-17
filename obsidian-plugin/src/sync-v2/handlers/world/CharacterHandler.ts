import type { Character, World } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { FrontmatterGenerator } from "../../generators/FrontmatterGenerator";
import { slugify } from "../../utils/slugify";
import { parseWorldEntityFile } from "../../parsers/worldEntityParser";

export class CharacterHandler {
	readonly entityType = "character";
	private readonly frontmatterGenerator = new FrontmatterGenerator();

	async pull(id: string, context: SyncContext): Promise<Character> {
		const character = await context.apiClient.getCharacter(id);
		const world = await context.apiClient.getWorld(character.world_id);
		const folderPath = context.fileManager.getWorldFolderPath(world.name);
		const charactersFolder = `${folderPath}/characters`;
		await context.fileManager.ensureFolderExists(charactersFolder);
		const filePath = `${charactersFolder}/${slugify(character.name)}.md`;
		await context.fileManager.writeFile(filePath, this.renderCharacter(character, world, context));
		return character;
	}

	async push(entity: Character, context: SyncContext): Promise<void> {
		const world = await context.apiClient.getWorld(entity.world_id);
		const folderPath = context.fileManager.getWorldFolderPath(world.name);
		const filePath = `${folderPath}/characters/${slugify(entity.name)}.md`;

		let localContent: string;
		try {
			localContent = await context.fileManager.readFile(filePath);
		} catch {
			return;
		}

		const parsed = parseWorldEntityFile(localContent);

		const normalizedDescription = parsed.description ?? undefined;
		const existingDescription = entity.description ?? undefined;

		if (parsed.name === entity.name && (normalizedDescription ?? "") === (existingDescription ?? "")) {
			return;
		}

		await context.apiClient.updateCharacter(entity.id, {
			name: parsed.name,
			description: normalizedDescription,
		});
	}

	async delete(id: string, context: SyncContext): Promise<void> {
		await context.apiClient.deleteCharacter(id);
	}

	private renderCharacter(character: Character, world: World, context: SyncContext): string {
		const baseFields = {
			id: character.id,
			world_id: character.world_id,
			class_level: character.class_level,
			archetype_id: character.archetype_id ?? null,
			current_class_id: character.current_class_id ?? null,
			created_at: character.created_at,
			updated_at: character.updated_at,
		};

		const frontmatter = this.frontmatterGenerator.generate(baseFields, undefined, {
			entityType: "character",
			worldName: world.name,
			date: character.created_at,
			idField: context.settings.frontmatterIdField,
		});

		return [
			frontmatter,
			"",
			`# ${character.name}`,
			"",
			"## Description",
			character.description || "_No description yet._",
			"",
			"## Metadata",
			`- Tenant: ${character.tenant_id}`,
			`- Archetype: ${character.archetype_id ?? "—"}`,
			`- Class Level: ${character.class_level}`,
			"",
		].join("\n");
	}
}

