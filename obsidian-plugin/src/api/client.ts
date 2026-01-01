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
} from "../types";

export class StoryEngineClient {
	constructor(
		private apiUrl: string,
		private apiKey: string
	) {}

	private async request<T>(
		method: string,
		endpoint: string,
		body?: unknown,
		tenantId?: string
	): Promise<T> {
		const url = `${this.apiUrl}${endpoint}`;
		const headers = new Headers();
		
		headers.set("Content-Type", "application/json");

		if (this.apiKey) {
			headers.set("Authorization", `Bearer ${this.apiKey}`);
		}

		if (tenantId) {
			const trimmedTenantId = tenantId.trim();
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

	async listStories(tenantId: string): Promise<Story[]> {
		const trimmedTenantId = tenantId.trim();
		if (!trimmedTenantId) {
			throw new Error("Tenant ID is required");
		}

		const response = await this.request<{ stories: Story[] }>(
			"GET",
			"/api/v1/stories",
			undefined,
			trimmedTenantId
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

	async createStory(tenantId: string, title: string): Promise<Story> {
		const trimmedTenantId = tenantId.trim();
		if (!trimmedTenantId) {
			throw new Error("Tenant ID is required");
		}

		const response = await this.request<{ story: Story }>(
			"POST",
			"/api/v1/stories",
			{
				title: title.trim(),
			},
			trimmedTenantId
		);
		return response.story;
	}

	async cloneStory(id: string, tenantId: string): Promise<Story> {
		const trimmedTenantId = tenantId.trim();
		if (!trimmedTenantId) {
			throw new Error("Tenant ID is required");
		}

		const response = await this.request<{ story: Story }>(
			"POST",
			`/api/v1/stories/${id}/clone`,
			{},
			trimmedTenantId
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
}

