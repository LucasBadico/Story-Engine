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
} from "../types";

export class StoryEngineClient {
	constructor(
		private apiUrl: string,
		private apiKey: string,
		private tenantId: string = ""
	) {}

	setTenantId(tenantId: string): void {
		this.tenantId = tenantId.trim();
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
		if (!this.tenantId || !this.tenantId.trim()) {
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

	async createStory(title: string): Promise<Story> {
		if (!this.tenantId || !this.tenantId.trim()) {
			throw new Error("Tenant ID is required");
		}

		const response = await this.request<{ story: Story }>(
			"POST",
			"/api/v1/stories",
			{
				title: title.trim(),
			}
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

	async createProseBlockReference(proseBlockId: string, entityType: string, entityId: string): Promise<ProseBlockReference> {
		const response = await this.request<{ reference: ProseBlockReference }>(
			"POST",
			`/api/v1/prose-blocks/${proseBlockId}/references`,
			{
				entity_type: entityType,
				entity_id: entityId,
			}
		);
		return response.reference;
	}

	async deleteProseBlockReference(id: string): Promise<void> {
		await this.request("DELETE", `/api/v1/prose-block-references/${id}`);
	}
}

