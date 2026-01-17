import type { SyncContext } from "../types/sync";
import { OutlineParser } from "../parsers/outlineParser";
import { parseFrontmatter } from "../utils/detectEntityMentions";
import { getFrontmatterId } from "../utils/frontmatterHelpers";

export interface OutlinePushAction {
	type: "chapter_reorder";
	chapterId: string;
	oldOrder: number;
	newOrder: number;
}

export interface OutlinePushResult {
	actions: OutlinePushAction[];
	warnings: string[];
}

export class OutlinePushHandler {
	constructor(private readonly parser = new OutlineParser()) {}

	async analyzeOutline(
		outlineFilePath: string,
		storyId: string,
		context: SyncContext
	): Promise<OutlinePushResult> {
		const warnings: string[] = [];
		const actions: OutlinePushAction[] = [];

		let outlineContent: string;
		try {
			outlineContent = await context.fileManager.readFile(outlineFilePath);
		} catch {
			warnings.push("Outline file not found");
			return { actions, warnings };
		}

		const entries = this.parser.parse(outlineContent).filter(
			(entry) => entry.type === "chapter" && entry.link
		);

		const story = await context.apiClient.getStory(storyId);
		const hierarchy = await context.apiClient.getStoryWithHierarchy(storyId);
		const storyFolder = context.fileManager.getStoryFolderPath(story.title);

		// Map chapter ID -> local order by reading each chapter file from the outline
		const localOrderByChapterId = new Map<string, number>();
		for (const entry of entries) {
			const chapterId = await this.resolveChapterId(storyFolder, entry.link!, context);
			if (chapterId) {
				localOrderByChapterId.set(chapterId, entry.order);
			} else {
				warnings.push(`Could not resolve chapter ID for link: ${entry.link}`);
			}
		}

		// Compare local order to remote order
		for (const chapterWithContent of hierarchy.chapters) {
			const remoteChapter = chapterWithContent.chapter;
			const localOrder = localOrderByChapterId.get(remoteChapter.id);
			if (localOrder !== undefined && localOrder !== remoteChapter.number) {
				actions.push({
					type: "chapter_reorder",
					chapterId: remoteChapter.id,
					oldOrder: remoteChapter.number,
					newOrder: localOrder,
				});
			}
		}

		return { actions, warnings };
	}

	/**
	 * Resolve a chapter link to its ID by reading the chapter file's frontmatter.
	 * Tries both `00-chapters/` and `chapters/` folder conventions.
	 */
	private async resolveChapterId(
		storyFolder: string,
		link: string,
		context: SyncContext
	): Promise<string | null> {
		const isFullPath = link.includes("/") || link.endsWith(".md");
		if (isFullPath) {
			const path = link.endsWith(".md") ? link : `${link}.md`;
			try {
				const content = await context.fileManager.readFile(path);
				const frontmatter = parseFrontmatter(content);
				const id = getFrontmatterId(frontmatter, context.settings.frontmatterIdField);
				if (id) return id;
			} catch {
				// fall through to legacy paths
			}
		}

		const possiblePaths = [
			`${storyFolder}/00-chapters/${link}.md`,
			`${storyFolder}/chapters/${link}.md`,
		];

		for (const path of possiblePaths) {
			try {
				const content = await context.fileManager.readFile(path);
				const frontmatter = parseFrontmatter(content);
				const id = getFrontmatterId(frontmatter, context.settings.frontmatterIdField);
				if (id) return id;
			} catch {
				// File not found, try next path
			}
		}

		return null;
	}
}
