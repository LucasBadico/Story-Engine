import { App, PluginSettingTab, Setting } from "obsidian";
import StoryEnginePlugin from "./main";

export class StoryEngineSettingTab extends PluginSettingTab {
	plugin: StoryEnginePlugin;

	constructor(app: App, plugin: StoryEnginePlugin) {
		super(app, plugin);
		this.plugin = plugin;
	}

	display(): void {
		const { containerEl } = this;

		containerEl.empty();

		containerEl.createEl("h2", { text: "Story Engine Settings" });

		new Setting(containerEl)
			.setName("API URL")
			.setDesc("The base URL of the Story Engine API")
			.addText((text) =>
				text
					.setPlaceholder("http://localhost:8080")
					.setValue(this.plugin.settings.apiUrl)
					.onChange(async (value) => {
						this.plugin.settings.apiUrl = value;
						await this.plugin.saveSettings();
					})
			);

		new Setting(containerEl)
			.setName("API Key")
			.setDesc("API key for authentication (optional for MVP)")
			.addText((text) => {
				text.setPlaceholder("Enter API key")
					.setValue(this.plugin.settings.apiKey)
					.inputEl.type = "password";
				text.onChange(async (value) => {
					this.plugin.settings.apiKey = value;
					await this.plugin.saveSettings();
				});
			});

		new Setting(containerEl)
			.setName("Tenant ID")
			.setDesc("Your workspace tenant ID (UUID format)")
			.addText((text) =>
				text
					.setPlaceholder("00000000-0000-0000-0000-000000000000")
					.setValue(this.plugin.settings.tenantId || "")
					.onChange(async (value) => {
						this.plugin.settings.tenantId = value.trim();
						// Update API client tenant ID immediately
						if (this.plugin.apiClient) {
							this.plugin.apiClient.setTenantId(value.trim());
						}
						await this.plugin.saveSettings();
					})
			);

		new Setting(containerEl)
			.setName("Sync Folder Path")
			.setDesc("Folder path where synced stories will be stored")
			.addText((text) =>
				text
					.setPlaceholder("Stories")
					.setValue(this.plugin.settings.syncFolderPath || "Stories")
					.onChange(async (value) => {
						this.plugin.settings.syncFolderPath = value.trim() || "Stories";
						await this.plugin.saveSettings();
					})
			);

		new Setting(containerEl)
			.setName("Auto Version Snapshots")
			.setDesc("Automatically create version snapshots when syncing")
			.addToggle((toggle) =>
				toggle
					.setValue(this.plugin.settings.autoVersionSnapshots ?? true)
					.onChange(async (value) => {
						this.plugin.settings.autoVersionSnapshots = value;
						await this.plugin.saveSettings();
					})
			);

		new Setting(containerEl)
			.setName("Conflict Resolution")
			.setDesc("How to resolve conflicts when both local and service have changes")
			.addDropdown((dropdown) =>
				dropdown
					.addOption("service", "Service Wins")
					.addOption("local", "Local Wins")
					.addOption("manual", "Manual (Newer Wins)")
					.setValue(this.plugin.settings.conflictResolution || "service")
					.onChange(async (value) => {
						this.plugin.settings.conflictResolution =
							value as "service" | "local" | "manual";
						await this.plugin.saveSettings();
					})
			);

		new Setting(containerEl)
			.setName("Unsplash Access Key")
			.setDesc("Application ID / Access Key for Unsplash API (get one at https://unsplash.com/developers)")
			.addText((text) => {
				text.setPlaceholder("Enter Unsplash access key")
					.setValue(this.plugin.settings.unsplashAccessKey || "")
					.inputEl.type = "password";
				text.onChange(async (value) => {
					this.plugin.settings.unsplashAccessKey = value.trim();
					await this.plugin.saveSettings();
				});
			});

		new Setting(containerEl)
			.setName("Unsplash Secret Key")
			.setDesc("Secret Key for Unsplash API (optional, needed for some operations)")
			.addText((text) => {
				text.setPlaceholder("Enter Unsplash secret key")
					.setValue(this.plugin.settings.unsplashSecretKey || "")
					.inputEl.type = "password";
				text.onChange(async (value) => {
					this.plugin.settings.unsplashSecretKey = value.trim();
					await this.plugin.saveSettings();
				});
			});

		// Test connection button
		new Setting(containerEl)
			.setName("Test Connection")
			.setDesc("Test connection to the Story Engine API")
			.addButton((button) =>
				button
					.setButtonText("Test")
					.onClick(async () => {
						button.setButtonText("Testing...");
						button.setDisabled(true);
						try {
							const result = await this.plugin.apiClient.testConnection();
							if (result) {
								button.setButtonText("Success!");
								setTimeout(() => {
									button.setButtonText("Test");
									button.setDisabled(false);
								}, 2000);
							} else {
								button.setButtonText("Failed");
								setTimeout(() => {
									button.setButtonText("Test");
									button.setDisabled(false);
								}, 2000);
							}
						} catch (err) {
							button.setButtonText("Error");
							setTimeout(() => {
								button.setButtonText("Test");
								button.setDisabled(false);
							}, 2000);
						}
					})
			);
	}
}



