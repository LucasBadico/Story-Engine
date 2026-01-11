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

		// Mode Selection
		let tenantIdSetting: Setting | null = null;
		let videoLinkSetting: Setting | null = null;
		let conflictResolutionSetting: Setting | null = null;

		const modeSetting = new Setting(containerEl)
			.setName("Connection Mode")
			.setDesc("Choose between local (offline) or remote (cloud) mode")
			.addDropdown((dropdown) =>
				dropdown
					.addOption("local", "Local")
					.addOption("remote", "Remote")
					.setValue(this.plugin.settings.mode || "local")
					.onChange(async (value) => {
						this.plugin.settings.mode = value as "local" | "remote";
						// Auto-set conflict resolution in local mode
						if (value === "local") {
							this.plugin.settings.conflictResolution = "local";
							if (conflictResolutionSetting) {
								conflictResolutionSetting.settingEl.style.display = "none";
							}
						} else {
							if (conflictResolutionSetting) {
								conflictResolutionSetting.settingEl.style.display = "";
							}
						}
						// Show/hide tenant ID
						if (tenantIdSetting) {
							tenantIdSetting.settingEl.style.display = value === "remote" ? "" : "none";
						}
						// Show/hide video link
						if (videoLinkSetting) {
							videoLinkSetting.settingEl.style.display = value === "local" ? "" : "none";
						}
						// Update API client mode
						if (this.plugin.apiClient) {
							this.plugin.apiClient.setMode(value as "local" | "remote");
						}
						await this.plugin.saveSettings();
					})
			);

		new Setting(containerEl)
			.setName("Sync Version")
			.setDesc(
				"Toggle between the legacy sync engine (v1) and the new modular pipeline (v2 - experimental)"
			)
			.addDropdown((dropdown) =>
				dropdown
					.addOption("v1", "Legacy (v1)")
					.addOption("v2", "Modular (v2, experimental)")
					.setValue(this.plugin.settings.syncVersion || "v1")
					.onChange(async (value) => {
						this.plugin.settings.syncVersion = value as "v1" | "v2";
						await this.plugin.saveSettings();
					})
			);

		// API URL
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
			.setName("LLM Gateway URL")
			.setDesc("The base URL of the LLM Gateway service")
			.addText((text) =>
				text
					.setPlaceholder("http://localhost:8081")
					.setValue(this.plugin.settings.llmGatewayUrl)
					.onChange(async (value) => {
						this.plugin.settings.llmGatewayUrl = value.trim();
						await this.plugin.saveSettings();
					})
			);

		// Tenant ID (conditional - only shown in remote mode)
		tenantIdSetting = new Setting(containerEl)
			.setName("Tenant ID")
			.setDesc("Your workspace tenant ID (UUID format) - Required in remote mode")
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
		if (this.plugin.settings.mode === "local") {
			tenantIdSetting.settingEl.style.display = "none";
		}

		// Video Link Section (conditional - only shown in local mode)
		const videoUrl = this.plugin.settings.localModeVideoUrl || "https://example.com/setup-video";
		videoLinkSetting = new Setting(containerEl)
			.setName("Setup Guide")
			.setDesc(`📹 Learn how to setup local mode: ${videoUrl}`)
			.addButton((button) => {
				button.setButtonText("Open Video").onClick(() => {
					window.open(videoUrl, "_blank");
				});
			});
		if (this.plugin.settings.mode === "remote") {
			videoLinkSetting.settingEl.style.display = "none";
		}

		// API Key
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
			.setName("Auto Sync on API Updates")
			.setDesc("Apply API update payloads directly to local files when available")
			.addToggle((toggle) =>
				toggle
					.setValue(this.plugin.settings.autoSyncOnApiUpdates ?? true)
					.onChange(async (value) => {
						this.plugin.settings.autoSyncOnApiUpdates = value;
						await this.plugin.saveSettings();
					})
			);

		new Setting(containerEl)
			.setName("Auto Push on Blur/Idle")
			.setDesc("When you leave a chapter/scene/beat or stay idle for 1 min, push changes upstream automatically")
			.addToggle((toggle) =>
				toggle
					.setValue(this.plugin.settings.autoPushOnFileBlur ?? true)
					.onChange(async (value) => {
						this.plugin.settings.autoPushOnFileBlur = value;
						await this.plugin.saveSettings();
					})
			);

		// Conflict Resolution (conditional - hidden in local mode)
		conflictResolutionSetting = new Setting(containerEl)
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
		if (this.plugin.settings.mode === "local") {
			conflictResolutionSetting.settingEl.style.display = "none";
		}

		// Show Help Box in MD Files
		new Setting(containerEl)
			.setName("Show Help Box in MD Files")
			.setDesc("Enable/disable info/help boxes in markdown files")
			.addToggle((toggle) =>
				toggle
					.setValue(this.plugin.settings.showHelpBox ?? true)
					.onChange(async (value) => {
						this.plugin.settings.showHelpBox = value;
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


