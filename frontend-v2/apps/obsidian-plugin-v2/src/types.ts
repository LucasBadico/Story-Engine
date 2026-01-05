export interface StoryEngineSettings {
  apiUrl: string;
  apiKey: string;
  tenantId: string;
  tenantName: string;
  syncFolderPath?: string;
}

export const DEFAULT_SETTINGS: StoryEngineSettings = {
  apiUrl: 'http://localhost:8080',
  apiKey: '',
  tenantId: '',
  tenantName: '',
  syncFolderPath: 'Stories',
};

