import { ItemView, WorkspaceLeaf, Notice, Modal, setIcon } from "obsidian";
import StoryEnginePlugin from "../main";
import { Story, Chapter, Scene, Beat, ContentBlock, ContentBlockReference, World, RPGSystem, Character, Location, Artifact, WorldEvent, Trait, Archetype, Faction, Lore, TimeConfig, CharacterTrait, CharacterRelationship, EventCharacter, SceneReference } from "../types";
import { ChapterModal } from "./modals/ChapterModal";
import { SceneModal } from "./modals/SceneModal";
import { BeatModal } from "./modals/BeatModal";
import { ContentBlockModal } from "./modals/ContentBlockModal";
import { CreateWorldModal } from "./CreateWorldModal";
import { WorldDetailsModal } from "./WorldDetailsModal";
import { TimelineModal } from "./TimelineModal";
import { CharacterDetailsView } from "./CharacterDetailsView";

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
	currentWorld: World | null = null;
	viewMode: "list" | "details" | "world-details" | "character-details" = "list";
	currentTab: "chapters" | "scenes" | "beats" | "contents" | "characters" = "chapters";
	storyCharacters: Character[] = [];
	worldTab: "characters" | "traits" | "archetypes" | "events" | "lore" | "locations" | "factions" | "artifacts" = "characters";
	listTab: "stories" | "worlds" | "rpg-systems" = "stories";
	expandedWorldId: string | null = null;
	// Character Details View
	characterDetailsView: CharacterDetailsView | null = null;
	chapters: Chapter[] = [];
	scenes: Scene[] = [];
	beats: Beat[] = [];
	contentBlocks: ContentBlock[] = [];
	contentBlockRefs: ContentBlockReference[] = [];
	loadingHierarchy: boolean = false;
	// World entities
	characters: Character[] = [];
	locations: Location[] = [];
	artifacts: Artifact[] = [];
	events: WorldEvent[] = [];
	traits: Trait[] = [];
	archetypes: Archetype[] = [];
	factions: Faction[] = [];
	lores: Lore[] = [];
	loadingWorldData: boolean = false;

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
		if (this.viewMode === "world-details" && this.currentWorld) {
			return this.currentWorld.name;
		}
		if (this.viewMode === "character-details" && this.characterDetailsView) {
			return this.characterDetailsView["character"].name;
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
		} else if (this.viewMode === "world-details" && this.currentWorld) {
			this.renderWorldDetails();
		} else if (this.viewMode === "character-details" && this.characterDetailsView) {
			this.characterDetailsView.render();
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

		// Settings button (gear icon) in header
		const settingsButton = tabsContainer.createEl("button", {
			cls: "story-engine-settings-btn story-engine-tab",
			attr: { "aria-label": "Open Settings" }
		});
		setIcon(settingsButton, "gear");
		settingsButton.onclick = () => {
			this.plugin.openSettings();
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
			createButtonAction = async () => {
				new CreateWorldModal(this.app, async (name: string, description: string, genre: string) => {
					try {
						new Notice(`Creating world "${name}"...`);
						const newWorld = await this.plugin.apiClient.createWorld(name, description, genre);
						new Notice(`World "${name}" created successfully`);
						// Refresh worlds list
						await this.loadStories();
					} catch (err) {
						const errorMessage = err instanceof Error ? err.message : "Failed to create world";
						new Notice(`Error: ${errorMessage}`, 5000);
					}
				}).open();
			};
		} else if (this.listTab === "rpg-systems") {
			createButtonText = "Create RPG System";
			createButtonAction = () => {
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
		if (!this.headerEl || !this.currentStory) return;

		this.headerEl.empty();
		
		const headerLeft = this.headerEl.createDiv({ cls: "story-engine-header-left" });
		
		const backButton = headerLeft.createEl("button", {
			text: "← Back",
			cls: "story-engine-back-btn",
		});
		backButton.onclick = () => {
			this.showList();
		};

		// Title with status pill and version
		const titleContainer = headerLeft.createDiv({ cls: "story-engine-title-container" });
		const titleRow = titleContainer.createDiv({ cls: "story-engine-title-row" });
		
		const titleH2 = titleRow.createEl("h2", { 
			text: this.currentStory.title,
			cls: "story-engine-title-header"
		});

		// Status pill
		const statusPill = titleRow.createSpan({ 
			cls: `story-engine-status-pill story-engine-status-${this.currentStory.status.toLowerCase().replace(/\s+/g, '-')}`
		});
		statusPill.textContent = this.currentStory.status;

		// Version
		const versionSpan = titleRow.createSpan({ cls: "story-engine-version" });
		versionSpan.textContent = `v.${this.currentStory.version_number}`;

		// UUID with copy icon
		const uuidRow = titleContainer.createDiv({ cls: "story-engine-uuid-row" });
		const uuidSpan = uuidRow.createSpan({ cls: "story-engine-uuid" });
		uuidSpan.textContent = this.currentStory.id;
		
		const copyUuidButton = uuidRow.createEl("button", {
			cls: "story-engine-copy-uuid-btn",
			attr: { "aria-label": "Copy UUID" }
		});
		setIcon(copyUuidButton, "copy");
		copyUuidButton.onclick = () => {
			this.copyStoryId();
		};

		const headerActions = this.headerEl.createDiv({ cls: "story-engine-header-actions" });
		
		// World button - only show if story has a world
		if (this.currentStory.world_id) {
			const world = this.worlds.find(w => w.id === this.currentStory!.world_id);
			if (world) {
				const worldButton = headerActions.createEl("button", {
					cls: "story-engine-world-btn",
					attr: { "aria-label": `Go to World: ${world.name}` }
				});
				setIcon(worldButton, "globe");
				worldButton.createSpan({ text: "World" });
				worldButton.onclick = () => {
					this.showWorldDetails(world);
				};
			}
		}
		
		// Context menu button with actions
		const contextButton = headerActions.createEl("button", {
			cls: "story-engine-context-btn",
			attr: { "aria-label": "Story Actions" }
		});
		setIcon(contextButton, "more-vertical");
		
		// Create dropdown menu
		const dropdownMenu = headerActions.createDiv({ cls: "story-engine-dropdown-menu" });
		dropdownMenu.style.display = "none";
		
		// Edit title option
		const editOption = dropdownMenu.createEl("button", {
			cls: "story-engine-dropdown-item",
		});
		setIcon(editOption, "pencil");
		editOption.createSpan({ text: "Edit Story Name" });
		editOption.onclick = () => {
			dropdownMenu.style.display = "none";
			this.showEditStoryNameModal();
		};
		
		// Clone option
		const cloneOption = dropdownMenu.createEl("button", {
			cls: "story-engine-dropdown-item",
		});
		setIcon(cloneOption, "copy");
		cloneOption.createSpan({ text: "Clone Story" });
		cloneOption.onclick = async () => {
			dropdownMenu.style.display = "none";
			await this.cloneStory();
		};
		
		// Pull option
		const pullOption = dropdownMenu.createEl("button", {
			cls: "story-engine-dropdown-item",
		});
		setIcon(pullOption, "download");
		pullOption.createSpan({ text: "Pull from Service" });
		pullOption.onclick = async () => {
			dropdownMenu.style.display = "none";
			if (!this.currentStory) return;
			try {
				new Notice(`Pulling story "${this.currentStory.title}"...`);
				await this.plugin.syncService.pullStory(this.currentStory.id);
				await this.loadHierarchy();
				this.renderTabContent();
				new Notice(`Story pulled successfully!`);
			} catch (err) {
				const errorMessage = err instanceof Error ? err.message : "Failed to pull story";
				new Notice(`Error: ${errorMessage}`, 5000);
			}
		};
		
		// Push option
		const pushOption = dropdownMenu.createEl("button", {
			cls: "story-engine-dropdown-item",
		});
		setIcon(pushOption, "upload");
		pushOption.createSpan({ text: "Push to Service" });
		pushOption.onclick = async () => {
			dropdownMenu.style.display = "none";
			if (!this.currentStory) return;
			try {
				const folderPath = this.plugin.fileManager.getStoryFolderPath(this.currentStory.title);
				new Notice(`Pushing story "${this.currentStory.title}"...`);
				await this.plugin.syncService.pushStory(folderPath);
				new Notice(`Story pushed successfully!`);
			} catch (err) {
				const errorMessage = err instanceof Error ? err.message : "Failed to push story";
				new Notice(`Error: ${errorMessage}`, 5000);
			}
		};
		
		// Toggle dropdown
		contextButton.onclick = (e) => {
			e.stopPropagation();
			const isVisible = dropdownMenu.style.display !== "none";
			dropdownMenu.style.display = isVisible ? "none" : "block";
		};
		
		// Close dropdown when clicking outside
		document.addEventListener("click", () => {
			dropdownMenu.style.display = "none";
		}, { once: true });
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
			
			// Click handler to show details - navigate to World View
			worldTitle.style.cursor = "pointer";
			worldTitle.onclick = () => {
				this.showWorldDetails(world);
			};
			
			// Accordion button
			const accordionButton = worldHeader.createEl("button", {
				text: worldStories.length > 0 ? `${worldStories.length} story${worldStories.length !== 1 ? 's' : ''}` : "No stories",
				cls: `story-engine-accordion-btn ${this.expandedWorldId === world.id ? "is-expanded" : ""}`,
			});
			accordionButton.onclick = (e) => {
				e.stopPropagation();
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
			if (this.plugin.settings.mode === "local") {
				this.contentEl.createEl("p", { text: "RPG systems are not available in local mode." });
			} else {
				this.contentEl.createEl("p", { text: "No RPG systems found." });
			}
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
			// Only validate tenant ID in remote mode
			if (this.plugin.settings.mode === "remote" && !this.plugin.settings.tenantId) {
				this.error = "Tenant ID not configured";
				this.loading = false;
				this.renderListContent();
				return;
			}

			// Load worlds first to group stories
			this.worlds = await this.plugin.apiClient.getWorlds();
			// Load stories
			this.stories = await this.plugin.apiClient.listStories();
			// Load RPG systems (optional - may not be available in local mode)
			try {
				this.rpgSystems = await this.plugin.apiClient.getRPGSystems();
			} catch (rpgErr) {
				// RPG systems not available (e.g., in local/offline mode) - not a fatal error
				console.warn("RPG systems not available:", rpgErr);
				this.rpgSystems = [];
			}
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

		const charactersTab = tabsContainer.createEl("button", {
			text: "Characters",
			cls: `story-engine-tab ${this.currentTab === "characters" ? "is-active" : ""}`,
		});
		charactersTab.onclick = async () => {
			this.currentTab = "characters";
			this.renderTabs();
			await this.loadStoryCharacters();
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
			case "characters":
				this.renderStoryCharactersTab(tabContent);
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
				}, this.scenes, undefined, chapter.id).open();
			};
		}

		// Render orphan scenes (no chapter)
		const orphanScenes = scenesByChapter.get(null) || [];
		if (orphanScenes.length > 0 || scenesByChapter.size === 0) {
			const group = list.createDiv({ cls: "story-engine-group" });
			const groupHeader = group.createDiv({ cls: "story-engine-group-header" });
			groupHeader.createEl("h3", { text: "Without Chapter" });

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
				}, this.scenes, undefined, null).open();
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
						}, this.beats, undefined, this.chapters, scene.id).open();
					};
				}
			}
		}

		// Render orphan beats (no scene or scene without chapter)
		const orphanBeats = this.beats.filter(b => {
			const scene = this.scenes.find(s => s.id === b.scene_id);
			return !scene || !scene.chapter_id;
		});
		
		// Show "Without Chapter" section if no chapters exist OR if there are orphan beats/scenes
		if (this.chapters.length === 0 || orphanBeats.length > 0 || this.scenes.some(s => !s.chapter_id)) {
			const orphanGroup = list.createDiv({ cls: "story-engine-chapter-group" });
			const orphanHeader = orphanGroup.createDiv({ cls: "story-engine-chapter-group-header" });
			orphanHeader.createEl("h2", { text: "Without Chapter" });

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
					}, this.beats, undefined, this.chapters, scene.id).open();
				};
			}

			// Beats without valid scene
			const beatsWithoutScene = orphanBeats.filter(b => {
				const scene = this.scenes.find(s => s.id === b.scene_id);
				return !scene;
			});
			
			// Show "Without Scene" section if there are beats without scene OR if there are no scenes but there are beats
			if (beatsWithoutScene.length > 0 || (orphanScenes.length === 0 && this.beats.length > 0 && this.scenes.length === 0)) {
				const sceneGroup = orphanContent.createDiv({ cls: "story-engine-group" });
				const sceneHeader = sceneGroup.createDiv({ cls: "story-engine-group-header" });
				sceneHeader.createEl("h3", { text: "Without Scene" });
				
				const sceneItems = sceneGroup.createDiv({ cls: "story-engine-group-items" });
				if (beatsWithoutScene.length > 0) {
					for (const beat of beatsWithoutScene.sort((a, b) => a.order_num - b.order_num)) {
						this.renderBeatItem(sceneItems, beat);
					}
				} else {
					sceneItems.createEl("p", { text: "No beats without scene." });
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
			}, this.beats, beat, this.chapters).open();
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

	showEditStoryNameModal() {
		if (!this.currentStory) return;
		
		const modal = new Modal(this.app);
		modal.titleEl.setText("Edit Story Name");
		
		const content = modal.contentEl;
		let title = this.currentStory.title;

		content.createEl("label", { text: "Story Name *" });
		const titleInput = content.createEl("input", { type: "text", cls: "story-engine-input", value: title });
		titleInput.oninput = () => { title = titleInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Save", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			if (!title.trim()) {
				new Notice("Story name is required", 3000);
				return;
			}
			try {
				const updatedStory = await this.plugin.apiClient.updateStory(this.currentStory!.id, title.trim());
				this.currentStory = updatedStory;
				await this.loadStories();
				this.renderDetailsHeader();
				modal.close();
				new Notice("Story name updated");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();

		modal.open();
		titleInput.focus();
		titleInput.select();
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

	showEditWorldModal() {
		if (!this.currentWorld) return;
		
		const modal = new Modal(this.app);
		modal.titleEl.setText("Edit World");
		
		const content = modal.contentEl;
		let name = this.currentWorld.name;
		let description = this.currentWorld.description;
		let genre = this.currentWorld.genre;

		content.createEl("label", { text: "World Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input", value: name });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Genre *" });
		const genreInput = content.createEl("input", { type: "text", cls: "story-engine-input", value: genre });
		genreInput.oninput = () => { genre = genreInput.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.value = description;
		descInput.oninput = () => { description = descInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Save", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("World name is required", 3000);
				return;
			}
			if (!genre.trim()) {
				new Notice("Genre is required", 3000);
				return;
			}
			try {
				const updatedWorld = await this.plugin.apiClient.updateWorld(
					this.currentWorld!.id,
					name.trim(),
					description.trim(),
					genre.trim()
				);
				this.currentWorld = updatedWorld;
				// Refresh world list to update it everywhere
				await this.loadStories();
				// Re-render the world details view
				this.renderWorldDetails();
				modal.close();
				new Notice("World updated");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();

		modal.open();
		nameInput.focus();
		nameInput.select();
	}

	copyStoryId() {
		if (!this.currentStory) return;

		navigator.clipboard.writeText(this.currentStory.id).then(() => {
			new Notice("UUID copied to clipboard");
		}).catch(() => {
			// Fallback for older browsers
			const textarea = document.createElement("textarea");
			textarea.value = this.currentStory!.id;
			textarea.style.position = "fixed";
			textarea.style.opacity = "0";
			textarea.style.left = "-9999px";
			document.body.appendChild(textarea);
			textarea.select();
			
			try {
				document.execCommand("copy");
				new Notice("UUID copied to clipboard");
			} catch (err) {
				new Notice("Failed to copy UUID", 3000);
			}
			
			document.body.removeChild(textarea);
		});
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
		// Show "Without Chapter" section if no chapters exist OR if there are orphan contents/scenes
		if (this.chapters.length === 0 || orphanContents.length > 0 || this.scenes.some(s => !s.chapter_id)) {
			const orphanGroup = list.createDiv({ cls: "story-engine-chapter-group" });
			const orphanHeader = orphanGroup.createDiv({ cls: "story-engine-chapter-group-header" });
			orphanHeader.createEl("h2", { text: "Without Chapter" });

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
				orphanContentsHeader.createEl("h3", { text: "Without Reference" });

				const orphanContentsItems = orphanContentsGroup.createDiv({ cls: "story-engine-group-items" });
				for (const block of orphanContents) {
					this.renderContentItem(orphanContentsItems, block, null, null);
				}
			}
		}
	}

	async loadStoryCharacters() {
		if (!this.currentStory || !this.currentStory.world_id) {
			this.storyCharacters = [];
			return;
		}

		try {
			this.storyCharacters = await this.plugin.apiClient.getCharacters(this.currentStory.world_id);
		} catch (err) {
			console.error("Error loading story characters:", err);
			this.storyCharacters = [];
			new Notice(`Error loading characters: ${err instanceof Error ? err.message : "Failed"}`, 5000);
		}
	}

	renderStoryCharactersTab(container: HTMLElement) {
		container.empty();

		if (!this.currentStory?.world_id) {
			const noWorldMsg = container.createDiv({ cls: "story-engine-empty-state" });
			noWorldMsg.createEl("p", { text: "This story is not linked to a world." });
			noWorldMsg.createEl("p", { text: "Link this story to a world to see its characters.", cls: "story-engine-hint" });
			return;
		}

		const list = container.createDiv({ cls: "story-engine-list" });

		if (this.storyCharacters.length === 0) {
			list.createEl("p", { text: "No characters found in this story's world." });
		} else {
			for (const character of this.storyCharacters.sort((a, b) => a.name.localeCompare(b.name))) {
				const item = list.createDiv({ cls: "story-engine-item" });
				
				const titleDiv = item.createDiv({ cls: "story-engine-title", text: character.name });
				titleDiv.style.cursor = "pointer";
				titleDiv.onclick = () => {
					this.showCharacterDetails(character);
				};
				
				const meta = item.createDiv({ cls: "story-engine-meta" });
				if (character.description) {
					meta.createEl("span", { 
						text: character.description.substring(0, 80) + (character.description.length > 80 ? "..." : "") 
					});
				}

				const actions = item.createDiv({ cls: "story-engine-item-actions" });
				
				// View details button
				const viewBtn = actions.createEl("button", { text: "View" });
				viewBtn.onclick = () => {
					this.showCharacterDetails(character);
				};
			}
		}

		// Footer with Create Character button
		const footer = container.createDiv({ cls: "story-engine-list-footer" });
		const createButton = footer.createEl("button", {
			text: "Create Character",
			cls: "mod-cta",
		});
		createButton.onclick = () => {
			this.showCreateCharacterModalForStory();
		};
	}

	showCreateCharacterModalForStory() {
		if (!this.currentStory?.world_id) {
			new Notice("This story is not linked to a world");
			return;
		}
		
		const modal = new Modal(this.app);
		modal.titleEl.setText("Create Character");
		
		const content = modal.contentEl;
		let name = "";
		let description = "";

		content.createEl("label", { text: "Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input" });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.oninput = () => { description = descInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const createBtn = buttonContainer.createEl("button", { text: "Create", cls: "mod-cta" });
		createBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.createCharacter(this.currentStory!.world_id!, { 
					name: name.trim(), 
					description: description.trim()
				});
				await this.loadStoryCharacters();
				this.renderTabContent();
				modal.close();
				new Notice("Character created");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();

		modal.open();
		nameInput.focus();
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

	// ==================== World View Methods ====================

	async showWorldDetails(world: World) {
		this.currentWorld = world;
		this.viewMode = "world-details";
		this.worldTab = "characters";
		await this.loadWorldData();
		this.renderWorldDetails();
	}

	async loadWorldData() {
		if (!this.currentWorld) return;

		this.loadingWorldData = true;
		try {
			this.characters = await this.plugin.apiClient.getCharacters(this.currentWorld.id);
			this.locations = await this.plugin.apiClient.getLocations(this.currentWorld.id);
			this.artifacts = await this.plugin.apiClient.getArtifacts(this.currentWorld.id);
			this.events = await this.plugin.apiClient.getEvents(this.currentWorld.id);
			// Traits are global, not per-world
			try {
				this.traits = await this.plugin.apiClient.getTraits();
			} catch (err) {
				console.warn("Traits not available:", err);
				this.traits = [];
			}
			// Archetypes are global per tenant
			try {
				this.archetypes = await this.plugin.apiClient.getArchetypes();
			} catch (err) {
				console.warn("Archetypes not available:", err);
				this.archetypes = [];
			}
			// Factions and Lore are per-world
			try {
				this.factions = await this.plugin.apiClient.getFactions(this.currentWorld.id);
			} catch (err) {
				console.warn("Factions not available:", err);
				this.factions = [];
			}
			try {
				this.lores = await this.plugin.apiClient.getLores(this.currentWorld.id);
			} catch (err) {
				console.warn("Lores not available:", err);
				this.lores = [];
			}
			// Reload world to get time_config
			try {
				this.currentWorld = await this.plugin.apiClient.getWorld(this.currentWorld.id);
			} catch (err) {
				console.warn("Failed to reload world:", err);
			}
		} catch (err) {
			const errorMessage = err instanceof Error ? err.message : "Failed to load world data";
			new Notice(`Error: ${errorMessage}`, 5000);
		} finally {
			this.loadingWorldData = false;
		}
	}

	renderWorldDetails() {
		if (!this.contentEl || !this.currentWorld) return;

		this.renderWorldDetailsHeader();
		this.contentEl.empty();

		// Tabs section
		this.renderWorldTabs();
		this.renderWorldTabContent();
	}

	renderWorldDetailsHeader() {
		if (!this.headerEl || !this.currentWorld) return;

		this.headerEl.empty();
		
		const headerLeft = this.headerEl.createDiv({ cls: "story-engine-header-left" });
		
		const backButton = headerLeft.createEl("button", {
			text: "← Back",
			cls: "story-engine-back-btn",
		});
		backButton.onclick = () => {
			this.currentWorld = null;
			this.viewMode = "list";
			this.listTab = "worlds";
			this.renderListHeader();
			this.renderListContent();
		};

		// Title container
		const titleContainer = headerLeft.createDiv({ cls: "story-engine-title-container" });
		const titleRow = titleContainer.createDiv({ cls: "story-engine-title-row" });
		
		titleRow.createEl("h2", { 
			text: this.currentWorld.name,
			cls: "story-engine-title-header"
		});

		// Genre pill
		if (this.currentWorld.genre) {
			const genrePill = titleRow.createSpan({ 
				cls: "story-engine-status-pill story-engine-status-draft"
			});
			genrePill.textContent = this.currentWorld.genre;
		}

		// UUID with copy icon
		const uuidRow = titleContainer.createDiv({ cls: "story-engine-uuid-row" });
		const uuidSpan = uuidRow.createSpan({ cls: "story-engine-uuid" });
		uuidSpan.textContent = this.currentWorld.id;
		
		const copyUuidButton = uuidRow.createEl("button", {
			cls: "story-engine-copy-uuid-btn",
			attr: { "aria-label": "Copy UUID" }
		});
		setIcon(copyUuidButton, "copy");
		copyUuidButton.onclick = () => {
			if (!this.currentWorld) return;
			navigator.clipboard.writeText(this.currentWorld.id).then(() => {
				new Notice("UUID copied to clipboard");
			}).catch(() => {
				new Notice("Failed to copy UUID", 3000);
			});
		};

		// Description if present
		if (this.currentWorld.description) {
			const descRow = titleContainer.createDiv({ cls: "story-engine-world-desc" });
			descRow.textContent = this.currentWorld.description;
		}

		const headerActions = this.headerEl.createDiv({ cls: "story-engine-header-actions" });
		
		// Context menu button with actions
		const contextButton = headerActions.createEl("button", {
			cls: "story-engine-context-btn",
			attr: { "aria-label": "World Actions" }
		});
		setIcon(contextButton, "more-vertical");
		
		// Create dropdown menu
		const dropdownMenu = headerActions.createDiv({ cls: "story-engine-dropdown-menu" });
		dropdownMenu.style.display = "none";
		
		// Edit option
		const editOption = dropdownMenu.createEl("button", {
			cls: "story-engine-dropdown-item",
		});
		setIcon(editOption, "pencil");
		editOption.createSpan({ text: "Edit World" });
		editOption.onclick = () => {
			dropdownMenu.style.display = "none";
			this.showEditWorldModal();
		};
		
		// Toggle dropdown
		contextButton.onclick = (e) => {
			e.stopPropagation();
			const isVisible = dropdownMenu.style.display !== "none";
			dropdownMenu.style.display = isVisible ? "none" : "block";
		};
		
		// Close dropdown when clicking outside
		document.addEventListener("click", () => {
			dropdownMenu.style.display = "none";
		}, { once: true });
	}

	renderWorldTabs() {
		if (!this.contentEl) return;

		// Remove existing tabs if any
		const existingTabs = this.contentEl.querySelector(".story-engine-tabs-container");
		if (existingTabs) {
			existingTabs.remove();
		}

		const tabsWrapper = this.contentEl.createDiv({ cls: "story-engine-tabs-container" });

		// First row: Entity tabs (Characters, Locations, Factions, Artifacts)
		const entityTabsContainer = tabsWrapper.createDiv({ cls: "story-engine-tabs" });
		const entityTabs: { key: "characters" | "traits" | "archetypes" | "events" | "lore" | "locations" | "factions" | "artifacts"; label: string }[] = [
			{ key: "characters", label: "Characters" },
			{ key: "locations", label: "Locations" },
			{ key: "factions", label: "Factions" },
			{ key: "artifacts", label: "Artifacts" },
		];

		for (const tab of entityTabs) {
			const tabButton = entityTabsContainer.createEl("button", {
				text: tab.label,
				cls: `story-engine-tab ${this.worldTab === tab.key ? "is-active" : ""}`,
			});
			tabButton.onclick = () => {
				this.worldTab = tab.key;
				this.renderWorldTabs();
				this.renderWorldTabContent();
			};
		}

		// Second row: Meta tabs (Traits, Archetypes, Events, Lore)
		const metaTabsContainer = tabsWrapper.createDiv({ cls: "story-engine-tabs" });
		const metaTabs: { key: "characters" | "traits" | "archetypes" | "events" | "lore" | "locations" | "factions" | "artifacts"; label: string }[] = [
			{ key: "traits", label: "Traits" },
			{ key: "archetypes", label: "Archetypes" },
			{ key: "events", label: "Events" },
			{ key: "lore", label: "Lore" },
		];

		for (const tab of metaTabs) {
			const tabButton = metaTabsContainer.createEl("button", {
				text: tab.label,
				cls: `story-engine-tab ${this.worldTab === tab.key ? "is-active" : ""}`,
			});
			tabButton.onclick = () => {
				this.worldTab = tab.key;
				this.renderWorldTabs();
				this.renderWorldTabContent();
			};
		}
	}

	renderWorldTabContent() {
		if (!this.contentEl) return;

		// Remove existing content
		const existingContent = this.contentEl.querySelector(".story-engine-tab-content");
		if (existingContent) {
			existingContent.remove();
		}

		const contentContainer = this.contentEl.createDiv({ cls: "story-engine-tab-content" });

		if (this.loadingWorldData) {
			contentContainer.createEl("p", { text: "Loading..." });
			return;
		}

		switch (this.worldTab) {
			case "characters":
				this.renderCharactersTab(contentContainer);
				break;
			case "traits":
				this.renderTraitsTab(contentContainer);
				break;
			case "archetypes":
				this.renderArchetypesTab(contentContainer);
				break;
			case "events":
				this.renderEventsTab(contentContainer);
				break;
			case "lore":
				this.renderLoreTab(contentContainer);
				break;
			case "locations":
				this.renderLocationsTab(contentContainer);
				break;
			case "factions":
				this.renderFactionsTab(contentContainer);
				break;
			case "artifacts":
				this.renderArtifactsTab(contentContainer);
				break;
		}

		// Actions bar
		this.renderWorldActionsBar(contentContainer);
	}

	renderWorldActionsBar(container: HTMLElement) {
		const actionsBar = container.createDiv({ cls: "story-engine-actions-bar" });
		
		let createButtonText = "Create Character";
		let createButtonAction = () => {
			this.showCreateCharacterModal();
		};

		switch (this.worldTab) {
			case "traits":
				createButtonText = "Create Trait";
				createButtonAction = () => this.showCreateTraitModal();
				break;
			case "archetypes":
				createButtonText = "Create Archetype";
				createButtonAction = () => this.showCreateArchetypeModal();
				break;
			case "events":
				createButtonText = "Create Event";
				createButtonAction = () => this.showCreateEventModal();
				break;
			case "lore":
				createButtonText = "Create Lore";
				createButtonAction = () => this.showCreateLoreModal();
				break;
			case "locations":
				createButtonText = "Create Location";
				createButtonAction = () => this.showCreateLocationModal();
				break;
			case "factions":
				createButtonText = "Create Faction";
				createButtonAction = () => this.showCreateFactionModal();
				break;
			case "artifacts":
				createButtonText = "Create Artifact";
				createButtonAction = () => this.showCreateArtifactModal();
				break;
		}

		const createButton = actionsBar.createEl("button", {
			text: createButtonText,
			cls: "mod-cta story-engine-create-btn",
		});
		createButton.onclick = createButtonAction;
	}

	renderCharactersTab(container: HTMLElement) {
		if (this.characters.length === 0) {
			container.createEl("p", { text: "No characters found. Create your first character!" });
			return;
		}

		const list = container.createDiv({ cls: "story-engine-list" });
		for (const character of this.characters) {
			const item = list.createDiv({ cls: "story-engine-item" });
			const titleDiv = item.createDiv({ cls: "story-engine-title", text: character.name });
			titleDiv.style.cursor = "pointer";
			titleDiv.onclick = () => {
				this.showCharacterDetails(character);
			};
			
			const meta = item.createDiv({ cls: "story-engine-meta" });
			if (character.description) {
				meta.createEl("span", { text: character.description.substring(0, 50) + (character.description.length > 50 ? "..." : "") });
			}

			const actions = item.createDiv({ cls: "story-engine-item-actions" });
			// Botão relacionamento
			const relBtn = actions.createEl("button");
			setIcon(relBtn, "users");
			relBtn.title = "Add Relationship";
			relBtn.onclick = () => this.showAddCharacterRelationshipModal(character);
			actions.createEl("button", { text: "Edit" }).onclick = () => {
				this.showEditCharacterModal(character);
			};
			actions.createEl("button", { text: "Delete" }).onclick = async () => {
				if (confirm(`Delete character "${character.name}"?`)) {
					try {
						await this.plugin.apiClient.deleteCharacter(character.id);
						await this.loadWorldData();
						this.renderWorldTabContent();
						new Notice("Character deleted");
					} catch (err) {
						new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
					}
				}
			};
		}
	}

	renderLocationsTab(container: HTMLElement) {
		if (this.locations.length === 0) {
			container.createEl("p", { text: "No locations found. Create your first location!" });
			return;
		}

		// Group by hierarchy level
		const rootLocations = this.locations.filter(l => !l.parent_id);
		const list = container.createDiv({ cls: "story-engine-list" });
		
		for (const location of rootLocations.sort((a, b) => a.name.localeCompare(b.name))) {
			this.renderLocationItem(list, location, 0);
		}
	}

	renderLocationItem(container: HTMLElement, location: Location, level: number) {
		const item = container.createDiv({ cls: "story-engine-item" });
		item.style.marginLeft = `${level * 1}rem`;
		
		const titleRow = item.createDiv({ cls: "story-engine-title" });
		titleRow.textContent = location.name;
		if (location.type) {
			const typeBadge = titleRow.createSpan({ cls: "story-engine-badge" });
			typeBadge.textContent = location.type;
		}
		
		const meta = item.createDiv({ cls: "story-engine-meta" });
		if (location.description) {
			meta.createEl("span", { text: location.description.substring(0, 50) + (location.description.length > 50 ? "..." : "") });
		}

		const actions = item.createDiv({ cls: "story-engine-item-actions" });
		actions.createEl("button", { text: "Edit" }).onclick = () => {
			this.showEditLocationModal(location);
		};
		actions.createEl("button", { text: "Delete" }).onclick = async () => {
			if (confirm(`Delete location "${location.name}"?`)) {
				try {
					await this.plugin.apiClient.deleteLocation(location.id);
					await this.loadWorldData();
					this.renderWorldTabContent();
					new Notice("Location deleted");
				} catch (err) {
					new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
				}
			}
		};

		// Render children
		const children = this.locations.filter(l => l.parent_id === location.id);
		for (const child of children.sort((a, b) => a.name.localeCompare(b.name))) {
			this.renderLocationItem(container, child, level + 1);
		}
	}

	renderArtifactsTab(container: HTMLElement) {
		if (this.artifacts.length === 0) {
			container.createEl("p", { text: "No artifacts found. Create your first artifact!" });
			return;
		}

		const list = container.createDiv({ cls: "story-engine-list" });
		for (const artifact of this.artifacts) {
			const item = list.createDiv({ cls: "story-engine-item" });
			
			const titleRow = item.createDiv({ cls: "story-engine-title" });
			titleRow.textContent = artifact.name;
			if (artifact.rarity) {
				const rarityBadge = titleRow.createSpan({ cls: `story-engine-badge story-engine-rarity-${artifact.rarity.toLowerCase()}` });
				rarityBadge.textContent = artifact.rarity;
			}
			
			const meta = item.createDiv({ cls: "story-engine-meta" });
			if (artifact.description) {
				meta.createEl("span", { text: artifact.description.substring(0, 50) + (artifact.description.length > 50 ? "..." : "") });
			}

			const actions = item.createDiv({ cls: "story-engine-item-actions" });
			actions.createEl("button", { text: "Edit" }).onclick = () => {
				this.showEditArtifactModal(artifact);
			};
			actions.createEl("button", { text: "Delete" }).onclick = async () => {
				if (confirm(`Delete artifact "${artifact.name}"?`)) {
					try {
						await this.plugin.apiClient.deleteArtifact(artifact.id);
						await this.loadWorldData();
						this.renderWorldTabContent();
						new Notice("Artifact deleted");
					} catch (err) {
						new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
					}
				}
			};
		}
	}

	renderEventsTab(container: HTMLElement) {
		// === SECAO 1: TIME CONFIG ===
		const timeConfigSection = container.createDiv({ cls: "story-engine-time-config-section" });
		
		if (this.currentWorld?.time_config) {
			// Mostrar configuracao atual
			this.renderTimeConfigDisplay(timeConfigSection, this.currentWorld.time_config);
			
			// Botao Edit
			const editBtn = timeConfigSection.createEl("button", { 
				text: "Edit Time Config",
				cls: "story-engine-btn-secondary"
			});
			editBtn.onclick = () => this.showTimeConfigModal(this.currentWorld!.time_config!);
		} else {
			// Mostrar botao para adicionar
			const addBtn = timeConfigSection.createEl("button", { 
				text: "Add Time Configuration",
				cls: "story-engine-btn-primary"
			});
			addBtn.onclick = () => this.showTimeConfigModal(null);
		}
		
		// === SECAO 2: BOTAO VER TIMELINE ===
		const timelineSection = container.createDiv({ cls: "story-engine-timeline-section" });
		const timelineBtn = timelineSection.createEl("button", {
			text: "View Timeline",
			cls: "story-engine-btn-secondary"
		});
		timelineBtn.onclick = () => this.showTimelineModal();
		
		// === SECAO 3: EPOCH EVENT (Year Zero) ===
		const epochSection = container.createDiv({ cls: "story-engine-epoch-section" });
		epochSection.createEl("h4", { text: "⏰ Epoch Event (Year Zero)", cls: "story-engine-section-title" });
		
		const epochEvent = this.events.find(e => e.is_epoch && e.timeline_position === 0);
		
		if (epochEvent) {
			// Epoch event exists - show it with edit option
			const epochItem = epochSection.createDiv({ cls: "story-engine-item story-engine-epoch-item" });
			
			const epochTitle = epochItem.createDiv({ cls: "story-engine-title" });
			epochTitle.createSpan({ text: epochEvent.name });
			epochTitle.createSpan({ cls: "story-engine-badge story-engine-badge-epoch", text: "EPOCH" });
			
			const epochMeta = epochItem.createDiv({ cls: "story-engine-meta" });
			epochMeta.createEl("span", { text: "Position: 0 (Year Zero)" });
			if (epochEvent.description) {
				epochMeta.createEl("span", { text: epochEvent.description.substring(0, 80) + (epochEvent.description.length > 80 ? "..." : "") });
			}
			
			const epochActions = epochItem.createDiv({ cls: "story-engine-item-actions" });
			epochActions.createEl("button", { text: "Edit Epoch" }).onclick = () => {
				this.showEditEventModal(epochEvent);
			};
		} else {
			// No epoch event - show create button
			const createEpochBtn = epochSection.createEl("button", { 
				text: "🌟 Create Epoch Event (Year Zero)",
				cls: "mod-cta story-engine-create-epoch-btn"
			});
			createEpochBtn.onclick = () => this.showCreateEpochEventModal();
			
			epochSection.createEl("p", { 
				text: "The Epoch Event marks Year Zero - all other events are dated relative to this point.",
				cls: "story-engine-hint"
			});
		}
		
		// === SECAO 4: LISTA DE EVENTOS ===
		const eventsSection = container.createDiv({ cls: "story-engine-events-list-section" });
		eventsSection.createEl("h4", { text: "📜 Events", cls: "story-engine-section-title" });
		
		// Filter out the epoch event from the list
		const regularEvents = this.events.filter(e => !(e.is_epoch && e.timeline_position === 0));
		
		if (regularEvents.length === 0) {
			eventsSection.createEl("p", { text: "No events yet. Create your first event!", cls: "story-engine-empty-hint" });
			return;
		}

		const list = eventsSection.createDiv({ cls: "story-engine-list" });
		
		// Sort by timeline_position (events closer to epoch first), then by importance
		const sortedEvents = regularEvents.sort((a, b) => {
			const posA = a.timeline_position ?? Number.MAX_SAFE_INTEGER;
			const posB = b.timeline_position ?? Number.MAX_SAFE_INTEGER;
			if (posA !== posB) return posA - posB;
			return b.importance - a.importance;
		});
		
		for (const event of sortedEvents) {
			const item = list.createDiv({ cls: "story-engine-item" });
			
			// Title row with name and badges
			const titleRow = item.createDiv({ cls: "story-engine-title" });
			titleRow.createSpan({ text: event.name });
			
			// Importance badge
			const importanceBadge = titleRow.createSpan({ cls: "story-engine-badge" });
			importanceBadge.textContent = `★${event.importance}`;
			
			// Type badge
			if (event.type) {
				titleRow.createSpan({ cls: "story-engine-badge", text: event.type });
			}
			
			// Epoch badge if is_epoch (but not year zero)
			if (event.is_epoch) {
				titleRow.createSpan({ cls: "story-engine-badge story-engine-badge-epoch", text: "EPOCH" });
			}
			
			// Meta info - clearer structure
			const meta = item.createDiv({ cls: "story-engine-meta" });
			
			// Timeline Position (Date since Year Zero)
			if (event.timeline_position !== undefined && event.timeline_position !== null) {
				const posText = event.timeline_position >= 0 
					? `Year ${event.timeline_position}` 
					: `${Math.abs(event.timeline_position)} years before Year Zero`;
				meta.createEl("span", { text: `📅 ${posText}`, cls: "story-engine-event-position" });
			}
			
			// Parent Event relationship
			if (event.parent_id) {
				const parentEvent = this.events.find(e => e.id === event.parent_id);
				if (parentEvent) {
					meta.createEl("span", { text: `↳ Related to: ${parentEvent.name}`, cls: "story-engine-event-parent" });
				}
			}
			
			// Description preview
			if (event.description) {
				meta.createEl("span", { 
					text: event.description.substring(0, 60) + (event.description.length > 60 ? "..." : ""),
					cls: "story-engine-event-description"
				});
			}

			// Actions
			const actions = item.createDiv({ cls: "story-engine-item-actions" });
			
			// Link to Entity
			const linkBtn = actions.createEl("button");
			setIcon(linkBtn, "link");
			linkBtn.title = "Link to Entity";
			linkBtn.onclick = () => this.showLinkEventToEntityModal(event);
			
			// Set Parent Event
			const eventLinkBtn = actions.createEl("button");
			setIcon(eventLinkBtn, "git-branch");
			eventLinkBtn.title = "Set Parent Event";
			eventLinkBtn.onclick = () => this.showSetEventParentModal(event);
			
			actions.createEl("button", { text: "Edit" }).onclick = () => {
				this.showEditEventModal(event);
			};
			actions.createEl("button", { text: "Delete" }).onclick = async () => {
				if (confirm(`Delete event "${event.name}"?`)) {
					try {
						await this.plugin.apiClient.deleteEvent(event.id);
						await this.loadWorldData();
						this.renderWorldTabContent();
						new Notice("Event deleted");
					} catch (err) {
						new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
					}
				}
			};
		}
	}

	showCreateEpochEventModal() {
		if (!this.currentWorld) return;
		const modal = new Modal(this.app);
		modal.titleEl.setText("Create Epoch Event (Year Zero)");
		
		const content = modal.contentEl;
		let name = "";
		let description = "";

		content.createEl("p", { 
			text: "The Epoch Event defines Year Zero in your world's timeline. All other events will be dated relative to this moment.",
			cls: "story-engine-modal-hint"
		});

		content.createEl("label", { text: "Event Name *" });
		const nameInput = content.createEl("input", { 
			type: "text", 
			cls: "story-engine-input",
			placeholder: "e.g., The Great Cataclysm, The Founding, Year of the Dragon"
		});
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { 
			cls: "story-engine-textarea",
			placeholder: "Describe what happened at this pivotal moment..."
		});
		descInput.oninput = () => { description = descInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const createBtn = buttonContainer.createEl("button", { text: "Create Epoch", cls: "mod-cta" });
		createBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required");
				return;
			}
			try {
				await this.plugin.apiClient.createEvent(this.currentWorld!.id, {
					name: name.trim(),
					description: description.trim() || undefined,
					timeline_position: 0,
					is_epoch: true,
					importance: 10, // Max importance for epoch
					type: "Epoch"
				});
				modal.close();
				await this.loadWorldData();
				this.renderWorldTabContent();
				new Notice("Epoch event created!");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		
		const cancelBtn = buttonContainer.createEl("button", { text: "Cancel" });
		cancelBtn.onclick = () => modal.close();

		modal.open();
	}

	renderTraitsTab(container: HTMLElement) {
		if (this.traits.length === 0) {
			container.createEl("p", { text: "No traits found. Create your first trait!" });
			return;
		}

		// Group by category
		const traitsByCategory = new Map<string, Trait[]>();
		for (const trait of this.traits) {
			const category = trait.category || "Uncategorized";
			if (!traitsByCategory.has(category)) {
				traitsByCategory.set(category, []);
			}
			traitsByCategory.get(category)!.push(trait);
		}

		const list = container.createDiv({ cls: "story-engine-list" });
		for (const [category, categoryTraits] of traitsByCategory.entries()) {
			const group = list.createDiv({ cls: "story-engine-group" });
			const groupHeader = group.createDiv({ cls: "story-engine-group-header" });
			groupHeader.createEl("h3", { text: category });

			const groupItems = group.createDiv({ cls: "story-engine-group-items" });
			for (const trait of categoryTraits.sort((a, b) => a.name.localeCompare(b.name))) {
				const item = groupItems.createDiv({ cls: "story-engine-item" });
				item.createDiv({ cls: "story-engine-title", text: trait.name });
				
				const meta = item.createDiv({ cls: "story-engine-meta" });
				if (trait.description) {
					meta.createEl("span", { text: trait.description.substring(0, 50) + (trait.description.length > 50 ? "..." : "") });
				}

				const actions = item.createDiv({ cls: "story-engine-item-actions" });
				actions.createEl("button", { text: "Edit" }).onclick = () => {
					this.showEditTraitModal(trait);
				};
				actions.createEl("button", { text: "Delete" }).onclick = async () => {
					if (confirm(`Delete trait "${trait.name}"?`)) {
						try {
							await this.plugin.apiClient.deleteTrait(trait.id);
							await this.loadWorldData();
							this.renderWorldTabContent();
							new Notice("Trait deleted");
						} catch (err) {
							new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
						}
					}
				};
			}
		}
	}

	renderArchetypesTab(container: HTMLElement) {
		if (this.archetypes.length === 0) {
			container.createEl("p", { text: "No archetypes found. Create your first archetype!" });
			return;
		}

		const list = container.createDiv({ cls: "story-engine-list" });
		for (const archetype of this.archetypes.sort((a, b) => a.name.localeCompare(b.name))) {
			const item = list.createDiv({ cls: "story-engine-item" });
			item.createDiv({ cls: "story-engine-title", text: archetype.name });
			
			const meta = item.createDiv({ cls: "story-engine-meta" });
			if (archetype.description) {
				meta.createEl("span", { text: archetype.description.substring(0, 50) + (archetype.description.length > 50 ? "..." : "") });
			}

			const actions = item.createDiv({ cls: "story-engine-item-actions" });
			actions.createEl("button", { text: "View Traits" }).onclick = async () => {
				try {
					const traits = await this.plugin.apiClient.getArchetypeTraits(archetype.id);
					this.showArchetypeTraitsModal(archetype, traits);
				} catch (err) {
					new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
				}
			};
			actions.createEl("button", { text: "Edit" }).onclick = () => {
				this.showEditArchetypeModal(archetype);
			};
			actions.createEl("button", { text: "Delete" }).onclick = async () => {
				if (confirm(`Delete archetype "${archetype.name}"?`)) {
					try {
						await this.plugin.apiClient.deleteArchetype(archetype.id);
						await this.loadWorldData();
						this.renderWorldTabContent();
						new Notice("Archetype deleted");
					} catch (err) {
						new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
					}
				}
			};
		}
	}

	renderLoreTab(container: HTMLElement) {
		if (this.lores.length === 0) {
			container.createEl("p", { text: "No lore found. Create your first lore!" });
			return;
		}

		// Group by hierarchy level
		const rootLores = this.lores.filter(l => !l.parent_id);
		const list = container.createDiv({ cls: "story-engine-list" });
		
		for (const lore of rootLores.sort((a, b) => a.name.localeCompare(b.name))) {
			this.renderLoreItem(list, lore, 0);
		}
	}

	renderLoreItem(container: HTMLElement, lore: Lore, level: number) {
		const item = container.createDiv({ cls: "story-engine-item" });
		item.style.marginLeft = `${level * 1}rem`;
		
		const titleRow = item.createDiv({ cls: "story-engine-title" });
		titleRow.textContent = lore.name;
		if (lore.category) {
			const categoryBadge = titleRow.createSpan({ cls: "story-engine-badge" });
			categoryBadge.textContent = lore.category;
		}
		
		const meta = item.createDiv({ cls: "story-engine-meta" });
		if (lore.description) {
			meta.createEl("span", { text: lore.description.substring(0, 50) + (lore.description.length > 50 ? "..." : "") });
		}

		const actions = item.createDiv({ cls: "story-engine-item-actions" });
		// Link to Entity
		const linkBtn = actions.createEl("button");
		setIcon(linkBtn, "link");
		linkBtn.title = "Link to Entity";
		linkBtn.onclick = () => this.showAddLoreReferenceModal(lore);
		// Create Sub-Lore
		const subBtn = actions.createEl("button");
		setIcon(subBtn, "folder-plus");
		subBtn.title = "Create Sub-Lore";
		subBtn.onclick = () => this.showCreateLoreModal(lore.id);
		actions.createEl("button", { text: "Edit" }).onclick = () => {
			this.showEditLoreModal(lore);
		};
		actions.createEl("button", { text: "Delete" }).onclick = async () => {
			if (confirm(`Delete lore "${lore.name}"?`)) {
				try {
					await this.plugin.apiClient.deleteLore(lore.id);
					await this.loadWorldData();
					this.renderWorldTabContent();
					new Notice("Lore deleted");
				} catch (err) {
					new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
				}
			}
		};

		// Render children
		const children = this.lores.filter(l => l.parent_id === lore.id);
		for (const child of children.sort((a, b) => a.name.localeCompare(b.name))) {
			this.renderLoreItem(container, child, level + 1);
		}
	}

	renderFactionsTab(container: HTMLElement) {
		if (this.factions.length === 0) {
			container.createEl("p", { text: "No factions found. Create your first faction!" });
			return;
		}

		// Group by hierarchy level
		const rootFactions = this.factions.filter(f => !f.parent_id);
		const list = container.createDiv({ cls: "story-engine-list" });
		
		for (const faction of rootFactions.sort((a, b) => a.name.localeCompare(b.name))) {
			this.renderFactionItem(list, faction, 0);
		}
	}

	renderFactionItem(container: HTMLElement, faction: Faction, level: number) {
		const item = container.createDiv({ cls: "story-engine-item" });
		item.style.marginLeft = `${level * 1}rem`;
		
		const titleRow = item.createDiv({ cls: "story-engine-title" });
		titleRow.textContent = faction.name;
		if (faction.type) {
			const typeBadge = titleRow.createSpan({ cls: "story-engine-badge" });
			typeBadge.textContent = faction.type;
		}
		
		const meta = item.createDiv({ cls: "story-engine-meta" });
		if (faction.description) {
			meta.createEl("span", { text: faction.description.substring(0, 50) + (faction.description.length > 50 ? "..." : "") });
		}

		const actions = item.createDiv({ cls: "story-engine-item-actions" });
		// Link to Entity
		const linkBtn = actions.createEl("button");
		setIcon(linkBtn, "link");
		linkBtn.title = "Link to Entity";
		linkBtn.onclick = () => this.showAddFactionReferenceModal(faction);
		// Create Sub-Faction
		const subBtn = actions.createEl("button");
		setIcon(subBtn, "folder-plus");
		subBtn.title = "Create Sub-Faction";
		subBtn.onclick = () => this.showCreateFactionModal(faction.id);
		actions.createEl("button", { text: "View Details" }).onclick = () => {
			this.showFactionDetailsModal(faction);
		};
		actions.createEl("button", { text: "Edit" }).onclick = () => {
			this.showEditFactionModal(faction);
		};
		actions.createEl("button", { text: "Delete" }).onclick = async () => {
			if (confirm(`Delete faction "${faction.name}"?`)) {
				try {
					await this.plugin.apiClient.deleteFaction(faction.id);
					await this.loadWorldData();
					this.renderWorldTabContent();
					new Notice("Faction deleted");
				} catch (err) {
					new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
				}
			}
		};

		// Render children
		const children = this.factions.filter(f => f.parent_id === faction.id);
		for (const child of children.sort((a, b) => a.name.localeCompare(b.name))) {
			this.renderFactionItem(container, child, level + 1);
		}
	}

	// Modal methods for World entities
	showCreateCharacterModal() {
		if (!this.currentWorld) return;
		const modal = new Modal(this.app);
		modal.titleEl.setText("Create Character");
		
		const content = modal.contentEl;
		let name = "";
		let description = "";

		content.createEl("label", { text: "Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input" });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.oninput = () => { description = descInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const createBtn = buttonContainer.createEl("button", { text: "Create", cls: "mod-cta" });
		createBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.createCharacter(this.currentWorld!.id, { name: name.trim(), description: description.trim() });
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Character created");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();

		modal.open();
		nameInput.focus();
	}

	showEditCharacterModal(character: Character) {
		const modal = new Modal(this.app);
		modal.titleEl.setText("Edit Character");
		
		const content = modal.contentEl;
		let name = character.name;
		let description = character.description;

		content.createEl("label", { text: "Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input", value: name });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.value = description;
		descInput.oninput = () => { description = descInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Save", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				const updated = await this.plugin.apiClient.updateCharacter(character.id, { name: name.trim(), description: description.trim() });
				// Update character details view if we're in character details view
				if (this.viewMode === "character-details" && this.characterDetailsView && this.characterDetailsView["character"].id === character.id) {
					this.characterDetailsView.updateCharacter(updated);
				} else {
					await this.loadWorldData();
					this.renderWorldTabContent();
				}
				modal.close();
				new Notice("Character updated");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();

		modal.open();
	}

	showCreateLocationModal() {
		if (!this.currentWorld) return;
		const modal = new Modal(this.app);
		modal.titleEl.setText("Create Location");
		
		const content = modal.contentEl;
		let name = "";
		let type = "";
		let description = "";
		let parentId: string | null = null;

		content.createEl("label", { text: "Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input" });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Type" });
		const typeInput = content.createEl("input", { type: "text", cls: "story-engine-input", placeholder: "e.g., City, Forest, Building" });
		typeInput.oninput = () => { type = typeInput.value; };

		content.createEl("label", { text: "Parent Location" });
		const parentSelect = content.createEl("select", { cls: "story-engine-select" });
		parentSelect.createEl("option", { value: "", text: "None (Root Location)" });
		for (const loc of this.locations.sort((a, b) => a.name.localeCompare(b.name))) {
			parentSelect.createEl("option", { value: loc.id, text: loc.name });
		}
		parentSelect.onchange = () => { parentId = parentSelect.value || null; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.oninput = () => { description = descInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const createBtn = buttonContainer.createEl("button", { text: "Create", cls: "mod-cta" });
		createBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.createLocation(this.currentWorld!.id, { 
					name: name.trim(), 
					type: type.trim(),
					description: description.trim(),
					parent_id: parentId
				});
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Location created");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();

		modal.open();
		nameInput.focus();
	}

	showEditLocationModal(location: Location) {
		const modal = new Modal(this.app);
		modal.titleEl.setText("Edit Location");
		
		const content = modal.contentEl;
		let name = location.name;
		let type = location.type;
		let description = location.description;

		content.createEl("label", { text: "Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input", value: name });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Type" });
		const typeInput = content.createEl("input", { type: "text", cls: "story-engine-input", value: type });
		typeInput.oninput = () => { type = typeInput.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.value = description;
		descInput.oninput = () => { description = descInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Save", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.updateLocation(location.id, { name: name.trim(), type: type.trim(), description: description.trim() });
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Location updated");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();

		modal.open();
	}

	showCreateArtifactModal() {
		if (!this.currentWorld) return;
		const modal = new Modal(this.app);
		modal.titleEl.setText("Create Artifact");
		
		const content = modal.contentEl;
		let name = "";
		let description = "";
		let rarity = "";

		content.createEl("label", { text: "Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input" });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Rarity" });
		const raritySelect = content.createEl("select", { cls: "story-engine-select" });
		raritySelect.createEl("option", { value: "", text: "Select Rarity" });
		["Common", "Uncommon", "Rare", "Epic", "Legendary", "Unique"].forEach(r => {
			raritySelect.createEl("option", { value: r.toLowerCase(), text: r });
		});
		raritySelect.onchange = () => { rarity = raritySelect.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.oninput = () => { description = descInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const createBtn = buttonContainer.createEl("button", { text: "Create", cls: "mod-cta" });
		createBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.createArtifact(this.currentWorld!.id, { 
					name: name.trim(), 
					description: description.trim(),
					rarity: rarity
				});
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Artifact created");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();

		modal.open();
		nameInput.focus();
	}

	showEditArtifactModal(artifact: Artifact) {
		const modal = new Modal(this.app);
		modal.titleEl.setText("Edit Artifact");
		
		const content = modal.contentEl;
		let name = artifact.name;
		let description = artifact.description;
		let rarity = artifact.rarity;

		content.createEl("label", { text: "Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input", value: name });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Rarity" });
		const raritySelect = content.createEl("select", { cls: "story-engine-select" });
		raritySelect.createEl("option", { value: "", text: "Select Rarity" });
		["Common", "Uncommon", "Rare", "Epic", "Legendary", "Unique"].forEach(r => {
			const opt = raritySelect.createEl("option", { value: r.toLowerCase(), text: r });
			if (rarity.toLowerCase() === r.toLowerCase()) opt.selected = true;
		});
		raritySelect.onchange = () => { rarity = raritySelect.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.value = description;
		descInput.oninput = () => { description = descInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Save", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.updateArtifact(artifact.id, { name: name.trim(), description: description.trim(), rarity: rarity });
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Artifact updated");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();

		modal.open();
	}

	showCreateEventModal() {
		if (!this.currentWorld) return;
		const modal = new Modal(this.app);
		modal.titleEl.setText("Create Event");
		
		const content = modal.contentEl;
		let name = "";
		let type = "";
		let description = "";
		let importance = 5;
		let timelinePosition: number | undefined = undefined;

		// === BASIC INFO ===
		content.createEl("h4", { text: "Basic Info", cls: "story-engine-modal-section" });
		
		content.createEl("label", { text: "Event Name *" });
		const nameInput = content.createEl("input", { 
			type: "text", 
			cls: "story-engine-input",
			placeholder: "e.g., The Battle of Crimson Fields"
		});
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Event Type" });
		const typeSelect = content.createEl("select", { cls: "story-engine-select" });
		typeSelect.createEl("option", { value: "", text: "Select type..." });
		["Battle", "Treaty", "Discovery", "Birth", "Death", "Coronation", "Disaster", "Migration", "Founding", "Other"].forEach(t => {
			typeSelect.createEl("option", { value: t, text: t });
		});
		typeSelect.onchange = () => { type = typeSelect.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { 
			cls: "story-engine-textarea",
			placeholder: "What happened during this event?"
		});
		descInput.oninput = () => { description = descInput.value; };

		// === TIMELINE ===
		content.createEl("h4", { text: "📅 Timeline Position", cls: "story-engine-modal-section" });
		
		content.createEl("label", { text: "Year (relative to Epoch/Year Zero)" });
		content.createEl("p", { 
			text: "Use positive numbers for years after Year Zero, negative for years before.",
			cls: "story-engine-hint"
		});
		const timelinePosInput = content.createEl("input", { 
			type: "number", 
			cls: "story-engine-input", 
			placeholder: "e.g., 100 (Year 100) or -50 (50 years before Year Zero)"
		});
		timelinePosInput.oninput = () => { 
			const val = timelinePosInput.value;
			timelinePosition = val ? parseInt(val) : undefined;
		};

		// === IMPORTANCE ===
		content.createEl("h4", { text: "⭐ Importance", cls: "story-engine-modal-section" });
		
		content.createEl("label", { text: "Importance Level (1-10)" });
		const importanceInput = content.createEl("input", { 
			type: "range", 
			cls: "story-engine-range",
			value: "5", 
			attr: { min: "1", max: "10" } 
		});
		const importanceValue = content.createEl("span", { text: " 5", cls: "story-engine-range-value" });
		importanceInput.oninput = () => { 
			importance = parseInt(importanceInput.value) || 5;
			importanceValue.textContent = ` ${importance}`;
		};

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const createBtn = buttonContainer.createEl("button", { text: "Create Event", cls: "mod-cta" });
		createBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.createEvent(this.currentWorld!.id, { 
					name: name.trim(), 
					type: type.trim() || undefined,
					description: description.trim() || undefined,
					importance: Math.max(1, Math.min(10, importance)),
					timeline_position: timelinePosition
				});
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Event created");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();

		modal.open();
		nameInput.focus();
	}

	showEditEventModal(event: WorldEvent) {
		const modal = new Modal(this.app);
		modal.titleEl.setText("Edit Event");
		
		const content = modal.contentEl;
		let name = event.name;
		let type = event.type || "";
		let description = event.description || "";
		let importance = event.importance;
		let timelinePosition: number | undefined = event.timeline_position ?? undefined;

		// === BASIC INFO ===
		content.createEl("h4", { text: "Basic Info", cls: "story-engine-modal-section" });
		
		content.createEl("label", { text: "Event Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input", value: name });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Event Type" });
		const typeSelect = content.createEl("select", { cls: "story-engine-select" });
		typeSelect.createEl("option", { value: "", text: "Select type..." });
		["Battle", "Treaty", "Discovery", "Birth", "Death", "Coronation", "Disaster", "Migration", "Founding", "Epoch", "Other"].forEach(t => {
			const opt = typeSelect.createEl("option", { value: t, text: t });
			if (type.toLowerCase() === t.toLowerCase()) opt.selected = true;
		});
		typeSelect.onchange = () => { type = typeSelect.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.value = description;
		descInput.oninput = () => { description = descInput.value; };

		// === TIMELINE ===
		content.createEl("h4", { text: "📅 Timeline Position", cls: "story-engine-modal-section" });
		
		content.createEl("label", { text: "Year (relative to Epoch/Year Zero)" });
		content.createEl("p", { 
			text: "Use positive numbers for years after Year Zero, negative for years before.",
			cls: "story-engine-hint"
		});
		const timelinePosInput = content.createEl("input", { 
			type: "number", 
			cls: "story-engine-input", 
			value: timelinePosition?.toString() ?? ""
		});
		timelinePosInput.placeholder = "e.g., 100 or -50";
		timelinePosInput.oninput = () => { 
			const val = timelinePosInput.value;
			timelinePosition = val ? parseInt(val) : undefined;
		};

		// === IMPORTANCE ===
		content.createEl("h4", { text: "⭐ Importance", cls: "story-engine-modal-section" });
		
		content.createEl("label", { text: "Importance Level (1-10)" });
		const importanceInput = content.createEl("input", { 
			type: "range", 
			cls: "story-engine-range",
			value: importance.toString(), 
			attr: { min: "1", max: "10" } 
		});
		const importanceValue = content.createEl("span", { text: ` ${importance}`, cls: "story-engine-range-value" });
		importanceInput.oninput = () => { 
			importance = parseInt(importanceInput.value) || 5;
			importanceValue.textContent = ` ${importance}`;
		};

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Save", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.updateEvent(event.id, { 
					name: name.trim(), 
					type: type.trim() || undefined,
					description: description.trim() || undefined,
					importance: Math.max(1, Math.min(10, importance)),
					timeline_position: timelinePosition
				});
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Event updated");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();

		modal.open();
	}

	showCreateTraitModal() {
		const modal = new Modal(this.app);
		modal.titleEl.setText("Create Trait");
		
		const content = modal.contentEl;
		let name = "";
		let category = "";
		let description = "";

		content.createEl("label", { text: "Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input" });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Category" });
		const categoryInput = content.createEl("input", { type: "text", cls: "story-engine-input", placeholder: "e.g., Personality, Physical, Background" });
		categoryInput.oninput = () => { category = categoryInput.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.oninput = () => { description = descInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const createBtn = buttonContainer.createEl("button", { text: "Create", cls: "mod-cta" });
		createBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.createTrait({ 
					name: name.trim(), 
					category: category.trim(),
					description: description.trim()
				});
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Trait created");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();

		modal.open();
		nameInput.focus();
	}

	showEditTraitModal(trait: Trait) {
		const modal = new Modal(this.app);
		modal.titleEl.setText("Edit Trait");
		
		const content = modal.contentEl;
		let name = trait.name;
		let category = trait.category;
		let description = trait.description;

		content.createEl("label", { text: "Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input", value: name });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Category" });
		const categoryInput = content.createEl("input", { type: "text", cls: "story-engine-input", value: category });
		categoryInput.oninput = () => { category = categoryInput.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.value = description;
		descInput.oninput = () => { description = descInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Save", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.updateTrait(trait.id, { name: name.trim(), category: category.trim(), description: description.trim() });
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Trait updated");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();

		modal.open();
	}

	renderTimeConfigDisplay(container: HTMLElement, config: TimeConfig) {
		container.createEl("h4", { text: "Time Configuration" });
		
		const grid = container.createDiv({ cls: "story-engine-time-config-grid" });
		
		// Mostrar campos
		grid.createDiv().setText(`Base Unit: ${config.base_unit}`);
		grid.createDiv().setText(`Hours/Day: ${config.hours_per_day}`);
		grid.createDiv().setText(`Days/Week: ${config.days_per_week}`);
		grid.createDiv().setText(`Days/Year: ${config.days_per_year}`);
		grid.createDiv().setText(`Months/Year: ${config.months_per_year}`);
		
		if (config.era_name) {
			grid.createDiv().setText(`Era: ${config.era_name}`);
		}
		if (config.month_names?.length) {
			grid.createDiv().setText(`Months: ${config.month_names.join(", ")}`);
		}
		if (config.year_zero !== undefined) {
			grid.createDiv().setText(`Year Zero: ${config.year_zero}`);
		}
	}

	showTimeConfigModal(existingConfig: TimeConfig | null) {
		if (!this.currentWorld) return;
		const modal = new Modal(this.app);
		modal.titleEl.setText(existingConfig ? "Edit Time Configuration" : "Create Time Configuration");
		
		const content = modal.contentEl;
		let baseUnit = existingConfig?.base_unit || "year";
		let hoursPerDay = existingConfig?.hours_per_day || 24;
		let daysPerWeek = existingConfig?.days_per_week || 7;
		let daysPerYear = existingConfig?.days_per_year || 365;
		let monthsPerYear = existingConfig?.months_per_year || 12;
		let eraName = existingConfig?.era_name || "";
		let yearZero = existingConfig?.year_zero?.toString() || "";

		content.createEl("label", { text: "Base Unit *" });
		const baseUnitSelect = content.createEl("select", { cls: "story-engine-select" });
		["year", "day", "hour", "custom"].forEach(unit => {
			const opt = baseUnitSelect.createEl("option", { value: unit, text: unit });
			if (baseUnit === unit) opt.selected = true;
		});
		baseUnitSelect.onchange = () => { baseUnit = baseUnitSelect.value; };

		content.createEl("label", { text: "Hours per Day" });
		const hoursPerDayInput = content.createEl("input", { 
			type: "number", 
			cls: "story-engine-input",
			value: hoursPerDay.toString()
		});
		hoursPerDayInput.oninput = () => { hoursPerDay = parseFloat(hoursPerDayInput.value) || 24; };

		content.createEl("label", { text: "Days per Week" });
		const daysPerWeekInput = content.createEl("input", { 
			type: "number", 
			cls: "story-engine-input",
			value: daysPerWeek.toString()
		});
		daysPerWeekInput.oninput = () => { daysPerWeek = parseInt(daysPerWeekInput.value) || 7; };

		content.createEl("label", { text: "Days per Year" });
		const daysPerYearInput = content.createEl("input", { 
			type: "number", 
			cls: "story-engine-input",
			value: daysPerYear.toString()
		});
		daysPerYearInput.oninput = () => { daysPerYear = parseInt(daysPerYearInput.value) || 365; };

		content.createEl("label", { text: "Months per Year" });
		const monthsPerYearInput = content.createEl("input", { 
			type: "number", 
			cls: "story-engine-input",
			value: monthsPerYear.toString()
		});
		monthsPerYearInput.oninput = () => { monthsPerYear = parseInt(monthsPerYearInput.value) || 12; };

		content.createEl("label", { text: "Era Name (optional)" });
		const eraNameInput = content.createEl("input", { 
			type: "text", 
			cls: "story-engine-input",
			value: eraName
		});
		eraNameInput.oninput = () => { eraName = eraNameInput.value; };

		content.createEl("label", { text: "Year Zero (optional)" });
		const yearZeroInput = content.createEl("input", { 
			type: "number", 
			cls: "story-engine-input",
			value: yearZero
		});
		yearZeroInput.oninput = () => { yearZero = yearZeroInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Save", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			try {
				const timeConfig: TimeConfig = {
					base_unit: baseUnit,
					hours_per_day: hoursPerDay,
					days_per_week: daysPerWeek,
					days_per_year: daysPerYear,
					months_per_year: monthsPerYear,
					era_name: eraName.trim() || undefined,
					year_zero: yearZero ? parseInt(yearZero) : undefined,
				};
				await this.plugin.apiClient.updateWorldTimeConfig(this.currentWorld!.id, timeConfig);
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Time configuration saved");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();

		modal.open();
	}

	async showTimelineModal() {
		if (!this.currentWorld) return;
		try {
			const events = await this.plugin.apiClient.getTimeline(this.currentWorld.id);
			const modal = new TimelineModal(this.app, events, this.currentWorld.time_config || null);
			modal.open();
		} catch (err) {
			new Notice(`Error: ${err instanceof Error ? err.message : "Failed to load timeline"}`, 5000);
		}
	}

	// Placeholder methods for modals that will be implemented later
	showArchetypeTraitsModal(archetype: Archetype, traits: any[]) {
		const modal = new Modal(this.app);
		modal.titleEl.setText(`Traits for ${archetype.name}`);
		const content = modal.contentEl;
		if (traits.length === 0) {
			content.createEl("p", { text: "No traits assigned to this archetype." });
		} else {
			const list = content.createDiv({ cls: "story-engine-list" });
			for (const trait of traits) {
				const item = list.createDiv({ cls: "story-engine-item" });
				item.createDiv({ cls: "story-engine-title", text: trait.trait_name || "Unknown" });
			}
		}
		modal.open();
	}

	showCreateArchetypeModal() {
		const modal = new Modal(this.app);
		modal.titleEl.setText("Create Archetype");
		const content = modal.contentEl;
		let name = "";
		let description = "";

		content.createEl("label", { text: "Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input" });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.oninput = () => { description = descInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const createBtn = buttonContainer.createEl("button", { text: "Create", cls: "mod-cta" });
		createBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.createArchetype({ name: name.trim(), description: description.trim() });
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Archetype created");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();
		modal.open();
		nameInput.focus();
	}

	showEditArchetypeModal(archetype: Archetype) {
		const modal = new Modal(this.app);
		modal.titleEl.setText("Edit Archetype");
		const content = modal.contentEl;
		let name = archetype.name;
		let description = archetype.description;

		content.createEl("label", { text: "Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input", value: name });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.value = description;
		descInput.oninput = () => { description = descInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Save", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.updateArchetype(archetype.id, { name: name.trim(), description: description.trim() });
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Archetype updated");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();
		modal.open();
	}

	showLoreDetailsModal(lore: Lore) {
		const modal = new Modal(this.app);
		modal.titleEl.setText(lore.name);
		const content = modal.contentEl;
		content.createEl("h4", { text: "Description" });
		content.createEl("p", { text: lore.description || "No description" });
		if (lore.rules) {
			content.createEl("h4", { text: "Rules" });
			content.createEl("p", { text: lore.rules });
		}
		if (lore.limitations) {
			content.createEl("h4", { text: "Limitations" });
			content.createEl("p", { text: lore.limitations });
		}
		if (lore.requirements) {
			content.createEl("h4", { text: "Requirements" });
			content.createEl("p", { text: lore.requirements });
		}
		modal.open();
	}

	showCreateLoreModal(parentId?: string | null) {
		if (!this.currentWorld) return;
		const modal = new Modal(this.app);
		modal.titleEl.setText("Create Lore");
		const content = modal.contentEl;
		let name = "";
		let category = "";
		let description = "";
		let rules = "";
		let limitations = "";
		let requirements = "";

		content.createEl("label", { text: "Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input" });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Category" });
		const categoryInput = content.createEl("input", { type: "text", cls: "story-engine-input" });
		categoryInput.oninput = () => { category = categoryInput.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.oninput = () => { description = descInput.value; };

		content.createEl("label", { text: "Rules" });
		const rulesInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		rulesInput.oninput = () => { rules = rulesInput.value; };

		content.createEl("label", { text: "Limitations" });
		const limitationsInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		limitationsInput.oninput = () => { limitations = limitationsInput.value; };

		content.createEl("label", { text: "Requirements" });
		const requirementsInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		requirementsInput.oninput = () => { requirements = requirementsInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const createBtn = buttonContainer.createEl("button", { text: "Create", cls: "mod-cta" });
		createBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.createLore(this.currentWorld!.id, {
					name: name.trim(),
					category: category.trim() || undefined,
					description: description.trim(),
					rules: rules.trim(),
					limitations: limitations.trim(),
					requirements: requirements.trim(),
					parent_id: parentId || undefined,
				});
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Lore created");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();
		modal.open();
		nameInput.focus();
	}

	showEditLoreModal(lore: Lore) {
		const modal = new Modal(this.app);
		modal.titleEl.setText("Edit Lore");
		const content = modal.contentEl;
		let name = lore.name;
		let category = lore.category || "";
		let description = lore.description;
		let rules = lore.rules;
		let limitations = lore.limitations;
		let requirements = lore.requirements;

		content.createEl("label", { text: "Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input", value: name });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Category" });
		const categoryInput = content.createEl("input", { type: "text", cls: "story-engine-input", value: category });
		categoryInput.oninput = () => { category = categoryInput.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.value = description;
		descInput.oninput = () => { description = descInput.value; };

		content.createEl("label", { text: "Rules" });
		const rulesInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		rulesInput.value = rules;
		rulesInput.oninput = () => { rules = rulesInput.value; };

		content.createEl("label", { text: "Limitations" });
		const limitationsInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		limitationsInput.value = limitations;
		limitationsInput.oninput = () => { limitations = limitationsInput.value; };

		content.createEl("label", { text: "Requirements" });
		const requirementsInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		requirementsInput.value = requirements;
		requirementsInput.oninput = () => { requirements = requirementsInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Save", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.updateLore(lore.id, {
					name: name.trim(),
					category: category.trim() || undefined,
					description: description.trim(),
					rules: rules.trim(),
					limitations: limitations.trim(),
					requirements: requirements.trim(),
				});
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Lore updated");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();
		modal.open();
	}

	showFactionDetailsModal(faction: Faction) {
		const modal = new Modal(this.app);
		modal.titleEl.setText(faction.name);
		const content = modal.contentEl;
		content.createEl("h4", { text: "Description" });
		content.createEl("p", { text: faction.description || "No description" });
		if (faction.beliefs) {
			content.createEl("h4", { text: "Beliefs" });
			content.createEl("p", { text: faction.beliefs });
		}
		if (faction.structure) {
			content.createEl("h4", { text: "Structure" });
			content.createEl("p", { text: faction.structure });
		}
		if (faction.symbols) {
			content.createEl("h4", { text: "Symbols" });
			content.createEl("p", { text: faction.symbols });
		}
		modal.open();
	}

	showCreateFactionModal(parentId?: string | null) {
		if (!this.currentWorld) return;
		const modal = new Modal(this.app);
		modal.titleEl.setText("Create Faction");
		const content = modal.contentEl;
		let name = "";
		let type = "";
		let description = "";
		let beliefs = "";
		let structure = "";
		let symbols = "";

		content.createEl("label", { text: "Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input" });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Type" });
		const typeInput = content.createEl("input", { type: "text", cls: "story-engine-input" });
		typeInput.oninput = () => { type = typeInput.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.oninput = () => { description = descInput.value; };

		content.createEl("label", { text: "Beliefs" });
		const beliefsInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		beliefsInput.oninput = () => { beliefs = beliefsInput.value; };

		content.createEl("label", { text: "Structure" });
		const structureInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		structureInput.oninput = () => { structure = structureInput.value; };

		content.createEl("label", { text: "Symbols" });
		const symbolsInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		symbolsInput.oninput = () => { symbols = symbolsInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const createBtn = buttonContainer.createEl("button", { text: "Create", cls: "mod-cta" });
		createBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.createFaction(this.currentWorld!.id, {
					name: name.trim(),
					type: type.trim() || undefined,
					description: description.trim(),
					beliefs: beliefs.trim(),
					structure: structure.trim(),
					symbols: symbols.trim(),
					parent_id: parentId || undefined,
				});
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Faction created");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();
		modal.open();
		nameInput.focus();
	}

	showEditFactionModal(faction: Faction) {
		const modal = new Modal(this.app);
		modal.titleEl.setText("Edit Faction");
		const content = modal.contentEl;
		let name = faction.name;
		let type = faction.type || "";
		let description = faction.description;
		let beliefs = faction.beliefs;
		let structure = faction.structure;
		let symbols = faction.symbols;

		content.createEl("label", { text: "Name *" });
		const nameInput = content.createEl("input", { type: "text", cls: "story-engine-input", value: name });
		nameInput.oninput = () => { name = nameInput.value; };

		content.createEl("label", { text: "Type" });
		const typeInput = content.createEl("input", { type: "text", cls: "story-engine-input", value: type });
		typeInput.oninput = () => { type = typeInput.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.value = description;
		descInput.oninput = () => { description = descInput.value; };

		content.createEl("label", { text: "Beliefs" });
		const beliefsInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		beliefsInput.value = beliefs;
		beliefsInput.oninput = () => { beliefs = beliefsInput.value; };

		content.createEl("label", { text: "Structure" });
		const structureInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		structureInput.value = structure;
		structureInput.oninput = () => { structure = structureInput.value; };

		content.createEl("label", { text: "Symbols" });
		const symbolsInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		symbolsInput.value = symbols;
		symbolsInput.oninput = () => { symbols = symbolsInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Save", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			if (!name.trim()) {
				new Notice("Name is required", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.updateFaction(faction.id, {
					name: name.trim(),
					type: type.trim() || undefined,
					description: description.trim(),
					beliefs: beliefs.trim(),
					structure: structure.trim(),
					symbols: symbols.trim(),
				});
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Faction updated");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();
		modal.open();
	}

	// ==================== Relationship Modals ====================

	showAddCharacterRelationshipModal(character: Character) {
		if (!this.currentWorld) return;
		const modal = new Modal(this.app);
		modal.titleEl.setText("Add Character Relationship");
		
		const content = modal.contentEl;
		let otherCharacterId = "";
		let relationshipType = "ally";
		let description = "";
		let bidirectional = true;

		content.createEl("label", { text: "Other Character *" });
		const characterSelect = content.createEl("select", { cls: "story-engine-select" });
		characterSelect.createEl("option", { value: "", text: "Select a character..." });
		for (const char of this.characters.filter(c => c.id !== character.id).sort((a, b) => a.name.localeCompare(b.name))) {
			characterSelect.createEl("option", { value: char.id, text: char.name });
		}
		characterSelect.onchange = () => { otherCharacterId = characterSelect.value; };

		content.createEl("label", { text: "Relationship Type *" });
		const typeSelect = content.createEl("select", { cls: "story-engine-select" });
		const relationshipTypes = ["ally", "enemy", "family", "lover", "rival", "mentor", "student"];
		for (const type of relationshipTypes) {
			typeSelect.createEl("option", { value: type, text: type.charAt(0).toUpperCase() + type.slice(1) });
		}
		typeSelect.onchange = () => { relationshipType = typeSelect.value; };

		content.createEl("label", { text: "Description" });
		const descInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		descInput.oninput = () => { description = descInput.value; };

		content.createEl("label", { text: "Bidirectional" });
		const bidirectionalCheckbox = content.createEl("input", { type: "checkbox" });
		bidirectionalCheckbox.checked = true;
		bidirectionalCheckbox.onchange = () => { bidirectional = bidirectionalCheckbox.checked; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const createBtn = buttonContainer.createEl("button", { text: "Create", cls: "mod-cta" });
		createBtn.onclick = async () => {
			if (!otherCharacterId) {
				new Notice("Please select a character", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.createCharacterRelationship(character.id, {
					character2_id: otherCharacterId,
					relationship_type: relationshipType,
					description: description.trim(),
					bidirectional: bidirectional,
				});
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Relationship created");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();
		modal.open();
	}

	showLinkEventToEntityModal(event: WorldEvent) {
		if (!this.currentWorld) return;
		const modal = new Modal(this.app);
		modal.titleEl.setText(`Link Event: ${event.name}`);
		
		const content = modal.contentEl;
		let entityType: "character" | "location" | "artifact" | "faction" | "lore" = "character";
		let entityId = "";
		let relationshipType = "";
		let notes = "";

		// Entity Type Selection
		content.createEl("label", { text: "Entity Type *" });
		const typeSelect = content.createEl("select", { cls: "story-engine-select" });
		typeSelect.createEl("option", { value: "character", text: "👤 Character" });
		typeSelect.createEl("option", { value: "location", text: "📍 Location" });
		typeSelect.createEl("option", { value: "faction", text: "🏴 Faction" });
		typeSelect.createEl("option", { value: "artifact", text: "⚔️ Artifact" });
		typeSelect.createEl("option", { value: "lore", text: "📜 Lore" });
		
		// Entity Selection (dynamic based on type)
		content.createEl("label", { text: "Entity *" });
		const entitySelect = content.createEl("select", { cls: "story-engine-select" });
		
		const populateEntitySelect = () => {
			entitySelect.empty();
			entitySelect.createEl("option", { value: "", text: `Select a ${entityType}...` });
			entityId = "";
			
			if (entityType === "character") {
				for (const char of this.characters.sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: char.id, text: char.name });
				}
			} else if (entityType === "location") {
				for (const loc of this.locations.sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: loc.id, text: loc.name });
				}
			} else if (entityType === "faction") {
				for (const faction of this.factions.sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: faction.id, text: faction.name });
				}
			} else if (entityType === "artifact") {
				for (const art of this.artifacts.sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: art.id, text: art.name });
				}
			} else if (entityType === "lore") {
				for (const lore of this.lores.sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: lore.id, text: lore.name });
				}
			}
		};
		
		// Initialize with characters
		populateEntitySelect();
		
		typeSelect.onchange = () => { 
			entityType = typeSelect.value as typeof entityType;
			populateEntitySelect();
		};
		entitySelect.onchange = () => { entityId = entitySelect.value; };

		content.createEl("label", { text: "Relationship Type" });
		const relTypeSelect = content.createEl("select", { cls: "story-engine-select" });
		relTypeSelect.createEl("option", { value: "", text: "Select relationship..." });
		["involved", "caused", "affected", "witnessed", "created", "destroyed", "participated", "led", "opposed"].forEach(rel => {
			relTypeSelect.createEl("option", { value: rel, text: rel.charAt(0).toUpperCase() + rel.slice(1) });
		});
		relTypeSelect.onchange = () => { relationshipType = relTypeSelect.value; };

		content.createEl("label", { text: "Notes" });
		const notesInput = content.createEl("textarea", { 
			cls: "story-engine-textarea",
			placeholder: "Additional details about this relationship..."
		});
		notesInput.oninput = () => { notes = notesInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const createBtn = buttonContainer.createEl("button", { text: "Link Entity", cls: "mod-cta" });
		createBtn.onclick = async () => {
			if (!entityId) {
				new Notice(`Please select a ${entityType}`, 3000);
				return;
			}
			try {
				await this.plugin.apiClient.addEventReference(event.id, entityType, entityId, relationshipType.trim() || undefined, notes.trim() || undefined);
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Entity linked to event");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();
		modal.open();
	}

	showSetEventParentModal(event: WorldEvent) {
		if (!this.currentWorld) return;
		const modal = new Modal(this.app);
		modal.titleEl.setText("Set Parent Event (Cause)");
		
		const content = modal.contentEl;
		let parentId: string | null = event.parent_id || null;

		content.createEl("p", { 
			text: "Link this event to its cause or parent event in the timeline hierarchy.",
			cls: "story-engine-hint"
		});

		content.createEl("label", { text: "Parent Event (Cause)" });
		const parentSelect = content.createEl("select", { cls: "story-engine-select" });
		parentSelect.createEl("option", { value: "", text: "None (Root Event)" });
		
		// Sort by timeline position, then by name
		const sortedEvents = this.events
			.filter(e => e.id !== event.id)
			.sort((a, b) => {
				const posA = a.timeline_position ?? Number.MAX_SAFE_INTEGER;
				const posB = b.timeline_position ?? Number.MAX_SAFE_INTEGER;
				if (posA !== posB) return posA - posB;
				return a.name.localeCompare(b.name);
			});
		
		for (const evt of sortedEvents) {
			const posLabel = evt.timeline_position !== undefined ? ` (Year ${evt.timeline_position})` : "";
			const opt = parentSelect.createEl("option", { value: evt.id, text: `${evt.name}${posLabel}` });
			if (event.parent_id === evt.id) {
				opt.selected = true;
			}
		}
		parentSelect.onchange = () => { parentId = parentSelect.value || null; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Save", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			try {
				await this.plugin.apiClient.moveEvent(event.id, parentId);
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Event parent updated");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();
		modal.open();
	}

	showAddFactionReferenceModal(faction: Faction) {
		if (!this.currentWorld) return;
		const modal = new Modal(this.app);
		modal.titleEl.setText("Link Faction to Entity");
		
		const content = modal.contentEl;
		let entityType = "character";
		let entityId = "";
		let role = "";
		let notes = "";

		content.createEl("label", { text: "Entity Type *" });
		const typeSelect = content.createEl("select", { cls: "story-engine-select" });
		typeSelect.createEl("option", { value: "character", text: "Character" });
		typeSelect.createEl("option", { value: "location", text: "Location" });
		typeSelect.createEl("option", { value: "artifact", text: "Artifact" });
		typeSelect.createEl("option", { value: "event", text: "Event" });
		typeSelect.createEl("option", { value: "faction", text: "Faction" });
		typeSelect.onchange = () => {
			entityType = typeSelect.value;
			// Update entity select options
			entitySelect.empty();
			entitySelect.createEl("option", { value: "", text: `Select a ${entityType}...` });
			loadEntitiesForType(entityType);
		};

		content.createEl("label", { text: "Entity *" });
		const entitySelect = content.createEl("select", { cls: "story-engine-select" });
		entitySelect.createEl("option", { value: "", text: "Select an entity..." });
		
		const loadEntitiesForType = (type: string) => {
			if (type === "character") {
				for (const char of this.characters.sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: char.id, text: char.name });
				}
			} else if (type === "location") {
				for (const loc of this.locations.sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: loc.id, text: loc.name });
				}
			} else if (type === "artifact") {
				for (const art of this.artifacts.sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: art.id, text: art.name });
				}
			} else if (type === "event") {
				for (const evt of this.events.sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: evt.id, text: evt.name });
				}
			} else if (type === "faction") {
				// Filter out the current faction to avoid self-reference
				for (const fac of this.factions.filter(f => f.id !== faction.id).sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: fac.id, text: fac.name });
				}
			}
		};
		loadEntitiesForType(entityType);
		entitySelect.onchange = () => { entityId = entitySelect.value; };

		content.createEl("label", { text: "Role" });
		const roleInput = content.createEl("input", { type: "text", cls: "story-engine-input", placeholder: "e.g., leader, member, ally, rival" });
		roleInput.oninput = () => { role = roleInput.value; };

		content.createEl("label", { text: "Notes" });
		const notesInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		notesInput.oninput = () => { notes = notesInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const createBtn = buttonContainer.createEl("button", { text: "Create", cls: "mod-cta" });
		createBtn.onclick = async () => {
			if (!entityId) {
				new Notice("Please select an entity", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.addFactionReference(faction.id, entityType, entityId, role.trim() || undefined, notes.trim() || undefined);
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Reference created");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();
		modal.open();
	}

	showAddLoreReferenceModal(lore: Lore) {
		if (!this.currentWorld) return;
		const modal = new Modal(this.app);
		modal.titleEl.setText("Link Lore to Entity");
		
		const content = modal.contentEl;
		let entityType = "character";
		let entityId = "";
		let relationshipType = "";
		let notes = "";

		content.createEl("label", { text: "Entity Type *" });
		const typeSelect = content.createEl("select", { cls: "story-engine-select" });
		typeSelect.createEl("option", { value: "character", text: "Character" });
		typeSelect.createEl("option", { value: "location", text: "Location" });
		typeSelect.createEl("option", { value: "artifact", text: "Artifact" });
		typeSelect.createEl("option", { value: "event", text: "Event" });
		typeSelect.createEl("option", { value: "faction", text: "Faction" });
		typeSelect.createEl("option", { value: "lore", text: "Lore" });
		typeSelect.onchange = () => {
			entityType = typeSelect.value;
			// Update entity select options
			entitySelect.empty();
			entitySelect.createEl("option", { value: "", text: `Select a ${entityType}...` });
			loadEntitiesForType(entityType);
		};

		content.createEl("label", { text: "Entity *" });
		const entitySelect = content.createEl("select", { cls: "story-engine-select" });
		entitySelect.createEl("option", { value: "", text: "Select an entity..." });
		
		const loadEntitiesForType = (type: string) => {
			if (type === "character") {
				for (const char of this.characters.sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: char.id, text: char.name });
				}
			} else if (type === "location") {
				for (const loc of this.locations.sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: loc.id, text: loc.name });
				}
			} else if (type === "artifact") {
				for (const art of this.artifacts.sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: art.id, text: art.name });
				}
			} else if (type === "event") {
				for (const evt of this.events.sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: evt.id, text: evt.name });
				}
			} else if (type === "faction") {
				for (const fac of this.factions.sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: fac.id, text: fac.name });
				}
			} else if (type === "lore") {
				// Filter out the current lore to avoid self-reference
				for (const l of this.lores.filter(l => l.id !== lore.id).sort((a, b) => a.name.localeCompare(b.name))) {
					entitySelect.createEl("option", { value: l.id, text: l.name });
				}
			}
		};
		loadEntitiesForType(entityType);
		entitySelect.onchange = () => { entityId = entitySelect.value; };

		content.createEl("label", { text: "Relationship Type" });
		const relTypeInput = content.createEl("input", { type: "text", cls: "story-engine-input", placeholder: "e.g., practitioner, origin, forbidden" });
		relTypeInput.oninput = () => { relationshipType = relTypeInput.value; };

		content.createEl("label", { text: "Notes" });
		const notesInput = content.createEl("textarea", { cls: "story-engine-textarea" });
		notesInput.oninput = () => { notes = notesInput.value; };

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const createBtn = buttonContainer.createEl("button", { text: "Create", cls: "mod-cta" });
		createBtn.onclick = async () => {
			if (!entityId) {
				new Notice("Please select an entity", 3000);
				return;
			}
			try {
				await this.plugin.apiClient.addLoreReference(lore.id, entityType, entityId, relationshipType.trim() || undefined, notes.trim() || undefined);
				await this.loadWorldData();
				this.renderWorldTabContent();
				modal.close();
				new Notice("Reference created");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
		buttonContainer.createEl("button", { text: "Cancel" }).onclick = () => modal.close();
		modal.open();
	}

	// ==================== Character Details View Methods ====================

	async showCharacterDetails(character: Character) {
		this.viewMode = "character-details";
		
		// Get world if available
		const world = character.world_id ? await this.plugin.apiClient.getWorld(character.world_id).catch(() => null) : null;
		
		// Get characters for relationships (from world if available, otherwise empty)
		const characters = world ? await this.plugin.apiClient.getCharacters(world.id).catch(() => []) : [];
		
		// Ensure traits and archetypes are loaded (they're global per tenant)
		let traits = this.traits;
		let archetypes = this.archetypes;
		let events = world ? this.events : [];
		
		// If traits or archetypes are empty, load them
		if (!traits || traits.length === 0) {
			try {
				traits = await this.plugin.apiClient.getTraits();
				this.traits = traits;
			} catch (err) {
				console.warn("Failed to load traits:", err);
				traits = [];
			}
		}
		
		if (!archetypes || archetypes.length === 0) {
			try {
				archetypes = await this.plugin.apiClient.getArchetypes();
				this.archetypes = archetypes;
			} catch (err) {
				console.warn("Failed to load archetypes:", err);
				archetypes = [];
			}
		}
		
		// If world is available but events are empty, load them
		if (world && (!events || events.length === 0)) {
			try {
				events = await this.plugin.apiClient.getEvents(world.id);
				this.events = events;
			} catch (err) {
				console.warn("Failed to load events:", err);
				events = [];
			}
		}
		
		// Create CharacterDetailsView instance
		this.characterDetailsView = new CharacterDetailsView(
			this.plugin,
			character,
			this.headerEl,
			this.contentEl,
			() => {
				// onBack callback
				this.characterDetailsView = null;
				if (this.currentWorld) {
					this.viewMode = "world-details";
					this.renderWorldDetails();
				} else if (this.currentStory) {
					this.viewMode = "details";
					this.renderDetails();
				} else {
					this.viewMode = "list";
					this.renderListHeader();
					this.renderListContent();
				}
			},
			(character) => {
				// onEditCharacter callback
				this.showEditCharacterModal(character);
			},
			world,
			characters,
			archetypes,
			traits,
			events
		);
		
		await this.characterDetailsView.render();
	}

}

