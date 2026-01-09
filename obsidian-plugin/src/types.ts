export interface StoryEngineSettings {
	apiUrl: string;
	llmGatewayUrl: string;
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

export interface ExtractSearchChunk {
	chunk_id: string;
	document_id: string;
	source_type: string;
	source_id: string;
	content: string;
	score: number;
	beat_type?: string;
	beat_intent?: string;
	characters?: string[];
	location_name?: string;
	timeline?: string;
	pov_character?: string;
	content_kind?: string;
}

export interface ExtractSearchResult {
	query: string;
	chunks: ExtractSearchChunk[];
	next_cursor?: string;
	received_at: string;
}

export interface ExtractEntityMatch {
	source_type: string;
	source_id: string;
	entity_name?: string;
	similarity: number;
	reason?: string;
}

export interface ExtractEntity {
	type: string;
	name: string;
	summary?: string;
	found: boolean;
	created?: boolean;
	match?: ExtractEntityMatch;
	candidates?: ExtractEntityMatch[];
}

export interface ExtractEntityResult {
	text: string;
	world_id: string;
	entities: ExtractEntity[];
	received_at: string;
}

export interface ExtractStreamEvent {
	type: string;
	phase?: string;
	message: string;
	data?: Record<string, any>;
	timestamp?: string;
}

export interface ExtractLogEntry {
	id: string;
	eventType: string;
	phase?: string;
	message: string;
	data?: Record<string, any>;
	timestamp: string;
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
	time_config?: TimeConfig | null;
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

export interface ContentAnchor {
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
	parent_id?: string | null;
	timeline_position?: number;
	is_epoch?: boolean;
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

// Archetype - global por tenant (NAO tem world_id)
export interface Archetype {
	id: string;
	tenant_id: string;
	name: string;
	description: string;
	created_at: string;
	updated_at: string;
}

// ArchetypeTrait - relaciona archetype com trait (template)
export interface ArchetypeTrait {
	id: string;
	archetype_id: string;
	trait_id: string;
	default_value: string;
	created_at: string;
}

// Faction - hierarquica, pertence a um world
export interface Faction {
	id: string;
	tenant_id: string;
	world_id: string;
	parent_id?: string | null;
	name: string;
	type?: string | null;
	description: string;
	beliefs: string;
	structure: string;
	symbols: string;
	hierarchy_level: number;
	created_at: string;
	updated_at: string;
}

// Lore - hierarquica, pertence a um world
export interface Lore {
	id: string;
	tenant_id: string;
	world_id: string;
	parent_id?: string | null;
	name: string;
	category?: string | null;
	description: string;
	rules: string;
	limitations: string;
	requirements: string;
	hierarchy_level: number;
	created_at: string;
	updated_at: string;
}

// CharacterTrait - trait atribuido a character com value customizado
export interface CharacterTrait {
	id: string;
	character_id: string;
	trait_id: string;
	trait_name: string;
	trait_category: string;
	trait_description: string;
	value: string;
	notes: string;
	created_at: string;
	updated_at: string;
}

// CharacterRelationship - auto-relacionamento entre characters
export interface CharacterRelationship {
	id: string;
	tenant_id: string;
	character1_id: string;
	character2_id: string;
	relationship_type: string; // "ally", "enemy", "family", "lover", "rival", "mentor", "student"
	description: string;
	bidirectional: boolean;
	created_at: string;
	updated_at: string;
}

// EventCharacter - relaciona evento com character
export interface EventCharacter {
	id: string;
	event_id: string;
	character_id: string;
	role?: string | null;
	created_at: string;
}

// EventReference - relaciona evento com entidades
export interface EventReference {
	id: string;
	event_id: string;
	entity_type: "character" | "location" | "artifact" | "faction";
	entity_id: string;
	relationship_type?: string | null;
	notes: string;
	created_at: string;
}

// SceneReference - relaciona scene com entity
export interface SceneReference {
	id: string;
	scene_id: string;
	entity_type: "character" | "location" | "artifact";
	entity_id: string;
	created_at: string;
}

// FactionReference - relaciona faction com outras entidades
export interface FactionReference {
	id: string;
	faction_id: string;
	entity_type: string;
	entity_id: string;
	role?: string | null;
	notes: string;
	created_at: string;
}

// LoreReference - relaciona lore com outras entidades
export interface LoreReference {
	id: string;
	lore_id: string;
	entity_type: string;
	entity_id: string;
	relationship_type?: string | null;
	notes: string;
	created_at: string;
}

// TimeConfig - configuracao de calendario do world
export interface TimeConfig {
	base_unit: string;
	hours_per_day: number;
	days_per_week: number;
	days_per_year: number;
	months_per_year: number;
	month_lengths?: number[];
	month_names?: string[];
	day_names?: string[];
	era_name?: string;
	year_zero?: number;
}

// WorldDate - data no calendario do world
export interface WorldDate {
	year: number;
	month: number;
	day: number;
	hour: number;
	minute: number;
}
