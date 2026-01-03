import { ItemView, WorkspaceLeaf, Notice, Modal, setIcon } from "obsidian";
import StoryEnginePlugin from "../main";
import { Story, Chapter, Scene, Beat, ContentBlock, ContentBlockReference, World, RPGSystem } from "../types";
import { ChapterModal } from "./modals/ChapterModal";
import { SceneModal } from "./modals/SceneModal";
import { BeatModal } from "./modals/BeatModal";
import { ContentBlockModal } from "./modals/ContentBlockModal";

export const STORY_LIST_VIEW_TYPE = "story-engine-list-view";

export class StoryListView extends ItemView {
	plugin: StoryEnginePlugin;
	stories: Story[] = [];
	worlds: World[] = [];
	rpgSystems: RPGSystem[] = [];
	loading: boolean = true;
	error: string | null = null;
	contentEl!: HTMLElement;
	headerEl!: HTMLElement;
	currentStory: Story | null = null;
	viewMode: "list" | "details" = "list";
	currentTab: "chapters" | "scenes" | "beats" | "contents" = "chapters";
	listTab: "stories" | "worlds" | "rpg-systems" = "stories";
	expandedWorldId: string | null = null;
	chapters: Chapter[] = [];
	scenes: Scene[] = [];
	beats: Beat[] = [];
	contentBlocks: ContentBlock[] = [];
	contentBlockRefs: ContentBlockReference[] = [];
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
			this.renderListContent();
		}
	}

	renderListHeader() {
		if (!this.headerEl) return;

		this.headerEl.empty();
		this.headerEl.createEl("h2", { text: "Stories" });

		// Tabs for list view
		const tabsContainer = this.headerEl.createDiv({ cls: "story-engine-tabs" });
		
		const storiesTab = tabsContainer.createEl("button", {
			text: "Stories",
			cls: `story-engine-tab ${this.listTab === "stories" ? "is-active" : ""}`,
		});
		storiesTab.onclick = () => {
			this.listTab = "stories";
			this.renderListHeader();
			this.renderListContent();
		};

		const worldsTab = tabsContainer.createEl("button", {
			text: "Worlds",
			cls: `story-engine-tab ${this.listTab === "worlds" ? "is-active" : ""}`,
		});
		worldsTab.onclick = () => {
			this.listTab = "worlds";
			this.renderListHeader();
			this.renderListContent();
		};

		const rpgSystemsTab = tabsContainer.createEl("button", {
			text: "RPG Systems",
			cls: `story-engine-tab ${this.listTab === "rpg-systems" ? "is-active" : ""}`,
		});
		rpgSystemsTab.onclick = () => {
			this.listTab = "rpg-systems";
			this.renderListHeader();
			this.renderListContent();
		};
	}

	renderListContent() {
		if (!this.contentEl) return;

		this.contentEl.empty();

		if (this.loading) {
			this.contentEl.createEl("p", { text: "Loading..." });
			return;
		}

		if (this.error) {
			this.contentEl.createEl("p", {
				text: `Error: ${this.error}`,
				cls: "story-engine-error",
			});
			// Still show actions bar even on error
			this.renderActionsBar();
			return;
		}

		switch (this.listTab) {
			case "stories":
				this.renderStoriesTab();
				break;
			case "worlds":
				this.renderWorldsTab();
				break;
			case "rpg-systems":
				this.renderRPGSystemsTab();
				break;
		}

		// Actions bar below list
		this.renderActionsBar();
	}

	renderActionsBar() {
		if (!this.contentEl) return;

		const actionsBar = this.contentEl.createDiv({ cls: "story-engine-actions-bar" });
		
		const refreshButton = actionsBar.createEl("button", {
			text: "Refresh",
			cls: "story-engine-refresh-btn",
		});
		refreshButton.onclick = async () => {
			await this.loadStories();
		};

		const syncAllButton = actionsBar.createEl("button", {
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

		// Create button based on active tab
		let createButtonText = "Create Story";
		let createButtonAction = () => {
			this.plugin.createStoryCommand();
		};

		if (this.listTab === "worlds") {
			createButtonText = "Create World";
			createButtonAction = () => {
				// TODO: Implement create world command
				new Notice("Create World - Coming soon", 3000);
			};
		} else if (this.listTab === "rpg-systems") {
			createButtonText = "Create RPG System";
			createButtonAction = () => {
				// TODO: Implement create RPG system command
				new Notice("Create RPG System - Coming soon", 3000);
			};
		}

		const createButton = actionsBar.createEl("button", {
			text: createButtonText,
			cls: "mod-cta story-engine-create-btn",
		});
		createButton.onclick = createButtonAction;
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

	renderStoriesTab() {
		if (this.stories.length === 0) {
			this.contentEl.createEl("p", { text: "No stories found." });
			return;
		}

		const storiesList = this.contentEl.createDiv({ cls: "story-engine-list" });

		for (const story of this.stories) {
			this.renderStoryItem(storiesList, story);
		}
	}

	renderWorldsTab() {
		if (this.worlds.length === 0) {
			this.contentEl.createEl("p", { text: "No worlds found." });
			return;
		}

		// Group stories by world
		const storiesByWorld = new Map<string | null, Story[]>();
		for (const story of this.stories) {
			const worldId = story.world_id || null;
			if (!storiesByWorld.has(worldId)) {
				storiesByWorld.set(worldId, []);
			}
			storiesByWorld.get(worldId)!.push(story);
		}

		const worldsList = this.contentEl.createDiv({ cls: "story-engine-list" });

		for (const world of this.worlds) {
			const worldStories = storiesByWorld.get(world.id) || [];
			const worldItem = worldsList.createDiv({ cls: "story-engine-world-item" });
			
			const worldHeader = worldItem.createDiv({ cls: "story-engine-world-item-header" });
			const worldTitle = worldHeader.createDiv({ cls: "story-engine-world-title" });
			worldTitle.createEl("h3", { text: world.name });
			if (world.description) {
				worldTitle.createEl("p", { 
					text: world.description,
					cls: "story-engine-world-description"
				});
			}
			
			// Accordion button
			const accordionButton = worldHeader.createEl("button", {
				text: worldStories.length > 0 ? `${worldStories.length} story${worldStories.length !== 1 ? 's' : ''}` : "No stories",
				cls: `story-engine-accordion-btn ${this.expandedWorldId === world.id ? "is-expanded" : ""}`,
			});
			accordionButton.onclick = () => {
				if (this.expandedWorldId === world.id) {
					this.expandedWorldId = null;
				} else {
					this.expandedWorldId = world.id;
				}
				this.renderListContent();
			};

			// Stories content (accordion)
			if (this.expandedWorldId === world.id && worldStories.length > 0) {
				const storiesContent = worldItem.createDiv({ cls: "story-engine-world-stories-content" });
				for (const story of worldStories) {
					this.renderStoryItem(storiesContent, story);
				}
			}
		}

		// Render stories without world
		const storiesWithoutWorld = storiesByWorld.get(null) || [];
		if (storiesWithoutWorld.length > 0) {
			const noWorldItem = worldsList.createDiv({ cls: "story-engine-world-item" });
			const noWorldHeader = noWorldItem.createDiv({ cls: "story-engine-world-item-header" });
			noWorldHeader.createEl("h3", { text: "No World" });
			
			const accordionButton = noWorldHeader.createEl("button", {
				text: `${storiesWithoutWorld.length} story${storiesWithoutWorld.length !== 1 ? 's' : ''}`,
				cls: `story-engine-accordion-btn ${this.expandedWorldId === "no-world" ? "is-expanded" : ""}`,
			});
			accordionButton.onclick = () => {
				if (this.expandedWorldId === "no-world") {
					this.expandedWorldId = null;
				} else {
					this.expandedWorldId = "no-world";
				}
				this.renderListContent();
			};

			if (this.expandedWorldId === "no-world") {
				const storiesContent = noWorldItem.createDiv({ cls: "story-engine-world-stories-content" });
				for (const story of storiesWithoutWorld) {
					this.renderStoryItem(storiesContent, story);
				}
			}
		}
	}

	renderRPGSystemsTab() {
		if (this.rpgSystems.length === 0) {
			this.contentEl.createEl("p", { text: "No RPG systems found." });
			return;
		}

		const rpgSystemsList = this.contentEl.createDiv({ cls: "story-engine-list" });

		for (const rpgSystem of this.rpgSystems) {
			const rpgSystemItem = rpgSystemsList.createDiv({
				cls: "story-engine-item",
			});

			const title = rpgSystemItem.createDiv({
				cls: "story-engine-title",
				text: rpgSystem.name,
			});

			const meta = rpgSystemItem.createDiv({
				cls: "story-engine-meta",
			});
			if (rpgSystem.description) {
				meta.createEl("span", {
					text: rpgSystem.description,
				});
			}
			if (rpgSystem.is_builtin) {
				meta.createEl("span", {
					text: "Built-in",
					cls: "story-engine-badge",
				});
			}
		}
	}

	renderStoryItem(container: HTMLElement, story: Story) {
		const storyItem = container.createDiv({
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
		// Show world if connected
		if (story.world_id) {
			const world = this.worlds.find(w => w.id === story.world_id);
			if (world) {
				meta.createEl("span", {
					text: `World: ${world.name}`,
				});
			}
		}

		storyItem.onclick = async () => {
			await this.showStoryDetails(story);
		};
	}

	async loadStories() {
		this.loading = true;
		this.error = null;
		try {
			if (!this.plugin.settings.tenantId) {
				this.error = "Tenant ID not configured";
				this.loading = false;
				this.renderListContent();
				return;
			}

			// Load worlds first to group stories
			this.worlds = await this.plugin.apiClient.getWorlds();
			// Load stories
			this.stories = await this.plugin.apiClient.listStories();
			// Load RPG systems
			this.rpgSystems = await this.plugin.apiClient.getRPGSystems();
		} catch (err) {
			this.error = err instanceof Error ? err.message : "Failed to load stories";
			console.error("Error loading stories:", err);
		} finally {
			this.loading = false;
			this.renderListContent();
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
			
			// Load content blocks and references
			const contentBlocksMap = new Map<string, ContentBlock>();
			this.contentBlockRefs = [];
			
			// Get content blocks by chapter
			for (const chapter of this.chapters) {
				const chapterBlocks = await this.plugin.apiClient.getContentBlocks(chapter.id);
				for (const block of chapterBlocks) {
					contentBlocksMap.set(block.id, block);
				}
			}
			
			// Get content blocks by scene
			for (const scene of this.scenes) {
				const sceneBlocks = await this.plugin.apiClient.getContentBlocksByScene(scene.id);
				for (const block of sceneBlocks) {
					contentBlocksMap.set(block.id, block);
				}
			}
			
			// Get content blocks by beat
			for (const beat of this.beats) {
				const beatBlocks = await this.plugin.apiClient.getContentBlocksByBeat(beat.id);
				for (const block of beatBlocks) {
					contentBlocksMap.set(block.id, block);
				}
			}
			
			// Convert map to array
			this.contentBlocks = Array.from(contentBlocksMap.values());
			
			// Load references for all content blocks
			for (const block of this.contentBlocks) {
				const refs = await this.plugin.apiClient.getContentBlockReferences(block.id);
				this.contentBlockRefs.push(...refs);
			}
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
		this.renderListContent();
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

		const contentsTab = tabsContainer.createEl("button", {
			text: "Contents",
			cls: `story-engine-tab ${this.currentTab === "contents" ? "is-active" : ""}`,
		});
		contentsTab.onclick = () => {
			this.currentTab = "contents";
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
			case "contents":
				this.renderContentsTab(tabContent);
				break;
		}
	}

	renderChaptersTab(container: HTMLElement) {
		container.empty();

		const list = container.createDiv({ cls: "story-engine-list" });
		
		if (this.chapters.length === 0) {
			list.createEl("p", { text: "No chapters found." });
		} else {
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
				// Up/Down buttons
				if (chapter.number > 1) {
					actions.createEl("button", { text: "Up" }).onclick = () => {
						this.moveChapterUp(chapter);
					};
				}
				if (chapter.number < this.chapters.length) {
					actions.createEl("button", { text: "Down" }).onclick = () => {
						this.moveChapterDown(chapter);
					};
				}
				// Add Content button (only in Contents tab)
				if (this.currentTab === "contents") {
					actions.createEl("button", { text: "+ Content" }).onclick = () => {
						this.createContentForEntity("chapter", chapter.id, chapter.id);
					};
				}
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

		// Footer with Create Chapter button
		const footer = container.createDiv({ cls: "story-engine-list-footer" });
		const createButton = footer.createEl("button", {
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

			const groupItems = group.createDiv({ cls: "story-engine-group-items" });
			if (chapterScenes.length === 0) {
				groupItems.createEl("p", { text: "No scenes in this chapter." });
			} else {
				for (const scene of chapterScenes.sort((a, b) => a.order_num - b.order_num)) {
					this.renderSceneItem(groupItems, scene);
				}
			}

			// Footer with Add Scene button (appears on hover)
			const groupFooter = group.createDiv({ cls: "story-engine-group-footer" });
			const addButton = groupFooter.createEl("button", {
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
		}

		// Render orphan scenes (no chapter)
		const orphanScenes = scenesByChapter.get(null) || [];
		if (orphanScenes.length > 0 || scenesByChapter.size === 0) {
			const group = list.createDiv({ cls: "story-engine-group" });
			const groupHeader = group.createDiv({ cls: "story-engine-group-header" });
			groupHeader.createEl("h3", { text: "Sem Chapter" });

			const groupItems = group.createDiv({ cls: "story-engine-group-items" });
			for (const scene of orphanScenes.sort((a, b) => a.order_num - b.order_num)) {
				this.renderSceneItem(groupItems, scene);
			}

			// Footer with Add Scene button (appears on hover)
			const groupFooter = group.createDiv({ cls: "story-engine-group-footer" });
			const addButton = groupFooter.createEl("button", {
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
		// Up/Down buttons
		const siblingScenes = this.scenes.filter(s => s.chapter_id === scene.chapter_id)
			.sort((a, b) => a.order_num - b.order_num);
		const minOrderNum = siblingScenes.length > 0 ? Math.min(...siblingScenes.map(s => s.order_num)) : scene.order_num;
		const maxOrderNum = siblingScenes.length > 0 ? Math.max(...siblingScenes.map(s => s.order_num)) : scene.order_num;
		if (scene.order_num > minOrderNum) {
			actions.createEl("button", { text: "Up" }).onclick = () => {
				this.moveSceneUp(scene);
			};
		}
		if (scene.order_num < maxOrderNum) {
			actions.createEl("button", { text: "Down" }).onclick = () => {
				this.moveSceneDown(scene);
			};
		}
		actions.createEl("button", { text: "Relinkar" }).onclick = async () => {
			await this.showMoveSceneModal(scene);
		};
		// Add Content button (only in Contents tab)
		if (this.currentTab === "contents") {
			const chapterId = scene.chapter_id || (this.chapters.length > 0 ? this.chapters[0].id : "");
			if (chapterId) {
				actions.createEl("button", { text: "+ Content" }).onclick = () => {
					this.createContentForEntity("scene", scene.id, chapterId);
				};
			}
		}
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

		// Group beats by scene
		const beatsByScene = new Map<string, Beat[]>();
		for (const beat of this.beats) {
			if (!beatsByScene.has(beat.scene_id)) {
				beatsByScene.set(beat.scene_id, []);
			}
			beatsByScene.get(beat.scene_id)!.push(beat);
		}

		const list = container.createDiv({ cls: "story-engine-list" });

		// Render beats grouped by chapter (with scenes as sub-groups)
		for (const chapter of this.chapters.sort((a, b) => a.number - b.number)) {
			const chapterScenes = this.scenes.filter(s => s.chapter_id === chapter.id)
				.sort((a, b) => a.order_num - b.order_num);
			
			// Only create chapter group if it has scenes with beats or empty scenes
			if (chapterScenes.length > 0) {
				const chapterGroup = list.createDiv({ cls: "story-engine-chapter-group" });
				const chapterHeader = chapterGroup.createDiv({ cls: "story-engine-chapter-group-header" });
				chapterHeader.createEl("h2", { 
					text: `Chapter ${chapter.number}: ${chapter.title}` 
				});

				const chapterContent = chapterGroup.createDiv({ cls: "story-engine-chapter-group-content" });

				// Render scenes as sub-groups within the chapter
				for (const scene of chapterScenes) {
					const sceneBeats = beatsByScene.get(scene.id) || [];
					const sceneGroup = chapterContent.createDiv({ cls: "story-engine-group" });
					const sceneHeader = sceneGroup.createDiv({ cls: "story-engine-group-header" });
					sceneHeader.createEl("h3", { 
						text: `Scene ${scene.order_num}: ${scene.goal || "Untitled"}` 
					});

					const sceneItems = sceneGroup.createDiv({ cls: "story-engine-group-items" });
					if (sceneBeats.length === 0) {
						sceneItems.createEl("p", { text: "No beats in this scene." });
					} else {
						for (const beat of sceneBeats.sort((a, b) => a.order_num - b.order_num)) {
							this.renderBeatItem(sceneItems, beat);
						}
					}

					// Footer with Add Beat button (appears on hover)
					const sceneFooter = sceneGroup.createDiv({ cls: "story-engine-group-footer" });
					const addButton = sceneFooter.createEl("button", {
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
				}
			}
		}

		// Render orphan beats (no scene or scene without chapter)
		const orphanBeats = this.beats.filter(b => {
			const scene = this.scenes.find(s => s.id === b.scene_id);
			return !scene || !scene.chapter_id;
		});
		
		if (orphanBeats.length > 0 || this.scenes.some(s => !s.chapter_id)) {
			const orphanGroup = list.createDiv({ cls: "story-engine-chapter-group" });
			const orphanHeader = orphanGroup.createDiv({ cls: "story-engine-chapter-group-header" });
			orphanHeader.createEl("h2", { text: "Sem Chapter" });

			const orphanContent = orphanGroup.createDiv({ cls: "story-engine-chapter-group-content" });

			// Group orphan beats by scene
			const orphanScenes = this.scenes.filter(s => !s.chapter_id)
				.sort((a, b) => a.order_num - b.order_num);

			for (const scene of orphanScenes) {
				const sceneBeats = beatsByScene.get(scene.id) || [];
				const sceneGroup = orphanContent.createDiv({ cls: "story-engine-group" });
				const sceneHeader = sceneGroup.createDiv({ cls: "story-engine-group-header" });
				sceneHeader.createEl("h3", { 
					text: `Scene ${scene.order_num}: ${scene.goal || "Untitled"}` 
				});

				const sceneItems = sceneGroup.createDiv({ cls: "story-engine-group-items" });
				if (sceneBeats.length === 0) {
					sceneItems.createEl("p", { text: "No beats in this scene." });
				} else {
					for (const beat of sceneBeats.sort((a, b) => a.order_num - b.order_num)) {
						this.renderBeatItem(sceneItems, beat);
					}
				}

				// Footer with Add Beat button (appears on hover)
				const sceneFooter = sceneGroup.createDiv({ cls: "story-engine-group-footer" });
				const addButton = sceneFooter.createEl("button", {
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
			}

			// Beats without valid scene
			const beatsWithoutScene = orphanBeats.filter(b => {
				const scene = this.scenes.find(s => s.id === b.scene_id);
				return !scene;
			});
			if (beatsWithoutScene.length > 0) {
				const sceneGroup = orphanContent.createDiv({ cls: "story-engine-group" });
				const sceneHeader = sceneGroup.createDiv({ cls: "story-engine-group-header" });
				sceneHeader.createEl("h3", { text: "Sem Scene" });
				
				const sceneItems = sceneGroup.createDiv({ cls: "story-engine-group-items" });
				for (const beat of beatsWithoutScene.sort((a, b) => a.order_num - b.order_num)) {
					this.renderBeatItem(sceneItems, beat);
				}
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
		// Up/Down buttons
		const siblingBeats = this.beats.filter(b => b.scene_id === beat.scene_id)
			.sort((a, b) => a.order_num - b.order_num);
		const minOrderNum = siblingBeats.length > 0 ? Math.min(...siblingBeats.map(b => b.order_num)) : beat.order_num;
		const maxOrderNum = siblingBeats.length > 0 ? Math.max(...siblingBeats.map(b => b.order_num)) : beat.order_num;
		if (beat.order_num > minOrderNum) {
			actions.createEl("button", { text: "Up" }).onclick = () => {
				this.moveBeatUp(beat);
			};
		}
		if (beat.order_num < maxOrderNum) {
			actions.createEl("button", { text: "Down" }).onclick = () => {
				this.moveBeatDown(beat);
			};
		}
		actions.createEl("button", { text: "Relinkar" }).onclick = async () => {
			await this.showMoveBeatModal(beat);
		};
		// Add Content button (only in Contents tab)
		if (this.currentTab === "contents") {
			const scene = this.scenes.find(s => s.id === beat.scene_id);
			const chapterId = scene?.chapter_id || (this.chapters.length > 0 ? this.chapters[0].id : "");
			if (chapterId) {
				actions.createEl("button", { text: "+ Content" }).onclick = () => {
					this.createContentForEntity("beat", beat.id, chapterId);
				};
			}
		}
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
				this.currentStory.id
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

	async moveChapterUp(chapter: Chapter) {
		const sortedChapters = [...this.chapters].sort((a, b) => a.number - b.number);
		const currentIndex = sortedChapters.findIndex(c => c.id === chapter.id);
		if (currentIndex <= 0) return;

		const previousChapter = sortedChapters[currentIndex - 1];
		const tempNumber = chapter.number;
		
		try {
			await this.plugin.apiClient.updateChapter(chapter.id, { number: previousChapter.number });
			await this.plugin.apiClient.updateChapter(previousChapter.id, { number: tempNumber });
			await this.loadHierarchy();
			this.renderTabContent();
			new Notice("Chapter moved up");
		} catch (err) {
			new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
		}
	}

	async moveChapterDown(chapter: Chapter) {
		const sortedChapters = [...this.chapters].sort((a, b) => a.number - b.number);
		const currentIndex = sortedChapters.findIndex(c => c.id === chapter.id);
		if (currentIndex < 0 || currentIndex >= sortedChapters.length - 1) return;

		const nextChapter = sortedChapters[currentIndex + 1];
		const tempNumber = chapter.number;
		
		try {
			await this.plugin.apiClient.updateChapter(chapter.id, { number: nextChapter.number });
			await this.plugin.apiClient.updateChapter(nextChapter.id, { number: tempNumber });
			await this.loadHierarchy();
			this.renderTabContent();
			new Notice("Chapter moved down");
		} catch (err) {
			new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
		}
	}

	async moveSceneUp(scene: Scene) {
		const siblingScenes = this.scenes.filter(s => s.chapter_id === scene.chapter_id)
			.sort((a, b) => a.order_num - b.order_num);
		const currentIndex = siblingScenes.findIndex(s => s.id === scene.id);
		if (currentIndex <= 0) return;

		const previousScene = siblingScenes[currentIndex - 1];
		const tempOrderNum = scene.order_num;
		
		try {
			await this.plugin.apiClient.updateScene(scene.id, { order_num: previousScene.order_num });
			await this.plugin.apiClient.updateScene(previousScene.id, { order_num: tempOrderNum });
			await this.loadHierarchy();
			this.renderTabContent();
			new Notice("Scene moved up");
		} catch (err) {
			new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
		}
	}

	async moveSceneDown(scene: Scene) {
		const siblingScenes = this.scenes.filter(s => s.chapter_id === scene.chapter_id)
			.sort((a, b) => a.order_num - b.order_num);
		const currentIndex = siblingScenes.findIndex(s => s.id === scene.id);
		if (currentIndex < 0 || currentIndex >= siblingScenes.length - 1) return;

		const nextScene = siblingScenes[currentIndex + 1];
		const tempOrderNum = scene.order_num;
		
		try {
			await this.plugin.apiClient.updateScene(scene.id, { order_num: nextScene.order_num });
			await this.plugin.apiClient.updateScene(nextScene.id, { order_num: tempOrderNum });
			await this.loadHierarchy();
			this.renderTabContent();
			new Notice("Scene moved down");
		} catch (err) {
			new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
		}
	}

	async moveBeatUp(beat: Beat) {
		const siblingBeats = this.beats.filter(b => b.scene_id === beat.scene_id)
			.sort((a, b) => a.order_num - b.order_num);
		const currentIndex = siblingBeats.findIndex(b => b.id === beat.id);
		if (currentIndex <= 0) return;

		const previousBeat = siblingBeats[currentIndex - 1];
		const tempOrderNum = beat.order_num;
		
		try {
			await this.plugin.apiClient.updateBeat(beat.id, { order_num: previousBeat.order_num });
			await this.plugin.apiClient.updateBeat(previousBeat.id, { order_num: tempOrderNum });
			await this.loadHierarchy();
			this.renderTabContent();
			new Notice("Beat moved up");
		} catch (err) {
			new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
		}
	}

	async moveBeatDown(beat: Beat) {
		const siblingBeats = this.beats.filter(b => b.scene_id === beat.scene_id)
			.sort((a, b) => a.order_num - b.order_num);
		const currentIndex = siblingBeats.findIndex(b => b.id === beat.id);
		if (currentIndex < 0 || currentIndex >= siblingBeats.length - 1) return;

		const nextBeat = siblingBeats[currentIndex + 1];
		const tempOrderNum = beat.order_num;
		
		try {
			await this.plugin.apiClient.updateBeat(beat.id, { order_num: nextBeat.order_num });
			await this.plugin.apiClient.updateBeat(nextBeat.id, { order_num: tempOrderNum });
			await this.loadHierarchy();
			this.renderTabContent();
			new Notice("Beat moved down");
		} catch (err) {
			new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
		}
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

	renderContentsTab(container: HTMLElement) {
		container.empty();

		// Group content blocks by their references (chapter, scene, beat)
		const contentsByChapter = new Map<string, ContentBlock[]>();
		const contentsByScene = new Map<string, ContentBlock[]>();
		const contentsByBeat = new Map<string, ContentBlock[]>();
		const orphanContents: ContentBlock[] = [];

		// Organize content blocks by their references
		for (const block of this.contentBlocks) {
			const refs = this.contentBlockRefs.filter(r => r.content_block_id === block.id);
			
			if (refs.length === 0) {
				// Content block without references (orphan)
				orphanContents.push(block);
				continue;
			}

			// Check each reference to categorize
			for (const ref of refs) {
				if (ref.entity_type === "chapter") {
					if (!contentsByChapter.has(ref.entity_id)) {
						contentsByChapter.set(ref.entity_id, []);
					}
					contentsByChapter.get(ref.entity_id)!.push(block);
				} else if (ref.entity_type === "scene") {
					if (!contentsByScene.has(ref.entity_id)) {
						contentsByScene.set(ref.entity_id, []);
					}
					contentsByScene.get(ref.entity_id)!.push(block);
				} else if (ref.entity_type === "beat") {
					if (!contentsByBeat.has(ref.entity_id)) {
						contentsByBeat.set(ref.entity_id, []);
					}
					contentsByBeat.get(ref.entity_id)!.push(block);
				}
			}
		}

		const list = container.createDiv({ cls: "story-engine-list" });

		// Render contents grouped by chapter (with scenes and beats as sub-groups)
		for (const chapter of this.chapters.sort((a, b) => a.number - b.number)) {
			const chapterContents = contentsByChapter.get(chapter.id) || [];
			const chapterScenes = this.scenes.filter(s => s.chapter_id === chapter.id)
				.sort((a, b) => a.order_num - b.order_num);

			// Always create chapter group to allow adding content
			if (chapterContents.length > 0 || chapterScenes.length > 0 || chapterScenes.some(s => {
				const sceneContents = contentsByScene.get(s.id) || [];
				const sceneBeats = this.beats.filter(b => b.scene_id === s.id);
				const beatContents = sceneBeats.flatMap(b => contentsByBeat.get(b.id) || []);
				return sceneContents.length > 0 || beatContents.length > 0;
			})) {
				const chapterGroup = list.createDiv({ cls: "story-engine-chapter-group" });
				const chapterHeader = chapterGroup.createDiv({ cls: "story-engine-chapter-group-header story-engine-hoverable-header" });
				chapterHeader.createEl("h2", { 
					text: `Chapter ${chapter.number}: ${chapter.title}` 
				});
				
				// Add Content buttons on hover
				const chapterHeaderActions = chapterHeader.createDiv({ cls: "story-engine-hover-header-actions" });
				
				// Text button
				const textBtn = chapterHeaderActions.createEl("button", { 
					cls: "story-engine-add-content-btn story-engine-add-text-btn",
					attr: { "aria-label": "Add text content" }
				});
				textBtn.innerHTML = '<span class="story-engine-icon">📝</span>';
				textBtn.onclick = () => {
					this.createContentForEntity("chapter", chapter.id, chapter.id, "text");
				};
				
				// Image button
				const imageBtn = chapterHeaderActions.createEl("button", { 
					cls: "story-engine-add-content-btn story-engine-add-image-btn",
					attr: { "aria-label": "Add image content" }
				});
				imageBtn.innerHTML = '<span class="story-engine-icon">🖼️</span>';
				imageBtn.onclick = () => {
					this.createContentForEntity("chapter", chapter.id, chapter.id, "image");
				};

				const chapterContent = chapterGroup.createDiv({ cls: "story-engine-chapter-group-content" });

				// Render chapter-level contents (always show group to allow adding content)
				const chapterContentsGroup = chapterContent.createDiv({ cls: "story-engine-group" });
				const chapterContentsHeader = chapterContentsGroup.createDiv({ cls: "story-engine-group-header" });
				chapterContentsHeader.createEl("h3", { text: "Chapter Contents" });

				const chapterContentsItems = chapterContentsGroup.createDiv({ cls: "story-engine-group-items" });
				if (chapterContents.length > 0) {
					for (const block of chapterContents) {
						this.renderContentItem(chapterContentsItems, block, "chapter", chapter.id);
					}
				} else {
					chapterContentsItems.createEl("p", { text: "No content in this chapter.", cls: "story-engine-empty-content" });
				}

				// Render scenes as sub-groups within the chapter
				for (const scene of chapterScenes) {
					const sceneContents = contentsByScene.get(scene.id) || [];
					const sceneBeats = this.beats.filter(b => b.scene_id === scene.id)
						.sort((a, b) => a.order_num - b.order_num);
					const beatContents = sceneBeats.flatMap(b => contentsByBeat.get(b.id) || []);

					// Always render scene group to allow adding content
					const sceneGroup = chapterContent.createDiv({ cls: "story-engine-group" });
					const sceneHeader = sceneGroup.createDiv({ cls: "story-engine-group-header story-engine-hoverable-header" });
					sceneHeader.createEl("h3", { 
						text: `Scene ${scene.order_num}: ${scene.goal || "Untitled"}` 
					});
					
					// Add Content buttons on hover
					const sceneHeaderActions = sceneHeader.createDiv({ cls: "story-engine-hover-header-actions" });
					const chapterId = scene.chapter_id || (this.chapters.length > 0 ? this.chapters[0].id : "");
					if (chapterId) {
						// Text button
						const textBtn = sceneHeaderActions.createEl("button", { 
							cls: "story-engine-add-content-btn story-engine-add-text-btn",
							attr: { "aria-label": "Add text content" }
						});
						textBtn.innerHTML = '<span class="story-engine-icon">📝</span>';
						textBtn.onclick = () => {
							this.createContentForEntity("scene", scene.id, chapterId, "text");
						};
						
						// Image button
						const imageBtn = sceneHeaderActions.createEl("button", { 
							cls: "story-engine-add-content-btn story-engine-add-image-btn",
							attr: { "aria-label": "Add image content" }
						});
						imageBtn.innerHTML = '<span class="story-engine-icon">🖼️</span>';
						imageBtn.onclick = () => {
							this.createContentForEntity("scene", scene.id, chapterId, "image");
						};
					}

					const sceneItems = sceneGroup.createDiv({ cls: "story-engine-group-items" });
					
					// Scene-level contents
					if (sceneContents.length > 0) {
						for (const block of sceneContents) {
							this.renderContentItem(sceneItems, block, "scene", scene.id);
						}
					} else {
						sceneItems.createEl("p", { text: "No content in this scene.", cls: "story-engine-empty-content" });
					}

					// Beat-level contents
					for (const beat of sceneBeats) {
						const beatContents = contentsByBeat.get(beat.id) || [];
						const beatSubGroup = sceneItems.createDiv({ cls: "story-engine-beat-subgroup" });
						const beatSubGroupTitle = beatSubGroup.createDiv({ cls: "story-engine-beat-subgroup-title-container story-engine-hoverable-header" });
						beatSubGroupTitle.createEl("h4", {
							text: `Beat ${beat.order_num}: ${beat.type}`,
							cls: "story-engine-beat-subgroup-title",
						});
						
								// Add Content buttons on hover
								const beatHeaderActions = beatSubGroupTitle.createDiv({ cls: "story-engine-hover-header-actions" });
								const beatChapterId = scene.chapter_id || (this.chapters.length > 0 ? this.chapters[0].id : "");
								if (beatChapterId) {
									// Text button
									const textBtn = beatHeaderActions.createEl("button", { 
										cls: "story-engine-add-content-btn story-engine-add-text-btn",
										attr: { "aria-label": "Add text content" }
									});
									textBtn.innerHTML = '<span class="story-engine-icon">📝</span>';
									textBtn.onclick = () => {
										this.createContentForEntity("beat", beat.id, beatChapterId, "text");
									};
									
									// Image button
									const imageBtn = beatHeaderActions.createEl("button", { 
										cls: "story-engine-add-content-btn story-engine-add-image-btn",
										attr: { "aria-label": "Add image content" }
									});
									imageBtn.innerHTML = '<span class="story-engine-icon">🖼️</span>';
									imageBtn.onclick = () => {
										this.createContentForEntity("beat", beat.id, beatChapterId, "image");
									};
								}
						
						if (beatContents.length > 0) {
							for (const block of beatContents) {
								this.renderContentItem(beatSubGroup, block, "beat", beat.id);
							}
						} else {
							beatSubGroup.createEl("p", { text: "No content in this beat.", cls: "story-engine-empty-beat-content" });
						}
					}
				}
			}
		}

		// Render orphan contents (no chapter/scene/beat)
		if (orphanContents.length > 0 || this.scenes.some(s => !s.chapter_id)) {
			const orphanGroup = list.createDiv({ cls: "story-engine-chapter-group" });
			const orphanHeader = orphanGroup.createDiv({ cls: "story-engine-chapter-group-header" });
			orphanHeader.createEl("h2", { text: "Sem Chapter" });

			const orphanContent = orphanGroup.createDiv({ cls: "story-engine-chapter-group-content" });

			// Orphan scenes
			const orphanScenes = this.scenes.filter(s => !s.chapter_id)
				.sort((a, b) => a.order_num - b.order_num);

			for (const scene of orphanScenes) {
				const sceneContents = contentsByScene.get(scene.id) || [];
				const sceneBeats = this.beats.filter(b => b.scene_id === scene.id)
					.sort((a, b) => a.order_num - b.order_num);
				const beatContents = sceneBeats.flatMap(b => contentsByBeat.get(b.id) || []);

				// Always render scene group to allow adding content
				const sceneGroup = orphanContent.createDiv({ cls: "story-engine-group" });
				const sceneHeader = sceneGroup.createDiv({ cls: "story-engine-group-header story-engine-hoverable-header" });
				sceneHeader.createEl("h3", { 
					text: `Scene ${scene.order_num}: ${scene.goal || "Untitled"}` 
				});
				
				// Add Content button on hover
				const sceneHeaderActions = sceneHeader.createDiv({ cls: "story-engine-header-actions" });
				const chapterId = this.chapters.length > 0 ? this.chapters[0].id : "";
				if (chapterId) {
					sceneHeaderActions.createEl("button", { 
						text: "+ Content",
						cls: "story-engine-add-content-btn"
					}).onclick = () => {
						this.createContentForEntity("scene", scene.id, chapterId);
					};
				}

				const sceneItems = sceneGroup.createDiv({ cls: "story-engine-group-items" });
				
				if (sceneContents.length > 0) {
					for (const block of sceneContents) {
						this.renderContentItem(sceneItems, block, "scene", scene.id);
					}
				} else {
					sceneItems.createEl("p", { text: "No content in this scene." });
				}

				for (const beat of sceneBeats) {
					const beatContents = contentsByBeat.get(beat.id) || [];
					const beatSubGroup = sceneItems.createDiv({ cls: "story-engine-beat-subgroup" });
					const beatSubGroupTitle = beatSubGroup.createDiv({ cls: "story-engine-beat-subgroup-title-container story-engine-hoverable-header" });
					beatSubGroupTitle.createEl("h4", {
						text: `Beat ${beat.order_num}: ${beat.type}`,
						cls: "story-engine-beat-subgroup-title",
					});
					
								// Add Content button on hover
								const beatHeaderActions = beatSubGroupTitle.createDiv({ cls: "story-engine-hover-header-actions" });
					const beatChapterId = this.chapters.length > 0 ? this.chapters[0].id : "";
					if (beatChapterId) {
						beatHeaderActions.createEl("button", { 
							text: "+ Content",
							cls: "story-engine-add-content-btn"
						}).onclick = () => {
							this.createContentForEntity("beat", beat.id, beatChapterId);
						};
					}
					
					if (beatContents.length > 0) {
						for (const block of beatContents) {
							this.renderContentItem(beatSubGroup, block, "beat", beat.id);
						}
					} else {
						beatSubGroup.createEl("p", { text: "No content in this beat.", cls: "story-engine-empty-beat-content" });
					}
				}
			}

			// Orphan contents (no references)
			if (orphanContents.length > 0) {
				const orphanContentsGroup = orphanContent.createDiv({ cls: "story-engine-group" });
				const orphanContentsHeader = orphanContentsGroup.createDiv({ cls: "story-engine-group-header" });
				orphanContentsHeader.createEl("h3", { text: "Sem Referência" });

				const orphanContentsItems = orphanContentsGroup.createDiv({ cls: "story-engine-group-items" });
				for (const block of orphanContents) {
					this.renderContentItem(orphanContentsItems, block, null, null);
				}
			}
		}
	}

	renderContentItem(container: HTMLElement, contentBlock: ContentBlock, entityType: string | null, entityId: string | null) {
		const item = container.createDiv({ cls: "story-engine-item story-engine-content-item" });
		
		// Icon and preview container
		const itemContent = item.createDiv({ cls: "story-engine-content-item-content" });
		
		// Type icon
		const iconContainer = itemContent.createDiv({ cls: "story-engine-content-icon" });
		const iconMap: Record<string, string> = {
			text: "file-text",
			image: "image",
			video: "video",
			audio: "music",
			embed: "code",
			link: "external-link",
		};
		const iconName = iconMap[contentBlock.type] || "file";
		setIcon(iconContainer, iconName);

		// Preview
		const preview = itemContent.createDiv({ cls: "story-engine-content-preview" });
		if (contentBlock.type === "text") {
			const textPreview = contentBlock.content || "";
			const truncated = textPreview.length > 100 ? textPreview.substring(0, 100) + "..." : textPreview;
			preview.createEl("span", { text: truncated });
		} else if (contentBlock.type === "image") {
			const imgContainer = preview.createDiv({ cls: "story-engine-image-container" });
			const img = imgContainer.createEl("img", {
				attr: { src: contentBlock.content || "", alt: contentBlock.metadata?.alt_text || "" },
				cls: "story-engine-content-thumbnail",
			});
			img.style.maxWidth = "100px";
			img.style.maxHeight = "60px";
			img.style.objectFit = "cover";
			img.style.borderRadius = "4px";
			
			// Show attribution if available
			if (contentBlock.metadata?.attribution) {
				const attribution = imgContainer.createDiv({ cls: "story-engine-unsplash-attribution" });
				attribution.createEl("span", { 
					text: contentBlock.metadata.attribution,
					cls: "story-engine-attribution-text"
				});
			}
		} else {
			preview.createEl("span", { text: contentBlock.content || "" });
		}

		// Meta info (removed type display as icon already shows it)

		// Actions
		const actions = item.createDiv({ cls: "story-engine-item-actions" });
		
		actions.createEl("button", { text: "Edit" }).onclick = () => {
			new ContentBlockModal(this.app, async (updatedContentBlock) => {
				try {
					await this.plugin.apiClient.updateContentBlock(contentBlock.id, updatedContentBlock);
					await this.loadHierarchy();
					this.renderTabContent();
					new Notice("Content block updated successfully");
				} catch (err) {
					throw err;
				}
			}, contentBlock, this.plugin).open();
		};

		actions.createEl("button", { text: "Move" }).onclick = async () => {
			await this.showMoveContentModal(contentBlock, entityType, entityId, "move");
		};

		actions.createEl("button", { text: "Link" }).onclick = async () => {
			await this.showMoveContentModal(contentBlock, entityType, entityId, "link");
		};

		actions.createEl("button", { text: "Delete" }).onclick = async () => {
			if (confirm("Delete this content block?")) {
				try {
					await this.plugin.apiClient.deleteContentBlock(contentBlock.id);
					await this.loadHierarchy();
					this.renderTabContent();
					new Notice("Content block deleted");
				} catch (err) {
					new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
				}
			}
		};
	}

	async showMoveContentModal(contentBlock: ContentBlock, currentEntityType: string | null, currentEntityId: string | null, mode: "move" | "link") {
		if (!this.currentStory) return;

		const modal = new Modal(this.app);
		modal.titleEl.setText(mode === "move" ? "Move Content Block" : "Link Content Block");

		const content = modal.contentEl;
		content.createEl("p", { 
			text: mode === "move" 
				? `Move content block to:` 
				: `Link content block to (will appear in both places):` 
		});

		const select = content.createEl("select", { cls: "story-engine-move-select" });
		
		// Add "No Reference" option
		const noRefOption = select.createEl("option", { text: "No Reference", value: "" });
		
		// Group by chapters
		for (const chapter of this.chapters.sort((a, b) => a.number - b.number)) {
			const option = select.createEl("option", {
				text: `Chapter ${chapter.number}: ${chapter.title}`,
				value: `chapter:${chapter.id}`,
			});
			if (currentEntityType === "chapter" && currentEntityId === chapter.id) {
				option.selected = true;
			}
		}

		// Group scenes by chapter
		for (const chapter of this.chapters.sort((a, b) => a.number - b.number)) {
			const chapterScenes = this.scenes
				.filter(s => s.chapter_id === chapter.id)
				.sort((a, b) => a.order_num - b.order_num);
			
			if (chapterScenes.length > 0) {
				const optgroup = select.createEl("optgroup");
				optgroup.label = `Chapter ${chapter.number}: ${chapter.title} - Scenes`;
				for (const scene of chapterScenes) {
					const option = optgroup.createEl("option", {
						text: `Scene ${scene.order_num}: ${scene.goal || "Untitled"}`,
						value: `scene:${scene.id}`,
					});
					if (currentEntityType === "scene" && currentEntityId === scene.id) {
						option.selected = true;
					}
				}
			}
		}

		// Add orphan scenes
		const orphanScenes = this.scenes.filter(s => !s.chapter_id);
		if (orphanScenes.length > 0) {
			const optgroup = select.createEl("optgroup");
			optgroup.label = "No Chapter - Scenes";
			for (const scene of orphanScenes.sort((a, b) => a.order_num - b.order_num)) {
				const option = optgroup.createEl("option", {
					text: `Scene ${scene.order_num}: ${scene.goal || "Untitled"}`,
					value: `scene:${scene.id}`,
				});
				if (currentEntityType === "scene" && currentEntityId === scene.id) {
					option.selected = true;
				}
			}
		}

		// Group beats by scene
		for (const scene of this.scenes.sort((a, b) => a.order_num - b.order_num)) {
			const sceneBeats = this.beats
				.filter(b => b.scene_id === scene.id)
				.sort((a, b) => a.order_num - b.order_num);
			
			if (sceneBeats.length > 0) {
				const chapter = this.chapters.find(c => c.id === scene.chapter_id);
				const chapterLabel = chapter ? `Chapter ${chapter.number}` : "No Chapter";
				const optgroup = select.createEl("optgroup");
				optgroup.label = `${chapterLabel} > Scene ${scene.order_num} - Beats`;
				for (const beat of sceneBeats) {
					const option = optgroup.createEl("option", {
						text: `Beat ${beat.order_num}: ${beat.type}`,
						value: `beat:${beat.id}`,
					});
					if (currentEntityType === "beat" && currentEntityId === beat.id) {
						option.selected = true;
					}
				}
			}
		}

		if (!currentEntityType) {
			noRefOption.selected = true;
		}

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const actionButton = buttonContainer.createEl("button", {
			text: mode === "move" ? "Move" : "Link",
			cls: "mod-cta",
		});
		actionButton.onclick = async () => {
			const selectedValue = select.value;
			
			try {
				// Find current reference to delete (if move mode)
				if (mode === "move" && currentEntityType && currentEntityId) {
					const currentRef = this.contentBlockRefs.find(
						r => r.content_block_id === contentBlock.id && 
						r.entity_type === currentEntityType && 
						r.entity_id === currentEntityId
					);
					if (currentRef) {
						await this.plugin.apiClient.deleteContentBlockReference(currentRef.id);
					}
				}

				// Create new reference
				if (selectedValue) {
					const [entityType, entityId] = selectedValue.split(":");
					await this.plugin.apiClient.createContentBlockReference(contentBlock.id, entityType, entityId);
				}

				await this.loadHierarchy();
				this.renderTabContent();
				modal.close();
				new Notice(mode === "move" ? "Content block moved successfully" : "Content block linked successfully");
			} catch (err) {
				const errorMessage = err instanceof Error ? err.message : `Failed to ${mode} content block`;
				new Notice(`Error: ${errorMessage}`, 5000);
			}
		};

		const cancelButton = buttonContainer.createEl("button", { text: "Cancel" });
		cancelButton.onclick = () => modal.close();

		modal.open();
	}

	async createContentForEntity(entityType: "chapter" | "scene" | "beat", entityId: string, chapterId: string, contentType: "text" | "image" = "text") {
		if (!this.currentStory) return;
		
		if (!chapterId) {
			if (this.chapters.length === 0) {
				new Notice("No chapter available. Please create a chapter first.", 5000);
				return;
			}
			chapterId = this.chapters[0].id;
		}

		// Calculate next order_num for this chapter
		const chapterBlocks = this.contentBlocks.filter(cb => {
			const refs = this.contentBlockRefs.filter(r => r.content_block_id === cb.id && r.entity_type === "chapter" && r.entity_id === chapterId);
			return refs.length > 0 || cb.chapter_id === chapterId;
		});
		const maxOrderNum = chapterBlocks.length > 0 
			? Math.max(...chapterBlocks.map(cb => cb.order_num || 0))
			: 0;
		const nextOrderNum = maxOrderNum + 1;

		// Create initial content block with the specified type
		const initialContentBlock: Partial<ContentBlock> = {
			type: contentType,
			kind: "final",
			content: "",
			metadata: {},
		};

		new ContentBlockModal(this.app, async (contentBlock) => {
			try {
				contentBlock.order_num = nextOrderNum;
				const created = await this.plugin.apiClient.createContentBlock(chapterId, contentBlock);
				await this.plugin.apiClient.createContentBlockReference(created.id, entityType, entityId);
				await this.loadHierarchy();
				this.renderTabContent();
				new Notice("Content block created successfully");
			} catch (err) {
				throw err;
			}
		}, initialContentBlock as ContentBlock, this.plugin).open();
	}
}

