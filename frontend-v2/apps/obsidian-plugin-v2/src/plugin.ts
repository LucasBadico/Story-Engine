import { Plugin } from 'obsidian';
import type { StoryEngineSettings } from './types';

export interface StoryEnginePlugin extends Plugin {
  settings: StoryEngineSettings;
  loadSettings(): Promise<void>;
  saveSettings(): Promise<void>;
  activateView(): Promise<void>;
  createStoryCommand(): Promise<void>;
}

