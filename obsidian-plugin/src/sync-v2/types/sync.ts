import type { App } from "obsidian";
import type { StoryEngineClient } from "../../api/client";
import type { FileManager } from "../../sync/fileManager";
import type { StoryEngineSettings } from "../../types";
import type { SyncEntityTarget } from "../../sync/entitySyncTypes";

export type SyncOperationType =
	| "pull_story"
	| "pull_all_stories"
	| "push_story"
	| "pull_chapter"
	| "pull_scene"
	| "pull_beat"
	| "pull_content_block"
	| "pull_world"
	| "pull_character"
	| "pull_location"
	| "pull_faction"
	| "pull_artifact"
	| "pull_event"
	| "pull_lore"
	| "pull_archetype"
	| "pull_trait";

export interface SyncStats {
	created?: number;
	updated?: number;
	deleted?: number;
	skipped?: number;
	durationMs?: number;
}

export interface SyncError {
	code: string;
	message: string;
	details?: unknown;
	recoverable?: boolean;
}

export interface SyncResult {
	success: boolean;
	message?: string;
	stats?: SyncStats;
	errors?: SyncError[];
	warnings?: SyncWarning[];
}

export interface SyncContext {
	app: App;
	apiClient: StoryEngineClient;
	fileManager: FileManager;
	settings: StoryEngineSettings;
	timestamp(): string;
	backupMode: "off" | "snapshots" | "git";
	emitWarning?: (warning: SyncWarning) => void;
}

export interface SyncWarning {
	code: string;
	message: string;
	details?: unknown;
	filePath?: string;
	severity?: "info" | "warning";
}

export interface PullStoryPayload {
	storyId: string;
	target?: SyncEntityTarget;
}

export interface PullChapterPayload {
	chapterId: string;
}

export interface PullScenePayload {
	sceneId: string;
}

export interface PullBeatPayload {
	beatId: string;
}

export interface PullContentBlockPayload {
	contentBlockId: string;
}

export interface PullWorldPayload {
	worldId: string;
}

export interface PullWorldEntityPayload {
	entityId: string;
}

export interface PullAllStoriesPayload {
	includeWorlds?: boolean;
}

export interface PushStoryPayload {
	folderPath: string;
	target?: SyncEntityTarget;
}

export type SyncOperation =
	| { type: "pull_story"; payload: PullStoryPayload }
	| { type: "pull_all_stories"; payload: PullAllStoriesPayload }
	| { type: "push_story"; payload: PushStoryPayload }
	| { type: "pull_chapter"; payload: PullChapterPayload }
	| { type: "pull_scene"; payload: PullScenePayload }
	| { type: "pull_beat"; payload: PullBeatPayload }
	| { type: "pull_content_block"; payload: PullContentBlockPayload }
	| { type: "pull_world"; payload: PullWorldPayload }
	| { type: "pull_character"; payload: PullWorldEntityPayload }
	| { type: "pull_location"; payload: PullWorldEntityPayload }
	| { type: "pull_faction"; payload: PullWorldEntityPayload }
	| { type: "pull_artifact"; payload: PullWorldEntityPayload }
	| { type: "pull_event"; payload: PullWorldEntityPayload }
	| { type: "pull_lore"; payload: PullWorldEntityPayload }
	| { type: "pull_archetype"; payload: PullWorldEntityPayload }
	| { type: "pull_trait"; payload: PullWorldEntityPayload };

