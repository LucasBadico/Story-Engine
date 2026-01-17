import { normalizePath } from "obsidian";
import type { Beat, Chapter, ContentBlock, Scene } from "../../types";

const CONTENT_FOLDER_BY_TYPE: Record<string, string> = {
	text: "00-texts",
	image: "01-images",
	video: "02-videos",
	audio: "03-audios",
	embed: "04-embeds",
	link: "05-links",
};

export class PathResolver {
	constructor(private readonly storyFolder: string) {}

	getChapterPath(chapter: Chapter, overrides?: { order?: number; title?: string }): string {
		const order = overrides?.order ?? chapter.number ?? 0;
		const title = overrides?.title ?? chapter.title ?? "chapter";
		return normalizePath(
			`${this.storyFolder}/00-chapters/${this.buildFileName("ch", order, title)}`
		);
	}

	getScenePath(
		scene: Scene,
		overrides?: { order?: number; goal?: string; chapterOrder?: number }
	): string {
		const order = overrides?.order ?? scene.order_num ?? 0;
		const chapterOrder = overrides?.chapterOrder ?? 0;
		const goal = overrides?.goal ?? scene.goal ?? "scene";
		return normalizePath(
			`${this.storyFolder}/01-scenes/${this.buildCompositeFileName(
				"sc",
				[chapterOrder, order],
				goal
			)}`
		);
	}

	getBeatPath(
		beat: Beat,
		overrides?: { order?: number; intent?: string; chapterOrder?: number; sceneOrder?: number }
	): string {
		const order = overrides?.order ?? beat.order_num ?? 0;
		const chapterOrder = overrides?.chapterOrder ?? 0;
		const sceneOrder = overrides?.sceneOrder ?? 0;
		const intent = overrides?.intent ?? beat.intent ?? "beat";
		return normalizePath(
			`${this.storyFolder}/02-beats/${this.buildCompositeFileName(
				"bt",
				[chapterOrder, sceneOrder, order],
				intent
			)}`
		);
	}

	getContentBlockPath(
		contentBlock: ContentBlock,
		overrides?: { order?: number; title?: string; type?: string }
	): string {
		const order = overrides?.order ?? contentBlock.order_num ?? 0;
		const title =
			overrides?.title ??
			contentBlock.metadata?.title ??
			contentBlock.kind ??
			contentBlock.type ??
			"content";
		const type = (overrides?.type ?? contentBlock.type ?? "text").toLowerCase();
		const folder = CONTENT_FOLDER_BY_TYPE[type] ?? "99-other";
		return normalizePath(
			`${this.storyFolder}/03-contents/${folder}/${this.buildFileName("cb", order, title)}`
		);
	}

	private buildFileName(prefix: string, order: number, label: string): string {
		return `${prefix}-${this.padOrder(order)}-${this.sanitize(label)}.md`;
	}

	private buildCompositeFileName(prefix: string, orders: number[], label: string): string {
		const orderTags = orders.map((order) => this.padOrder(order)).join("-");
		return `${prefix}-${orderTags}-${this.sanitize(label)}.md`;
	}

	private padOrder(order?: number): string {
		const value =
			typeof order === "number" && Number.isFinite(order) ? Math.max(0, order) : 0;
		return String(value).padStart(4, "0");
	}

	private sanitize(value: string): string {
		const sanitized = value
			.normalize("NFKD")
			.replace(/[^\w\s-]/g, "")
			.trim()
			.replace(/\s+/g, "-")
			.replace(/-+/g, "-")
			.replace(/^-|-$/g, "")
			.toLowerCase();
		return sanitized || "untitled";
	}
}

