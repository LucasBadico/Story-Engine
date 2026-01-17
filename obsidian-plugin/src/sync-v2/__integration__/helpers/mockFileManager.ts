import type { Story, Chapter } from "../../../types";

export class MockFileManager {
	private readonly files = new Map<string, string>();
	private readonly folders = new Set<string>();

	constructor(private readonly baseFolder: string) {}

	getStoryFolderPath(storyTitle: string): string {
		return `${this.baseFolder}/${storyTitle}`;
	}

	getWorldsRootPath(): string {
		return `${this.baseFolder}/worlds`;
	}

	getWorldFolderPath(worldName: string): string {
		return `${this.baseFolder}/worlds/${worldName}`;
	}

	async ensureFolderExists(path: string): Promise<void> {
		this.folders.add(path);
	}

	async writeFile(filePath: string, content: string): Promise<void> {
		const folderPath = filePath.split("/").slice(0, -1).join("/");
		if (folderPath) {
			await this.ensureFolderExists(folderPath);
		}
		this.files.set(filePath, content);
	}

	async readFile(filePath: string): Promise<string> {
		const content = this.files.get(filePath);
		if (content === undefined) {
			throw new Error(`File not found: ${filePath}`);
		}
		return content;
	}

	async writeStoryMetadata(story: Story, folderPath: string, chapters: Chapter[]): Promise<void> {
		const content = [
			"---",
			`id: ${story.id}`,
			`title: ${story.title}`,
			`status: ${story.status}`,
			`version: ${story.version_number}`,
			`root_story_id: ${story.root_story_id}`,
			`previous_version_id: ${story.previous_story_id ?? ""}`,
			`created_at: ${story.created_at}`,
			`updated_at: ${story.updated_at}`,
			"---",
			"",
			`# ${story.title}`,
			"",
			`Chapters: ${chapters.length}`,
			"",
		].join("\n");
		await this.writeFile(`${folderPath}/story.md`, content);
	}

	async writeChapterFile(
		chapterWithContent: { chapter: Chapter },
		filePath: string,
		_storyName?: string
	): Promise<void> {
		const { chapter } = chapterWithContent;
		const content = [
			"---",
			`id: ${chapter.id}`,
			`title: ${chapter.title}`,
			"---",
			"",
			`# ${chapter.title}`,
			"",
		].join("\n");
		await this.writeFile(filePath, content);
	}

	async writeSceneFile(
		sceneWithBeats: { scene: { id: string; goal?: string | null } },
		filePath: string,
		_storyName?: string
	): Promise<void> {
		const { scene } = sceneWithBeats;
		const content = [
			"---",
			`id: ${scene.id}`,
			"---",
			"",
			`# ${scene.goal ?? "Scene"}`,
			"",
		].join("\n");
		await this.writeFile(filePath, content);
	}

	async writeBeatFile(
		beat: { id: string; intent?: string | null },
		filePath: string,
		_storyName?: string
	): Promise<void> {
		const content = [
			"---",
			`id: ${beat.id}`,
			"---",
			"",
			`# ${beat.intent ?? "Beat"}`,
			"",
		].join("\n");
		await this.writeFile(filePath, content);
	}

	async readStoryMetadata(folderPath: string, idField?: string): Promise<{ frontmatter: any; content: string }> {
		const content = await this.readFile(`${folderPath}/story.md`);
		const frontmatter = this.parseFrontmatter(content);
		const effectiveIdField = idField || "id";
		return {
			frontmatter: {
				id: frontmatter[effectiveIdField] || frontmatter.id,
				title: frontmatter.title,
				status: frontmatter.status,
				version: parseInt(frontmatter.version, 10),
				root_story_id: frontmatter.root_story_id,
				previous_version_id: frontmatter.previous_version_id || null,
				created_at: frontmatter.created_at,
				updated_at: frontmatter.updated_at,
			},
			content: content.split("---").slice(2).join("---").trim(),
		};
	}

	async writeWorldMetadata(world: { id: string; name: string; description: string }, folderPath: string): Promise<void> {
		const content = [
			"---",
			`id: ${world.id}`,
			"---",
			"",
			`# ${world.name}`,
			"",
			world.description || "_No description yet._",
			"",
		].join("\n");
		await this.writeFile(`${folderPath}/world.md`, content);
	}

	async writeContentBlockFile(contentBlock: { id: string; content: string }, filePath: string): Promise<void> {
		const content = [
			"---",
			`id: ${contentBlock.id}`,
			"---",
			"",
			contentBlock.content,
			"",
		].join("\n");
		await this.writeFile(filePath, content);
	}

	setFile(filePath: string, content: string): void {
		this.files.set(filePath, content);
	}

	getFile(filePath: string): string | undefined {
		return this.files.get(filePath);
	}

	private parseFrontmatter(content: string): Record<string, string> {
		const match = content.match(/^---\n([\s\S]*?)\n---/);
		if (!match) return {};
		const lines = match[1].split("\n");
		const result: Record<string, string> = {};
		for (const line of lines) {
			const [key, ...rest] = line.split(":");
			if (!key) continue;
			result[key.trim()] = rest.join(":").trim();
		}
		return result;
	}
}
