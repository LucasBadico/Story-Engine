import { Notice, Plugin } from "obsidian";
import { StoryEngineSettings } from "./types";
import { StoryEngineClient } from "./api/client";
import { StoryEngineSettingTab } from "./settings";
import { registerCommands } from "./commands";
import { StoryDetailsModal } from "./views/StoryDetailsModal";
import { CreateStoryModal } from "./views/CreateStoryModal";
import { FileManager } from "./sync/fileManager";
import { SyncService } from "./sync/syncService";
import { StoryListView, STORY_LIST_VIEW_TYPE } from "./views/StoryListView";

const DEFAULT_SETTINGS: StoryEngineSettings = {
	apiUrl: "http://localhost:8080",
	apiKey: "",
	tenantId: "",
	tenantName: "",
	syncFolderPath: "Stories",
	autoVersionSnapshots: true,
	conflictResolution: "service",
};

export default class StoryEnginePlugin extends Plugin {
	settings!: StoryEngineSettings;
	apiClient!: StoryEngineClient;
	fileManager!: FileManager;
	syncService!: SyncService;

	async onload() {
		await this.loadSettings();

		this.apiClient = new StoryEngineClient(
			this.settings.apiUrl,
			this.settings.apiKey
		);

		this.fileManager = new FileManager(
			this.app.vault,
			this.settings.syncFolderPath || "Stories"
		);

		this.syncService = new SyncService(
			this.apiClient,
			this.fileManager,
			this.settings,
			this.app
		);

		this.addSettingTab(new StoryEngineSettingTab(this.app, this));

		// Register the story list view
		this.registerView(
			STORY_LIST_VIEW_TYPE,
			(leaf) => new StoryListView(leaf, this)
		);

		// Add ribbon icon to open the view
		this.addRibbonIcon("book-open", "Story Engine", () => {
			this.activateView();
		});

		registerCommands(this);
	}

	async onunload() {
		// Detach all leaves of the story list view
		this.app.workspace.detachLeavesOfType(STORY_LIST_VIEW_TYPE);
	}

	async loadSettings() {
		this.settings = Object.assign(
			{},
			DEFAULT_SETTINGS,
			await this.loadData()
		);
	}

	async saveSettings() {
		await this.saveData(this.settings);
		// Update API client when settings change
		this.apiClient = new StoryEngineClient(
			this.settings.apiUrl,
			this.settings.apiKey
		);
		// Update file manager base path
		this.fileManager = new FileManager(
			this.app.vault,
			this.settings.syncFolderPath || "Stories"
		);
		// Update sync service
		this.syncService = new SyncService(
			this.apiClient,
			this.fileManager,
			this.settings,
			this.app
		);
	}

	async createStoryCommand() {
		// Validate and trim tenant ID first
		const tenantId = this.settings.tenantId?.trim();
		if (!tenantId) {
			new Notice("Please configure Tenant ID in settings", 5000);
			return;
		}

		// Basic UUID format validation
		const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
		if (!uuidRegex.test(tenantId)) {
			new Notice("Invalid Tenant ID format. Please check your settings.", 5000);
			return;
		}

		// Open modal to get title and sync preference
		new CreateStoryModal(this.app, async (title: string, shouldSync: boolean) => {
			try {
				new Notice(`Creating story "${title}"...`);
				
				const story = await this.apiClient.createStory(tenantId, title);
				
				new Notice(`Story "${title}" created successfully`);

				// Sync to Obsidian if requested
				if (shouldSync) {
					try {
						new Notice(`Syncing story to Obsidian...`);
						await this.syncService.pullStory(story.id);
						new Notice(`Story synced to your vault!`);
					} catch (syncErr) {
						const syncErrorMessage = syncErr instanceof Error 
							? syncErr.message 
							: "Failed to sync story";
						new Notice(`Story created but sync failed: ${syncErrorMessage}`, 5000);
					}
				}

				// Update the view if it's open
				const openView = this.app.workspace.getLeavesOfType(STORY_LIST_VIEW_TYPE)[0];
				if (openView) {
					const view = openView.view as StoryListView;
					await view.refresh();
					if (!shouldSync) {
						// Show details in the view
						await view.showStoryDetails(story);
					}
				}
			} catch (err) {
				const errorMessage = err instanceof Error ? err.message : "Failed to create story";
				new Notice(`Error: ${errorMessage}`, 5000);
			}
		}).open();
	}

	async activateView() {
		const { workspace } = this.app;

		let leaf = workspace.getLeavesOfType(STORY_LIST_VIEW_TYPE)[0];

		if (!leaf) {
			// Create new leaf in the right panel
			const rightLeaf = workspace.getRightLeaf(false);
			if (!rightLeaf) {
				new Notice("Could not create view. Please try again.", 3000);
				return;
			}
			leaf = rightLeaf;
			await leaf.setViewState({
				type: STORY_LIST_VIEW_TYPE,
				active: true,
			});
		}

		workspace.revealLeaf(leaf);
	}
}

