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
	ContentBlockReference,
	World,
	RPGSystem,
	Character,
	Location,
	Artifact,
	WorldEvent,
	Trait,
} from "../types";

export class StoryEngineClient {
	private mode: "local" | "remote" = "remote";

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
		return response.chapter;
	}

	async updateChapter(id: string, chapter: Partial<Chapter>): Promise<Chapter> {
		const response = await this.request<{ chapter: Chapter }>(
			"PUT",
			`/api/v1/chapters/${id}`,
			chapter
		);
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
		return response.scene;
	}

	async updateScene(id: string, scene: Partial<Scene>): Promise<Scene> {
		const response = await this.request<{ scene: Scene }>(
			"PUT",
			`/api/v1/scenes/${id}`,
			scene
		);
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
		const response = await this.request<{ beat: Beat }>(
			"POST",
			"/api/v1/beats",
			beat
		);
		return response.beat;
	}

	async updateBeat(id: string, beat: Partial<Beat>): Promise<Beat> {
		const response = await this.request<{ beat: Beat }>(
			"PUT",
			`/api/v1/beats/${id}`,
			beat
		);
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
		return response.content_block;
	}

	async updateContentBlock(id: string, contentBlock: Partial<ContentBlock>): Promise<ContentBlock> {
		const response = await this.request<{ content_block: ContentBlock }>(
			"PUT",
			`/api/v1/content-blocks/${id}`,
			contentBlock
		);
		return response.content_block;
	}

	async deleteContentBlock(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/content-blocks/${id}`);
	}

	async getContentBlockReferences(contentBlockId: string): Promise<ContentBlockReference[]> {
		const response = await this.request<{ references: ContentBlockReference[] }>(
			"GET",
			`/api/v1/content-blocks/${contentBlockId}/references`
		);
		return response.references || [];
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

	async createContentBlockReference(contentBlockId: string, entityType: string, entityId: string): Promise<ContentBlockReference> {
		const response = await this.request<{ reference: ContentBlockReference }>(
			"POST",
			`/api/v1/content-blocks/${contentBlockId}/references`,
			{
				entity_type: entityType,
				entity_id: entityId,
			}
		);
		return response.reference;
	}

	async deleteContentBlockReference(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/content-block-references/${id}`);
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
}

