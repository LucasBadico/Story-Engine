/**
 * Shared TypeScript types for Story Engine
 */

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

export interface ErrorResponse {
  error: string;
  message: string;
  code: string;
}

