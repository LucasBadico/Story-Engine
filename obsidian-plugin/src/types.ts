export interface StoryEngineSettings {
	apiUrl: string;
	apiKey: string;
	tenantId: string;
	tenantName: string;
	syncFolderPath: string;
	autoVersionSnapshots: boolean;
	conflictResolution: "service" | "local" | "manual";
	unsplashAccessKey?: string;
	unsplashSecretKey?: string;
	mode: "local" | "remote";
	showHelpBox: boolean;
	localModeVideoUrl?: string;
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
	world_id?: string | null;
	created_by_user_id: string;
	created_at: string;
	updated_at: string;
}

export interface World {
	id: string;
	tenant_id: string;
	name: string;
	description: string;
	genre: string;
	is_implicit: boolean;
	rpg_system_id?: string | null;
	created_at: string;
	updated_at: string;
}

export interface RPGSystem {
	id: string;
	tenant_id?: string | null;
	name: string;
	description?: string | null;
	base_stats_schema: Record<string, any>;
	derived_stats_schema?: Record<string, any> | null;
	progression_schema?: Record<string, any> | null;
	is_builtin: boolean;
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

export interface ContentMetadata {
	word_count?: number;
	alt_text?: string;
	caption?: string;
	width?: number;
	height?: number;
	mime_type?: string;
	provider?: string;
	video_id?: string;
	duration?: number;
	thumbnail_url?: string;
	transcript?: string;
	html?: string;
	title?: string;
	description?: string;
	image_url?: string;
	site_name?: string;
	// Image attribution
	attribution?: string;
	attribution_url?: string;
	author_name?: string;
	source?: "unsplash" | "internet link" | "local";
}

export interface ContentBlock {
	id: string;
	chapter_id?: string | null;
	order_num?: number | null;
	type: "text" | "image" | "video" | "audio" | "embed" | "link";
	kind: string;
	content: string;
	metadata: ContentMetadata;
	created_at: string;
	updated_at: string;
}

export interface ContentBlockReference {
	id: string;
	content_block_id: string;
	entity_type: string;
	entity_id: string;
	created_at: string;
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

// World entities
export interface Character {
	id: string;
	tenant_id: string;
	world_id: string;
	archetype_id?: string | null;
	current_class_id?: string | null;
	class_level: number;
	name: string;
	description: string;
	created_at: string;
	updated_at: string;
}

export interface Location {
	id: string;
	tenant_id: string;
	world_id: string;
	parent_id?: string | null;
	name: string;
	type: string;
	description: string;
	hierarchy_level: number;
	created_at: string;
	updated_at: string;
}

export interface Artifact {
	id: string;
	tenant_id: string;
	world_id: string;
	name: string;
	description: string;
	rarity: string;
	created_at: string;
	updated_at: string;
}

export interface WorldEvent {
	id: string;
	tenant_id: string;
	world_id: string;
	name: string;
	type?: string | null;
	description?: string | null;
	timeline?: string | null;
	importance: number;
	created_at: string;
	updated_at: string;
}

export interface Trait {
	id: string;
	tenant_id: string;
	name: string;
	category: string;
	description: string;
	created_at: string;
	updated_at: string;
}

