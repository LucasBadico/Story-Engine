import { App, PluginSettingTab, Setting } from "obsidian";
import StoryEnginePlugin from "./main";
import { StoryEngineSettings } from "./types";

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
			.setDesc("Your workspace tenant ID")
			.addText((text) =>
				text
					.setPlaceholder("Enter tenant ID")
					.setValue(this.plugin.settings.tenantId)
					.onChange(async (value) => {
						this.plugin.settings.tenantId = value;
						await this.plugin.saveSettings();
					})
			);

		new Setting(containerEl)
			.setName("Test Connection")
			.setDesc("Test the connection to the Story Engine API")
			.addButton((button) =>
				button.setButtonText("Test").onClick(async () => {
					button.setButtonText("Testing...");
					button.setDisabled(true);

					const success = await this.plugin.apiClient.testConnection();

					if (success) {
						button.setButtonText("✓ Connected");
						button.buttonEl.style.color = "green";
					} else {
						button.setButtonText("✗ Failed");
						button.buttonEl.style.color = "red";
					}

					setTimeout(() => {
						button.setButtonText("Test");
						button.setDisabled(false);
						button.buttonEl.style.color = "";
					}, 3000);
				})
			);
	}
}

