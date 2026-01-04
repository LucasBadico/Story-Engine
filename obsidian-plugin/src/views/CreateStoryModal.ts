import { App, Modal, Notice, Setting } from "obsidian";
import StoryEnginePlugin from "../main";
import { World } from "../types";
import { CreateWorldModal } from "./CreateWorldModal";

export class CreateStoryModal extends Modal {
	plugin: StoryEnginePlugin;
	title: string = "";
	selectedWorldId: string = "";
	shouldSync: boolean = true;
	worlds: World[] = [];
	onSubmit: (title: string, worldId: string | undefined, shouldSync: boolean) => void;

	constructor(app: App, plugin: StoryEnginePlugin, onSubmit: (title: string, worldId: string | undefined, shouldSync: boolean) => void) {
		super(app);
		this.plugin = plugin;
		this.onSubmit = onSubmit;
	}

	async onOpen() {
		const { contentEl } = this;

		contentEl.createEl("h2", { text: "Create New Story" });

		// Load worlds
		try {
			this.worlds = await this.plugin.apiClient.getWorlds();
		} catch (err) {
			console.error("Failed to load worlds:", err);
			this.worlds = [];
		}

		// Title input
		new Setting(contentEl)
			.setName("Story Title")
			.setDesc("Enter the title for your new story")
			.addText((text) =>
				text
					.setPlaceholder("My New Story")
					.setValue(this.title)
					.onChange((value) => {
						this.title = value;
					})
					.inputEl.addEventListener("keypress", (e: KeyboardEvent) => {
						if (e.key === "Enter") {
							this.submit();
						}
					})
			);

		// World selection
		const worldOptions: Record<string, string> = {
			"": "No World",
		};
		for (const world of this.worlds) {
			worldOptions[world.id] = world.name;
		}
		worldOptions["__create_new__"] = "Create new world...";

		new Setting(contentEl)
			.setName("World")
			.setDesc("Select a world for this story (optional)")
			.addDropdown((dropdown) => {
				for (const [value, label] of Object.entries(worldOptions)) {
					dropdown.addOption(value, label);
				}
				dropdown.setValue(this.selectedWorldId || "");
				dropdown.onChange(async (value) => {
					if (value === "__create_new__") {
						// Open create world modal
						new CreateWorldModal(this.app, async (name: string, description: string, genre: string) => {
							try {
								const newWorld = await this.plugin.apiClient.createWorld(name, description, genre);
								this.worlds.push(newWorld);
								// Refresh dropdown
								this.onClose();
								this.onOpen();
								// Re-select the newly created world
								this.selectedWorldId = newWorld.id;
							} catch (err) {
								const errorMessage = err instanceof Error ? err.message : "Failed to create world";
								new Notice(`Error: ${errorMessage}`, 5000);
							}
						}).open();
					} else {
						this.selectedWorldId = value;
					}
				});
			});

		// Sync checkbox
		new Setting(contentEl)
			.setName("Sync to Obsidian")
			.setDesc("Automatically sync the story files to your vault after creation")
			.addToggle((toggle) =>
				toggle
					.setValue(this.shouldSync)
					.onChange((value) => {
						this.shouldSync = value;
					})
			);

		// Buttons
		const buttonContainer = contentEl.createDiv({ cls: "modal-button-container" });
		
		const createButton = buttonContainer.createEl("button", {
			text: "Create",
			cls: "mod-cta",
		});
		createButton.addEventListener("click", () => this.submit());

		const cancelButton = buttonContainer.createEl("button", {
			text: "Cancel",
		});
		cancelButton.addEventListener("click", () => this.close());

		// Focus on title input
		const titleInput = (contentEl as HTMLElement).querySelector("input") as HTMLInputElement | null;
		if (titleInput) {
			titleInput.focus();
		}
	}

	submit() {
		const trimmedTitle = this.title.trim();
		
		if (!trimmedTitle) {
			new Notice("Please enter a story title", 3000);
			return;
		}

		this.close();
		const worldId = this.selectedWorldId && this.selectedWorldId !== "__create_new__" ? this.selectedWorldId : undefined;
		this.onSubmit(trimmedTitle, worldId, this.shouldSync);
	}

	onClose() {
		const { contentEl } = this;
		contentEl.empty();
	}
}

