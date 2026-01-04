import type { Storage } from '@story-engine/shared-ts';
import type { StoryEnginePlugin } from '../plugin';

export class ObsidianStorage implements Storage {
  constructor(private plugin: StoryEnginePlugin) {}

  async get<T>(key: string): Promise<T | null> {
    const data = this.plugin.settings as any;
    return data[key] !== undefined ? (data[key] as T) : null;
  }

  async set<T>(key: string, value: T): Promise<void> {
    (this.plugin.settings as any)[key] = value;
    await this.plugin.saveSettings();
  }

  async remove(key: string): Promise<void> {
    delete (this.plugin.settings as any)[key];
    await this.plugin.saveSettings();
  }

  async clear(): Promise<void> {
    // Clear only custom keys, not core settings
    const coreKeys = ['apiUrl', 'apiKey', 'tenantId', 'tenantName', 'syncFolderPath'];
    for (const key in this.plugin.settings) {
      if (!coreKeys.includes(key)) {
        delete (this.plugin.settings as any)[key];
      }
    }
    await this.plugin.saveSettings();
  }

  async getAllKeys(): Promise<string[]> {
    return Object.keys(this.plugin.settings);
  }
}

