import { ItemView, WorkspaceLeaf, Notice } from "obsidian";
import StoryEnginePlugin from "../main";
import { Story } from "../types";
import { StoryDetailsModal } from "./StoryDetailsModal";

export const STORY_LIST_VIEW_TYPE = "story-engine-list-view";

export class StoryListView extends ItemView {
	plugin: StoryEnginePlugin;
	stories: Story[] = [];
	loading: boolean = true;
	error: string | null = null;
	contentEl!: HTMLElement;

	constructor(leaf: WorkspaceLeaf, plugin: StoryEnginePlugin) {
		super(leaf);
		this.plugin = plugin;
	}

	getViewType(): string {
		return STORY_LIST_VIEW_TYPE;
	}

	getDisplayText(): string {
		return "Stories";
	}

	getIcon(): string {
		return "book-open";
	}

	async onOpen() {
		const container = this.containerEl.children[1];
		container.empty();
		container.addClass("story-engine-view-container");
		
		await this.render(container);
		await this.loadStories();
	}

	async onClose() {
		// Cleanup if needed
	}

	async render(container: HTMLElement) {
		container.empty();

		// Header with title and create button
		const header = container.createDiv({ cls: "story-engine-view-header" });
		header.createEl("h2", { text: "Stories" });

		const headerActions = header.createDiv({ cls: "story-engine-header-actions" });
		
		const refreshButton = headerActions.createEl("button", {
			text: "Refresh",
			cls: "story-engine-refresh-btn",
		});
		refreshButton.onclick = async () => {
			await this.loadStories();
		};

		const syncAllButton = headerActions.createEl("button", {
			text: "Sync All",
			cls: "story-engine-sync-all-btn",
		});
		syncAllButton.onclick = async () => {
			if (!this.plugin.settings.tenantId) {
				new Notice("Please configure Tenant ID in settings", 5000);
				return;
			}

			try {
				new Notice("Syncing all stories...");
				await this.plugin.syncService.pullAllStories();
				await this.loadStories(); // Refresh the list after sync
			} catch (err) {
				const errorMessage =
					err instanceof Error ? err.message : "Failed to sync stories";
				new Notice(`Error: ${errorMessage}`, 5000);
			}
		};

		const createButton = headerActions.createEl("button", {
			text: "Create Story",
			cls: "mod-cta story-engine-create-btn",
		});
		createButton.onclick = () => {
			this.plugin.createStoryCommand();
		};

		// Content area for stories list
		this.contentEl = container.createDiv({ cls: "story-engine-view-content" });
	}

	renderStories() {
		if (!this.contentEl) return;

		this.contentEl.empty();

		if (this.loading) {
			this.contentEl.createEl("p", { text: "Loading stories..." });
			return;
		}

		if (this.error) {
			this.contentEl.createEl("p", {
				text: `Error: ${this.error}`,
				cls: "story-engine-error",
			});
			return;
		}

		if (this.stories.length === 0) {
			this.contentEl.createEl("p", { text: "No stories found." });
			return;
		}

		const storiesList = this.contentEl.createDiv({ cls: "story-engine-list" });

		for (const story of this.stories) {
			const storyItem = storiesList.createDiv({
				cls: "story-engine-item",
			});

			const title = storyItem.createDiv({
				cls: "story-engine-title",
				text: story.title,
			});

			const meta = storyItem.createDiv({
				cls: "story-engine-meta",
			});
			meta.createEl("span", {
				text: `Version ${story.version_number}`,
			});
			meta.createEl("span", {
				text: `Status: ${story.status}`,
			});

			storyItem.onclick = () => {
				new StoryDetailsModal(this.plugin, story).open();
			};
		}
	}

	async loadStories() {
		this.loading = true;
		this.error = null;
		this.renderStories();

		try {
			if (!this.plugin.settings.tenantId) {
				this.error = "Tenant ID not configured";
				this.loading = false;
				this.renderStories();
				return;
			}

			this.stories = await this.plugin.apiClient.listStories(
				this.plugin.settings.tenantId
			);
		} catch (err) {
			this.error = err instanceof Error ? err.message : "Unknown error";
		} finally {
			this.loading = false;
			this.renderStories();
		}
	}

	// Method to refresh the view
	async refresh() {
		await this.loadStories();
	}
}

