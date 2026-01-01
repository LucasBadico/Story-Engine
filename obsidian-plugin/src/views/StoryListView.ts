import { ItemView, WorkspaceLeaf, Notice } from "obsidian";
import StoryEnginePlugin from "../main";
import { Story } from "../types";

export const STORY_LIST_VIEW_TYPE = "story-engine-list-view";

export class StoryListView extends ItemView {
	plugin: StoryEnginePlugin;
	stories: Story[] = [];
	loading: boolean = true;
	error: string | null = null;
	contentEl!: HTMLElement;
	headerEl!: HTMLElement;
	currentStory: Story | null = null;
	viewMode: "list" | "details" = "list";

	constructor(leaf: WorkspaceLeaf, plugin: StoryEnginePlugin) {
		super(leaf);
		this.plugin = plugin;
	}

	getViewType(): string {
		return STORY_LIST_VIEW_TYPE;
	}

	getDisplayText(): string {
		if (this.viewMode === "details" && this.currentStory) {
			return this.currentStory.title;
		}
		return "Stories";
	}

	getIcon(): string {
		return "book-open";
	}

	async onOpen() {
		const container = this.containerEl.children[1] as HTMLElement;
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

		// Header - will be updated based on view mode
		this.headerEl = container.createDiv({ cls: "story-engine-view-header" });
		
		// Content area
		this.contentEl = container.createDiv({ cls: "story-engine-view-content" });

		// Render based on current mode
		if (this.viewMode === "details" && this.currentStory) {
			this.renderDetails();
		} else {
			this.renderListHeader();
		}
	}

	renderListHeader() {
		if (!this.headerEl) return;

		this.headerEl.empty();
		this.headerEl.createEl("h2", { text: "Stories" });

		const headerActions = this.headerEl.createDiv({ cls: "story-engine-header-actions" });
		
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
	}

	renderDetailsHeader() {
		if (!this.headerEl) return;

		this.headerEl.empty();
		
		const headerLeft = this.headerEl.createDiv({ cls: "story-engine-header-left" });
		
		const backButton = headerLeft.createEl("button", {
			text: "← Back",
			cls: "story-engine-back-btn",
		});
		backButton.onclick = () => {
			this.showList();
		};

		headerLeft.createEl("h2", { text: this.currentStory?.title || "Story Details" });

		const headerActions = this.headerEl.createDiv({ cls: "story-engine-header-actions" });
		
		if (this.currentStory) {
			const cloneButton = headerActions.createEl("button", {
				text: "Clone Story",
				cls: "mod-cta story-engine-clone-btn",
			});
			cloneButton.onclick = async () => {
				await this.cloneStory();
			};

			const copyIdButton = headerActions.createEl("button", {
				text: "Copy ID",
				cls: "story-engine-copy-id-btn",
			});
			copyIdButton.onclick = () => {
				this.copyStoryId();
			};
		}
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
				this.showStoryDetails(story);
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

	showStoryDetails(story: Story) {
		this.currentStory = story;
		this.viewMode = "details";
		this.renderDetails();
	}

	showList() {
		this.currentStory = null;
		this.viewMode = "list";
		this.renderListHeader();
		this.renderStories();
	}

	renderDetails() {
		if (!this.contentEl || !this.currentStory) return;

		this.renderDetailsHeader();
		this.contentEl.empty();

		const story = this.currentStory;
		const details = this.contentEl.createDiv({ cls: "story-engine-details" });

		details.createEl("p", {
			text: `Status: ${story.status}`,
		});

		details.createEl("p", {
			text: `Version: ${story.version_number}`,
		});

		details.createEl("p", {
			text: `Created: ${new Date(story.created_at).toLocaleString()}`,
		});

		details.createEl("p", {
			text: `Updated: ${new Date(story.updated_at).toLocaleString()}`,
		});

		details.createEl("p", {
			text: `ID: ${story.id}`,
			cls: "story-engine-id",
		});

		// Action buttons section
		const actionsSection = this.contentEl.createDiv({ cls: "story-engine-details-actions" });
		
		const syncButton = actionsSection.createEl("button", {
			text: "Sync from Service",
			cls: "story-engine-sync-btn",
		});
		syncButton.onclick = async () => {
			try {
				new Notice(`Syncing story "${story.title}"...`);
				await this.plugin.syncService.pullStory(story.id);
				new Notice(`Story synced successfully!`);
			} catch (err) {
				const errorMessage =
					err instanceof Error ? err.message : "Failed to sync story";
				new Notice(`Error: ${errorMessage}`, 5000);
			}
		};

		const pushButton = actionsSection.createEl("button", {
			text: "Push to Service",
			cls: "story-engine-push-btn",
		});
		pushButton.onclick = async () => {
			try {
				const folderPath = this.plugin.fileManager.getStoryFolderPath(story.title);
				new Notice(`Pushing story "${story.title}"...`);
				await this.plugin.syncService.pushStory(folderPath);
				new Notice(`Story pushed successfully!`);
			} catch (err) {
				const errorMessage =
					err instanceof Error ? err.message : "Failed to push story";
				new Notice(`Error: ${errorMessage}`, 5000);
			}
		};
	}

	async cloneStory() {
		if (!this.currentStory) return;

		const cloneButton = this.headerEl?.querySelector(".story-engine-clone-btn") as HTMLButtonElement;
		if (cloneButton) {
			cloneButton.disabled = true;
			cloneButton.setText("Cloning...");
		}

		try {
			if (!this.plugin.settings.tenantId) {
				throw new Error("Tenant ID not configured");
			}

			const clonedStory = await this.plugin.apiClient.cloneStory(
				this.currentStory.id,
				this.plugin.settings.tenantId
			);
			
			new Notice(`Story "${clonedStory.title}" cloned successfully!`);
			
			// Refresh the list and show the cloned story
			await this.loadStories();
			this.showStoryDetails(clonedStory);
		} catch (err) {
			const errorMessage = err instanceof Error ? err.message : "Clone failed";
			new Notice(`Error: ${errorMessage}`, 5000);
			if (cloneButton) {
				cloneButton.setText("Clone Story");
				cloneButton.disabled = false;
			}
		}
	}

	copyStoryId() {
		if (!this.currentStory) return;

		const textarea = document.createElement("textarea");
		textarea.value = this.currentStory.id;
		textarea.style.position = "fixed";
		textarea.style.opacity = "0";
		textarea.style.left = "-9999px";
		document.body.appendChild(textarea);
		textarea.select();
		
		try {
			const doc = document;
			if (doc.execCommand) {
				doc.execCommand("copy");
				const copyButton = this.headerEl?.querySelector(".story-engine-copy-id-btn") as HTMLButtonElement;
				if (copyButton) {
					copyButton.setText("Copied!");
					setTimeout(() => {
						copyButton.setText("Copy ID");
					}, 2000);
				}
				new Notice("Story ID copied to clipboard");
			}
		} catch (err) {
			console.error("Failed to copy ID:", err);
			new Notice("Failed to copy ID", 3000);
		}
		
		document.body.removeChild(textarea);
	}
}

