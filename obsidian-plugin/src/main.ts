import { Notice, Plugin } from "obsidian";
import { StoryEngineSettings } from "./types";
import { StoryEngineClient } from "./api/client";
import { StoryEngineSettingTab } from "./settings";
import { registerCommands } from "./commands";
import { StoryDetailsModal } from "./views/StoryDetailsModal";
import { PromptModal } from "./views/PromptModal";

const DEFAULT_SETTINGS: StoryEngineSettings = {
	apiUrl: "http://localhost:8080",
	apiKey: "",
	tenantId: "",
	tenantName: "",
};

export default class StoryEnginePlugin extends Plugin {
	settings: StoryEngineSettings;
	apiClient: StoryEngineClient;

	async onload() {
		await this.loadSettings();

		this.apiClient = new StoryEngineClient(
			this.settings.apiUrl,
			this.settings.apiKey
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

