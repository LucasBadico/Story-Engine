import { Notice, Plugin } from "obsidian";
import {
	ExtractEntityResult,
	ExtractLogEntry,
	ExtractStreamEvent,
	StoryEngineSettings,
} from "./types";
import { StoryEngineClient } from "./api/client";
import { StoryEngineSettingTab } from "./settings";
import { registerCommands } from "./commands";
import { StoryDetailsModal } from "./views/StoryDetailsModal";
import { CreateStoryModal } from "./views/CreateStoryModal";
import { ExtractConfigModal } from "./views/ExtractConfigModal";
import { FileManager } from "./sync/fileManager";
import { AutoSyncManager } from "./sync/autoSyncManager";
import { AutoSyncManagerV2 } from "./sync-v2/autoSync/AutoSyncManagerV2";
import { StoryListView, STORY_LIST_VIEW_TYPE } from "./views/StoryListView";
import type { ModularSyncEngine } from "./sync-v2/core/ModularSyncEngine";
import {
	StoryEngineExtractView,
	STORY_ENGINE_EXTRACT_VIEW_TYPE,
} from "./views/StoryEngineExtractView";
import { createSyncEngine } from "./sync/createSyncEngine";
import type { SyncEngine } from "./sync/engine";

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
	syncVersion: "v1",
	showHelpBox: true,
	localModeVideoUrl: "https://example.com/setup-video",
	autoSyncOnApiUpdates: true,
	autoPushOnFileBlur: true,
	backupMode: "snapshots",
	backupRetentionDays: 7,
};

export default class StoryEnginePlugin extends Plugin {
	settings!: StoryEngineSettings;
	apiClient!: StoryEngineClient;
	fileManager!: FileManager;
	syncService!: SyncEngine;
	private autoSyncManager?: AutoSyncManager | AutoSyncManagerV2;
	extractResult: ExtractEntityResult | null = null;
	extractLogs: ExtractLogEntry[] = [];
	extractStatus: "idle" | "running" | "done" | "error" | "canceled" = "idle";
	private extractAbortController: AbortController | null = null;

	async onload() {
		await this.loadSettings();

		this.apiClient = new StoryEngineClient(
			this.settings.apiUrl,
			this.settings.apiKey,
			this.settings.tenantId || ""
		);
		this.apiClient.setMode(this.settings.mode || "local");
		this.apiClient.setAutoSyncOnApiUpdates(
			this.settings.autoSyncOnApiUpdates ?? true
		);

		this.fileManager = new FileManager(
			this.app.vault,
			this.settings.syncFolderPath || "Stories"
		);

		this.syncService = createSyncEngine(
			{
				app: this.app,
				apiClient: this.apiClient,
				fileManager: this.fileManager,
				settings: this.settings,
			},
			this.settings.syncVersion || "v1"
		);
		this.initializeAutoSyncManager();

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
					item.setTitle("Story Engine: Extract Entities and Relations");
					item.setIcon("search");
					item.onClick(() => {
						this.extractSelectionCommand(selection, true);
					});
				});

				menu.addItem((item) => {
					item.setTitle("Story Engine: Extract Entities Only");
					item.setIcon("search");
					item.onClick(() => {
						this.extractSelectionCommand(selection, false);
					});
				});
			})
		);
	}

	async onunload() {
		// Detach all leaves of the story list view
		this.app.workspace.detachLeavesOfType(STORY_LIST_VIEW_TYPE);
		this.app.workspace.detachLeavesOfType(STORY_ENGINE_EXTRACT_VIEW_TYPE);
		if (this.syncService) {
			this.syncService.dispose();
		}
		if (this.autoSyncManager) {
			this.autoSyncManager.dispose();
			this.autoSyncManager = undefined;
		}
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
			this.apiClient.setAutoSyncOnApiUpdates(
				this.settings.autoSyncOnApiUpdates ?? true
			);
		} else {
			this.apiClient = new StoryEngineClient(
				this.settings.apiUrl,
				this.settings.apiKey,
				this.settings.tenantId || ""
			);
			this.apiClient.setMode(this.settings.mode || "local");
			this.apiClient.setAutoSyncOnApiUpdates(
				this.settings.autoSyncOnApiUpdates ?? true
			);
		}
		// Update file manager base path
		this.fileManager = new FileManager(
			this.app.vault,
			this.settings.syncFolderPath || "Stories"
		);
		// Update sync service
		if (this.syncService) {
			this.syncService.dispose();
		}
		this.syncService = createSyncEngine(
			{
				app: this.app,
				apiClient: this.apiClient,
				fileManager: this.fileManager,
				settings: this.settings,
			},
			this.settings.syncVersion || "v1"
		);
		this.initializeAutoSyncManager();
	}

	private initializeAutoSyncManager(): void {
		if (this.autoSyncManager) {
			this.autoSyncManager.dispose();
			this.autoSyncManager = undefined;
		}

		if (!this.settings.autoPushOnFileBlur) {
			return;
		}

		const syncVersion = this.settings.syncVersion || "v1";

		if (syncVersion === "v2") {
			// Use AutoSyncManagerV2 with SyncOrchestrator
			const modularEngine = this.syncService as ModularSyncEngine;
			if (modularEngine && "getOrchestrator" in modularEngine) {
				const orchestrator = modularEngine.getOrchestrator();
				const context = modularEngine.getContext();
				this.autoSyncManager = new AutoSyncManagerV2(this, orchestrator, context);
			} else {
				console.warn(
					"[AutoSyncManager] Sync V2 selected but ModularSyncEngine not available, falling back to V1"
				);
				// Fallback to V1 if V2 not available
				this.autoSyncManager = new AutoSyncManager(this);
			}
		} else {
			// Use AutoSyncManager V1
			this.autoSyncManager = new AutoSyncManager(this);
		}
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

	private async getActiveWorldId(): Promise<string | null> {
		const activeFile = this.app.workspace.getActiveFile();
		if (!activeFile) {
			return null;
		}

		const fileContent = await this.app.vault.read(activeFile);
		const frontmatter = this.fileManager.parseFrontmatter(fileContent);
		const storyId =
			frontmatter.story_id ||
			(activeFile.name === "story.md" ? frontmatter.id : "");

		if (!storyId) {
			return null;
		}

		const story = await this.apiClient.getStory(storyId);
		return story.world_id ?? null;
	}

	async extractSelectionCommand(selection: string, includeRelations = true) {
		const trimmedSelection = selection.trim();
		if (!trimmedSelection) {
			new Notice(
				includeRelations
					? "Select text to extract entities and relations"
					: "Select text to extract entities",
				3000
			);
			return;
		}

		const defaultTypes = ["character", "location", "artefact", "faction", "event"];
		const config = await ExtractConfigModal.open(
			this.app,
			defaultTypes,
			includeRelations
		);
		if (!config) {
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

		let worldId: string | null = null;
		try {
			worldId = await this.getActiveWorldId();
		} catch (err) {
			const errorMessage =
				err instanceof Error ? err.message : "Failed to resolve story";
			new Notice(`Error: ${errorMessage}`, 5000);
			return;
		}

		if (!worldId) {
			new Notice(
				"Open a synced story document before extracting entities.",
				5000
			);
			return;
		}

		this.resetExtractState(trimmedSelection, worldId, config.includeRelations);
		await this.activateView();
		await this.activateExtractView();
		this.updateExtractViews();

		try {
			if (navigator?.clipboard?.writeText) {
				await navigator.clipboard.writeText(trimmedSelection);
			}
		} catch {
			// Clipboard is optional; ignore failures.
		}

		new Notice("Starting extraction stream...", 3000);
		await this.startExtractStream({
			tenantId,
			gatewayUrl,
			worldId,
			text: trimmedSelection,
			includeRelations: config.includeRelations,
			entityTypes: config.entityTypes,
		});
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
			view.setLogs(this.extractLogs, this.extractStatus);
		}
	}

	private resetExtractState(
		text: string,
		worldId: string,
		includeRelations = true
	) {
		this.cancelExtractStream();
		this.extractResult = {
			text,
			world_id: worldId,
			entities: [],
			relations: [],
			received_at: new Date().toISOString(),
			include_relations: includeRelations,
		};
		this.extractLogs = [];
		this.extractStatus = "running";
	}

	cancelExtractStream() {
		if (this.extractAbortController) {
			this.extractAbortController.abort();
			this.extractAbortController = null;
			this.extractStatus = "canceled";
			this.appendExtractLog({
				type: "client.cancel",
				message: "Extraction canceled by user.",
			});
			this.updateExtractViews();
		}
	}

	private appendExtractLog(event: ExtractStreamEvent) {
		const timestamp = event.timestamp
			? new Date(event.timestamp).toISOString()
			: new Date().toISOString();
		this.extractLogs.push({
			id: `${timestamp}-${this.extractLogs.length}`,
			eventType: event.type,
			phase: event.phase,
			message: event.message,
			data: event.data,
			timestamp,
		});
	}

	private async startExtractStream(params: {
		tenantId: string;
		gatewayUrl: string;
		worldId: string;
		text: string;
		includeRelations: boolean;
		entityTypes: string[];
	}) {
		const { tenantId, gatewayUrl, worldId, text, includeRelations, entityTypes } =
			params;
		const controller = new AbortController();
		this.extractAbortController = controller;
		this.appendExtractLog({
			type: "client.start",
			message: "Opening extraction stream.",
		});
		this.updateExtractViews();

		try {
			const response = await fetch(
				`${gatewayUrl.replace(/\/$/, "")}/api/v1/entity-extract/stream`,
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
						text,
						world_id: worldId,
						include_relations: includeRelations,
						entity_types: entityTypes,
					}),
					signal: controller.signal,
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

			if (!response.body) {
				throw new Error("No response stream available.");
			}

			const reader = response.body.getReader();
			const decoder = new TextDecoder("utf-8");
			let buffer = "";
			let currentEvent = "message";
			let dataLines: string[] = [];

			const flushEvent = () => {
				if (!dataLines.length) return;
				const rawData = dataLines.join("\n").trim();
				dataLines = [];
				if (!rawData) return;

				let parsed: ExtractStreamEvent | null = null;
				try {
					parsed = JSON.parse(rawData) as ExtractStreamEvent;
				} catch {
					this.appendExtractLog({
						type: "parse.error",
						message: "Failed to parse stream event payload.",
						data: { raw: rawData, event: currentEvent },
					});
					this.updateExtractViews();
					return;
				}

				if (!parsed.type) {
					parsed.type = currentEvent || "message";
				}

				if (parsed.type === "relation.error") {
					parsed.message = `❌ ${parsed.message}`;
				} else if (parsed.type === "relation.success") {
					parsed.message = `✅ ${parsed.message}`;
				}

				this.appendExtractLog(parsed);

				if (parsed.type === "result_entities" && parsed.data?.entities) {
					const entities = parsed.data.entities as ExtractEntityResult["entities"];
					if (this.extractResult) {
						this.extractResult.entities = entities ?? [];
						this.extractResult.received_at = new Date().toISOString();
					}
					const foundCount = this.extractResult?.entities.filter(
						(entity) => entity.found
					).length;
					new Notice(
						`Entities extracted: ${foundCount ?? 0}/${this.extractResult?.entities.length ?? 0} matched`,
						4000
					);
				}

				if (parsed.type === "result_relations" && parsed.data?.relations) {
					const relations = parsed.data.relations as ExtractEntityResult["relations"];
					if (this.extractResult) {
						this.extractResult.relations = relations ?? [];
						this.extractResult.received_at = new Date().toISOString();
					}
				}

				if (parsed.type === "error") {
					this.extractStatus = "error";
				}

				this.updateExtractViews();
			};

			while (true) {
				const { value, done } = await reader.read();
				if (done) break;
				buffer += decoder.decode(value, { stream: true });

				let lineEnd = buffer.indexOf("\n");
				while (lineEnd !== -1) {
					const line = buffer.slice(0, lineEnd).replace(/\r$/, "");
					buffer = buffer.slice(lineEnd + 1);
					lineEnd = buffer.indexOf("\n");

					if (!line) {
						flushEvent();
						currentEvent = "message";
						continue;
					}

					if (line.startsWith("event:")) {
						currentEvent = line.replace("event:", "").trim();
						continue;
					}

					if (line.startsWith("data:")) {
						dataLines.push(line.replace("data:", "").trim());
					}
				}
			}

			this.extractAbortController = null;
			if (this.extractStatus === "running") {
				this.extractStatus = "done";
			}
		} catch (err) {
			if (err instanceof DOMException && err.name === "AbortError") {
				return;
			}
			const errorMessage =
				err instanceof Error ? err.message : "Failed to extract entities";
			this.extractStatus = "error";
			this.appendExtractLog({
				type: "error",
				message: errorMessage,
			});
			new Notice(`Error: ${errorMessage}`, 5000);
			this.updateExtractViews();
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
