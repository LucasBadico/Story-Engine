import type { SyncEntityTarget } from "./entitySyncTypes";

export interface SyncEngine {
	pullStory(storyId: string, target?: SyncEntityTarget): Promise<void>;
	pullAllStories(): Promise<void>;
	pushStory(folderPath: string, target?: SyncEntityTarget): Promise<void>;
	dispose(): void;
}

export interface SyncEngineDependencies {
	app: import("obsidian").App;
	apiClient: import("../api/client").StoryEngineClient;
	fileManager: import("./fileManager").FileManager;
	settings: import("../types").StoryEngineSettings;
}

