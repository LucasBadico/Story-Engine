import { Notice, Modal, setIcon } from "obsidian";
import StoryEnginePlugin from "../main";
import {
	Character,
	CharacterTrait,
	CharacterRelationship,
	WorldEvent,
	Scene,
	Story,
	Archetype,
	Trait,
	World,
} from "../types";

export class CharacterDetailsView {
	private plugin: StoryEnginePlugin;
	character: Character; // Made public for access from StoryListView
	private world: World | null;
	private characters: Character[];
	private archetypes: Archetype[];
	private traits: Trait[];
	private events: WorldEvent[];
	private headerEl: HTMLElement;
	private contentEl: HTMLElement;
	private onBack: () => void;
	private onEditCharacter: (character: Character) => void;

	// State
	private characterTab: "overview" | "traits" | "events" | "scenes" | "relationships" = "overview";
	private characterTraits: CharacterTrait[] = [];
	private characterEvents: { event: WorldEvent; role?: string }[] = [];
	private characterScenes: { scene: Scene; story: Story; type: "pov" | "coadjuvante" }[] = [];
	private characterRelationships: CharacterRelationship[] = [];

	constructor(
		plugin: StoryEnginePlugin,
		character: Character,
		headerEl: HTMLElement,
		contentEl: HTMLElement,
		onBack: () => void,
		onEditCharacter: (character: Character) => void,
		world: World | null = null,
		characters: Character[] = [],
		archetypes: Archetype[] = [],
		traits: Trait[] = [],
		events: WorldEvent[] = []
	) {
		this.plugin = plugin;
		this.character = character;
		this.world = world;
		this.characters = characters;
		this.archetypes = archetypes;
		this.traits = traits;
		this.events = events;
		this.headerEl = headerEl;
		this.contentEl = contentEl;
		this.onBack = onBack;
		this.onEditCharacter = onEditCharacter;
	}

	async render() {
		await this.loadCharacterData();
		this.renderHeader();
		this.renderContent();
	}

	private renderContent() {
		// Clear content completely
		this.contentEl.empty();
		
		// Create main container for tabs and content
		const mainContainer = this.contentEl.createDiv({ cls: "character-details-container" });
		
		// Render tabs
		this.renderTabs(mainContainer);
		
		// Render tab content
		this.renderTabContent(mainContainer);
	}

	async loadCharacterData() {
		try {
			// Load character traits
			this.characterTraits = await this.plugin.apiClient.getCharacterTraits(this.character.id);

			// Load character events
			const eventCharacters = await this.plugin.apiClient.getCharacterEvents(this.character.id);
			this.characterEvents = await Promise.all(
				eventCharacters.map(async (ec) => {
					const event = await this.plugin.apiClient.getEvent(ec.event_id);
					return { event, role: ec.role || undefined };
				})
			);

			// Load character relationships
			this.characterRelationships = await this.plugin.apiClient.getCharacterRelationships(this.character.id);

			// Load character scenes (workaround)
			this.characterScenes = [];
			if (this.world) {
				// Get all stories for this world
				const allStories = await this.plugin.apiClient.listStories();
				const worldStories = allStories.filter((s) => s.world_id === this.world!.id);

				for (const story of worldStories) {
					// Get all scenes for this story
					const scenes = await this.plugin.apiClient.getScenesByStory(story.id);

					// Filter POV scenes
					const povScenes = scenes.filter((s) => s.pov_character_id === this.character.id);
					for (const scene of povScenes) {
						this.characterScenes.push({ scene, story, type: "pov" });
					}

					// Filter coadjuvante scenes (via SceneReferences)
					for (const scene of scenes) {
						const refs = await this.plugin.apiClient.getSceneReferences(scene.id);
						const charRef = refs.find(
							(r) => r.entity_type === "character" && r.entity_id === this.character.id
						);
						if (charRef) {
							this.characterScenes.push({ scene, story, type: "coadjuvante" });
						}
					}
				}
			}
		} catch (err) {
			const errorMessage = err instanceof Error ? err.message : "Failed to load character data";
			new Notice(`Error: ${errorMessage}`, 5000);
		}
	}

	renderHeader() {
		if (!this.headerEl) return;

		this.headerEl.empty();

		const headerLeft = this.headerEl.createDiv({ cls: "story-engine-header-left" });

		const backButton = headerLeft.createEl("button", {
			text: "← Back",
			cls: "story-engine-back-btn",
		});
		backButton.onclick = () => {
			this.onBack();
		};

		const titleDiv = headerLeft.createDiv({ cls: "story-engine-header-title" });
		const titleRow = titleDiv.createDiv({ cls: "story-engine-title-row" });
		titleRow.createEl("h2", { text: this.character.name });

		// Archetype pill
		if (this.character.archetype_id) {
			const archetype = this.archetypes.find((a) => a.id === this.character.archetype_id);
			if (archetype) {
				const archetypePill = titleRow.createSpan({ cls: "story-engine-badge" });
				archetypePill.textContent = archetype.name;
			}
		}

		// Copy ID button with tooltip
		const copyIdBtn = titleRow.createEl("button", { 
			cls: "story-engine-copy-id-btn",
			attr: { 
				"aria-label": "Copy ID",
				"title": `ID: ${this.character.id}\nClick to copy`
			}
		});
		setIcon(copyIdBtn, "copy");
		copyIdBtn.onclick = () => {
			navigator.clipboard.writeText(this.character.id);
			new Notice("UUID copied to clipboard");
		};

		const headerRight = this.headerEl.createDiv({ cls: "story-engine-header-right" });
		const menuButton = headerRight.createEl("button", { text: "Edit Character" });
		menuButton.onclick = () => {
			this.onEditCharacter(this.character);
		};
	}

	private renderTabs(container: HTMLElement) {
		const tabsContainer = container.createDiv({ cls: "character-details-tabs" });
		
		type TabKey = "overview" | "traits" | "events" | "scenes" | "relationships";
		const tabs: { key: TabKey; label: string }[] = [
			{ key: "overview", label: "Overview" },
			{ key: "traits", label: "Traits" },
			{ key: "events", label: "Events" },
			{ key: "scenes", label: "Scenes" },
			{ key: "relationships", label: "Relationships" },
		];

		tabs.forEach(tab => {
			const isActive = this.characterTab === tab.key;
			const tabButton = tabsContainer.createEl("button", {
				text: tab.label,
				cls: isActive ? "character-tab character-tab-active" : "character-tab",
			});
			
			tabButton.onclick = () => {
				if (this.characterTab !== tab.key) {
					this.characterTab = tab.key;
					this.renderContent();
				}
			};
		});
	}

	private renderTabContent(container: HTMLElement) {
		const contentContainer = container.createDiv({ cls: "character-details-content-inner" });

		switch (this.characterTab) {
			case "overview":
				this.renderOverviewTab(contentContainer);
				break;
			case "traits":
				this.renderTraitsTab(contentContainer);
				break;
			case "events":
				this.renderEventsTab(contentContainer);
				break;
			case "scenes":
				this.renderScenesTab(contentContainer);
				break;
			case "relationships":
				this.renderRelationshipsTab(contentContainer);
				break;
		}
	}

	renderOverviewTab(container: HTMLElement) {
		// Description (editable textarea)
		const descSection = container.createDiv({ cls: "story-engine-section" });
		descSection.createEl("h3", { text: "Description" });
		const descTextarea = descSection.createEl("textarea", {
			cls: "story-engine-textarea",
			attr: { rows: "5" },
		});
		descTextarea.value = this.character.description || "";
		descTextarea.placeholder = "Enter character description...";

		const descActions = descSection.createDiv({ cls: "story-engine-actions" });
		const saveDescBtn = descActions.createEl("button", { text: "Save Description", cls: "mod-cta" });
		saveDescBtn.onclick = async () => {
			try {
				const updated = await this.plugin.apiClient.updateCharacter(this.character.id, {
					description: descTextarea.value,
				});
				this.character = updated;
				new Notice("Description saved");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};

		// Archetype dropdown
		const archetypeSection = container.createDiv({ cls: "story-engine-section" });
		archetypeSection.createEl("h3", { text: "Archetype" });
		const archetypeSelect = archetypeSection.createEl("select", { cls: "story-engine-select" });

		// Add "None" option
		const noneOption = archetypeSelect.createEl("option", { text: "None", value: "" });
		if (!this.character.archetype_id) {
			noneOption.selected = true;
		}

		// Add archetype options
		for (const archetype of this.archetypes) {
			const option = archetypeSelect.createEl("option", {
				text: archetype.name,
				value: archetype.id,
			});
			if (this.character.archetype_id === archetype.id) {
				option.selected = true;
			}
		}

		const archetypeActions = archetypeSection.createDiv({ cls: "story-engine-actions" });
		const saveArchetypeBtn = archetypeActions.createEl("button", { text: "Save Archetype", cls: "mod-cta" });
		saveArchetypeBtn.onclick = async () => {
			try {
				const archetypeId = archetypeSelect.value || null;
				const updated = await this.plugin.apiClient.updateCharacter(this.character.id, {
					archetype_id: archetypeId,
				});
				this.character = updated;
				await this.loadCharacterData();
				this.renderHeader();
				this.renderContent();
				new Notice("Archetype saved");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};
	}

	renderTraitsTab(container: HTMLElement) {
		const actionsBar = container.createDiv({ cls: "story-engine-actions-bar" });
		const addBtn = actionsBar.createEl("button", { text: "Add Trait", cls: "mod-cta" });
		addBtn.onclick = () => {
			this.showAddTraitModal();
		};

		if (this.characterTraits.length === 0) {
			container.createEl("p", { text: "No traits assigned. Add a trait to get started!" });
			return;
		}

		const list = container.createDiv({ cls: "story-engine-list" });
		for (const charTrait of this.characterTraits) {
			const item = list.createDiv({ cls: "story-engine-item" });

			const titleRow = item.createDiv({ cls: "story-engine-title" });
			titleRow.createEl("strong", { text: charTrait.trait_name });
			const categoryPill = titleRow.createSpan({ cls: "story-engine-badge" });
			categoryPill.textContent = charTrait.trait_category;

			const meta = item.createDiv({ cls: "story-engine-meta" });
			meta.createEl("span", { text: `Value: ${charTrait.value || "N/A"}` });
			if (charTrait.notes) {
				meta.createEl("span", { text: ` | Notes: ${charTrait.notes}` });
			}

			const actions = item.createDiv({ cls: "story-engine-item-actions" });
			actions.createEl("button", { text: "Edit" }).onclick = () => {
				this.showEditTraitModal(charTrait);
			};
			actions.createEl("button", { text: "Delete" }).onclick = async () => {
				if (confirm(`Remove trait "${charTrait.trait_name}"?`)) {
					try {
						await this.plugin.apiClient.removeCharacterTrait(this.character.id, charTrait.trait_id);
						await this.loadCharacterData();
						this.renderContent();
						new Notice("Trait removed");
					} catch (err) {
						new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
					}
				}
			};
		}
	}

	renderEventsTab(container: HTMLElement) {
		const actionsBar = container.createDiv({ cls: "story-engine-actions-bar" });
		const addBtn = actionsBar.createEl("button", { text: "Add Event", cls: "mod-cta" });
		addBtn.onclick = () => {
			this.showAddEventModal();
		};

		if (this.characterEvents.length === 0) {
			container.createEl("p", { text: "No events assigned. Add an event to get started!" });
			return;
		}

		const list = container.createDiv({ cls: "story-engine-list" });
		for (const { event, role } of this.characterEvents) {
			const item = list.createDiv({ cls: "story-engine-item" });
			item.createDiv({ cls: "story-engine-title", text: event.name });

			const meta = item.createDiv({ cls: "story-engine-meta" });
			if (role) {
				meta.createEl("span", { text: `Role: ${role}` });
			}

			const actions = item.createDiv({ cls: "story-engine-item-actions" });
			actions.createEl("button", { text: "Edit Role" }).onclick = () => {
				this.showEditEventRoleModal(event, role);
			};
			actions.createEl("button", { text: "Remove" }).onclick = async () => {
				if (confirm(`Remove character from event "${event.name}"?`)) {
					try {
						await this.plugin.apiClient.removeEventCharacter(event.id, this.character.id);
						await this.loadCharacterData();
						this.renderContent();
						new Notice("Character removed from event");
					} catch (err) {
						new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
					}
				}
			};
		}
	}

	renderScenesTab(container: HTMLElement) {
		if (this.characterScenes.length === 0) {
			container.createEl("p", { text: "No scenes found for this character." });
			return;
		}

		const list = container.createDiv({ cls: "story-engine-list" });
		for (const { scene, story, type } of this.characterScenes) {
			const item = list.createDiv({ cls: "story-engine-item" });

			const titleRow = item.createDiv({ cls: "story-engine-title" });
			titleRow.createEl("span", { text: scene.goal || "No goal" });

			const typePill = titleRow.createSpan({
				cls:
					type === "pov"
						? "story-engine-badge story-engine-badge-green"
						: "story-engine-badge story-engine-badge-blue",
			});
			typePill.textContent = type === "pov" ? "POV" : "Coadjuvante";

			const meta = item.createDiv({ cls: "story-engine-meta" });
			meta.createEl("span", { text: `Story: ${story.title}` });
		}
	}

	renderRelationshipsTab(container: HTMLElement) {
		const actionsBar = container.createDiv({ cls: "story-engine-actions-bar" });
		const addBtn = actionsBar.createEl("button", { text: "Add Relationship", cls: "mod-cta" });
		addBtn.onclick = () => {
			this.showAddRelationshipModal();
		};

		if (this.characterRelationships.length === 0) {
			container.createEl("p", { text: "No relationships defined. Add a relationship to get started!" });
			return;
		}

		const list = container.createDiv({ cls: "story-engine-list" });
		for (const rel of this.characterRelationships) {
			// Determine the other character
			const otherCharId =
				rel.character1_id === this.character.id ? rel.character2_id : rel.character1_id;
			const otherChar = this.characters.find((c) => c.id === otherCharId);
			const otherCharName = otherChar ? otherChar.name : "Unknown";

			const item = list.createDiv({ cls: "story-engine-item" });

			const titleRow = item.createDiv({ cls: "story-engine-title" });
			titleRow.createEl("strong", { text: otherCharName });

			const typePill = titleRow.createSpan({ cls: "story-engine-badge" });
			typePill.textContent = rel.relationship_type;

			const directionIcon = titleRow.createSpan({ cls: "story-engine-direction" });
			directionIcon.textContent = rel.bidirectional ? "↔" : "→";

			const meta = item.createDiv({ cls: "story-engine-meta" });
			if (rel.description) {
				meta.createEl("span", { text: rel.description });
			}

			const actions = item.createDiv({ cls: "story-engine-item-actions" });
			actions.createEl("button", { text: "Edit" }).onclick = () => {
				this.showEditRelationshipModal(rel);
			};
			actions.createEl("button", { text: "Delete" }).onclick = async () => {
				if (confirm(`Delete relationship with "${otherCharName}"?`)) {
					try {
						await this.plugin.apiClient.deleteCharacterRelationship(rel.id);
						await this.loadCharacterData();
						this.renderContent();
						new Notice("Relationship deleted");
					} catch (err) {
						new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
					}
				}
			};
		}
	}

	// Modal methods
	showAddTraitModal() {
		const modal = new Modal(this.plugin.app);
		modal.titleEl.textContent = "Add Trait";

		const content = modal.contentEl;
		content.createEl("p", { text: "Select a trait to add:" });

		const traitSelect = content.createEl("select", { cls: "story-engine-select" });
		const availableTraits = this.traits.filter(
			(t) => !this.characterTraits.some((ct) => ct.trait_id === t.id)
		);

		if (availableTraits.length === 0) {
			content.createEl("p", { text: "No available traits. Create a trait first!" });
			const buttonContainer = content.createDiv({ cls: "modal-button-container" });
			const closeBtn = buttonContainer.createEl("button", { text: "Close" });
			closeBtn.onclick = () => modal.close();
			return;
		}

		for (const trait of availableTraits) {
			traitSelect.createEl("option", { text: `${trait.name} (${trait.category})`, value: trait.id });
		}

		const valueInput = content.createEl("input", {
			cls: "story-engine-input",
			attr: { type: "text", placeholder: "Value (optional)" },
		});

		const notesInput = content.createEl("textarea", {
			cls: "story-engine-textarea",
			attr: { rows: "3", placeholder: "Notes (optional)" },
		});

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Add", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			try {
				await this.plugin.apiClient.addCharacterTrait(
					this.character.id,
					traitSelect.value,
					valueInput.value || undefined,
					notesInput.value || undefined
				);
				await this.loadCharacterData();
				this.renderContent();
				modal.close();
				new Notice("Trait added");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};

		const cancelBtn = buttonContainer.createEl("button", { text: "Cancel" });
		cancelBtn.onclick = () => modal.close();

		modal.open();
	}

	showEditTraitModal(charTrait: CharacterTrait) {
		const modal = new Modal(this.plugin.app);
		modal.titleEl.textContent = "Edit Trait";

		const content = modal.contentEl;
		content.createEl("p", { text: `Editing: ${charTrait.trait_name}` });

		const valueInput = content.createEl("input", {
			cls: "story-engine-input",
			attr: { type: "text", placeholder: "Value" },
		});
		valueInput.value = charTrait.value || "";

		const notesInput = content.createEl("textarea", {
			cls: "story-engine-textarea",
			attr: { rows: "3", placeholder: "Notes" },
		});
		notesInput.value = charTrait.notes || "";

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Save", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			try {
				await this.plugin.apiClient.updateCharacterTrait(
					this.character.id,
					charTrait.trait_id,
					valueInput.value || undefined,
					notesInput.value || undefined
				);
				await this.loadCharacterData();
				this.renderContent();
				modal.close();
				new Notice("Trait updated");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};

		const cancelBtn = buttonContainer.createEl("button", { text: "Cancel" });
		cancelBtn.onclick = () => modal.close();

		modal.open();
	}

	showAddEventModal() {
		if (!this.world) return;

		const modal = new Modal(this.plugin.app);
		modal.titleEl.textContent = "Add Event";

		const content = modal.contentEl;
		content.createEl("p", { text: "Select an event:" });

		const eventSelect = content.createEl("select", { cls: "story-engine-select" });
		const availableEvents = this.events.filter(
			(e) => !this.characterEvents.some((ce) => ce.event.id === e.id)
		);

		if (availableEvents.length === 0) {
			content.createEl("p", { text: "No available events. Create an event first!" });
			const buttonContainer = content.createDiv({ cls: "modal-button-container" });
			const closeBtn = buttonContainer.createEl("button", { text: "Close" });
			closeBtn.onclick = () => modal.close();
			return;
		}

		for (const event of availableEvents) {
			eventSelect.createEl("option", { text: event.name, value: event.id });
		}

		const roleInput = content.createEl("input", {
			cls: "story-engine-input",
			attr: { type: "text", placeholder: "Role (optional)" },
		});

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Add", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			try {
				await this.plugin.apiClient.addEventCharacter(
					eventSelect.value,
					this.character.id,
					roleInput.value || undefined
				);
				await this.loadCharacterData();
				this.renderContent();
				modal.close();
				new Notice("Character added to event");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};

		const cancelBtn = buttonContainer.createEl("button", { text: "Cancel" });
		cancelBtn.onclick = () => modal.close();

		modal.open();
	}

	showEditEventRoleModal(event: WorldEvent, currentRole?: string) {
		const modal = new Modal(this.plugin.app);
		modal.titleEl.textContent = "Edit Role";

		const content = modal.contentEl;
		content.createEl("p", { text: `Event: ${event.name}` });

		const roleInput = content.createEl("input", {
			cls: "story-engine-input",
			attr: { type: "text", placeholder: "Role" },
			value: currentRole || "",
		});

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Save", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			try {
				await this.plugin.apiClient.removeEventCharacter(event.id, this.character.id);
				await this.plugin.apiClient.addEventCharacter(
					event.id,
					this.character.id,
					roleInput.value || undefined
				);
				await this.loadCharacterData();
				this.renderContent();
				modal.close();
				new Notice("Role updated");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};

		const cancelBtn = buttonContainer.createEl("button", { text: "Cancel" });
		cancelBtn.onclick = () => modal.close();

		modal.open();
	}

	showAddRelationshipModal() {
		const modal = new Modal(this.plugin.app);
		modal.titleEl.textContent = "Add Relationship";

		const content = modal.contentEl;
		content.createEl("p", { text: "Select another character:" });

		const characterSelect = content.createEl("select", { cls: "story-engine-select" });
		const availableCharacters = this.characters.filter((c) => c.id !== this.character.id);

		if (availableCharacters.length === 0) {
			content.createEl("p", { text: "No other characters available." });
			const buttonContainer = content.createDiv({ cls: "modal-button-container" });
			const closeBtn = buttonContainer.createEl("button", { text: "Close" });
			closeBtn.onclick = () => modal.close();
			return;
		}

		for (const char of availableCharacters) {
			characterSelect.createEl("option", { text: char.name, value: char.id });
		}

		const relationshipTypeSelect = content.createEl("select", { cls: "story-engine-select" });
		relationshipTypeSelect.createEl("option", { text: "Ally", value: "ally" });
		relationshipTypeSelect.createEl("option", { text: "Enemy", value: "enemy" });
		relationshipTypeSelect.createEl("option", { text: "Family", value: "family" });
		relationshipTypeSelect.createEl("option", { text: "Lover", value: "lover" });
		relationshipTypeSelect.createEl("option", { text: "Rival", value: "rival" });
		relationshipTypeSelect.createEl("option", { text: "Mentor", value: "mentor" });
		relationshipTypeSelect.createEl("option", { text: "Student", value: "student" });

		const descriptionInput = content.createEl("textarea", {
			cls: "story-engine-textarea",
			attr: { rows: "3", placeholder: "Description (optional)" },
		});

		const bidirectionalCheckbox = content.createEl("input", {
			attr: { type: "checkbox" },
		});
		bidirectionalCheckbox.checked = true;
		const bidirectionalLabel = content.createEl("label");
		bidirectionalLabel.appendChild(bidirectionalCheckbox);
		bidirectionalLabel.appendChild(document.createTextNode(" Bidirectional"));

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Add", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			try {
				await this.plugin.apiClient.createCharacterRelationship(this.character.id, {
					character1_id: this.character.id,
					character2_id: characterSelect.value,
					relationship_type: relationshipTypeSelect.value,
					description: descriptionInput.value || "",
					bidirectional: bidirectionalCheckbox.checked,
				});
				await this.loadCharacterData();
			this.renderContent();
				modal.close();
				new Notice("Relationship added");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};

		const cancelBtn = buttonContainer.createEl("button", { text: "Cancel" });
		cancelBtn.onclick = () => modal.close();

		modal.open();
	}

	showEditRelationshipModal(rel: CharacterRelationship) {
		const modal = new Modal(this.plugin.app);
		modal.titleEl.textContent = "Edit Relationship";

		const content = modal.contentEl;

		const relationshipTypeSelect = content.createEl("select", { cls: "story-engine-select" });
		relationshipTypeSelect.createEl("option", { text: "Ally", value: "ally" });
		relationshipTypeSelect.createEl("option", { text: "Enemy", value: "enemy" });
		relationshipTypeSelect.createEl("option", { text: "Family", value: "family" });
		relationshipTypeSelect.createEl("option", { text: "Lover", value: "lover" });
		relationshipTypeSelect.createEl("option", { text: "Rival", value: "rival" });
		relationshipTypeSelect.createEl("option", { text: "Mentor", value: "mentor" });
		relationshipTypeSelect.createEl("option", { text: "Student", value: "student" });
		relationshipTypeSelect.value = rel.relationship_type;

		const descriptionInput = content.createEl("textarea", {
			cls: "story-engine-textarea",
			attr: { rows: "3", placeholder: "Description" },
			text: rel.description || "",
		});

		const bidirectionalCheckbox = content.createEl("input", {
			attr: { type: "checkbox" },
		});
		bidirectionalCheckbox.checked = rel.bidirectional;
		const bidirectionalLabel = content.createEl("label");
		bidirectionalLabel.appendChild(bidirectionalCheckbox);
		bidirectionalLabel.appendChild(document.createTextNode(" Bidirectional"));

		const buttonContainer = content.createDiv({ cls: "modal-button-container" });
		const saveBtn = buttonContainer.createEl("button", { text: "Save", cls: "mod-cta" });
		saveBtn.onclick = async () => {
			try {
				await this.plugin.apiClient.updateCharacterRelationship(rel.id, {
					relationship_type: relationshipTypeSelect.value,
					description: descriptionInput.value || "",
					bidirectional: bidirectionalCheckbox.checked,
				});
				await this.loadCharacterData();
				this.renderContent();
				modal.close();
				new Notice("Relationship updated");
			} catch (err) {
				new Notice(`Error: ${err instanceof Error ? err.message : "Failed"}`, 5000);
			}
		};

		const cancelBtn = buttonContainer.createEl("button", { text: "Cancel" });
		cancelBtn.onclick = () => modal.close();

		modal.open();
	}

	// Method to update character after edit
	updateCharacter(character: Character) {
		this.character = character;
		this.renderHeader();
		this.renderContent();
	}
}

