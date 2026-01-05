/**
 * Shared TypeScript exports
 * Barrel file for all shared types and interfaces
 */

// Types
export * from './types';

// Ports (interfaces)
export * from './ports/storyEngineClient';
export * from './ports/storage';
export * from './ports/notifier';
export * from './ports/navigator';

// Services type
import type { StoryEngineClient } from './ports/storyEngineClient';
import type { Storage } from './ports/storage';
import type { Notifier } from './ports/notifier';
import type { Navigator } from './ports/navigator';

export interface Services {
  storyEngine: StoryEngineClient;
  storage: Storage;
  notify: Notifier;
  navigator: Navigator;
}

