import { App, PluginSettingTab, Setting } from 'obsidian';
import type { StoryEnginePlugin } from './plugin';

export class StoryEngineSettingTab extends PluginSettingTab {
  plugin: StoryEnginePlugin;

  constructor(app: App, plugin: StoryEnginePlugin) {
    super(app, plugin);
    this.plugin = plugin;
  }

  display(): void {
    const { containerEl } = this;

    containerEl.empty();

    containerEl.createEl('h2', { text: 'Story Engine Settings' });

    new Setting(containerEl)
      .setName('API URL')
      .setDesc('Base URL for the Story Engine API')
      .addText((text) =>
        text
          .setPlaceholder('http://localhost:8080')
          .setValue(this.plugin.settings.apiUrl)
          .onChange(async (value) => {
            this.plugin.settings.apiUrl = value;
            await this.plugin.saveSettings();
          })
      );

    new Setting(containerEl)
      .setName('API Key')
      .setDesc('API key for authentication')
      .addText((text) => {
        text
          .setPlaceholder('Enter API key')
          .setValue(this.plugin.settings.apiKey);
        text.inputEl.type = 'password';
        text.onChange(async (value) => {
          this.plugin.settings.apiKey = value;
          await this.plugin.saveSettings();
        });
      });

    new Setting(containerEl)
      .setName('Tenant ID')
      .setDesc('Tenant ID for multi-tenancy')
      .addText((text) =>
        text
          .setPlaceholder('Enter tenant ID')
          .setValue(this.plugin.settings.tenantId)
          .onChange(async (value) => {
            this.plugin.settings.tenantId = value;
            await this.plugin.saveSettings();
          })
      );

    new Setting(containerEl)
      .setName('Tenant Name')
      .setDesc('Display name for the tenant')
      .addText((text) =>
        text
          .setPlaceholder('Enter tenant name')
          .setValue(this.plugin.settings.tenantName)
          .onChange(async (value) => {
            this.plugin.settings.tenantName = value;
            await this.plugin.saveSettings();
          })
      );

    new Setting(containerEl)
      .setName('Sync Folder Path')
      .setDesc('Folder path for syncing stories (optional)')
      .addText((text) =>
        text
          .setPlaceholder('Stories')
          .setValue(this.plugin.settings.syncFolderPath || 'Stories')
          .onChange(async (value) => {
            this.plugin.settings.syncFolderPath = value;
            await this.plugin.saveSettings();
          })
      );
  }
}

