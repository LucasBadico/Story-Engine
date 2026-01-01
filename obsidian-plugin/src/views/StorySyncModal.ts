import { Modal, Notice } from "obsidian";
import StoryEnginePlugin from "../main";
import { Story } from "../types";

export class StorySyncModal extends Modal {
	plugin: StoryEnginePlugin;
	mode: "pull" | "push";
	stories: Story[] = [];
	loading: boolean = true;
	error: string | null = null;

	constructor(plugin: StoryEnginePlugin, mode: "pull" | "push") {
		super(plugin.app);
		this.plugin = plugin;
		this.mode = mode;
	}

	async onOpen() {
		const { contentEl } = this;
		contentEl.empty();

		const title =
			this.mode === "pull"
				? "Sync Story from Service"
				: "Push Story to Service";
		contentEl.createEl("h2", { text: title });

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

			storyItem.onclick = async () => {
				this.close();
				try {
					if (this.mode === "pull") {
						await this.plugin.syncService.pullStory(story.id);
					} else {
						const folderPath = this.plugin.fileManager.getStoryFolderPath(
							story.title
						);
						await this.plugin.syncService.pushStory(folderPath);
					}
				} catch (err) {
					const errorMessage =
						err instanceof Error ? err.message : "Failed to sync story";
					new Notice(`Error: ${errorMessage}`, 5000);
				}
			};
		}
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


