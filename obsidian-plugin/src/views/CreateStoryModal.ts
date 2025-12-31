import { App, Modal, Notice, Setting } from "obsidian";

export class CreateStoryModal extends Modal {
	title: string = "";
	shouldSync: boolean = true;
	onSubmit: (title: string, shouldSync: boolean) => void;

	constructor(app: App, onSubmit: (title: string, shouldSync: boolean) => void) {
		super(app);
		this.onSubmit = onSubmit;
	}

	onOpen() {
		const { contentEl } = this;

		contentEl.createEl("h2", { text: "Create New Story" });

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
					.inputEl.addEventListener("keypress", (e) => {
						if (e.key === "Enter") {
							this.submit();
						}
					})
			);

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
		const titleInput = contentEl.querySelector("input");
		if (titleInput) {
			(titleInput as HTMLInputElement).focus();
		}
	}

	submit() {
		const trimmedTitle = this.title.trim();
		
		if (!trimmedTitle) {
			new Notice("Please enter a story title", 3000);
			return;
		}

		this.close();
		this.onSubmit(trimmedTitle, this.shouldSync);
	}

	onClose() {
		const { contentEl } = this;
		contentEl.empty();
	}
}

