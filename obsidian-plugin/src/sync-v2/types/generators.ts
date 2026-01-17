import type {
	Beat,
	Chapter,
	ChapterWithContent,
	ContentBlock,
	Scene,
	SceneWithBeats,
	Story,
	StoryWithHierarchy,
} from "../../types";

export interface GeneratorMetadata {
	syncedAt?: string;
	showHelpBox?: boolean;
	/** Custom ID field name (from settings), defaults to "id" */
	idField?: string;
	/** Full path to story folder for link generation */
	storyFolderPath?: string;
	/** Full path to world folder for link generation */
	worldFolderPath?: string;
}

export interface OutlineGeneratorInput {
	story: StoryWithHierarchy;
	options?: GeneratorMetadata;
}

export interface StoryContentsInput {
	story: Story;
	chapters: ChapterWithContent[];
	chapterContentBlocks?: Map<string, ContentBlock[]>;
	sceneContentBlocks?: Map<string, ContentBlock[]>;
	beatContentBlocks?: Map<string, ContentBlock[]>;
	options?: GeneratorMetadata;
}

export interface RelationsEntityMeta {
	id: string;
	name: string;
	type: string;
	worldId?: string;
	worldName?: string;
}

export interface ParsedRelationEntry {
	targetType: string;
	targetId: string;
	targetName: string;
	relationType: string;
	summary?: string;
	attributes?: Record<string, unknown>;
	contextLabel?: string;
}

export interface RelationsGeneratorInput {
	entity: RelationsEntityMeta;
	relations: ParsedRelationEntry[];
	options?: GeneratorMetadata;
}

export interface ParsedCitationEntry {
	storyId: string;
	storyTitle: string;
	storyPath?: string;
	relationType: string;
	sourceType: "chapter" | "scene" | "beat" | "content_block";
	sourceId: string;
	sourceTitle: string;
	sourcePath?: string;
	chapterTitle?: string;
	summary?: string;
	attributes?: Record<string, unknown>;
}

export interface CitationsGeneratorInput {
	entity: RelationsEntityMeta;
	citations: ParsedCitationEntry[];
	options?: GeneratorMetadata;
}

export type OutlineFormatter = (
	story: StoryWithHierarchy,
	options?: GeneratorMetadata
) => string;

