import type { Artifact, World } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { FrontmatterGenerator } from "../../generators/FrontmatterGenerator";
import { slugify } from "../../utils/slugify";
import { parseWorldEntityFile } from "../../parsers/worldEntityParser";

export class ArtifactHandler {
	readonly entityType = "artifact";
	private readonly frontmatterGenerator = new FrontmatterGenerator();

	async pull(id: string, context: SyncContext): Promise<Artifact> {
		const artifact = await context.apiClient.getArtifact(id);
		const world = await context.apiClient.getWorld(artifact.world_id);
		const folderPath = context.fileManager.getWorldFolderPath(world.name);
		const artifactsFolder = `${folderPath}/artifacts`;
		await context.fileManager.ensureFolderExists(artifactsFolder);
		const filePath = `${artifactsFolder}/${slugify(artifact.name)}.md`;
		await context.fileManager.writeFile(filePath, this.renderArtifact(artifact, world, context));
		return artifact;
	}

	async push(entity: Artifact, context: SyncContext): Promise<void> {
		const world = await context.apiClient.getWorld(entity.world_id);
		const folderPath = context.fileManager.getWorldFolderPath(world.name);
		const filePath = `${folderPath}/artifacts/${slugify(entity.name)}.md`;

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

		await context.apiClient.updateArtifact(entity.id, {
			name: parsed.name,
			description,
		});
	}

	async delete(id: string, context: SyncContext): Promise<void> {
		await context.apiClient.deleteArtifact(id);
	}

	private renderArtifact(artifact: Artifact, world: World, context: SyncContext): string {
		const baseFields = {
			id: artifact.id,
			world_id: artifact.world_id,
			rarity: artifact.rarity,
			created_at: artifact.created_at,
			updated_at: artifact.updated_at,
		};

		const frontmatter = this.frontmatterGenerator.generate(baseFields, undefined, {
			entityType: "artifact",
			worldName: world.name,
			date: artifact.created_at,
			idField: context.settings.frontmatterIdField,
		});

		return [
			frontmatter,
			"",
			`# ${artifact.name}`,
			"",
			"## Description",
			artifact.description || "_No description yet._",
			"",
		].join("\n");
	}
}

