import { Modal } from "obsidian";
import StoryEnginePlugin from "../main";
import { Story } from "../types";
import { StoryDetailsModal } from "./StoryDetailsModal";

export class StoryListModal extends Modal {
	plugin: StoryEnginePlugin;
	stories: Story[] = [];
	loading: boolean = true;
	error: string | null = null;

	constructor(plugin: StoryEnginePlugin) {
		super(plugin.app);
		this.plugin = plugin;
	}

	async onOpen() {
		const { contentEl } = this;
		contentEl.empty();

		contentEl.createEl("h2", { text: "Stories" });

		await this.loadStories();

		if (this.loading) {
			contentEl.createEl("p", { text: "Loading stories..." });
			return;
		}

		if (this.error) {
			contentEl.createEl("p", {
				text: `Error: ${this.error}`,
				cls: "story-engine-error",
			});
			return;
		}

		if (this.stories.length === 0) {
			contentEl.createEl("p", { text: "No stories found." });
			const createButton = contentEl.createEl("button", {
				text: "Create Story",
			});
			createButton.onclick = () => {
				this.close();
				this.plugin.createStoryCommand();
			};
			return;
		}

		const storiesList = contentEl.createEl("div", { cls: "story-engine-list" });

		for (const story of this.stories) {
			const storyItem = storiesList.createEl("div", {
				cls: "story-engine-item",
			});

			const title = storyItem.createEl("div", {
				cls: "story-engine-title",
				text: story.title,
			});

			const meta = storyItem.createEl("div", {
				cls: "story-engine-meta",
			});
			meta.createEl("span", {
				text: `Version ${story.version_number}`,
			});
			meta.createEl("span", {
				text: `Status: ${story.status}`,
			});

			storyItem.onclick = () => {
				this.close();
				new StoryDetailsModal(this.plugin, story).open();
			};
		}

		const createButton = contentEl.createEl("button", {
			text: "Create New Story",
			cls: "mod-cta",
		});
		createButton.onclick = () => {
			this.close();
			this.plugin.createStoryCommand();
		};
	}

	async loadStories() {
		this.loading = true;
		this.error = null;

		try {
			if (!this.plugin.settings.tenantId) {
				this.error = "Tenant ID not configured";
				this.loading = false;
				return;
			}

			this.stories = await this.plugin.apiClient.listStories(
				this.plugin.settings.tenantId
			);
		} catch (err) {
			this.error = err instanceof Error ? err.message : "Unknown error";
		} finally {
			this.loading = false;
		}
	}

	onClose() {
		const { contentEl } = this;
		contentEl.empty();
	}
}

