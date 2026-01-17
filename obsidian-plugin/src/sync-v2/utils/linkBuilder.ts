import { slugify } from "./slugify";

const STORY_FOLDER_MAP: Record<string, { folder: string; prefix: string }> = {
	chapter: { folder: "00-chapters", prefix: "ch" },
	scene: { folder: "01-scenes", prefix: "sc" },
	beat: { folder: "02-beats", prefix: "bt" },
};

const WORLD_FOLDER_MAP: Record<string, string> = {
	character: "characters",
	location: "locations",
	faction: "factions",
	artifact: "artifacts",
	event: "events",
	lore: "lore",
};

export function buildWikiLink(path: string, label: string): string {
	return `[[${path}|${label}]]`;
}

export function buildStoryEntityPath(
	storyFolderPath: string,
	type: keyof typeof STORY_FOLDER_MAP,
	order: number,
	title: string,
	overrides?: { chapterOrder?: number; sceneOrder?: number }
): string {
	const entry = STORY_FOLDER_MAP[type];
	const slug = slugify(title);
	const orderTag = String(order).padStart(4, "0");
	if (type === "scene") {
		const chapterTag = String(overrides?.chapterOrder ?? 0).padStart(4, "0");
		return `${storyFolderPath}/${entry.folder}/${entry.prefix}-${chapterTag}-${orderTag}-${slug}.md`;
	}
	if (type === "beat") {
		const chapterTag = String(overrides?.chapterOrder ?? 0).padStart(4, "0");
		const sceneTag = String(overrides?.sceneOrder ?? 0).padStart(4, "0");
		return `${storyFolderPath}/${entry.folder}/${entry.prefix}-${chapterTag}-${sceneTag}-${orderTag}-${slug}.md`;
	}
	return `${storyFolderPath}/${entry.folder}/${entry.prefix}-${orderTag}-${slug}.md`;
}

export function buildWorldEntityPath(
	worldFolderPath: string,
	entityType: keyof typeof WORLD_FOLDER_MAP,
	name: string
): string {
	const folder = WORLD_FOLDER_MAP[entityType];
	const slug = slugify(name);
	return `${worldFolderPath}/${folder}/${slug}.md`;
}

export function buildWorldFolderLink(worldFolderPath: string): string {
	return `${worldFolderPath}/world.md`;
}

export function resolveLinkBasename(link: string): string {
	const trimmed = link.split("|")[0].trim();
	const base = trimmed.split("/").pop() ?? trimmed;
	return base.endsWith(".md") ? base.slice(0, -3) : base;
}
