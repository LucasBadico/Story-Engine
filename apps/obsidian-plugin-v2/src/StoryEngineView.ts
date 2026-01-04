import { ItemView, WorkspaceLeaf } from 'obsidian';
import { createRoot, Root } from 'react-dom/client';
import React from 'react';
import { ServicesProvider, StoryList } from '@story-engine/ui-package';
import type { Services } from '@story-engine/shared-ts';
import type { StoryEnginePlugin } from './plugin';
import { ObsidianStoryEngineClient } from './adapters/obsidianStoryEngineClient';
import { ObsidianStorage } from './adapters/obsidianStorage';
import { ObsidianNotifier } from './adapters/obsidianNotifier';
import { ObsidianNavigator } from './adapters/obsidianNavigator';

export const STORY_ENGINE_VIEW_TYPE = 'story-engine-view-v2';

export class StoryEngineView extends ItemView {
  private reactRoot: Root | null = null;
  private plugin: StoryEnginePlugin;

  constructor(leaf: WorkspaceLeaf, plugin: StoryEnginePlugin) {
    super(leaf);
    this.plugin = plugin;
  }

  getViewType(): string {
    return STORY_ENGINE_VIEW_TYPE;
  }

  getDisplayText(): string {
    return 'Story Engine';
  }

  getIcon(): string {
    return 'book-open';
  }

  async onOpen() {
    const rootEl = this.containerEl.createDiv('se-root se-obsidian');
    
    // Tokens CSS will be imported via styles.css

    this.reactRoot = createRoot(rootEl);

    // Create adapters
    const services: Services = {
      storyEngine: new ObsidianStoryEngineClient(
        this.plugin.settings.apiUrl,
        this.plugin.settings.apiKey,
        this.plugin.settings.tenantId
      ),
      storage: new ObsidianStorage(this.plugin),
      notify: new ObsidianNotifier(),
      navigator: new ObsidianNavigator(this.app),
    };

    this.reactRoot.render(
      React.createElement(
        ServicesProvider,
        { services },
        React.createElement(StoryList, {
          onCreateStory: () => {
            this.plugin.createStoryCommand();
          },
        })
      )
    );
  }

  async onClose() {
    if (this.reactRoot) {
      this.reactRoot.unmount();
      this.reactRoot = null;
    }
  }
}

