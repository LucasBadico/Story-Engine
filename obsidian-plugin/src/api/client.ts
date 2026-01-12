import {
	ErrorResponse,
	Story,
	Tenant,
	Chapter,
	Scene,
	Beat,
	StoryWithHierarchy,
	ChapterWithContent,
	SceneWithBeats,
	ContentBlock,
	ContentAnchor,
	World,
	RPGSystem,
	Character,
	Location,
	Artifact,
	WorldEvent,
	Trait,
	Archetype,
	ArchetypeTrait,
	Faction,
	FactionReference,
	Lore,
	LoreReference,
	CharacterTrait,
	CharacterRelationship,
	EventCharacter,
	EventReference,
	SceneReference,
	TimeConfig,
	EntityRelation,
	CreateEntityRelationInput,
} from "../types";
import { SyncEntityPayload } from "../sync/entitySyncTypes";
import { apiUpdateNotifier } from "../sync/apiUpdateNotifier";

export class StoryEngineClient {
	private mode: "local" | "remote" = "remote";
	private autoSyncOnApiUpdates = true;

	constructor(
		private apiUrl: string,
		private apiKey: string,
		private tenantId: string = ""
	) {}

	setTenantId(tenantId: string): void {
		this.tenantId = tenantId.trim();
	}

	setMode(mode: "local" | "remote"): void {
		this.mode = mode;
	}

	setAutoSyncOnApiUpdates(enabled: boolean): void {
		this.autoSyncOnApiUpdates = enabled;
	}

	private async request<T>(
		method: string,
		endpoint: string,
		body?: unknown,
		tenantIdOverride?: string
	): Promise<T> {
		const url = `${this.apiUrl}${endpoint}`;
		const headers = new Headers();
		
		headers.set("Content-Type", "application/json");

		if (this.apiKey) {
			headers.set("Authorization", `Bearer ${this.apiKey}`);
		}

		// Use override if provided, otherwise use instance tenantId
		const effectiveTenantId = tenantIdOverride ?? this.tenantId;
		if (effectiveTenantId) {
			const trimmedTenantId = effectiveTenantId.trim();
			if (trimmedTenantId) {
				headers.set("X-Tenant-ID", trimmedTenantId);
			}
		}

		const options: RequestInit = {
			method,
			headers,
		};

		if (body) {
			options.body = JSON.stringify(body);
		}

		const response = await fetch(url, options);

		if (!response.ok) {
			let error: ErrorResponse;
			try {
				error = (await response.json()) as ErrorResponse;
			} catch {
				error = {
					error: "unknown_error",
					message: `HTTP ${response.status}: ${response.statusText}`,
					code: "HTTP_ERROR",
				};
			}
			const errorMessage = error.message || error.error || `HTTP ${response.status}: ${response.statusText}`;
			throw new Error(errorMessage);
		}

		return response.json() as Promise<T>;
	}

	async listStories(): Promise<Story[]> {
		if (this.mode === "remote" && (!this.tenantId || !this.tenantId.trim())) {
			throw new Error("Tenant ID is required");
		}

		const response = await this.request<{ stories: Story[] }>(
			"GET",
			"/api/v1/stories"
		);
		return response.stories || [];
	}

	async getStory(id: string): Promise<Story> {
		const response = await this.request<{ story: Story }>(
			"GET",
			`/api/v1/stories/${id}`
		);
		return response.story;
	}

	async createStory(title: string, worldId?: string): Promise<Story> {
		if (this.mode === "remote" && (!this.tenantId || !this.tenantId.trim())) {
			throw new Error("Tenant ID is required");
		}

		const body: Record<string, string> = { title: title.trim() };
		if (worldId) {
			body.world_id = worldId;
		}

		const response = await this.request<{ story: Story }>(
			"POST",
			"/api/v1/stories",
			body
		);
		return response.story;
	}

	async cloneStory(id: string): Promise<Story> {
		if (!this.tenantId || !this.tenantId.trim()) {
			throw new Error("Tenant ID is required");
		}

		const response = await this.request<{ story: Story }>(
			"POST",
			`/api/v1/stories/${id}/clone`,
			{}
		);
		return response.story;
	}

	async getTenant(id: string): Promise<Tenant> {
		const response = await this.request<{ tenant: Tenant }>(
			"GET",
			`/api/v1/tenants/${id}`
		);
		return response.tenant;
	}

	async testConnection(): Promise<boolean> {
		try {
			await this.request<{ status: string }>("GET", "/health");
			return true;
		} catch {
			return false;
		}
	}

	async updateStory(id: string, title: string, status?: string): Promise<Story> {
		const body: Record<string, string> = { title: title.trim() };
		if (status) {
			body.status = status;
		}
		const response = await this.request<{ story: Story }>(
			"PUT",
			`/api/v1/stories/${id}`,
			body
		);
		return response.story;
	}

	async getStoryWithHierarchy(id: string): Promise<StoryWithHierarchy> {
		const story = await this.getStory(id);
		const chapters = await this.getChapters(id);

		const chaptersWithContent: ChapterWithContent[] = await Promise.all(
			chapters.map(async (chapter) => {
				const scenes = await this.getScenes(chapter.id);
				const scenesWithBeats: SceneWithBeats[] = await Promise.all(
					scenes.map(async (scene) => {
						const beats = await this.getBeats(scene.id);
						return { scene, beats };
					})
				);
				return { chapter, scenes: scenesWithBeats };
			})
		);

		return {
			story,
			chapters: chaptersWithContent,
		};
	}

	async createChapter(storyId: string, chapter: Partial<Chapter>): Promise<Chapter> {
		const response = await this.request<{ chapter: Chapter }>(
			"POST",
			"/api/v1/chapters",
			{
				story_id: storyId,
				number: chapter.number,
				title: chapter.title,
				status: chapter.status,
			}
		);
		void this.publishChapterUpdate(response.chapter.id);
		return response.chapter;
	}

	async updateChapter(id: string, chapter: Partial<Chapter>): Promise<Chapter> {
		const response = await this.request<{ chapter: Chapter }>(
			"PUT",
			`/api/v1/chapters/${id}`,
			chapter
		);
		void this.publishChapterUpdate(response.chapter.id);
		return response.chapter;
	}

	async getChapters(storyId: string): Promise<Chapter[]> {
		const response = await this.request<{ chapters: Chapter[] }>(
			"GET",
			`/api/v1/stories/${storyId}/chapters`
		);
		return response.chapters || [];
	}

	async getChapter(id: string): Promise<Chapter> {
		const response = await this.request<{ chapter: Chapter }>(
			"GET",
			`/api/v1/chapters/${id}`
		);
		return response.chapter;
	}

	async deleteChapter(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/chapters/${id}`);
	}

	async createScene(scene: Partial<Scene>): Promise<Scene> {
		const response = await this.request<{ scene: Scene }>(
			"POST",
			"/api/v1/scenes",
			scene
		);
		void this.publishSceneTree(response.scene.id);
		return response.scene;
	}

	async updateScene(id: string, scene: Partial<Scene>): Promise<Scene> {
		const response = await this.request<{ scene: Scene }>(
			"PUT",
			`/api/v1/scenes/${id}`,
			scene
		);
		void this.publishSceneTree(response.scene.id);
		return response.scene;
	}

	async getScenes(chapterId: string): Promise<Scene[]> {
		const response = await this.request<{ scenes: Scene[] }>(
			"GET",
			`/api/v1/chapters/${chapterId}/scenes`
		);
		return response.scenes || [];
	}

	async getScene(id: string): Promise<Scene> {
		const response = await this.request<{ scene: Scene }>(
			"GET",
			`/api/v1/scenes/${id}`
		);
		return response.scene;
	}

	async deleteScene(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/scenes/${id}`);
	}

	async createBeat(beat: Partial<Beat>): Promise<Beat> {
		if (!beat.scene_id) {
			throw new Error("scene_id is required to create a beat");
		}

		const payload: Partial<Beat> = { ...beat };

		if (!payload.order_num || payload.order_num <= 0) {
			const existingBeats = await this.getBeats(beat.scene_id);
			const nextOrder =
				existingBeats.length > 0
					? Math.max(...existingBeats.map((b) => b.order_num || 0)) + 1
					: 1;
			payload.order_num = nextOrder;
		}

		const response = await this.request<{ beat: Beat }>(
			"POST",
			"/api/v1/beats",
			payload
		);
		void this.publishSceneTree(response.beat.scene_id);
		return response.beat;
	}

	async updateBeat(id: string, beat: Partial<Beat>): Promise<Beat> {
		const response = await this.request<{ beat: Beat }>(
			"PUT",
			`/api/v1/beats/${id}`,
			beat
		);
		void this.publishSceneTree(response.beat.scene_id);
		return response.beat;
	}

	async getBeats(sceneId: string): Promise<Beat[]> {
		const response = await this.request<{ beats: Beat[] }>(
			"GET",
			`/api/v1/scenes/${sceneId}/beats`
		);
		return response.beats || [];
	}

	async getBeat(id: string): Promise<Beat> {
		const response = await this.request<{ beat: Beat }>(
			"GET",
			`/api/v1/beats/${id}`
		);
		return response.beat;
	}

	async deleteBeat(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/beats/${id}`);
	}

	// Get all versions of a story (for version history)
	async getStoryVersions(rootStoryId: string): Promise<Story[]> {
		const response = await this.request<{ stories: Story[] }>(
			"GET",
			`/api/v1/stories/${rootStoryId}/versions`
		);
		return response.stories || [];
	}

	async getScenesByStory(storyId: string): Promise<Scene[]> {
		const response = await this.request<{ scenes: Scene[] }>(
			"GET",
			`/api/v1/stories/${storyId}/scenes`
		);
		return response.scenes || [];
	}

	async moveScene(sceneId: string, chapterId: string | null): Promise<Scene> {
		const body: { chapter_id?: string | null } = {};
		if (chapterId !== null) {
			body.chapter_id = chapterId;
		} else {
			body.chapter_id = null;
		}
		const response = await this.request<{ scene: Scene }>(
			"PUT",
			`/api/v1/scenes/${sceneId}/move`,
			body
		);
		void this.publishSceneTree(response.scene.id);
		return response.scene;
	}

	async getBeatsByStory(storyId: string): Promise<Beat[]> {
		const response = await this.request<{ beats: Beat[] }>(
			"GET",
			`/api/v1/stories/${storyId}/beats`
		);
		return response.beats || [];
	}

	async moveBeat(beatId: string, sceneId: string): Promise<Beat> {
		const response = await this.request<{ beat: Beat }>(
			"PUT",
			`/api/v1/beats/${beatId}/move`,
			{ scene_id: sceneId }
		);
		void this.publishSceneTree(response.beat.scene_id);
		return response.beat;
	}

	async getContentBlocks(chapterId: string): Promise<ContentBlock[]> {
		const response = await this.request<{ content_blocks: ContentBlock[] }>(
			"GET",
			`/api/v1/chapters/${chapterId}/content-blocks`
		);
		return response.content_blocks || [];
	}

	async getContentBlock(id: string): Promise<ContentBlock> {
		const response = await this.request<{ content_block: ContentBlock }>(
			"GET",
			`/api/v1/content-blocks/${id}`
		);
		return response.content_block;
	}

	async createContentBlock(chapterId: string, contentBlock: Partial<ContentBlock>): Promise<ContentBlock> {
		const response = await this.request<{ content_block: ContentBlock }>(
			"POST",
			`/api/v1/chapters/${chapterId}/content-blocks`,
			contentBlock
		);
		void this.publishContentBlockUpdate(response.content_block.id);
		return response.content_block;
	}

	async updateContentBlock(id: string, contentBlock: Partial<ContentBlock>): Promise<ContentBlock> {
		const response = await this.request<{ content_block: ContentBlock }>(
			"PUT",
			`/api/v1/content-blocks/${id}`,
			contentBlock
		);
		void this.publishContentBlockUpdate(response.content_block.id);
		return response.content_block;
	}

	async deleteContentBlock(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/content-blocks/${id}`);
	}

	async getContentAnchors(contentBlockId: string): Promise<ContentAnchor[]> {
		const response = await this.request<{ anchors: ContentAnchor[] }>(
			"GET",
			`/api/v1/content-blocks/${contentBlockId}/anchors`
		);
		return response.anchors || [];
	}

	async getContentBlocksByScene(sceneId: string): Promise<ContentBlock[]> {
		const response = await this.request<{ content_blocks: ContentBlock[] }>(
			"GET",
			`/api/v1/scenes/${sceneId}/content-blocks`
		);
		return response.content_blocks || [];
	}

	async getContentBlocksByBeat(beatId: string): Promise<ContentBlock[]> {
		const response = await this.request<{ content_blocks: ContentBlock[] }>(
			"GET",
			`/api/v1/beats/${beatId}/content-blocks`
		);
		return response.content_blocks || [];
	}

	async createContentAnchor(contentBlockId: string, entityType: string, entityId: string): Promise<ContentAnchor> {
		const response = await this.request<{ anchor: ContentAnchor }>(
			"POST",
			`/api/v1/content-blocks/${contentBlockId}/anchors`,
			{
				entity_type: entityType,
				entity_id: entityId,
			}
		);
		return response.anchor;
	}

	async deleteContentAnchor(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/content-anchors/${id}`);
	}

	async getWorlds(): Promise<World[]> {
		const response = await this.request<{ worlds: World[] }>(
			"GET",
			"/api/v1/worlds"
		);
		return response.worlds || [];
	}

	async getWorld(id: string): Promise<World> {
		const response = await this.request<{ world: World }>(
			"GET",
			`/api/v1/worlds/${id}`
		);
		return response.world;
	}

	async createWorld(name: string, description: string, genre: string): Promise<World> {
		const response = await this.request<{ world: World }>(
			"POST",
			"/api/v1/worlds",
			{
				name: name.trim(),
				description: description.trim(),
				genre: genre.trim(),
			}
		);
		return response.world;
	}

	async updateWorld(id: string, name: string, description: string, genre: string): Promise<World> {
		const response = await this.request<{ world: World }>(
			"PUT",
			`/api/v1/worlds/${id}`,
			{
				name: name.trim(),
				description: description.trim(),
				genre: genre.trim(),
			}
		);
		return response.world;
	}

	async getRPGSystems(): Promise<RPGSystem[]> {
		const response = await this.request<{ rpg_systems: RPGSystem[] }>(
			"GET",
			"/api/v1/rpg-systems"
		);
		return response.rpg_systems || [];
	}

	async getRPGSystem(id: string): Promise<RPGSystem> {
		const response = await this.request<{ rpg_system: RPGSystem }>(
			"GET",
			`/api/v1/rpg-systems/${id}`
		);
		return response.rpg_system;
	}

	// Character methods
	async getCharacters(worldId: string): Promise<Character[]> {
		const response = await this.request<{ characters: Character[] }>(
			"GET",
			`/api/v1/worlds/${worldId}/characters`
		);
		return response.characters || [];
	}

	async getCharacter(id: string): Promise<Character> {
		const response = await this.request<{ character: Character }>(
			"GET",
			`/api/v1/characters/${id}`
		);
		return response.character;
	}

	async createCharacter(worldId: string, data: Partial<Character>): Promise<Character> {
		const response = await this.request<{ character: Character }>(
			"POST",
			`/api/v1/worlds/${worldId}/characters`,
			data
		);
		return response.character;
	}

	async updateCharacter(id: string, data: Partial<Character>): Promise<Character> {
		const response = await this.request<{ character: Character }>(
			"PUT",
			`/api/v1/characters/${id}`,
			data
		);
		return response.character;
	}

	async deleteCharacter(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/characters/${id}`);
	}

	// Location methods
	async getLocations(worldId: string): Promise<Location[]> {
		const response = await this.request<{ locations: Location[] }>(
			"GET",
			`/api/v1/worlds/${worldId}/locations`
		);
		return response.locations || [];
	}

	async getLocation(id: string): Promise<Location> {
		const response = await this.request<{ location: Location }>(
			"GET",
			`/api/v1/locations/${id}`
		);
		return response.location;
	}

	async createLocation(worldId: string, data: Partial<Location>): Promise<Location> {
		const response = await this.request<{ location: Location }>(
			"POST",
			`/api/v1/worlds/${worldId}/locations`,
			data
		);
		return response.location;
	}

	async updateLocation(id: string, data: Partial<Location>): Promise<Location> {
		const response = await this.request<{ location: Location }>(
			"PUT",
			`/api/v1/locations/${id}`,
			data
		);
		return response.location;
	}

	async deleteLocation(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/locations/${id}`);
	}

	// Artifact methods
	async getArtifacts(worldId: string): Promise<Artifact[]> {
		const response = await this.request<{ artifacts: Artifact[] }>(
			"GET",
			`/api/v1/worlds/${worldId}/artifacts`
		);
		return response.artifacts || [];
	}

	async getArtifact(id: string): Promise<Artifact> {
		const response = await this.request<{ artifact: Artifact }>(
			"GET",
			`/api/v1/artifacts/${id}`
		);
		return response.artifact;
	}

	async createArtifact(worldId: string, data: Partial<Artifact>): Promise<Artifact> {
		const response = await this.request<{ artifact: Artifact }>(
			"POST",
			`/api/v1/worlds/${worldId}/artifacts`,
			data
		);
		return response.artifact;
	}

	async updateArtifact(id: string, data: Partial<Artifact>): Promise<Artifact> {
		const response = await this.request<{ artifact: Artifact }>(
			"PUT",
			`/api/v1/artifacts/${id}`,
			data
		);
		return response.artifact;
	}

	async deleteArtifact(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/artifacts/${id}`);
	}

	// Event methods
	async getEvents(worldId: string): Promise<WorldEvent[]> {
		const response = await this.request<{ events: WorldEvent[] }>(
			"GET",
			`/api/v1/worlds/${worldId}/events`
		);
		return response.events || [];
	}

	async getEvent(id: string): Promise<WorldEvent> {
		const response = await this.request<{ event: WorldEvent }>(
			"GET",
			`/api/v1/events/${id}`
		);
		return response.event;
	}

	async createEvent(worldId: string, data: Partial<WorldEvent>): Promise<WorldEvent> {
		const response = await this.request<{ event: WorldEvent }>(
			"POST",
			`/api/v1/worlds/${worldId}/events`,
			data
		);
		return response.event;
	}

	async updateEvent(id: string, data: Partial<WorldEvent>): Promise<WorldEvent> {
		const response = await this.request<{ event: WorldEvent }>(
			"PUT",
			`/api/v1/events/${id}`,
			data
		);
		return response.event;
	}

	async deleteEvent(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/events/${id}`);
	}

	async moveEvent(id: string, parentId: string | null): Promise<WorldEvent> {
		const response = await this.request<{ event: WorldEvent }>(
			"PUT",
			`/api/v1/events/${id}/move`,
			{ parent_id: parentId }
		);
		return response.event;
	}

	// Trait methods
	async getTraits(): Promise<Trait[]> {
		const response = await this.request<{ traits: Trait[] }>(
			"GET",
			"/api/v1/traits"
		);
		return response.traits || [];
	}

	async getTrait(id: string): Promise<Trait> {
		const response = await this.request<{ trait: Trait }>(
			"GET",
			`/api/v1/traits/${id}`
		);
		return response.trait;
	}

	async createTrait(data: Partial<Trait>): Promise<Trait> {
		const response = await this.request<{ trait: Trait }>(
			"POST",
			"/api/v1/traits",
			data
		);
		return response.trait;
	}

	async updateTrait(id: string, data: Partial<Trait>): Promise<Trait> {
		const response = await this.request<{ trait: Trait }>(
			"PUT",
			`/api/v1/traits/${id}`,
			data
		);
		return response.trait;
	}

	async deleteTrait(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/traits/${id}`);
	}

	// Archetype methods
	async getArchetypes(): Promise<Archetype[]> {
		const response = await this.request<{ archetypes: Archetype[] }>(
			"GET",
			"/api/v1/archetypes"
		);
		return response.archetypes || [];
	}

	async getArchetype(id: string): Promise<Archetype> {
		const response = await this.request<{ archetype: Archetype }>(
			"GET",
			`/api/v1/archetypes/${id}`
		);
		return response.archetype;
	}

	async createArchetype(data: Partial<Archetype>): Promise<Archetype> {
		const response = await this.request<{ archetype: Archetype }>(
			"POST",
			"/api/v1/archetypes",
			data
		);
		return response.archetype;
	}

	async updateArchetype(id: string, data: Partial<Archetype>): Promise<Archetype> {
		const response = await this.request<{ archetype: Archetype }>(
			"PUT",
			`/api/v1/archetypes/${id}`,
			data
		);
		return response.archetype;
	}

	async deleteArchetype(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/archetypes/${id}`);
	}

	async getArchetypeTraits(archetypeId: string): Promise<ArchetypeTrait[]> {
		const response = await this.request<{ archetype_traits: ArchetypeTrait[] }>(
			"GET",
			`/api/v1/archetypes/${archetypeId}/traits`
		);
		return response.archetype_traits || [];
	}

	async addArchetypeTrait(archetypeId: string, traitId: string, defaultValue?: string): Promise<void> {
		await this.request(
			"POST",
			`/api/v1/archetypes/${archetypeId}/traits`,
			{
				trait_id: traitId,
				default_value: defaultValue,
			}
		);
	}

	async removeArchetypeTrait(archetypeId: string, traitId: string): Promise<void> {
		await this.request(
			"DELETE",
			`/api/v1/archetypes/${archetypeId}/traits/${traitId}`
		);
	}

	// Faction methods
	async getFactions(worldId: string): Promise<Faction[]> {
		const response = await this.request<{ factions: Faction[] }>(
			"GET",
			`/api/v1/worlds/${worldId}/factions`
		);
		return response.factions || [];
	}

	async getFaction(id: string): Promise<Faction> {
		const response = await this.request<{ faction: Faction }>(
			"GET",
			`/api/v1/factions/${id}`
		);
		return response.faction;
	}

	async createFaction(worldId: string, data: Partial<Faction>): Promise<Faction> {
		const response = await this.request<{ faction: Faction }>(
			"POST",
			`/api/v1/worlds/${worldId}/factions`,
			data
		);
		return response.faction;
	}

	async updateFaction(id: string, data: Partial<Faction>): Promise<Faction> {
		const response = await this.request<{ faction: Faction }>(
			"PUT",
			`/api/v1/factions/${id}`,
			data
		);
		return response.faction;
	}

	async deleteFaction(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/factions/${id}`);
	}

	async getFactionChildren(id: string): Promise<Faction[]> {
		const response = await this.request<{ factions: Faction[] }>(
			"GET",
			`/api/v1/factions/${id}/children`
		);
		return response.factions || [];
	}

	async getFactionReferences(factionId: string): Promise<FactionReference[]> {
		const response = await this.request<{ references: FactionReference[] }>(
			"GET",
			`/api/v1/factions/${factionId}/references`
		);
		return response.references || [];
	}

	async addFactionReference(factionId: string, entityType: string, entityId: string, role?: string, notes?: string): Promise<FactionReference> {
		const response = await this.request<{ reference: FactionReference }>(
			"POST",
			`/api/v1/factions/${factionId}/references`,
			{
				entity_type: entityType,
				entity_id: entityId,
				role: role,
				notes: notes,
			}
		);
		return response.reference;
	}

	async updateFactionReference(id: string, role?: string, notes?: string): Promise<FactionReference> {
		const response = await this.request<{ reference: FactionReference }>(
			"PUT",
			`/api/v1/faction-references/${id}`,
			{
				role: role,
				notes: notes,
			}
		);
		return response.reference;
	}

	async removeFactionReference(factionId: string, entityType: string, entityId: string): Promise<void> {
		await this.request(
			"DELETE",
			`/api/v1/factions/${factionId}/references/${entityType}/${entityId}`
		);
	}

	// Lore methods
	async getLores(worldId: string): Promise<Lore[]> {
		const response = await this.request<{ lores: Lore[] }>(
			"GET",
			`/api/v1/worlds/${worldId}/lores`
		);
		return response.lores || [];
	}

	async getLore(id: string): Promise<Lore> {
		const response = await this.request<{ lore: Lore }>(
			"GET",
			`/api/v1/lores/${id}`
		);
		return response.lore;
	}

	async createLore(worldId: string, data: Partial<Lore>): Promise<Lore> {
		const response = await this.request<{ lore: Lore }>(
			"POST",
			`/api/v1/worlds/${worldId}/lores`,
			data
		);
		return response.lore;
	}

	async updateLore(id: string, data: Partial<Lore>): Promise<Lore> {
		const response = await this.request<{ lore: Lore }>(
			"PUT",
			`/api/v1/lores/${id}`,
			data
		);
		return response.lore;
	}

	async deleteLore(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/lores/${id}`);
	}

	async getLoreChildren(id: string): Promise<Lore[]> {
		const response = await this.request<{ lores: Lore[] }>(
			"GET",
			`/api/v1/lores/${id}/children`
		);
		return response.lores || [];
	}

	async getLoreReferences(loreId: string): Promise<LoreReference[]> {
		const response = await this.request<{ references: LoreReference[] }>(
			"GET",
			`/api/v1/lores/${loreId}/references`
		);
		return response.references || [];
	}

	async addLoreReference(loreId: string, entityType: string, entityId: string, relationshipType?: string, notes?: string): Promise<LoreReference> {
		const response = await this.request<{ reference: LoreReference }>(
			"POST",
			`/api/v1/lores/${loreId}/references`,
			{
				entity_type: entityType,
				entity_id: entityId,
				relationship_type: relationshipType,
				notes: notes,
			}
		);
		return response.reference;
	}

	async updateLoreReference(id: string, relationshipType?: string, notes?: string): Promise<LoreReference> {
		const response = await this.request<{ reference: LoreReference }>(
			"PUT",
			`/api/v1/lore-references/${id}`,
			{
				relationship_type: relationshipType,
				notes: notes,
			}
		);
		return response.reference;
	}

	async removeLoreReference(loreId: string, entityType: string, entityId: string): Promise<void> {
		await this.request(
			"DELETE",
			`/api/v1/lores/${loreId}/references/${entityType}/${entityId}`
		);
	}

	// Character Traits methods
	async getCharacterTraits(characterId: string): Promise<CharacterTrait[]> {
		const response = await this.request<{ character_traits: CharacterTrait[] }>(
			"GET",
			`/api/v1/characters/${characterId}/traits`
		);
		return response.character_traits || [];
	}

	async addCharacterTrait(characterId: string, traitId: string, value?: string, notes?: string): Promise<CharacterTrait> {
		const response = await this.request<{ character_trait: CharacterTrait }>(
			"POST",
			`/api/v1/characters/${characterId}/traits`,
			{
				trait_id: traitId,
				value: value,
				notes: notes,
			}
		);
		return response.character_trait;
	}

	async updateCharacterTrait(characterId: string, traitId: string, value?: string, notes?: string): Promise<CharacterTrait> {
		const response = await this.request<{ character_trait: CharacterTrait }>(
			"PUT",
			`/api/v1/characters/${characterId}/traits/${traitId}`,
			{
				value: value,
				notes: notes,
			}
		);
		return response.character_trait;
	}

	async removeCharacterTrait(characterId: string, traitId: string): Promise<void> {
		await this.request(
			"DELETE",
			`/api/v1/characters/${characterId}/traits/${traitId}`
		);
	}

	// Character Events methods
	async getCharacterEvents(characterId: string): Promise<EventCharacter[]> {
		const response = await this.request<{ event_characters: EventCharacter[] }>(
			"GET",
			`/api/v1/characters/${characterId}/events`
		);
		return response.event_characters || [];
	}

	// Character Relationships methods
	async getCharacterRelationships(characterId: string): Promise<CharacterRelationship[]> {
		const response = await this.request<{ relationships: CharacterRelationship[] }>(
			"GET",
			`/api/v1/characters/${characterId}/relationships`
		);
		return response.relationships || [];
	}

	async createCharacterRelationship(characterId: string, data: Partial<CharacterRelationship>): Promise<CharacterRelationship> {
		const response = await this.request<{ relationship: CharacterRelationship }>(
			"POST",
			`/api/v1/characters/${characterId}/relationships`,
			data
		);
		return response.relationship;
	}

	async updateCharacterRelationship(id: string, data: Partial<CharacterRelationship>): Promise<CharacterRelationship> {
		const response = await this.request<{ relationship: CharacterRelationship }>(
			"PUT",
			`/api/v1/character-relationships/${id}`,
			data
		);
		return response.relationship;
	}

	async deleteCharacterRelationship(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/character-relationships/${id}`);
	}

	async createEntityRelation(input: CreateEntityRelationInput): Promise<EntityRelation> {
		const response = await this.request<{ relation: EntityRelation }>(
			"POST",
			"/api/v1/relations",
			{
				world_id: input.world_id,
				source_type: input.source_type,
				source_id: input.source_id,
				target_type: input.target_type,
				target_id: input.target_id,
				relation_type: input.relation_type,
				summary: input.summary,
				create_mirror: input.create_mirror ?? false,
			}
		);
		return response.relation;
	}

	// Event Characters/References methods
	async getEventCharacters(eventId: string): Promise<EventCharacter[]> {
		const response = await this.request<{ event_characters: EventCharacter[] }>(
			"GET",
			`/api/v1/events/${eventId}/characters`
		);
		return response.event_characters || [];
	}

	async addEventCharacter(eventId: string, characterId: string, role?: string): Promise<EventCharacter> {
		const response = await this.request<{ event_character: EventCharacter }>(
			"POST",
			`/api/v1/events/${eventId}/characters`,
			{
				character_id: characterId,
				role: role,
			}
		);
		return response.event_character;
	}

	async removeEventCharacter(eventId: string, characterId: string): Promise<void> {
		await this.request(
			"DELETE",
			`/api/v1/events/${eventId}/characters/${characterId}`
		);
	}

	async getEventReferences(eventId: string): Promise<EventReference[]> {
		const response = await this.request<{ references: EventReference[] }>(
			"GET",
			`/api/v1/events/${eventId}/references`
		);
		return response.references || [];
	}

	async addEventReference(eventId: string, entityType: string, entityId: string, relationshipType?: string, notes?: string): Promise<EventReference> {
		const response = await this.request<{ reference: EventReference }>(
			"POST",
			`/api/v1/events/${eventId}/references`,
			{
				entity_type: entityType,
				entity_id: entityId,
				relationship_type: relationshipType,
				notes: notes,
			}
		);
		return response.reference;
	}

	async updateEventReference(id: string, relationshipType?: string, notes?: string): Promise<EventReference> {
		const response = await this.request<{ reference: EventReference }>(
			"PUT",
			`/api/v1/event-references/${id}`,
			{
				relationship_type: relationshipType,
				notes: notes,
			}
		);
		return response.reference;
	}

	async removeEventReference(eventId: string, entityType: string, entityId: string): Promise<void> {
		await this.request(
			"DELETE",
			`/api/v1/events/${eventId}/references/${entityType}/${entityId}`
		);
	}

	// Scene References methods
	async getSceneReferences(sceneId: string): Promise<SceneReference[]> {
		const response = await this.request<{ references: SceneReference[] }>(
			"GET",
			`/api/v1/scenes/${sceneId}/references`
		);
		return response.references || [];
	}

	async addSceneReference(sceneId: string, entityType: string, entityId: string): Promise<SceneReference> {
		const response = await this.request<{ reference: SceneReference }>(
			"POST",
			`/api/v1/scenes/${sceneId}/references`,
			{
				entity_type: entityType,
				entity_id: entityId,
			}
		);
		return response.reference;
	}

	async removeSceneReference(sceneId: string, entityType: string, entityId: string): Promise<void> {
		await this.request(
			"DELETE",
			`/api/v1/scenes/${sceneId}/references/${entityType}/${entityId}`
		);
	}

	// Timeline methods
	async getTimeline(worldId: string, fromPos?: number, toPos?: number): Promise<WorldEvent[]> {
		const queryParams = new URLSearchParams();
		if (fromPos !== undefined) {
			queryParams.append("from_pos", fromPos.toString());
		}
		if (toPos !== undefined) {
			queryParams.append("to_pos", toPos.toString());
		}
		const queryString = queryParams.toString();
		const endpoint = `/api/v1/worlds/${worldId}/timeline${queryString ? `?${queryString}` : ""}`;
		const response = await this.request<{ events: WorldEvent[] }>(
			"GET",
			endpoint
		);
		return response.events || [];
	}

	// World TimeConfig methods
	async updateWorldTimeConfig(worldId: string, timeConfig: TimeConfig): Promise<World> {
		const response = await this.request<{ world: World }>(
			"PUT",
			`/api/v1/worlds/${worldId}/time-config`,
			timeConfig
		);
		return response.world;
	}

	private isAutoSyncEnabled(): boolean {
		return this.autoSyncOnApiUpdates;
	}

	private async publishChapterUpdate(chapterId: string): Promise<void> {
		if (!this.isAutoSyncEnabled()) {
			return;
		}

		try {
			const chapter = await this.getChapter(chapterId);
			const story = await this.getStory(chapter.story_id);
			const scenes = await this.getScenes(chapterId);
			const scenesWithBeats: SceneWithBeats[] = await Promise.all(
				scenes.map(async (scene) => {
					const beats = await this.getBeats(scene.id);
					return { scene, beats };
				})
			);
			const contentBlocks = await this.getContentBlocks(chapterId);
			const contentBlockRefs: ContentAnchor[] = [];
			for (const block of contentBlocks) {
				const refs = await this.getContentAnchors(block.id);
				contentBlockRefs.push(...refs);
			}

			await this.notifyEntityUpdate({
				type: "chapter",
				story,
				chapter,
				scenes: scenesWithBeats,
				contentBlocks,
				contentBlockRefs,
			});
		} catch (err) {
			console.error("Failed to auto-sync chapter update", err);
		}
	}

	private async publishSceneTree(sceneId: string): Promise<void> {
		if (!this.isAutoSyncEnabled()) {
			return;
		}

		try {
			const scene = await this.getScene(sceneId);
			if (scene.chapter_id) {
				await this.publishChapterUpdate(scene.chapter_id);
				return;
			}

			const story = await this.getStory(scene.story_id);
			const beats = await this.getBeats(scene.id);
			const sceneContentBlocks = await this.getContentBlocksByScene(scene.id);
			const beatContentBlocks: Record<string, ContentBlock[]> = {};
			for (const beat of beats) {
				beatContentBlocks[beat.id] = await this.getContentBlocksByBeat(beat.id);
			}

			await this.notifyEntityUpdate({
				type: "scene",
				story,
				scene,
				beats,
				sceneContentBlocks,
				beatContentBlocks,
			});
		} catch (err) {
			console.error("Failed to auto-sync scene update", err);
		}
	}

	private async publishContentBlockUpdate(contentBlockId: string): Promise<void> {
		if (!this.isAutoSyncEnabled()) {
			return;
		}

		try {
			const contentBlock = await this.getContentBlock(contentBlockId);
			let story: Story | null = null;
			if (contentBlock.chapter_id) {
				const chapter = await this.getChapter(contentBlock.chapter_id);
				story = await this.getStory(chapter.story_id);
			} else {
				const anchors = await this.getContentAnchors(contentBlock.id);
				const sceneAnchor = anchors.find((anchor) => anchor.entity_type === "scene");
				if (sceneAnchor) {
					const scene = await this.getScene(sceneAnchor.entity_id);
					story = await this.getStory(scene.story_id);
				}
			}

			if (!story) {
				console.warn("Unable to resolve story for content block auto-sync", contentBlockId);
				return;
			}

			await this.notifyEntityUpdate({
				type: "content",
				story,
				contentBlock,
			});
		} catch (err) {
			console.error("Failed to auto-sync content block update", err);
		}
	}

	async notifyEntityUpdate(payload: SyncEntityPayload): Promise<void> {
		await apiUpdateNotifier.notify(payload);
	}
}
