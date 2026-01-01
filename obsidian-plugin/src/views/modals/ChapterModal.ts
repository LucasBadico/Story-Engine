import { App, Modal, Notice, Setting } from "obsidian";
import { Chapter } from "../../types";

export class ChapterModal extends Modal {
	chapter: Partial<Chapter> = {
		title: "",
		status: "draft",
	};
	isEdit: boolean = false;
	existingChapters: Chapter[] = [];
	onSubmit: (chapter: Partial<Chapter>) => Promise<void>;

	constructor(
		app: App,
		onSubmit: (chapter: Partial<Chapter>) => Promise<void>,
		existingChapters: Chapter[] = [],
		chapter?: Chapter
	) {
		super(app);
		this.onSubmit = onSubmit;
		this.existingChapters = existingChapters;
		if (chapter) {
			this.isEdit = true;
			this.chapter = {
				number: chapter.number,
				title: chapter.title,
				status: chapter.status,
			};
		}
	}

	onOpen() {
		const { contentEl } = this;
		contentEl.empty();

		contentEl.createEl("h2", {
			text: this.isEdit ? "Edit Chapter" : "Create Chapter",
		});

		if (this.isEdit) {
			new Setting(contentEl)
				.setName("Chapter Number")
				.setDesc("The chapter number")
				.addText((text) =>
					text
						.setPlaceholder("1")
						.setValue(this.chapter.number?.toString() || "1")
						.onChange((value) => {
							const num = parseInt(value);
							if (!isNaN(num) && num > 0) {
								this.chapter.number = num;
							}
						})
				);
		}

		new Setting(contentEl)
			.setName("Title")
			.setDesc("Chapter title")
			.addText((text) =>
				text
					.setPlaceholder("Chapter Title")
					.setValue(this.chapter.title || "")
					.onChange((value) => {
						this.chapter.title = value;
					})
					.inputEl.addEventListener("keypress", (e: KeyboardEvent) => {
						if (e.key === "Enter") {
							this.submit();
						}
					})
			);

		new Setting(contentEl)
			.setName("Status")
			.setDesc("Chapter status")
			.addDropdown((dropdown) =>
				dropdown
					.addOption("draft", "Draft")
					.addOption("in_progress", "In Progress")
					.addOption("completed", "Completed")
					.setValue(this.chapter.status || "draft")
					.onChange((value) => {
						this.chapter.status = value;
					})
			);

		const buttonContainer = contentEl.createDiv({ cls: "modal-button-container" });

		const submitButton = buttonContainer.createEl("button", {
			text: this.isEdit ? "Update" : "Create",
			cls: "mod-cta",
		});
		submitButton.addEventListener("click", () => this.submit());

		const cancelButton = buttonContainer.createEl("button", {
			text: "Cancel",
		});
		cancelButton.addEventListener("click", () => this.close());

		const titleInput = contentEl.querySelector("input[placeholder='Chapter Title']") as HTMLInputElement | null;
		if (titleInput) {
			titleInput.focus();
		}
	}

	async submit() {
		if (!this.chapter.title?.trim()) {
			new Notice("Please enter a chapter title", 3000);
			return;
		}

		// Auto-calculate chapter number if creating
		if (!this.isEdit) {
			const maxNumber = this.existingChapters.length > 0
				? Math.max(...this.existingChapters.map(c => c.number))
				: 0;
			this.chapter.number = maxNumber + 1;
		} else {
			if (!this.chapter.number || this.chapter.number < 1) {
				new Notice("Chapter number must be greater than 0", 3000);
				return;
			}
		}

		try {
			await this.onSubmit(this.chapter);
			this.close();
		} catch (err) {
			const errorMessage = err instanceof Error ? err.message : "Failed to save chapter";
			new Notice(`Error: ${errorMessage}`, 5000);
		}
	}

	onClose() {
		const { contentEl } = this;
		contentEl.empty();
	}
}

