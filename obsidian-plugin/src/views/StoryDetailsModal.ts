import { Modal } from "obsidian";
import StoryEnginePlugin from "../main";
import { Story } from "../types";

export class StoryDetailsModal extends Modal {
	plugin: StoryEnginePlugin;
	story: Story;

	constructor(plugin: StoryEnginePlugin, story: Story) {
		super(plugin.app);
		this.plugin = plugin;
		this.story = story;
	}

	onOpen() {
		const { contentEl } = this;
		contentEl.empty();

		contentEl.createEl("h2", { text: this.story.title });

		const details = contentEl.createEl("div", { cls: "story-engine-details" });

		details.createEl("p", {
			text: `Status: ${this.story.status}`,
		});

		details.createEl("p", {
			text: `Version: ${this.story.version_number}`,
		});

		details.createEl("p", {
			text: `Created: ${new Date(this.story.created_at).toLocaleString()}`,
		});

		details.createEl("p", {
			text: `Updated: ${new Date(this.story.updated_at).toLocaleString()}`,
		});

		details.createEl("p", {
			text: `ID: ${this.story.id}`,
			cls: "story-engine-id",
		});

		const buttonContainer = contentEl.createEl("div", {
			cls: "story-engine-buttons",
		});

		const cloneButton = buttonContainer.createEl("button", {
			text: "Clone Story",
			cls: "mod-cta",
		});
		cloneButton.onclick = async () => {
			cloneButton.disabled = true;
			cloneButton.setText("Cloning...");

			try {
				const clonedStory = await this.plugin.apiClient.cloneStory(
					this.story.id
				);
				this.close();
				new StoryDetailsModal(this.plugin, clonedStory).open();
			} catch (err) {
				cloneButton.setText(
					err instanceof Error ? err.message : "Clone failed"
				);
				setTimeout(() => {
					cloneButton.disabled = false;
					cloneButton.setText("Clone Story");
				}, 3000);
			}
		};

		const copyIdButton = buttonContainer.createEl("button", {
			text: "Copy ID",
		});
		copyIdButton.onclick = () => {
			navigator.clipboard.writeText(this.story.id);
			copyIdButton.setText("Copied!");
			setTimeout(() => {
				copyIdButton.setText("Copy ID");
			}, 2000);
		};
	}

	onClose() {
		const { contentEl } = this;
		contentEl.empty();
	}
}

