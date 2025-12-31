import { Notice, Plugin } from "obsidian";
import { StoryEngineSettings } from "./types";
import { StoryEngineClient } from "./api/client";
import { StoryEngineSettingTab } from "./settings";
import { registerCommands } from "./commands";
import { StoryDetailsModal } from "./views/StoryDetailsModal";

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
		const title = await this.app.prompt("Enter story title:", {
			placeholder: "My New Story",
		});

		if (!title) {
			return;
		}

		if (!this.settings.tenantId) {
			new Notice("Please configure Tenant ID in settings", 5000);
			return;
		}

		try {
			const story = await this.apiClient.createStory(
				this.settings.tenantId,
				title
			);
			new StoryDetailsModal(this, story).open();
			// Show notification using notice API
			new Notice(`Story "${title}" created successfully`);
		} catch (err) {
			new Notice(
				err instanceof Error ? err.message : "Failed to create story",
				5000
			);
		}
	}
}

