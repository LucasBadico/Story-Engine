import { Plugin, Notice } from 'obsidian';
import { StoryEngineSettingTab } from './settings';
import { StoryEngineView, STORY_ENGINE_VIEW_TYPE } from './StoryEngineView';
import { DEFAULT_SETTINGS, StoryEngineSettings } from './types';

export default class StoryEnginePlugin extends Plugin {
  settings!: StoryEngineSettings;

  async onload() {
    await this.loadSettings();

    this.addSettingTab(new StoryEngineSettingTab(this.app, this));

    this.registerView(
      STORY_ENGINE_VIEW_TYPE,
      (leaf) => new StoryEngineView(leaf, this)
    );

    this.addRibbonIcon('book-open', 'Story Engine', () => {
      this.activateView();
    });

    this.addCommand({
      id: 'open-story-engine',
      name: 'Open Story Engine',
      callback: () => {
        this.activateView();
      },
    });

    this.addCommand({
      id: 'create-story',
      name: 'Create Story',
      callback: () => {
        this.createStoryCommand();
      },
    });
  }

  async onunload() {
    this.app.workspace.detachLeavesOfType(STORY_ENGINE_VIEW_TYPE);
  }

  async loadSettings() {
    this.settings = Object.assign({}, DEFAULT_SETTINGS, await this.loadData());
  }

  async saveSettings() {
    await this.saveData(this.settings);
  }

  async activateView() {
    const { workspace } = this.app;

    let leaf = workspace.getLeavesOfType(STORY_ENGINE_VIEW_TYPE)[0];

    if (!leaf) {
      const rightLeaf = workspace.getRightLeaf(false);
      if (!rightLeaf) {
        new Notice('Could not create view. Please try again.', 3000);
        return;
      }
      leaf = rightLeaf;
      await leaf.setViewState({
        type: STORY_ENGINE_VIEW_TYPE,
        active: true,
      });
    }

    workspace.revealLeaf(leaf);
  }

  async createStoryCommand() {
    const tenantId = this.settings.tenantId?.trim();
    if (!tenantId) {
      new Notice('Please configure Tenant ID in settings', 5000);
      return;
    }

    // TODO: Implement create story modal or inline form
    new Notice('Create story feature coming soon', 3000);
  }
}

