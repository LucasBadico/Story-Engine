import { Notice, Plugin } from "obsidian";
import { ExtractSearchResult, StoryEngineSettings } from "./types";
import { StoryEngineClient } from "./api/client";
import { StoryEngineSettingTab } from "./settings";
import { registerCommands } from "./commands";
import { StoryDetailsModal } from "./views/StoryDetailsModal";
import { CreateStoryModal } from "./views/CreateStoryModal";
import { FileManager } from "./sync/fileManager";
import { SyncService } from "./sync/syncService";
import { StoryListView, STORY_LIST_VIEW_TYPE } from "./views/StoryListView";
import {
	StoryEngineExtractView,
	STORY_ENGINE_EXTRACT_VIEW_TYPE,
} from "./views/StoryEngineExtractView";

const DEFAULT_SETTINGS: StoryEngineSettings = {
	apiUrl: "http://localhost:8080",
	llmGatewayUrl: "http://localhost:8081",
	apiKey: "",
	tenantId: "",
	tenantName: "",
	syncFolderPath: "Stories",
	autoVersionSnapshots: true,
	conflictResolution: "service",
	mode: "local",
	showHelpBox: true,
	localModeVideoUrl: "https://example.com/setup-video",
};

export default class StoryEnginePlugin extends Plugin {
	settings!: StoryEngineSettings;
	apiClient!: StoryEngineClient;
	fileManager!: FileManager;
	syncService!: SyncService;
	extractResult: ExtractSearchResult | null = null;

	async onload() {
		await this.loadSettings();

		this.apiClient = new StoryEngineClient(
			this.settings.apiUrl,
			this.settings.apiKey,
			this.settings.tenantId || ""
		);
		this.apiClient.setMode(this.settings.mode || "local");

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

		this.registerView(
			STORY_ENGINE_EXTRACT_VIEW_TYPE,
			(leaf) => new StoryEngineExtractView(leaf, this)
		);

		// Add ribbon icon to open the view
		this.addRibbonIcon("book-open", "Story Engine", () => {
			this.activateView();
		});

		registerCommands(this);

		this.registerEvent(
			this.app.workspace.on("editor-menu", (menu, editor) => {
				const selection = editor.getSelection().trim();
				if (!selection) {
					return;
				}

				menu.addItem((item) => {
					item.setTitle("Story Engine: Extract entities");
					item.setIcon("search");
					item.onClick(() => {
						this.extractSelectionCommand(selection);
					});
				});
			})
		);
	}

	async onunload() {
		// Detach all leaves of the story list view
		this.app.workspace.detachLeavesOfType(STORY_LIST_VIEW_TYPE);
		this.app.workspace.detachLeavesOfType(STORY_ENGINE_EXTRACT_VIEW_TYPE);
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
		if (this.apiClient) {
			this.apiClient.setTenantId(this.settings.tenantId || "");
			this.apiClient.setMode(this.settings.mode || "local");
		} else {
			this.apiClient = new StoryEngineClient(
				this.settings.apiUrl,
				this.settings.apiKey,
				this.settings.tenantId || ""
			);
			this.apiClient.setMode(this.settings.mode || "local");
		}
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
		// Validate tenant ID only in remote mode
		if (this.settings.mode === "remote") {
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
		}

		// Open modal to get title, world, and sync preference
		new CreateStoryModal(this.app, this, async (title: string, worldId: string | undefined, shouldSync: boolean) => {
			try {
				new Notice(`Creating story "${title}"...`);
				
				const story = await this.apiClient.createStory(title, worldId);
				
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

	async activateExtractView() {
		const { workspace } = this.app;

		let leaf = workspace.getLeavesOfType(STORY_ENGINE_EXTRACT_VIEW_TYPE)[0];

		if (!leaf) {
			leaf = workspace.getLeavesOfType(STORY_LIST_VIEW_TYPE)[0];
		}

		if (!leaf) {
			new Notice("Could not open extract view. Please try again.", 3000);
			return;
		}

		await leaf.setViewState({
			type: STORY_ENGINE_EXTRACT_VIEW_TYPE,
			active: true,
		});

		workspace.revealLeaf(leaf);
	}

	async extractSelectionCommand(selection: string) {
		const trimmedSelection = selection.trim();
		if (!trimmedSelection) {
			new Notice("Select text to extract entities", 3000);
			return;
		}

		if (this.settings.mode !== "remote") {
			new Notice("Extraction requires the full remote version.", 5000);
			return;
		}

		const tenantId = this.settings.tenantId?.trim();
		if (!tenantId) {
			new Notice("Please configure Tenant ID in settings", 5000);
			return;
		}

		const gatewayUrl = this.settings.llmGatewayUrl?.trim();
		if (!gatewayUrl) {
			new Notice("Please configure LLM Gateway URL in settings", 5000);
			return;
		}

		try {
			if (navigator?.clipboard?.writeText) {
				await navigator.clipboard.writeText(trimmedSelection);
			}
		} catch {
			// Clipboard is optional; ignore failures.
		}

		new Notice("Sending text to extraction...", 3000);

		try {
			const response = await fetch(
				`${gatewayUrl.replace(/\/$/, "")}/api/v1/search`,
				{
					method: "POST",
					headers: {
						"Content-Type": "application/json",
						"X-Tenant-ID": tenantId,
						...(this.settings.apiKey
							? { Authorization: `Bearer ${this.settings.apiKey}` }
							: {}),
					},
					body: JSON.stringify({
						query: trimmedSelection,
						limit: 10,
					}),
				}
			);

			if (!response.ok) {
				let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
				try {
					const errorBody = (await response.json()) as { error?: string };
					if (errorBody?.error) {
						errorMessage = errorBody.error;
					}
				} catch {
					// Ignore JSON parse errors.
				}
				throw new Error(errorMessage);
			}

			const payload = (await response.json()) as {
				chunks: ExtractSearchResult["chunks"];
				next_cursor?: string;
			};

			this.extractResult = {
				query: trimmedSelection,
				chunks: payload.chunks ?? [],
				next_cursor: payload.next_cursor,
				received_at: new Date().toISOString(),
			};

			await this.activateView();
			await this.activateExtractView();
			this.updateExtractViews();

			new Notice(
				`Extraction complete: ${this.extractResult.chunks.length} matches`,
				4000
			);
		} catch (err) {
			const errorMessage =
				err instanceof Error ? err.message : "Failed to extract entities";
			new Notice(`Error: ${errorMessage}`, 5000);
		}
	}

	updateExtractViews() {
		const listLeaf =
			this.app.workspace.getLeavesOfType(STORY_LIST_VIEW_TYPE)[0];
		if (listLeaf) {
			const view = listLeaf.view as StoryListView;
			if (view.viewMode === "list") {
				view.renderListContent();
			}
		}

		const extractLeaf =
			this.app.workspace.getLeavesOfType(STORY_ENGINE_EXTRACT_VIEW_TYPE)[0];
		if (extractLeaf) {
			const view = extractLeaf.view as StoryEngineExtractView;
			view.setResult(this.extractResult);
		}
	}

	openSettings() {
		const setting = (this.app as any).setting;
		if (setting) {
			setting.open();
			setTimeout(() => {
				setting.openTabById(this.manifest.id);
			}, 100);
		}
	}
}
