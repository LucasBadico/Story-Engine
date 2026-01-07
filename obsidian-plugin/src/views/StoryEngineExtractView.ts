import { ItemView, WorkspaceLeaf } from "obsidian";
import StoryEnginePlugin from "../main";
import { ExtractSearchChunk, ExtractSearchResult } from "../types";

export const STORY_ENGINE_EXTRACT_VIEW_TYPE = "story-engine-extract-view";

export class StoryEngineExtractView extends ItemView {
	plugin: StoryEnginePlugin;
	private result: ExtractSearchResult | null = null;
	private contentRoot!: HTMLElement;

	constructor(leaf: WorkspaceLeaf, plugin: StoryEnginePlugin) {
		super(leaf);
		this.plugin = plugin;
	}

	getViewType(): string {
		return STORY_ENGINE_EXTRACT_VIEW_TYPE;
	}

	getDisplayText(): string {
		return "Extract";
	}

	getIcon(): string {
		return "search";
	}

	async onOpen() {
		const container = this.containerEl.children[1] as HTMLElement;
		container.empty();
		container.addClass("story-engine-extract-container");
		this.contentRoot = container;
		this.setResult(this.plugin.extractResult);
	}

	setResult(result: ExtractSearchResult | null) {
		this.result = result;
		this.render();
	}

	private render() {
		if (!this.contentRoot) return;
		this.contentRoot.empty();

		if (!this.result) {
			const header = this.contentRoot.createDiv({
				cls: "story-engine-extract-header",
			});
			header.createEl("h2", { text: "Extract" });

			const headerActions = header.createDiv({
				cls: "story-engine-extract-header-actions",
			});
			const backButton = headerActions.createEl("button", {
				text: "Back to Stories",
				cls: "story-engine-extract-back",
			});
			backButton.onclick = () => {
				this.plugin.activateView();
			};

			this.contentRoot.createEl("p", {
				text: "No extraction results yet.",
				cls: "story-engine-extract-empty",
			});
			return;
		}

		const header = this.contentRoot.createDiv({
			cls: "story-engine-extract-header",
		});
		header.createEl("h2", { text: "Extract Results" });

		const headerMeta = header.createDiv({
			cls: "story-engine-extract-meta",
		});
		headerMeta.createEl("div", {
			text: `Matches: ${this.result.chunks.length}`,
		});
		headerMeta.createEl("div", {
			text: `Query length: ${this.result.query.length}`,
		});

		const headerActions = header.createDiv({
			cls: "story-engine-extract-header-actions",
		});
		const backButton = headerActions.createEl("button", {
			text: "Back to Stories",
			cls: "story-engine-extract-back",
		});
		backButton.onclick = () => {
			this.plugin.activateView();
		};

		const queryBlock = this.contentRoot.createDiv({
			cls: "story-engine-extract-query",
		});
		queryBlock.createEl("div", {
			text: "Query",
			cls: "story-engine-extract-label",
		});
		queryBlock.createEl("div", {
			text: this.result.query,
			cls: "story-engine-extract-query-text",
		});

		const actions = this.contentRoot.createDiv({
			cls: "story-engine-extract-actions",
		});
		const clearButton = actions.createEl("button", {
			text: "Clear Results",
			cls: "story-engine-extract-clear",
		});
		clearButton.onclick = () => {
			this.plugin.extractResult = null;
			this.setResult(null);
			this.plugin.updateExtractViews();
		};

		const list = this.contentRoot.createDiv({
			cls: "story-engine-extract-list",
		});

		if (!this.result.chunks.length) {
			list.createEl("p", {
				text: "No matches returned from search.",
				cls: "story-engine-extract-empty",
			});
			return;
		}

		this.result.chunks.forEach((chunk, index) => {
			this.renderChunk(list, chunk, index);
		});
	}

	private renderChunk(container: HTMLElement, chunk: ExtractSearchChunk, index: number) {
		const item = container.createDiv({ cls: "story-engine-extract-item" });
		const header = item.createDiv({ cls: "story-engine-extract-item-header" });
		header.createEl("div", {
			text: `#${index + 1}`,
			cls: "story-engine-extract-rank",
		});
		header.createEl("div", {
			text: `Score: ${chunk.score.toFixed(3)}`,
			cls: "story-engine-extract-score",
		});

		const meta = item.createDiv({ cls: "story-engine-extract-item-meta" });
		meta.createEl("div", { text: `Source: ${chunk.source_type}` });
		meta.createEl("div", { text: `Source ID: ${chunk.source_id}` });
		if (chunk.content_kind) {
			meta.createEl("div", { text: `Kind: ${chunk.content_kind}` });
		}
		if (chunk.location_name) {
			meta.createEl("div", { text: `Location: ${chunk.location_name}` });
		}
		if (chunk.timeline) {
			meta.createEl("div", { text: `Timeline: ${chunk.timeline}` });
		}
		if (chunk.pov_character) {
			meta.createEl("div", { text: `POV: ${chunk.pov_character}` });
		}
		if (chunk.characters && chunk.characters.length) {
			meta.createEl("div", {
				text: `Characters: ${chunk.characters.join(", ")}`,
			});
		}

		item.createEl("div", {
			text: chunk.content,
			cls: "story-engine-extract-content",
		});
	}
}
