import {
	Story,
	Chapter,
	SceneWithBeats,
	ContentBlock,
	ContentAnchor,
	Scene,
	Beat,
} from "../types";

export type SyncEntityType = "chapter" | "scene" | "content";

export interface SyncEntityTarget {
	type: SyncEntityType;
	id: string;
}

export interface ChapterEntityPayload {
	type: "chapter";
	story: Story;
	chapter: Chapter;
	scenes: SceneWithBeats[];
	contentBlocks: ContentBlock[];
	contentBlockRefs: ContentAnchor[];
}

export interface SceneEntityPayload {
	type: "scene";
	story: Story;
	scene: Scene;
	beats: Beat[];
	sceneContentBlocks?: ContentBlock[];
	beatContentBlocks?: Record<string, ContentBlock[]>;
}

export interface ContentEntityPayload {
	type: "content";
	story: Story;
	contentBlock: ContentBlock;
}

export type SyncEntityPayload =
	| ChapterEntityPayload
	| SceneEntityPayload
	| ContentEntityPayload;



