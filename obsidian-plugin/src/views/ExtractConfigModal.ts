import { App, Modal, Notice } from "obsidian";

export interface ExtractConfigResult {
	entityTypes: string[];
	includeRelations: boolean;
}

const ENTITY_TYPE_OPTIONS = [
	{ value: "character", label: "Character" },
	{ value: "location", label: "Location" },
	{ value: "artefact", label: "Artefact" },
	{ value: "faction", label: "Faction" },
	{ value: "event", label: "Event" },
];

export class ExtractConfigModal extends Modal {
	private resolve: ((value: ExtractConfigResult | null) => void) | null = null;
	private selectedTypes = new Set<string>();
	private includeRelations = true;

	constructor(
		app: App,
		defaultTypes: string[],
		defaultIncludeRelations: boolean
	) {
		super(app);
		defaultTypes.forEach((type) => this.selectedTypes.add(type));
		this.includeRelations = defaultIncludeRelations;
	}

	static open(
		app: App,
		defaultTypes: string[],
		defaultIncludeRelations: boolean
	): Promise<ExtractConfigResult | null> {
		return new Promise((resolve) => {
			const modal = new ExtractConfigModal(
				app,
				defaultTypes,
				defaultIncludeRelations
			);
			modal.resolve = resolve;
			modal.open();
		});
	}

	onOpen() {
		const { contentEl } = this;
		contentEl.empty();

		contentEl.createEl("h3", { text: "Extract Configuration" });
		contentEl.createEl("p", {
			text: "Choose which entity types to extract and whether to include relations.",
		});

		const typeBlock = contentEl.createDiv({ cls: "story-engine-extract-config" });
		typeBlock.createEl("div", {
			text: "Entity types",
			cls: "story-engine-extract-label",
		});

		ENTITY_TYPE_OPTIONS.forEach((option) => {
			const row = typeBlock.createDiv({
				cls: "story-engine-extract-config-row",
			});
			const checkbox = row.createEl("input", {
				type: "checkbox",
			});
			checkbox.checked = this.selectedTypes.has(option.value);
			checkbox.onchange = () => {
				if (checkbox.checked) {
					this.selectedTypes.add(option.value);
				} else {
					this.selectedTypes.delete(option.value);
				}
			};
			row.createEl("span", { text: option.label });
		});

		const relationRow = contentEl.createDiv({
			cls: "story-engine-extract-config-row",
		});
		const relationsCheckbox = relationRow.createEl("input", {
			type: "checkbox",
		});
		relationsCheckbox.checked = this.includeRelations;
		relationsCheckbox.onchange = () => {
			this.includeRelations = relationsCheckbox.checked;
		};
		relationRow.createEl("span", { text: "Include relations" });

		const buttonContainer = contentEl.createDiv({
			cls: "modal-button-container",
		});
		const submitButton = buttonContainer.createEl("button", {
			text: "Start Extraction",
			cls: "mod-cta",
		});
		submitButton.onclick = () => this.submit();

		const cancelButton = buttonContainer.createEl("button", {
			text: "Cancel",
		});
		cancelButton.onclick = () => this.cancel();
	}

	private submit() {
		const types = Array.from(this.selectedTypes);
		if (types.length === 0) {
			new Notice("Select at least one entity type.", 3000);
			return;
		}
		if (this.resolve) {
			this.resolve({
				entityTypes: types,
				includeRelations: this.includeRelations,
			});
		}
		this.close();
	}

	private cancel() {
		if (this.resolve) {
			this.resolve(null);
		}
		this.close();
	}

	onClose() {
		const { contentEl } = this;
		contentEl.empty();
	}
}
