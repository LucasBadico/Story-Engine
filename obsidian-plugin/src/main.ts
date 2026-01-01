import { Notice, Plugin } from "obsidian";
import { StoryEngineSettings } from "./types";
import { StoryEngineClient } from "./api/client";
import { StoryEngineSettingTab } from "./settings";
import { registerCommands } from "./commands";
import { StoryDetailsModal } from "./views/StoryDetailsModal";
import { CreateStoryModal } from "./views/CreateStoryModal";
import { FileManager } from "./sync/fileManager";
import { SyncService } from "./sync/syncService";

const DEFAULT_SETTINGS: StoryEngineSettings = {
	apiUrl: "http://localhost:8080",
	apiKey: "",
	tenantId: "",
	tenantName: "",
	syncFolderPath: "Stories",
	autoVersionSnapshots: true,
	conflictResolution: "service",
};

export default class StoryEnginePlugin extends Plugin {
	settings!: StoryEngineSettings;
	apiClient!: StoryEngineClient;
	fileManager!: FileManager;
	syncService!: SyncService;

	async onload() {
		await this.loadSettings();

		this.apiClient = new StoryEngineClient(
			this.settings.apiUrl,
			this.settings.apiKey
		);

		this.fileManager = new FileManager(
			this.app.vault,
			this.settings.syncFolderPath || "Stories"
		);

		this.syncService = new SyncService(
			this.apiClient,
			this.fileManager,
			this.settings
		);

		this.addSettingTab(new StoryEngineSettingTab(this.app, this));

		registerCommands(this);
	}

	async onunload() {}

	async loadSettings() {
		this.settings = Object.assign(
			{},
			DEFAULT_SETTINGS,
			await this.loadData()
		);
	}

	async saveSettings() {
		await this.saveData(this.settings);
		// Update API client when settings change
		this.apiClient = new StoryEngineClient(
			this.settings.apiUrl,
			this.settings.apiKey
		);
		// Update file manager base path
		this.fileManager = new FileManager(
			this.app.vault,
			this.settings.syncFolderPath || "Stories"
		);
		// Update sync service
		this.syncService = new SyncService(
			this.apiClient,
			this.fileManager,
			this.settings
		);
	}

	async createStoryCommand() {
		// Validate and trim tenant ID first
		const tenantId = this.settings.tenantId?.trim();
		if (!tenantId) {
			new Notice("Please configure Tenant ID in settings", 5000);
			return;
		}

		// Basic UUID format validation
		const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
		if (!uuidRegex.test(tenantId)) {
			new Notice("Invalid Tenant ID format. Please check your settings.", 5000);
			return;
		}

		// Open modal to get title and sync preference
		new CreateStoryModal(this.app, async (title: string, shouldSync: boolean) => {
			try {
				new Notice(`Creating story "${title}"...`);
				
				const story = await this.apiClient.createStory(tenantId, title);
				
				new Notice(`Story "${title}" created successfully`);

				// Sync to Obsidian if requested
				if (shouldSync) {
					try {
						new Notice(`Syncing story to Obsidian...`);
						await this.syncService.pullStory(story.id);
						new Notice(`Story synced to your vault!`);
					} catch (syncErr) {
						const syncErrorMessage = syncErr instanceof Error 
							? syncErr.message 
							: "Failed to sync story";
						new Notice(`Story created but sync failed: ${syncErrorMessage}`, 5000);
					}
				} else {
					// Show details modal if not syncing
					new StoryDetailsModal(this, story).open();
				}
			} catch (err) {
				const errorMessage = err instanceof Error ? err.message : "Failed to create story";
				new Notice(`Error: ${errorMessage}`, 5000);
			}
		}).open();
	}
}

