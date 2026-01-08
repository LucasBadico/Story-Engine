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
	private activeTab: "entities" | "progress" = "progress";
	private expandedLogs = new Set<string>();

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

		if (this.activeTab === "entities" && this.status === "running") {
			this.activeTab = "progress";
		}

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

		const tabs = this.contentRoot.createDiv({
			cls: "story-engine-extract-tabs",
		});

		const hasEntities = this.result.entities.length > 0;
		const entitiesDisabled = !hasEntities || this.status === "running";
		const entitiesTab = tabs.createEl("button", {
			text: "Entities Found",
			cls: `story-engine-extract-tab ${
				this.activeTab === "entities" ? "is-active" : ""
			}`,
		});
		if (entitiesDisabled) {
			entitiesTab.disabled = true;
			entitiesTab.addClass("is-disabled");
			if (this.activeTab === "entities") {
				this.activeTab = "progress";
			}
		}
		entitiesTab.onclick = () => {
			if (entitiesDisabled) return;
			this.activeTab = "entities";
			this.render();
		};

		const progressTab = tabs.createEl("button", {
			text: "Progress",
			cls: `story-engine-extract-tab ${
				this.activeTab === "progress" ? "is-active" : ""
			}`,
		});
		progressTab.onclick = () => {
			this.activeTab = "progress";
			this.render();
		};

		const panels = this.contentRoot.createDiv({
			cls: "story-engine-extract-panels",
		});

		const entitiesPanel = panels.createDiv({
			cls: `story-engine-extract-panel ${
				this.activeTab === "entities" ? "is-active" : ""
			}`,
		});

		const progressPanel = panels.createDiv({
			cls: `story-engine-extract-panel ${
				this.activeTab === "progress" ? "is-active" : ""
			}`,
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

		this.renderLogs(progressPanel);

		if (hasEntities) {
			const list = entitiesPanel.createDiv({
				cls: "story-engine-extract-list",
			});
			this.result.entities.forEach((entity, index) => {
				this.renderEntity(list, entity, index);
			});
		} else {
			entitiesPanel.createEl("p", {
				text: "No entities returned from extraction.",
				cls: "story-engine-extract-empty",
			});
		}
	}

	private renderLogs(container?: HTMLElement) {
		const root = container ?? this.contentRoot;
		const logBlock = root.createDiv({
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
		const lastIndex = this.logs.length - 1;
		this.logs.forEach((entry, index) => {
			const eventLabel = entry.phase
				? `${entry.phase} · ${entry.eventType}`
				: entry.eventType;
			const item = list.createDiv({
				cls: `story-engine-extract-log-item ${
					index === lastIndex ? "is-latest" : ""
				}`,
			});
			const header = item.createDiv({
				cls: "story-engine-extract-log-header",
			});
			header.createSpan({
				text: `${entry.timestamp} | ${eventLabel}`,
				cls: "story-engine-extract-log-title",
			});

			const label = item.createDiv({
				cls: "story-engine-extract-log-label",
			});
			label.createSpan({ text: entry.message });
			const toggle = label.createSpan({
				text: index === lastIndex
					? "..."
					: this.expandedLogs.has(entry.id)
						? "−"
						: "+",
				cls: "story-engine-extract-log-toggle",
			});
			label.onclick = () => {
				if (this.expandedLogs.has(entry.id)) {
					this.expandedLogs.delete(entry.id);
				} else {
					this.expandedLogs.add(entry.id);
				}
				this.render();
			};
			if (index === lastIndex) {
				toggle.addClass("is-typing");
			}

			if (this.expandedLogs.has(entry.id) && entry.data) {
				item.createEl("pre", {
					text: JSON.stringify(entry.data, null, 2),
					cls: "story-engine-extract-log-data",
				});
			}
		});
		list.scrollTop = list.scrollHeight;
	}

	private renderEntity(container: HTMLElement, entity: ExtractEntity, index: number) {
		const item = container.createDiv({ cls: "story-engine-extract-item" });
		const header = item.createDiv({ cls: "story-engine-extract-item-header" });
		header.createEl("div", {
			text: `#${index + 1}`,
			cls: "story-engine-extract-rank",
		});
		const statusText = entity.created
			? "Created"
			: entity.found
				? "Found"
				: "New";
		const statusClass = entity.created
			? "is-created"
			: entity.found
				? "is-found"
				: "is-new";
		header.createEl("div", {
			text: statusText,
			cls: `story-engine-extract-status ${statusClass}`,
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

		if (!entity.created) {
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

		entity.found = false;
		entity.created = true;
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
