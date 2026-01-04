import type { Story, World, RPGSystem, ErrorResponse } from '../types';

/**
 * Interface for Story Engine API client
 * Implementations: HttpStoryEngineClient (web), ObsidianStoryEngineClient (obsidian)
 */
export interface StoryEngineClient {
  // Stories
  listStories(): Promise<Story[]>;
  getStory(id: string): Promise<Story>;
  createStory(title: string, worldId?: string | null): Promise<Story>;
  updateStory(id: string, patch: Partial<Story>): Promise<Story>;
  deleteStory(id: string): Promise<void>;
  cloneStory(id: string): Promise<Story>;

  // Worlds
  listWorlds(): Promise<World[]>;
  getWorld(id: string): Promise<World>;
  createWorld(data: {
    name: string;
    description: string;
    genre: string;
    isImplicit?: boolean;
  }): Promise<World>;
  updateWorld(id: string, patch: Partial<World>): Promise<World>;
  deleteWorld(id: string): Promise<void>;

  // RPG Systems
  listRPGSystems(): Promise<RPGSystem[]>;
  getRPGSystem(id: string): Promise<RPGSystem>;
  createRPGSystem(data: {
    name: string;
    description?: string;
    baseStatsSchema: Record<string, any>;
    derivedStatsSchema?: Record<string, any>;
    progressionSchema?: Record<string, any>;
  }): Promise<RPGSystem>;
  updateRPGSystem(id: string, patch: Partial<RPGSystem>): Promise<RPGSystem>;
  deleteRPGSystem(id: string): Promise<void>;
}

