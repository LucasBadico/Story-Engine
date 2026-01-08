import { ItemView, Notice, WorkspaceLeaf } from "obsidian";
import StoryEnginePlugin from "../main";
import {
	ExtractEntity,
	ExtractEntityMatch,
	ExtractEntityResult,
} from "../types";

export const STORY_ENGINE_EXTRACT_VIEW_TYPE = "story-engine-extract-view";

export class StoryEngineExtractView extends ItemView {
	plugin: StoryEnginePlugin;
	private result: ExtractEntityResult | null = null;
	private contentRoot!: HTMLElement;
	private logs = [] as StoryEnginePlugin["extractLogs"];
	private status: StoryEnginePlugin["extractStatus"] = "idle";

	constructor(leaf: WorkspaceLeaf, plugin: StoryEnginePlugin) {
		super(leaf);
		this.plugin = plugin;
		this.logs = plugin.extractLogs;
		this.status = plugin.extractStatus;
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
		this.setLogs(this.plugin.extractLogs, this.plugin.extractStatus);
	}

	setResult(result: ExtractEntityResult | null) {
		this.result = result;
		this.render();
	}

	setLogs(logs: typeof this.logs, status: typeof this.status) {
		this.logs = logs;
		this.status = status;
		this.render();
	}

	private render() {
		if (!this.contentRoot) return;
		this.contentRoot.empty();

		const header = this.contentRoot.createDiv({
			cls: "story-engine-extract-header",
		});
		header.createEl("h2", { text: "Extract" });

		const headerMeta = header.createDiv({
			cls: "story-engine-extract-meta",
		});
		headerMeta.createEl("div", {
			text: `Status: ${this.status}`,
		});
		if (this.result) {
			const foundCount = this.result.entities.filter((entity) => entity.found)
				.length;
			headerMeta.createEl("div", {
				text: `Entities: ${this.result.entities.length}`,
			});
			headerMeta.createEl("div", {
				text: `Found: ${foundCount}`,
			});
			headerMeta.createEl("div", {
				text: `Text length: ${this.result.text.length}`,
			});
		}

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
		if (this.status === "running") {
			const cancelButton = headerActions.createEl("button", {
				text: "Cancel",
				cls: "story-engine-extract-cancel",
			});
			cancelButton.onclick = () => {
				this.plugin.cancelExtractStream();
			};
		}

		if (!this.result) {
			this.renderLogs();
			this.contentRoot.createEl("p", {
				text: "No extraction results yet.",
				cls: "story-engine-extract-empty",
			});
			return;
		}

		const queryBlock = this.contentRoot.createDiv({
			cls: "story-engine-extract-query",
		});
		queryBlock.createEl("div", {
			text: "Text",
			cls: "story-engine-extract-label",
		});
		queryBlock.createEl("div", {
			text: this.result.text,
			cls: "story-engine-extract-query-text",
		});

		this.renderLogs();

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

		if (!this.result.entities.length) {
			list.createEl("p", {
				text: "No entities returned from extraction.",
				cls: "story-engine-extract-empty",
			});
			return;
		}

		this.result.entities.forEach((entity, index) => {
			this.renderEntity(list, entity, index);
		});
	}

	private renderLogs() {
		const logBlock = this.contentRoot.createDiv({
			cls: "story-engine-extract-logs",
		});
		logBlock.createEl("div", {
			text: "Progress",
			cls: "story-engine-extract-label",
		});

		if (!this.logs.length) {
			logBlock.createEl("p", {
				text: "No events yet.",
				cls: "story-engine-extract-empty",
			});
			return;
		}

		const list = logBlock.createDiv({ cls: "story-engine-extract-log-list" });
		this.logs.forEach((entry) => {
			const item = list.createDiv({ cls: "story-engine-extract-log-item" });
			item.createDiv({
				text: entry.timestamp,
				cls: "story-engine-extract-log-time",
			});
			item.createDiv({
				text: entry.phase ? `${entry.phase} · ${entry.eventType}` : entry.eventType,
				cls: "story-engine-extract-log-type",
			});
			item.createDiv({
				text: entry.message,
				cls: "story-engine-extract-log-message",
			});
			if (entry.data && Object.keys(entry.data).length) {
				item.createEl("pre", {
					text: JSON.stringify(entry.data, null, 2),
					cls: "story-engine-extract-log-data",
				});
			}
		});
	}

	private renderEntity(container: HTMLElement, entity: ExtractEntity, index: number) {
		const item = container.createDiv({ cls: "story-engine-extract-item" });
		const header = item.createDiv({ cls: "story-engine-extract-item-header" });
		header.createEl("div", {
			text: `#${index + 1}`,
			cls: "story-engine-extract-rank",
		});
		header.createEl("div", {
			text: entity.found ? "Found" : "New",
			cls: `story-engine-extract-status ${
				entity.found ? "is-found" : "is-new"
			}`,
		});

		const title = item.createDiv({ cls: "story-engine-extract-item-title" });
		title.createEl("div", {
			text: entity.name,
			cls: "story-engine-extract-entity-name",
		});
		title.createEl("div", {
			text: entity.type,
			cls: "story-engine-extract-entity-type",
		});

		if (entity.summary) {
			item.createEl("div", {
				text: entity.summary,
				cls: "story-engine-extract-content",
			});
		}

		if (entity.match) {
			item.appendChild(this.renderMatch("Match", entity.match));
		}

		if (entity.candidates && entity.candidates.length) {
			const list = item.createDiv({ cls: "story-engine-extract-candidates" });
			list.createEl("div", {
				text: "Candidates",
				cls: "story-engine-extract-label",
			});
			entity.candidates.forEach((candidate) => {
				list.appendChild(this.renderMatch("", candidate));
			});
		}

		const actions = item.createDiv({ cls: "story-engine-extract-actions" });
		const actionButton = actions.createEl("button", {
			text: entity.found ? "Update Entity" : "Create Entity",
			cls: "story-engine-extract-action",
		});
		actionButton.onclick = async () => {
			actionButton.disabled = true;
			try {
				if (entity.found) {
					await this.updateEntity(entity);
				} else {
					await this.createEntity(entity);
				}
			} finally {
				actionButton.disabled = false;
			}
		};
	}

	private renderMatch(label: string, match: ExtractEntityMatch): HTMLElement {
		const wrapper = document.createElement("div");
		wrapper.className = "story-engine-extract-match";
		const parts = [];
		if (label) {
			parts.push(label);
		}
		parts.push(`${match.source_type}:${match.source_id}`);
		if (match.entity_name) {
			parts.push(match.entity_name);
		}
		parts.push(`sim ${match.similarity.toFixed(3)}`);
		if (match.reason) {
			parts.push(match.reason);
		}
		wrapper.textContent = parts.join(" · ");
		return wrapper;
	}

	private async createEntity(entity: ExtractEntity) {
		if (!this.result?.world_id) {
			new Notice("World ID missing. Select a story or world first.", 4000);
			return;
		}

		const description = entity.summary ?? "";
		let createdId = "";

		switch (entity.type) {
			case "character": {
				const created = await this.plugin.apiClient.createCharacter(
					this.result.world_id,
					{ name: entity.name, description }
				);
				createdId = created.id;
				break;
			}
			case "location": {
				const created = await this.plugin.apiClient.createLocation(
					this.result.world_id,
					{ name: entity.name, description }
				);
				createdId = created.id;
				break;
			}
			case "artefact": {
				const created = await this.plugin.apiClient.createArtifact(
					this.result.world_id,
					{ name: entity.name, description }
				);
				createdId = created.id;
				break;
			}
			case "faction": {
				const created = await this.plugin.apiClient.createFaction(
					this.result.world_id,
					{ name: entity.name, description }
				);
				createdId = created.id;
				break;
			}
			case "event": {
				const created = await this.plugin.apiClient.createEvent(
					this.result.world_id,
					{ name: entity.name, description }
				);
				createdId = created.id;
				break;
			}
			default:
				new Notice(`Unsupported type: ${entity.type}`, 4000);
				return;
		}

		entity.found = true;
		entity.match = {
			source_type: entity.type,
			source_id: createdId,
			entity_name: entity.name,
			similarity: 1,
			reason: "Created from extract",
		};
		this.render();
		new Notice(`Created ${entity.type}: ${entity.name}`, 3000);
	}

	private async updateEntity(entity: ExtractEntity) {
		if (!entity.match?.source_id) {
			new Notice("No match available to update.", 4000);
			return;
		}

		const description = entity.summary ?? "";
		switch (entity.type) {
			case "character":
				await this.plugin.apiClient.updateCharacter(entity.match.source_id, {
					name: entity.name,
					description,
				});
				break;
			case "location":
				await this.plugin.apiClient.updateLocation(entity.match.source_id, {
					name: entity.name,
					description,
				});
				break;
			case "artefact":
				await this.plugin.apiClient.updateArtifact(entity.match.source_id, {
					name: entity.name,
					description,
				});
				break;
			case "faction":
				await this.plugin.apiClient.updateFaction(entity.match.source_id, {
					name: entity.name,
					description,
				});
				break;
			case "event":
				await this.plugin.apiClient.updateEvent(entity.match.source_id, {
					name: entity.name,
					description,
				});
				break;
			default:
				new Notice(`Unsupported type: ${entity.type}`, 4000);
				return;
		}

		new Notice(`Updated ${entity.type}: ${entity.name}`, 3000);
	}
}
