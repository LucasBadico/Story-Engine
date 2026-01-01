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
				if (!this.plugin.settings.tenantId) {
					throw new Error("Tenant ID not configured");
				}

				const clonedStory = await this.plugin.apiClient.cloneStory(
					this.story.id,
					this.plugin.settings.tenantId
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
			// Use Obsidian's copy text functionality
			const textarea = contentEl.createEl("textarea");
			textarea.value = this.story.id;
			textarea.style.position = "fixed";
			textarea.style.opacity = "0";
			textarea.style.left = "-9999px";
			textarea.select();
			try {
				// Use execCommand for clipboard copy (works in Obsidian's Electron environment)
				const doc = (contentEl as any).ownerDocument || (globalThis as any).document;
				if (doc && doc.execCommand) {
					doc.execCommand("copy");
					copyIdButton.setText("Copied!");
					setTimeout(() => {
						copyIdButton.setText("Copy ID");
					}, 2000);
				}
			} catch (err) {
				console.error("Failed to copy ID:", err);
			}
			textarea.remove();
		};
	}

	onClose() {
		const { contentEl } = this;
		contentEl.empty();
	}
}

