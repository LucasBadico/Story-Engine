import type { StoryEngineClient } from '@story-engine/shared-ts';
import type { Story, World, RPGSystem } from '@story-engine/shared-ts';

export class ObsidianStoryEngineClient implements StoryEngineClient {
  constructor(
    private baseUrl: string,
    private apiKey: string,
    private tenantId: string
  ) {}

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey,
        'X-Tenant-ID': this.tenantId,
        ...options.headers,
      },
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({
        message: 'Request failed',
      }));
      throw new Error(error.message || `HTTP ${response.status}`);
    }

    const data = await response.json();
    return data.story || data.world || data.rpg_system || data.stories || data.worlds || data.rpg_systems || data;
  }

  // Stories
  async listStories(): Promise<Story[]> {
    return this.request<{ stories: Story[] }>('/api/v1/stories').then(
      (data) => (Array.isArray(data) ? data : data.stories || [])
    );
  }

  async getStory(id: string): Promise<Story> {
    return this.request<{ story: Story }>(`/api/v1/stories/${id}`).then(
      (data) => (data.story || data)
    );
  }

  async createStory(title: string, worldId?: string | null): Promise<Story> {
    return this.request<{ story: Story }>('/api/v1/stories', {
      method: 'POST',
      body: JSON.stringify({ title, world_id: worldId }),
    }).then((data) => data.story || data);
  }

  async updateStory(id: string, patch: Partial<Story>): Promise<Story> {
    return this.request<{ story: Story }>(`/api/v1/stories/${id}`, {
      method: 'PUT',
      body: JSON.stringify(patch),
    }).then((data) => data.story || data);
  }

  async deleteStory(id: string): Promise<void> {
    await this.request(`/api/v1/stories/${id}`, {
      method: 'DELETE',
    });
  }

  async cloneStory(id: string): Promise<Story> {
    return this.request<{ story: Story }>(`/api/v1/stories/${id}/clone`, {
      method: 'POST',
    }).then((data) => data.story || data);
  }

  // Worlds
  async listWorlds(): Promise<World[]> {
    return this.request<{ worlds: World[] }>('/api/v1/worlds').then(
      (data) => (Array.isArray(data) ? data : data.worlds || [])
    );
  }

  async getWorld(id: string): Promise<World> {
    return this.request<{ world: World }>(`/api/v1/worlds/${id}`).then(
      (data) => (data.world || data)
    );
  }

  async createWorld(data: {
    name: string;
    description: string;
    genre: string;
    isImplicit?: boolean;
  }): Promise<World> {
    return this.request<{ world: World }>('/api/v1/worlds', {
      method: 'POST',
      body: JSON.stringify({
        name: data.name,
        description: data.description,
        genre: data.genre,
        is_implicit: data.isImplicit || false,
      }),
    }).then((data) => data.world || data);
  }

  async updateWorld(id: string, patch: Partial<World>): Promise<World> {
    return this.request<{ world: World }>(`/api/v1/worlds/${id}`, {
      method: 'PUT',
      body: JSON.stringify(patch),
    }).then((data) => data.world || data);
  }

  async deleteWorld(id: string): Promise<void> {
    await this.request(`/api/v1/worlds/${id}`, {
      method: 'DELETE',
    });
  }

  // RPG Systems
  async listRPGSystems(): Promise<RPGSystem[]> {
    return this.request<{ rpg_systems: RPGSystem[] }>('/api/v1/rpg-systems').then(
      (data) => (Array.isArray(data) ? data : data.rpg_systems || [])
    );
  }

  async getRPGSystem(id: string): Promise<RPGSystem> {
    return this.request<{ rpg_system: RPGSystem }>(`/api/v1/rpg-systems/${id}`).then(
      (data) => (data.rpg_system || data)
    );
  }

  async createRPGSystem(data: {
    name: string;
    description?: string;
    baseStatsSchema: Record<string, any>;
    derivedStatsSchema?: Record<string, any>;
    progressionSchema?: Record<string, any>;
  }): Promise<RPGSystem> {
    return this.request<{ rpg_system: RPGSystem }>('/api/v1/rpg-systems', {
      method: 'POST',
      body: JSON.stringify({
        name: data.name,
        description: data.description,
        base_stats_schema: data.baseStatsSchema,
        derived_stats_schema: data.derivedStatsSchema,
        progression_schema: data.progressionSchema,
      }),
    }).then((data) => data.rpg_system || data);
  }

  async updateRPGSystem(
    id: string,
    patch: Partial<RPGSystem>
  ): Promise<RPGSystem> {
    return this.request<{ rpg_system: RPGSystem }>(`/api/v1/rpg-systems/${id}`, {
      method: 'PUT',
      body: JSON.stringify(patch),
    }).then((data) => data.rpg_system || data);
  }

  async deleteRPGSystem(id: string): Promise<void> {
    await this.request(`/api/v1/rpg-systems/${id}`, {
      method: 'DELETE',
    });
  }
}

