import { Notice, Plugin } from "obsidian";
import { StoryEngineSettings } from "./types";
import { StoryEngineClient } from "./api/client";
import { StoryEngineSettingTab } from "./settings";
import { registerCommands } from "./commands";
import { StoryDetailsModal } from "./views/StoryDetailsModal";
import { PromptModal } from "./views/PromptModal";
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
	settings: StoryEngineSettings;
	apiClient: StoryEngineClient;
	fileManager: FileManager;
	syncService: SyncService;

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
		const title = await PromptModal.prompt(
			this.app,
			"Enter story title:",
			"My New Story"
		);

		if (!title) {
			return;
		}

		// Validate and trim tenant ID
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

		try {
			const story = await this.apiClient.createStory(
				tenantId,
				title.trim()
			);
			new StoryDetailsModal(this, story).open();
			// Show notification using notice API
			new Notice(`Story "${title}" created successfully`);
		} catch (err) {
			const errorMessage = err instanceof Error ? err.message : "Failed to create story";
			new Notice(`Error: ${errorMessage}`, 5000);
		}
	}
}

