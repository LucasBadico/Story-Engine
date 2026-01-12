import { ItemView, Notice, WorkspaceLeaf } from "obsidian";
import StoryEnginePlugin from "../main";
import {
	ExtractEntity,
	ExtractEntityMatch,
	ExtractEntityResult,
	ExtractRelation,
	ExtractRelationNode,
} from "../types";

export const STORY_ENGINE_EXTRACT_VIEW_TYPE = "story-engine-extract-view";

export class StoryEngineExtractView extends ItemView {
	plugin: StoryEnginePlugin;
	private result: ExtractEntityResult | null = null;
	private contentRoot!: HTMLElement;
	private logs = [] as StoryEnginePlugin["extractLogs"];
	private status: StoryEnginePlugin["extractStatus"] = "idle";
	private activeTab: "entities" | "relations" | "progress" = "progress";
	private expandedLogs = new Set<string>();
	private expandedRelationIndex: number | null = null;
	private pendingRelationScrollIndex: number | null = null;
	private pendingTabScroll = false;

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
				text: `Relations: ${this.result.relations.length}`,
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
		const hasRelations = this.result.relations.length > 0;
		const includeRelations = this.result.include_relations !== false;
		const entitiesDisabled = !hasEntities && this.status === "idle";
		const relationsDisabled = (!hasRelations && this.status === "idle")
			|| !includeRelations;

		const progressTab = tabs.createEl("button", {
			text: "Progress",
			cls: `story-engine-extract-tab ${
				this.activeTab === "progress" ? "is-active" : ""
			}`,
		});
		progressTab.onclick = () => {
			this.setActiveTab("progress");
		};

		const entitiesTab = tabs.createEl("button", {
			text: "Entities",
			cls: `story-engine-extract-tab ${
				this.activeTab === "entities" ? "is-active" : ""
			}`,
		});
		if (entitiesDisabled) {
			entitiesTab.disabled = true;
			entitiesTab.addClass("is-disabled");
		}
		entitiesTab.onclick = () => {
			if (entitiesDisabled) return;
			this.setActiveTab("entities");
		};

		const relationsTab = tabs.createEl("button", {
			text: "Relations",
			cls: `story-engine-extract-tab ${
				this.activeTab === "relations" ? "is-active" : ""
			}`,
		});
		if (relationsDisabled) {
			relationsTab.disabled = true;
			relationsTab.addClass("is-disabled");
		}
		relationsTab.onclick = () => {
			if (relationsDisabled) return;
			this.setActiveTab("relations");
		};

		const panels = this.contentRoot.createDiv({
			cls: "story-engine-extract-panels",
		});

		const entitiesPanel = panels.createDiv({
			cls: `story-engine-extract-panel ${
				this.activeTab === "entities" ? "is-active" : ""
			}`,
		});

		const relationsPanel = panels.createDiv({
			cls: `story-engine-extract-panel ${
				this.activeTab === "relations" ? "is-active" : ""
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
			const emptyText = this.status === "running"
				? "Extracting entities..."
				: "No entities returned from extraction.";
			entitiesPanel.createEl("p", {
				text: emptyText,
				cls: "story-engine-extract-empty",
			});
		}

		if (hasRelations) {
			const list = relationsPanel.createDiv({
				cls: "story-engine-extract-list",
			});
			this.result.relations.forEach((relation, index) => {
				this.renderRelation(list, relation, index);
			});
		} else {
			let emptyText = "No relations returned from extraction.";
			if (!includeRelations) {
				emptyText = "Relations not requested.";
			} else if (!hasEntities) {
				emptyText = "Waiting for entity extraction.";
			} else if (this.status === "running") {
				emptyText = "Extracting relations...";
			}
			relationsPanel.createEl("p", {
				text: emptyText,
				cls: "story-engine-extract-empty",
			});
		}

		this.applyPendingScroll();
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
		this.renderEntityCard(container, entity, index);
	}

	private renderEntityCard(
		container: HTMLElement,
		entity: ExtractEntity,
		index: number,
		onCreated?: (createdId: string) => void
	) {
		const item = container.createDiv({ cls: "story-engine-extract-item" });
		item.dataset.relationIndex = String(index);
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
						const createdId = await this.createEntity(entity);
						if (createdId && onCreated) {
							onCreated(createdId);
						}
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

	private async createEntity(entity: ExtractEntity): Promise<string> {
		if (!this.result?.world_id) {
			new Notice("World ID missing. Select a story or world first.", 4000);
			return "";
		}

		const description = entity.summary?.trim() || entity.name;
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
				return "";
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
		return createdId;
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

	private renderRelation(container: HTMLElement, relation: ExtractRelation, index: number) {
		const item = container.createDiv({ cls: "story-engine-extract-item" });
		const header = item.createDiv({ cls: "story-engine-extract-item-header" });
		header.createEl("div", {
			text: `#${index + 1}`,
			cls: "story-engine-extract-rank",
		});
		const statusText = relation.status === "pending_entities"
			? "Pending"
			: relation.status;
		const statusClass = relation.status === "pending_entities"
			? "is-new"
			: "is-found";
		header.createEl("div", {
			text: statusText,
			cls: `story-engine-extract-status ${statusClass}`,
		});

		const title = item.createDiv({ cls: "story-engine-extract-item-title" });
		const sourceLabel = relation.source.name || relation.source.ref;
		const targetLabel = relation.target.name || relation.target.ref;
		const relationLine = `#${index + 1} ${relation.relation_type} | ${sourceLabel} -> ${targetLabel}`;
		title.createEl("div", {
			text: relationLine,
			cls: "story-engine-extract-entity-name",
		});
		title.createEl("div", {
			text: relation.status,
			cls: "story-engine-extract-entity-type",
		});

		if (relation.summary) {
			item.createEl("div", {
				text: relation.summary,
				cls: "story-engine-extract-content",
			});
		}

		const hasPending = this.hasTempNode(relation.source)
			|| this.hasTempNode(relation.target);
		const relationCreated = relation.status === "created";

		const actions = item.createDiv({ cls: "story-engine-extract-actions" });
		const createRelationButton = actions.createEl("button", {
			text: relationCreated ? "Relation Created" : "Create Relation",
			cls: "story-engine-extract-action",
		});
		if (relationCreated) {
			createRelationButton.disabled = true;
			createRelationButton.addClass("is-disabled");
			createRelationButton.title = "Relation already created.";
		} else if (hasPending) {
			createRelationButton.disabled = true;
			createRelationButton.addClass("is-disabled");
			createRelationButton.title = "Create pending entities first.";
		} else {
			createRelationButton.onclick = async () => {
				createRelationButton.disabled = true;
				try {
					await this.createRelation(relation);
				} finally {
					createRelationButton.disabled = false;
				}
			};
		}
		if (hasPending) {
			const createButton = actions.createEl("button", {
				text: "Create Pending",
				cls: "story-engine-extract-action",
			});
			createButton.onclick = async () => {
				createButton.disabled = true;
				try {
					await this.createPendingEntities(relation);
				} finally {
					createButton.disabled = false;
				}
			};
		}

		const accordionButton = actions.createEl("button", {
			text: "See Entities",
			cls: "story-engine-extract-action",
		});
		accordionButton.onclick = () => {
			const shouldExpand = this.expandedRelationIndex !== index;
			this.expandedRelationIndex = shouldExpand ? index : null;
			this.pendingRelationScrollIndex = shouldExpand ? index : null;
			this.render();
		};

		if (this.expandedRelationIndex === index) {
			const entitiesWrap = item.createDiv({
				cls: "story-engine-extract-relations-entities",
			});
			entitiesWrap.dataset.relationIndex = String(index);
			const entities = [
				this.resolveRelationEntity(relation.source),
				this.resolveRelationEntity(relation.target),
			];
			entities.forEach((entity, entityIndex) => {
				this.renderEntityCard(entitiesWrap, entity, entityIndex, (createdId) => {
					this.applyCreatedIdToRelation(relation, entity, createdId);
				});
			});
		}
	}

	private resolveRelationEntity(node: ExtractRelationNode): ExtractEntity {
		if (!this.result) {
			return this.buildRelationEntity(node);
		}

		const existing = this.result.entities.find((entity) => {
			if (node.id && entity.match?.source_id === node.id) {
				return true;
			}
			return entity.type === node.type && entity.name === node.name;
		});

		return existing ?? this.buildRelationEntity(node);
	}

	private buildRelationEntity(node: ExtractRelationNode): ExtractEntity {
		const hasId = !!(node.id && node.id.trim());
		return {
			type: node.type,
			name: node.name || node.ref,
			found: hasId,
			match: hasId
				? {
					source_type: node.type,
					source_id: node.id as string,
					entity_name: node.name,
					similarity: 1,
					reason: "Relation match",
				}
				: undefined,
		};
	}

	private hasTempNode(node: ExtractRelationNode): boolean {
		if (!node.id || !node.id.trim()) {
			return true;
		}
		if (node.id.startsWith("temp")) {
			return true;
		}
		return node.ref.startsWith("finding:");
	}

	private async createPendingEntities(relation: ExtractRelation) {
		const nodes = [relation.source, relation.target];
		for (const node of nodes) {
			if (!this.hasTempNode(node)) {
				continue;
			}
			const entity = this.buildRelationEntity(node);
			const createdId = await this.createEntity(entity);
			if (createdId) {
				this.applyCreatedIdToRelation(relation, entity, createdId);
			}
		}
		this.render();
	}

	private async createRelation(relation: ExtractRelation) {
		if (!this.result?.world_id) {
			new Notice("World ID missing. Select a story or world first.", 4000);
			return;
		}
		if (!relation.source.id || !relation.target.id) {
			new Notice("Create the related entities before creating the relation.", 4000);
			return;
		}

		const created = await this.plugin.apiClient.createEntityRelation({
			world_id: this.result.world_id,
			source_type: relation.source.type,
			source_id: relation.source.id,
			target_type: relation.target.type,
			target_id: relation.target.id,
			relation_type: relation.relation_type,
			summary: relation.summary,
			create_mirror: relation.create_mirror ?? false,
		});

		relation.status = "created";
		relation.created_id = created.id;
		this.render();
		new Notice(`Relation created: ${relation.relation_type}`, 3000);
	}

	private applyCreatedIdToRelation(
		relation: ExtractRelation,
		entity: ExtractEntity,
		createdId: string
	) {
		const updateNode = (node: ExtractRelationNode) => {
			if (node.type !== entity.type) return false;
			if (entity.name && node.name && node.name !== entity.name) return false;
			node.id = createdId;
			node.name = entity.name;
			return true;
		};
		if (!updateNode(relation.source)) {
			updateNode(relation.target);
		}
		if (!this.hasTempNode(relation.source) && !this.hasTempNode(relation.target)) {
			relation.status = "ready";
		}
	}

	private setActiveTab(tab: "progress" | "entities" | "relations") {
		this.activeTab = tab;
		this.pendingTabScroll = true;
		this.render();
	}

	private applyPendingScroll() {
		if (!this.pendingTabScroll && this.pendingRelationScrollIndex === null) {
			return;
		}
		requestAnimationFrame(() => {
			if (this.pendingTabScroll) {
				const panel = this.contentRoot.querySelector(
					".story-engine-extract-panel.is-active"
				) as HTMLElement | null;
				if (panel) {
					panel.scrollTop = 0;
					panel.scrollIntoView({ block: "start" });
				}
				this.pendingTabScroll = false;
			}

			if (this.pendingRelationScrollIndex !== null) {
				const selector = `.story-engine-extract-relations-entities[data-relation-index="${this.pendingRelationScrollIndex}"]`;
				const entityBlock = this.contentRoot.querySelector(selector) as HTMLElement | null;
				if (entityBlock) {
					entityBlock.scrollIntoView({ block: "start" });
				}
				this.pendingRelationScrollIndex = null;
			}
		});
	}
}
