import { ItemView, WorkspaceLeaf, Notice, Modal } from "obsidian";
import StoryEnginePlugin from "../main";
import { Story, Chapter, Scene, Beat } from "../types";
import { ChapterModal } from "./modals/ChapterModal";
import { SceneModal } from "./modals/SceneModal";
import { BeatModal } from "./modals/BeatModal";

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
	currentTab: "chapters" | "scenes" | "beats" = "chapters";
	chapters: Chapter[] = [];
	scenes: Scene[] = [];
	beats: Beat[] = [];
	loadingHierarchy: boolean = false;

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

			storyItem.onclick = async () => {
				await this.showStoryDetails(story);
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

	async showStoryDetails(story: Story) {
		this.currentStory = story;
		this.viewMode = "details";
		this.currentTab = "chapters";
		await this.loadHierarchy();
		this.renderDetails();
	}

	async loadHierarchy() {
		if (!this.currentStory) return;

		this.loadingHierarchy = true;
		try {
			this.chapters = await this.plugin.apiClient.getChapters(this.currentStory.id);
			this.scenes = await this.plugin.apiClient.getScenesByStory(this.currentStory.id);
			this.beats = await this.plugin.apiClient.getBeatsByStory(this.currentStory.id);
		} catch (err) {
			const errorMessage = err instanceof Error ? err.message : "Failed to load hierarchy";
			new Notice(`Error: ${errorMessage}`, 5000);
		} finally {
			this.loadingHierarchy = false;
		}
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
				await this.loadHierarchy();
				this.renderTabContent();
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

		// Tabs section
		this.renderTabs();
		this.renderTabContent();
	}

	renderTabs() {
		if (!this.contentEl) return;

		// Remove existing tabs if any
		const existingTabs = this.contentEl.querySelector(".story-engine-tabs");
		if (existingTabs) {
			existingTabs.remove();
		}

		const tabsContainer = this.contentEl.createDiv({ cls: "story-engine-tabs" });
		
		const chaptersTab = tabsContainer.createEl("button", {
			text: "Chapters",
			cls: `story-engine-tab ${this.currentTab === "chapters" ? "is-active" : ""}`,
		});
		chaptersTab.onclick = () => {
			this.currentTab = "chapters";
			this.renderTabs();
			this.renderTabContent();
		};

		const scenesTab = tabsContainer.createEl("button", {
			text: "Scenes",
			cls: `story-engine-tab ${this.currentTab === "scenes" ? "is-active" : ""}`,
		});
		scenesTab.onclick = () => {
			this.currentTab = "scenes";
			this.renderTabs();
			this.renderTabContent();
		};

		const beatsTab = tabsContainer.createEl("button", {
			text: "Beats",
			cls: `story-engine-tab ${this.currentTab === "beats" ? "is-active" : ""}`,
		});
		beatsTab.onclick = () => {
			this.currentTab = "beats";
			this.renderTabs();
			this.renderTabContent();
		};
	}

	renderTabContent() {
		if (!this.contentEl) return;

		// Remove existing tab content
		const existingContent = this.contentEl.querySelector(".story-engine-tab-content");
		if (existingContent) {
			existingContent.remove();
		}

		const tabContent = this.contentEl.createDiv({ cls: "story-engine-tab-content" });

		if (this.loadingHierarchy) {
			tabContent.createEl("p", { text: "Loading..." });
			return;
		}

		switch (this.currentTab) {
			case "chapters":
				this.renderChaptersTab(tabContent);
				break;
			case "scenes":
				this.renderScenesTab(tabContent);
				break;
			case "beats":
				this.renderBeatsTab(tabContent);
				break;
		}
	}

	renderChaptersTab(container: HTMLElement) {
		container.empty();

		const header = container.createDiv({ cls: "story-engine-tab-header" });
		const createButton = header.createEl("button", {
			text: "Create Chapter",
			cls: "mod-cta",
		});
		createButton.onclick = () => {
			if (!this.currentStory) return;
			new ChapterModal(this.app, async (chapter) => {
				try {
					await this.plugin.apiClient.createChapter(this.currentStory!.id, chapter);
					await this.loadHierarchy();
					this.renderTabContent();
					new Notice("Chapter created successfully");
				} catch (err) {
					throw err;
				}
			}, this.chapters).open();
		};

		const list = container.createDiv({ cls: "story-engine-list" });
		
		if (this.chapters.length === 0) {
			list.createEl("p", { text: "No chapters found." });
			return;
		}

		for (const chapter of this.chapters.sort((a, b) => a.number - b.number)) {
			const item = list.createDiv({ cls: "story-engine-item" });
			item.createDiv({
				cls: "story-engine-title",
				text: `Chapter ${chapter.number}: ${chapter.title}`,
			});
			const meta = item.createDiv({ cls: "story-engine-meta" });
			meta.createEl("span", { text: `Status: ${chapter.status}` });
			
			// Hover actions
			const actions = item.createDiv({ cls: "story-engine-item-actions" });
			actions.createEl("button", { text: "Edit" }).onclick = () => {
				new ChapterModal(this.app, async (updatedChapter) => {
					try {
						await this.plugin.apiClient.updateChapter(chapter.id, updatedChapter);
						await this.loadHierarchy();
						this.renderTabContent();
						new Notice("Chapter updated successfully");
					} catch (err) {
						throw err;
					}
				}, this.chapters, chapter).open();
			};
			actions.createEl("button", { text: "Delete" }).onclick = async () => {
				if (confirm("Delete this chapter?")) {
					try {
						await this.plugin.apiClient.deleteChapter(chapter.id);
						await this.loadHierarchy();
						this.renderTabContent();
						new Notice("Chapter deleted");
					} catch (err) {
						new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
					}
				}
			};
		}
	}

	renderScenesTab(container: HTMLElement) {
		container.empty();

		// Group scenes by chapter
		const scenesByChapter = new Map<string | null, Scene[]>();
		for (const scene of this.scenes) {
			const chapterId = scene.chapter_id || null;
			if (!scenesByChapter.has(chapterId)) {
				scenesByChapter.set(chapterId, []);
			}
			scenesByChapter.get(chapterId)!.push(scene);
		}

		const list = container.createDiv({ cls: "story-engine-list" });

		// Render scenes grouped by chapter
		for (const chapter of this.chapters.sort((a, b) => a.number - b.number)) {
			const chapterScenes = scenesByChapter.get(chapter.id) || [];
			const group = list.createDiv({ cls: "story-engine-group" });
			const groupHeader = group.createDiv({ cls: "story-engine-group-header" });
			groupHeader.createEl("h3", { text: `Chapter ${chapter.number}: ${chapter.title}` });
			
			const addButton = groupHeader.createEl("button", {
				text: "+ Add Scene",
				cls: "story-engine-add-btn",
			});
			addButton.onclick = () => {
				if (!this.currentStory) return;
				new SceneModal(this.app, this.currentStory.id, this.chapters, async (scene) => {
					try {
						scene.chapter_id = chapter.id;
						await this.plugin.apiClient.createScene(scene);
						await this.loadHierarchy();
						this.renderTabContent();
						new Notice("Scene created successfully");
					} catch (err) {
						throw err;
					}
				}, this.scenes).open();
			};

			const groupItems = group.createDiv({ cls: "story-engine-group-items" });
			if (chapterScenes.length === 0) {
				groupItems.createEl("p", { text: "No scenes in this chapter." });
			} else {
				for (const scene of chapterScenes.sort((a, b) => a.order_num - b.order_num)) {
					this.renderSceneItem(groupItems, scene);
				}
			}
		}

		// Render orphan scenes (no chapter)
		const orphanScenes = scenesByChapter.get(null) || [];
		if (orphanScenes.length > 0 || scenesByChapter.size === 0) {
			const group = list.createDiv({ cls: "story-engine-group" });
			const groupHeader = group.createDiv({ cls: "story-engine-group-header" });
			groupHeader.createEl("h3", { text: "Sem Chapter" });
			
			const addButton = groupHeader.createEl("button", {
				text: "+ Add Scene",
				cls: "story-engine-add-btn",
			});
			addButton.onclick = () => {
				if (!this.currentStory) return;
				new SceneModal(this.app, this.currentStory.id, this.chapters, async (scene) => {
					try {
						scene.chapter_id = null;
						await this.plugin.apiClient.createScene(scene);
						await this.loadHierarchy();
						this.renderTabContent();
						new Notice("Scene created successfully");
					} catch (err) {
						throw err;
					}
				}, this.scenes).open();
			};

			const groupItems = group.createDiv({ cls: "story-engine-group-items" });
			for (const scene of orphanScenes.sort((a, b) => a.order_num - b.order_num)) {
				this.renderSceneItem(groupItems, scene);
			}
		}
	}

	renderSceneItem(container: HTMLElement, scene: Scene) {
		const item = container.createDiv({ cls: "story-engine-item" });
		item.createDiv({
			cls: "story-engine-title",
			text: `Scene ${scene.order_num}: ${scene.goal || "Untitled"}`,
		});
		const meta = item.createDiv({ cls: "story-engine-meta" });
		if (scene.time_ref) {
			meta.createEl("span", { text: `Time: ${scene.time_ref}` });
		}

		const actions = item.createDiv({ cls: "story-engine-item-actions" });
		actions.createEl("button", { text: "Edit" }).onclick = () => {
			if (!this.currentStory) return;
			new SceneModal(this.app, this.currentStory.id, this.chapters, async (updatedScene) => {
				try {
					await this.plugin.apiClient.updateScene(scene.id, updatedScene);
					await this.loadHierarchy();
					this.renderTabContent();
					new Notice("Scene updated successfully");
				} catch (err) {
					throw err;
				}
			}, this.scenes, scene).open();
		};
		actions.createEl("button", { text: "Move" }).onclick = async () => {
			await this.showMoveSceneModal(scene);
		};
		actions.createEl("button", { text: "Delete" }).onclick = async () => {
			if (confirm("Delete this scene?")) {
				try {
					await this.plugin.apiClient.deleteScene(scene.id);
					await this.loadHierarchy();
					this.renderTabContent();
					new Notice("Scene deleted");
				} catch (err) {
					new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
				}
			}
		};
	}

	renderBeatsTab(container: HTMLElement) {
		container.empty();

		// Group beats by scene (which is in a chapter)
		const beatsByScene = new Map<string, Beat[]>();
		for (const beat of this.beats) {
			if (!beatsByScene.has(beat.scene_id)) {
				beatsByScene.set(beat.scene_id, []);
			}
			beatsByScene.get(beat.scene_id)!.push(beat);
		}

		const list = container.createDiv({ cls: "story-engine-list" });

		// Render beats grouped by chapter > scene
		for (const chapter of this.chapters.sort((a, b) => a.number - b.number)) {
			const chapterScenes = this.scenes.filter(s => s.chapter_id === chapter.id)
				.sort((a, b) => a.order_num - b.order_num);
			
			for (const scene of chapterScenes) {
				const sceneBeats = beatsByScene.get(scene.id) || [];
				const group = list.createDiv({ cls: "story-engine-group" });
				const groupHeader = group.createDiv({ cls: "story-engine-group-header" });
				groupHeader.createEl("h3", { 
					text: `Chapter ${chapter.number} > Scene ${scene.order_num}: ${scene.goal || "Untitled"}` 
				});
				
				const addButton = groupHeader.createEl("button", {
					text: "+ Add Beat",
					cls: "story-engine-add-btn",
				});
				addButton.onclick = () => {
					if (!this.currentStory) return;
					new BeatModal(this.app, this.currentStory.id, this.scenes, async (beat) => {
						try {
							beat.scene_id = scene.id;
							await this.plugin.apiClient.createBeat(beat);
							await this.loadHierarchy();
							this.renderTabContent();
							new Notice("Beat created successfully");
						} catch (err) {
							throw err;
						}
					}, this.beats).open();
				};

				const groupItems = group.createDiv({ cls: "story-engine-group-items" });
				if (sceneBeats.length === 0) {
					groupItems.createEl("p", { text: "No beats in this scene." });
				} else {
					for (const beat of sceneBeats.sort((a, b) => a.order_num - b.order_num)) {
						this.renderBeatItem(groupItems, beat);
					}
				}
			}
		}

		// Render orphan beats (no scene)
		const orphanBeats = this.beats.filter(b => {
			const scene = this.scenes.find(s => s.id === b.scene_id);
			return !scene;
		});
		if (orphanBeats.length > 0) {
			const group = list.createDiv({ cls: "story-engine-group" });
			const groupHeader = group.createDiv({ cls: "story-engine-group-header" });
			groupHeader.createEl("h3", { text: "Sem Scene" });
			
			const addButton = groupHeader.createEl("button", {
				text: "+ Add Beat",
				cls: "story-engine-add-btn",
			});
			addButton.onclick = () => {
				if (!this.currentStory) return;
				new BeatModal(this.app, this.currentStory.id, this.scenes, async (beat) => {
					try {
						// For orphan beats, scene_id will be set in modal
						await this.plugin.apiClient.createBeat(beat);
						await this.loadHierarchy();
						this.renderTabContent();
						new Notice("Beat created successfully");
					} catch (err) {
						throw err;
					}
				}, this.beats).open();
			};

			const groupItems = group.createDiv({ cls: "story-engine-group-items" });
			for (const beat of orphanBeats.sort((a, b) => a.order_num - b.order_num)) {
				this.renderBeatItem(groupItems, beat);
			}
		}
	}

	renderBeatItem(container: HTMLElement, beat: Beat) {
		const item = container.createDiv({ cls: "story-engine-item" });
		item.createDiv({
			cls: "story-engine-title",
			text: `Beat ${beat.order_num}: ${beat.type}`,
		});
		const meta = item.createDiv({ cls: "story-engine-meta" });
		if (beat.intent) {
			meta.createEl("span", { text: `Intent: ${beat.intent}` });
		}
		if (beat.outcome) {
			meta.createEl("span", { text: `Outcome: ${beat.outcome}` });
		}

		const actions = item.createDiv({ cls: "story-engine-item-actions" });
		actions.createEl("button", { text: "Edit" }).onclick = () => {
			if (!this.currentStory) return;
			new BeatModal(this.app, this.currentStory.id, this.scenes, async (updatedBeat) => {
				try {
					await this.plugin.apiClient.updateBeat(beat.id, updatedBeat);
					await this.loadHierarchy();
					this.renderTabContent();
					new Notice("Beat updated successfully");
				} catch (err) {
					throw err;
				}
			}, this.beats, beat).open();
		};
		actions.createEl("button", { text: "Move" }).onclick = async () => {
			await this.showMoveBeatModal(beat);
		};
		actions.createEl("button", { text: "Delete" }).onclick = async () => {
			if (confirm("Delete this beat?")) {
				try {
					await this.plugin.apiClient.deleteBeat(beat.id);
					await this.loadHierarchy();
					this.renderTabContent();
					new Notice("Beat deleted");
				} catch (err) {
					new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
				}
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
			await this.showStoryDetails(clonedStory);
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

	async showMoveSceneModal(scene: Scene) {
		if (!this.currentStory) return;

		const modal = new Modal(this.app);
		modal.titleEl.setText("Move Scene");

		const content = modal.contentEl;
		content.createEl("p", { text: `Move scene "${scene.goal || `Scene ${scene.order_num}`}" to:` });

		const select = content.createEl("select", { cls: "story-engine-move-select" });
		const noChapterOption = select.createEl("option", { text: "No Chapter", value: "" });
		
		for (const chapter of this.chapters.sort((a, b) => a.number - b.number)) {
			const option = select.createEl("option", {
				text: `Chapter ${chapter.number}: ${chapter.title}`,
				value: chapter.id,
			});
			if (scene.chapter_id === chapter.id) {
				option.selected = true;
			}
		}

		if (!scene.chapter_id) {
			noChapterOption.selected = true;
		}

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const moveButton = buttonContainer.createEl("button", {
			text: "Move",
			cls: "mod-cta",
		});
		moveButton.onclick = async () => {
			const selectedChapterId = select.value || null;
			try {
				await this.plugin.apiClient.moveScene(scene.id, selectedChapterId);
				await this.loadHierarchy();
				this.renderTabContent();
				modal.close();
				new Notice("Scene moved successfully");
			} catch (err) {
				const errorMessage = err instanceof Error ? err.message : "Failed to move scene";
				new Notice(`Error: ${errorMessage}`, 5000);
			}
		};

		const cancelButton = buttonContainer.createEl("button", { text: "Cancel" });
		cancelButton.onclick = () => modal.close();

		modal.open();
	}

	async showMoveBeatModal(beat: Beat) {
		if (!this.currentStory) return;

		const modal = new Modal(this.app);
		modal.titleEl.setText("Move Beat");

		const content = modal.contentEl;
		content.createEl("p", { text: `Move beat "${beat.type}" to:` });

		const select = content.createEl("select", { cls: "story-engine-move-select" });
		
		// Group scenes by chapter for better UX
		for (const chapter of this.chapters.sort((a, b) => a.number - b.number)) {
			const chapterScenes = this.scenes
				.filter(s => s.chapter_id === chapter.id)
				.sort((a, b) => a.order_num - b.order_num);
			
			if (chapterScenes.length > 0) {
				const optgroup = select.createEl("optgroup");
				optgroup.label = `Chapter ${chapter.number}: ${chapter.title}`;
				for (const scene of chapterScenes) {
					const option = optgroup.createEl("option", {
						text: `Scene ${scene.order_num}: ${scene.goal || "Untitled"}`,
						value: scene.id,
					});
					if (beat.scene_id === scene.id) {
						option.selected = true;
					}
				}
			}
		}

		// Add orphan scenes
		const orphanScenes = this.scenes.filter(s => !s.chapter_id);
		if (orphanScenes.length > 0) {
			const optgroup = select.createEl("optgroup");
			optgroup.label = "No Chapter";
			for (const scene of orphanScenes.sort((a, b) => a.order_num - b.order_num)) {
				const option = optgroup.createEl("option", {
					text: `Scene ${scene.order_num}: ${scene.goal || "Untitled"}`,
					value: scene.id,
				});
				if (beat.scene_id === scene.id) {
					option.selected = true;
				}
			}
		}

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const moveButton = buttonContainer.createEl("button", {
			text: "Move",
			cls: "mod-cta",
		});
		moveButton.onclick = async () => {
			const selectedSceneId = select.value;
			if (!selectedSceneId) {
				new Notice("Please select a scene", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.moveBeat(beat.id, selectedSceneId);
				await this.loadHierarchy();
				this.renderTabContent();
				modal.close();
				new Notice("Beat moved successfully");
			} catch (err) {
				const errorMessage = err instanceof Error ? err.message : "Failed to move beat";
				new Notice(`Error: ${errorMessage}`, 5000);
			}
		};

		const cancelButton = buttonContainer.createEl("button", { text: "Cancel" });
		cancelButton.onclick = () => modal.close();

		modal.open();
	}
}

