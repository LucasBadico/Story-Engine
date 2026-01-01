export interface StoryEngineSettings {
	apiUrl: string;
	apiKey: string;
	tenantId: string;
	tenantName: string;
	syncFolderPath: string;
	autoVersionSnapshots: boolean;
	conflictResolution: "service" | "local" | "manual";
}

export interface ErrorResponse {
	error: string;
	message: string;
	code: string;
}

export interface Tenant {
	id: string;
	name: string;
	created_at: string;
	updated_at: string;
}

export interface Story {
	id: string;
	tenant_id: string;
	title: string;
	status: string;
	version_number: number;
	root_story_id: string;
	previous_story_id: string | null;
	created_by_user_id: string;
	created_at: string;
	updated_at: string;
}

export interface Chapter {
	id: string;
	story_id: string;
	number: number;
	title: string;
	status: string;
	created_at: string;
	updated_at: string;
}

export interface Scene {
	id: string;
	story_id: string;
	chapter_id?: string | null;
	order_num: number;
	pov_character_id?: string | null;
	location_id?: string | null;
	time_ref: string;
	goal: string;
	created_at: string;
	updated_at: string;
}

export interface Beat {
	id: string;
	scene_id: string;
	order_num: number;
	type: string;
	intent: string;
	outcome: string;
	created_at: string;
	updated_at: string;
}

export interface User {
	id: string;
	tenant_id: string;
	email: string;
	display_name: string;
	created_at: string;
	updated_at: string;
}

export interface Membership {
	id: string;
	tenant_id: string;
	user_id: string;
	role: string;
	created_at: string;
	updated_at: string;
}

export interface SceneWithBeats {
	scene: Scene;
	beats: Beat[];
}

export interface ChapterWithContent {
	chapter: Chapter;
	scenes: SceneWithBeats[];
}

export interface StoryWithHierarchy {
	story: Story;
	chapters: ChapterWithContent[];
}

export interface StoryMetadata {
	frontmatter: {
		id: string;
		title: string;
		status: string;
		version: number;
		root_story_id: string;
		previous_version_id: string | null;
		created_at: string;
		updated_at: string;
	};
	content: string;
}

