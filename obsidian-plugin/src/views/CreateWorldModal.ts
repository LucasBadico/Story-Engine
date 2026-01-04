import { App, Modal, Notice, Setting } from "obsidian";

export class CreateWorldModal extends Modal {
	name: string = "";
	description: string = "";
	genre: string = "";
	onSubmit: (name: string, description: string, genre: string) => void;

	constructor(app: App, onSubmit: (name: string, description: string, genre: string) => void) {
		super(app);
		this.onSubmit = onSubmit;
	}

	onOpen() {
		const { contentEl } = this;

		contentEl.createEl("h2", { text: "Create New World" });

		// Name input
		new Setting(contentEl)
			.setName("World Name")
			.setDesc("Enter the name for your new world")
			.addText((text) =>
				text
					.setPlaceholder("My New World")
					.setValue(this.name)
					.onChange((value) => {
						this.name = value;
					})
					.inputEl.addEventListener("keypress", (e: KeyboardEvent) => {
						if (e.key === "Enter") {
							const descriptionInput = contentEl.querySelector("textarea") as HTMLTextAreaElement | null;
							if (descriptionInput) {
								descriptionInput.focus();
							}
						}
					})
			);

		// Description input
		new Setting(contentEl)
			.setName("Description")
			.setDesc("Enter a description for your world")
			.addTextArea((text) =>
				text
					.setPlaceholder("A brief description of your world...")
					.setValue(this.description)
					.onChange((value) => {
						this.description = value;
					})
			);

		// Genre input
		new Setting(contentEl)
			.setName("Genre")
			.setDesc("Enter the genre of your world")
			.addText((text) =>
				text
					.setPlaceholder("Fantasy, Sci-Fi, Contemporary, etc.")
					.setValue(this.genre)
					.onChange((value) => {
						this.genre = value;
					})
					.inputEl.addEventListener("keypress", (e: KeyboardEvent) => {
						if (e.key === "Enter") {
							this.submit();
						}
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

		// Focus on name input
		const nameInput = (contentEl as HTMLElement).querySelector("input") as HTMLInputElement | null;
		if (nameInput) {
			nameInput.focus();
		}
	}

	submit() {
		const trimmedName = this.name.trim();
		const trimmedDescription = this.description.trim();
		const trimmedGenre = this.genre.trim();
		
		if (!trimmedName) {
			new Notice("Please enter a world name", 3000);
			return;
		}

		if (!trimmedGenre) {
			new Notice("Please enter a genre", 3000);
			return;
		}

		this.close();
		this.onSubmit(trimmedName, trimmedDescription, trimmedGenre);
	}

	onClose() {
		const { contentEl } = this;
		contentEl.empty();
	}
}

