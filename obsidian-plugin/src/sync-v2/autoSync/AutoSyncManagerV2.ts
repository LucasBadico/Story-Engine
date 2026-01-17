import type StoryEnginePlugin from "../../main";
import { EventRef, TFile, TFolder } from "obsidian";
import type { SyncOrchestrator } from "../core/SyncOrchestrator";
import type { SyncOperation, SyncContext, SyncResult } from "../types/sync";

type SyncReason = "blur" | "idle" | "typing_pause";

const TYPING_PAUSE_DELAY_MS = 1_000; // 1s typing pause
const IDLE_DELAY_MS = 5_000; // 5s idle

interface PendingOperation {
	filePath: string;
	operation: SyncOperation;
	reason: SyncReason;
	timestamp: number;
}

export class AutoSyncManagerV2 {
	private leafChangeRef?: EventRef;
	private editorChangeRef?: EventRef;
	private typingPauseTimeoutId: number | null = null;
	private idleTimeoutId: number | null = null;
	private activeFile: TFile | null = null;
	private lastEditTs = 0;
	private dirtyFiles = new Set<string>();
	private pendingOperations = new Map<string, PendingOperation>();
	private operationQueue: PendingOperation[] = [];
	private isProcessingQueue = false;
	private lastOnlineCheck = 0;
	private onlineCheckInFlight = false;
	private isOnlineCached = true;

	constructor(
		private plugin: StoryEnginePlugin,
		private orchestrator: SyncOrchestrator,
		private context: SyncContext
	) {
		this.activeFile = this.plugin.app.workspace.getActiveFile();
		this.lastEditTs = Date.now();
		this.registerEvents();
	}

	dispose(): void {
		if (this.leafChangeRef) {
			this.plugin.app.workspace.offref(this.leafChangeRef);
			this.leafChangeRef = undefined;
		}
		if (this.editorChangeRef) {
			this.plugin.app.workspace.offref(this.editorChangeRef);
			this.editorChangeRef = undefined;
		}
		if (this.typingPauseTimeoutId !== null) {
			const clearFn = typeof window !== "undefined" ? window.clearTimeout : globalThis.clearTimeout;
			clearFn(this.typingPauseTimeoutId);
			this.typingPauseTimeoutId = null;
		}
		if (this.idleTimeoutId !== null) {
			const clearFn = typeof window !== "undefined" ? window.clearTimeout : globalThis.clearTimeout;
			clearFn(this.idleTimeoutId);
			this.idleTimeoutId = null;
		}
		this.dirtyFiles.clear();
		this.pendingOperations.clear();
		this.operationQueue = [];
	}

	private registerEvents(): void {
		// Active leaf change (blur event)
		this.leafChangeRef = this.plugin.app.workspace.on(
			"active-leaf-change",
			this.handleActiveLeafChange
		);
		this.plugin.registerEvent(this.leafChangeRef);

		// Editor change (typing pause)
		this.editorChangeRef = this.plugin.app.workspace.on("editor-change", () => {
			const file = this.plugin.app.workspace.getActiveFile();
			if (!file || !this.isStoryEntityFile(file)) {
				return;
			}

			this.activeFile = file;
			this.lastEditTs = Date.now();
			this.dirtyFiles.add(file.path);
			this.resetTypingPauseTimer(file);
			this.resetIdleTimer();
		});
		this.plugin.registerEvent(this.editorChangeRef);
	}

	private handleActiveLeafChange = (): void => {
		const previousFile = this.activeFile;
		const newFile = this.plugin.app.workspace.getActiveFile();

		if (
			previousFile &&
			(!newFile || previousFile.path !== newFile.path) &&
			this.dirtyFiles.has(previousFile.path) &&
			this.isStoryEntityFile(previousFile)
		) {
			void this.triggerSyncForFile(previousFile, "blur");
		}

		this.activeFile = newFile;
		this.lastEditTs = Date.now();
		this.resetTypingPauseTimer(newFile);
		this.resetIdleTimer();
	};

	private resetTypingPauseTimer(file: TFile | null): void {
		if (this.typingPauseTimeoutId !== null) {
			const clearFn = typeof window !== "undefined" ? window.clearTimeout : globalThis.clearTimeout;
			clearFn(this.typingPauseTimeoutId);
		}

		if (!file || !this.isStoryEntityFile(file)) {
			this.typingPauseTimeoutId = null;
			return;
		}

		const setTimeoutFn = typeof window !== "undefined" ? window.setTimeout : globalThis.setTimeout;
		this.typingPauseTimeoutId = setTimeoutFn(async () => {
			if (this.dirtyFiles.has(file.path)) {
				await this.triggerSyncForFile(file, "typing_pause");
			}
		}, TYPING_PAUSE_DELAY_MS) as unknown as number;
	}

	private resetIdleTimer(): void {
		if (this.idleTimeoutId !== null) {
			const clearFn = typeof window !== "undefined" ? window.clearTimeout : globalThis.clearTimeout;
			clearFn(this.idleTimeoutId);
		}

		if (!this.activeFile || !this.isStoryEntityFile(this.activeFile)) {
			this.idleTimeoutId = null;
			return;
		}

		const setTimeoutFn = typeof window !== "undefined" ? window.setTimeout : globalThis.setTimeout;
		this.idleTimeoutId = setTimeoutFn(async () => {
			const file = this.activeFile;
			if (!file) {
				return;
			}

			const idleDuration = Date.now() - this.lastEditTs;
			if (idleDuration >= IDLE_DELAY_MS && this.dirtyFiles.has(file.path)) {
				await this.triggerSyncForFile(file, "idle");
			} else {
				this.resetIdleTimer();
			}
		}, IDLE_DELAY_MS) as unknown as number;
	}

	private async triggerSyncForFile(file: TFile, reason: SyncReason): Promise<void> {
		const path = file.path;

		if (this.pendingOperations.has(path)) {
			// Operation already pending, update reason if needed (blur > idle > typing_pause)
			const existing = this.pendingOperations.get(path)!;
			const priority = { blur: 3, idle: 2, typing_pause: 1 };
			if (priority[reason] > priority[existing.reason]) {
				existing.reason = reason;
				existing.timestamp = Date.now();
			}
			return;
		}

		const operation = await this.resolveOperation(file);
		if (!operation) {
			// Not a syncable file, remove from dirty files
			this.dirtyFiles.delete(path);
			return;
		}

		const pendingOp: PendingOperation = {
			filePath: path,
			operation,
			reason,
			timestamp: Date.now(),
		};

		this.pendingOperations.set(path, pendingOp);
		this.operationQueue.push(pendingOp);
		this.dirtyFiles.delete(path);

		// Process queue if not already processing
		if (!this.isProcessingQueue) {
			void this.processOperationQueue();
		}

		if (reason === "idle") {
			this.lastEditTs = Date.now();
			this.resetIdleTimer();
		}
	}

	private async processOperationQueue(): Promise<void> {
		if (this.isProcessingQueue) {
			return;
		}

		if (this.operationQueue.length === 0) {
			return;
		}

		this.isProcessingQueue = true;

		try {
			// Batch operations by story folder (group by folderPath)
			// Get a snapshot of current queue to avoid processing items added during processing
			const queueSnapshot = [...this.operationQueue];
			this.operationQueue = []; // Clear queue immediately to prevent duplicates
			
			const batches = this.batchOperationsFromSnapshot(queueSnapshot);
			const processedFiles = new Set<string>();
			
			for (const batch of batches) {
				// Process batch (for now, process sequentially - can be parallelized later)
				for (const pendingOp of batch) {
					processedFiles.add(pendingOp.filePath);
					try {
						const result = await this.orchestrator.run(pendingOp.operation);
						this.handleSyncResult(pendingOp, result);
					} catch (err) {
						console.error(`Auto sync failed for ${pendingOp.filePath}`, err);
						this.context.emitWarning?.({
							code: "auto_sync_error",
							message: `Auto sync failed for ${pendingOp.filePath}: ${err instanceof Error ? err.message : String(err)}`,
							filePath: pendingOp.filePath,
							details: err,
						});
						// Re-add to dirty files if error is network-related (for retry later)
						if (this.isNetworkError(err)) {
							this.dirtyFiles.add(pendingOp.filePath);
						}
					} finally {
						this.pendingOperations.delete(pendingOp.filePath);
					}
				}
			}
		} catch (err) {
			console.error("Error processing operation queue", err);
		} finally {
			this.isProcessingQueue = false;

			// Process remaining queue items if any were added during processing
			if (this.operationQueue.length > 0) {
				void this.processOperationQueue();
			}
		}
	}

	private batchOperationsFromSnapshot(snapshot: PendingOperation[]): PendingOperation[][] {
		// Group operations by story folder (folderPath in PushStoryPayload)
		// For push_story operations, only keep one per folder (most recent)
		// For other operations, keep all
		const storyBatches = new Map<string, PendingOperation>();
		const otherOperations: PendingOperation[] = [];

		for (const pendingOp of snapshot) {
			if (pendingOp.operation.type === "push_story") {
				const folderPath = pendingOp.operation.payload.folderPath;
				// Keep only the most recent operation per folder (prioritize by reason: blur > idle > typing_pause)
				const existing = storyBatches.get(folderPath);
				if (!existing) {
					storyBatches.set(folderPath, pendingOp);
				} else {
					const priority = { blur: 3, idle: 2, typing_pause: 1 };
					if (
						priority[pendingOp.reason] > priority[existing.reason] ||
						(priority[pendingOp.reason] === priority[existing.reason] &&
							pendingOp.timestamp > existing.timestamp)
					) {
						storyBatches.set(folderPath, pendingOp);
					}
				}
			} else {
				// For non-push_story operations, process individually
				otherOperations.push(pendingOp);
			}
		}

		// Convert to batches array (one operation per batch)
		const batches: PendingOperation[][] = [];
		for (const op of storyBatches.values()) {
			batches.push([op]);
		}
		for (const op of otherOperations) {
			batches.push([op]);
		}

		return batches;
	}

	private isNetworkError(err: unknown): boolean {
		if (err instanceof Error) {
			return (
				err.message.includes("network") ||
				err.message.includes("fetch") ||
				err.message.includes("timeout") ||
				err.message.includes("ECONNREFUSED") ||
				err.message.includes("ENOTFOUND")
			);
		}
		return false;
	}


	private handleSyncResult(pendingOp: PendingOperation, result: SyncResult): void {
		if (!result.success) {
			console.error(`Sync failed for ${pendingOp.filePath}:`, result.errors);
			// Re-add to dirty files if error is recoverable
			if (result.errors?.some((e) => e.recoverable)) {
				this.dirtyFiles.add(pendingOp.filePath);
			}
		} else {
			console.log(`Sync succeeded for ${pendingOp.filePath}:`, result.message);
		}

		// Emit warnings
		if (result.warnings) {
			for (const warning of result.warnings) {
				this.context.emitWarning?.(warning);
			}
		}
	}

	private async resolveOperation(file: TFile): Promise<SyncOperation | null> {
		const folderPath = this.findStoryFolderPath(file);
		if (!folderPath) {
			return null;
		}

		// For V2, always push the entire story (push_story)
		// This is simpler and more reliable than pushing individual entities
		if (file.name === "story.md" || this.isStoryEntityFile(file)) {
			return {
				type: "push_story",
				payload: {
					folderPath,
				},
			};
		}

		return null;
	}

	private isStoryEntityFile(file: TFile | null): boolean {
		if (!file) {
			return false;
		}

		// Check if file is in a story folder structure
		const path = file.path;
		return (
			path.includes("/00-chapters/") ||
			path.includes("/01-scenes/") ||
			path.includes("/02-beats/") ||
			path.includes("/03-contents/") ||
			path.includes("/worlds/")
		);
	}

	private findStoryFolderPath(file: TFile): string | null {
		let current: TFolder | null = file.parent;

		while (current) {
			const storyFilePath = `${current.path}/story.md`;
			const maybeStoryFile = this.plugin.app.vault.getAbstractFileByPath(storyFilePath);
			// Check if it's a file (has path but no children property, or is instance of TFile)
			if (maybeStoryFile && ("path" in maybeStoryFile) && !("children" in maybeStoryFile)) {
				return current.path;
			}

			// For world entities, check for world.md
			const worldFilePath = `${current.path}/world.md`;
			const maybeWorldFile = this.plugin.app.vault.getAbstractFileByPath(worldFilePath);
			if (maybeWorldFile && ("path" in maybeWorldFile) && !("children" in maybeWorldFile)) {
				// For world entities, we need to find the world root (parent of world folder)
				const worldRoot = current.parent;
				if (worldRoot && "children" in worldRoot) {
					return worldRoot.path;
				}
				return current.path;
			}

			const parent = current.parent;
			if (parent && "children" in parent) {
				current = parent as TFolder;
			} else {
				current = null;
			}
		}

		return null;
	}

	/**
	 * Add operation to queue (for manual triggers or offline mode)
	 * This method allows adding operations that will be processed when online
	 */
	async enqueueOperation(operation: SyncOperation, filePath: string): Promise<void> {
		const pendingOp: PendingOperation = {
			filePath,
			operation,
			reason: "typing_pause",
			timestamp: Date.now(),
		};

		// Check if operation already exists for this file
		const existingIndex = this.operationQueue.findIndex((op) => op.filePath === filePath);
		if (existingIndex !== -1) {
			// Replace existing operation with newer one
			this.operationQueue[existingIndex] = pendingOp;
		} else {
			this.operationQueue.push(pendingOp);
		}

		// Mark file as dirty to ensure it's tracked
		this.dirtyFiles.add(filePath);

		// Process queue if not already processing and we're online
		// (In offline mode, queue will be processed when connection is restored)
		if (!this.isProcessingQueue && this.isOnline()) {
			void this.processOperationQueue();
		}
	}

	/**
	 * Check if we're online (can make API calls)
	 * For now, always return true - can be enhanced with actual network detection
	 */
	private isOnline(): boolean {
		if (typeof navigator !== "undefined" && navigator.onLine === false) {
			this.isOnlineCached = false;
			return false;
		}

		const now = Date.now();
		if (now - this.lastOnlineCheck > 15_000 && !this.onlineCheckInFlight) {
			void this.refreshOnlineStatus();
		}

		return this.isOnlineCached;
	}

	private async refreshOnlineStatus(): Promise<void> {
		this.onlineCheckInFlight = true;
		try {
			await this.context.apiClient.listStories();
			this.isOnlineCached = true;
		} catch {
			this.isOnlineCached = false;
		} finally {
			this.lastOnlineCheck = Date.now();
			this.onlineCheckInFlight = false;
		}
	}

	/**
	 * Process pending queue (call this when connection is restored)
	 */
	async processPendingQueue(): Promise<void> {
		if (this.operationQueue.length > 0 && !this.isProcessingQueue) {
			void this.processOperationQueue();
		}
	}

	/**
	 * Get pending operations count (for UI/status)
	 */
	getPendingOperationsCount(): number {
		return this.operationQueue.length + this.pendingOperations.size;
	}

	/**
	 * Clear pending operations (for reset/cleanup)
	 */
	clearPendingOperations(): void {
		this.operationQueue = [];
		this.pendingOperations.clear();
		this.dirtyFiles.clear();
	}
}

