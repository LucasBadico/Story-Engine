import type StoryEnginePlugin from "../main";
import { EventRef, TAbstractFile, TFile, TFolder } from "obsidian";
import { SyncEntityTarget } from "./entitySyncTypes";

type SyncReason = "blur" | "idle";

const IDLE_DELAY_MS = 60_000;

export class AutoSyncManager {
	private leafChangeRef?: EventRef;
	private editorChangeRef?: EventRef;
	private idleTimeoutId: number | null = null;
	private activeFile: TFile | null = null;
	private lastEditTs = 0;
	private dirtyFiles = new Set<string>();
	private pendingSyncs = new Set<string>();

	constructor(private plugin: StoryEnginePlugin) {
		this.activeFile = this.plugin.app.workspace.getActiveFile();
		this.lastEditTs = Date.now();
		this.registerEvents();
		this.resetIdleTimer();
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
		if (this.idleTimeoutId !== null) {
			window.clearTimeout(this.idleTimeoutId);
			this.idleTimeoutId = null;
		}
		this.dirtyFiles.clear();
		this.pendingSyncs.clear();
	}

	private registerEvents(): void {
		this.leafChangeRef = this.plugin.app.workspace.on(
			"active-leaf-change",
			this.handleActiveLeafChange
		);
		this.plugin.registerEvent(this.leafChangeRef);

		this.editorChangeRef = this.plugin.app.workspace.on("editor-change", () => {
			const file = this.plugin.app.workspace.getActiveFile();
			if (!file) {
				return;
			}

			this.activeFile = file;
			this.lastEditTs = Date.now();
			this.dirtyFiles.add(file.path);
			this.resetIdleTimer();
		});
		this.plugin.registerEvent(this.editorChangeRef);
	}

	private handleActiveLeafChange = (leaf: unknown): void => {
		const previousFile = this.activeFile;
		const newFile = this.plugin.app.workspace.getActiveFile();

		if (
			previousFile &&
			(!newFile || previousFile.path !== newFile.path) &&
			this.dirtyFiles.has(previousFile.path)
		) {
			void this.triggerSyncForFile(previousFile, "blur");
		}

		this.activeFile = newFile;
		this.lastEditTs = Date.now();
		this.resetIdleTimer();
	};

	private resetIdleTimer(): void {
		if (this.idleTimeoutId !== null) {
			window.clearTimeout(this.idleTimeoutId);
		}

		if (!this.activeFile) {
			this.idleTimeoutId = null;
			return;
		}

		this.idleTimeoutId = window.setTimeout(() => {
			const file = this.activeFile;
			if (!file) {
				return;
			}

			const idleDuration = Date.now() - this.lastEditTs;
			if (idleDuration >= IDLE_DELAY_MS && this.dirtyFiles.has(file.path)) {
				void this.triggerSyncForFile(file, "idle");
			} else {
				this.resetIdleTimer();
			}
		}, IDLE_DELAY_MS);
	}

	private async triggerSyncForFile(file: TFile, reason: SyncReason): Promise<void> {
		const path = file.path;

		if (this.pendingSyncs.has(path)) {
			return;
		}

		this.pendingSyncs.add(path);

		try {
			const target = await this.resolveTarget(file);
			if (!target) {
				return;
			}

			if (target.pushWholeStory) {
				await this.plugin.syncService.pushStory(target.folderPath);
			} else if (target.syncTarget) {
				await this.plugin.syncService.pushStory(target.folderPath, target.syncTarget);
			}

			this.dirtyFiles.delete(path);
			if (reason === "idle") {
				this.lastEditTs = Date.now();
				this.resetIdleTimer();
			}
		} catch (err) {
			console.error(`Auto push failed for ${file.path}`, err);
		} finally {
			this.pendingSyncs.delete(path);
		}
	}

	private async resolveTarget(file: TFile): Promise<{
		folderPath: string;
		pushWholeStory: boolean;
		syncTarget?: SyncEntityTarget;
	} | null> {
		const folderPath = this.findStoryFolderPath(file);
		if (!folderPath) {
			return null;
		}

		if (file.name === "story.md") {
			return { folderPath, pushWholeStory: true };
		}

		const content = await this.plugin.app.vault.read(file);
		const frontmatter = this.plugin.fileManager.parseFrontmatter(content);

		const entityId = frontmatter?.id;
		if (!entityId) {
			return null;
		}

		const sceneId = frontmatter.scene_id;
		const chapterId = frontmatter.chapter_id;
		const isContentFile = file.path.includes("/03-contents/");
		const isChapterFile = file.path.includes("/00-chapters/");
		const isSceneFile = file.path.includes("/01-scenes/");

		if (frontmatter.story_id && frontmatter.number && isChapterFile) {
			return {
				folderPath,
				pushWholeStory: false,
				syncTarget: { type: "chapter", id: entityId },
			};
		}

		if (frontmatter.story_id && (isSceneFile || chapterId)) {
			return {
				folderPath,
				pushWholeStory: false,
				syncTarget: { type: "scene", id: entityId },
			};
		}

		if (sceneId) {
			return {
				folderPath,
				pushWholeStory: false,
				syncTarget: { type: "scene", id: sceneId },
			};
		}

		if (isContentFile) {
			return {
				folderPath,
				pushWholeStory: false,
				syncTarget: { type: "content", id: entityId },
			};
		}

		return null;
	}

	private findStoryFolderPath(file: TFile): string | null {
		let current: TFolder | null = file.parent;

		while (current) {
			const storyFilePath = `${current.path}/story.md`;
			const maybeStoryFile = this.plugin.app.vault.getAbstractFileByPath(storyFilePath);
			if (maybeStoryFile instanceof TFile) {
				return current.path;
			}

			const parent = current.parent;
			if (parent instanceof TFolder) {
				current = parent;
			} else {
				current = null;
			}
		}

		return null;
	}
}

