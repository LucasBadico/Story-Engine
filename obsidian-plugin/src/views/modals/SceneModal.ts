import { App, Modal, Notice, Setting } from "obsidian";
import { Scene, Chapter } from "../../types";

export class SceneModal extends Modal {
	scene: Partial<Scene> = {
		time_ref: "",
		goal: "",
	};
	isEdit: boolean = false;
	chapters: Chapter[] = [];
	existingScenes: Scene[] = [];
	storyId: string;
	defaultChapterId?: string | null;
	onSubmit: (scene: Partial<Scene>) => Promise<void>;

	constructor(
		app: App,
		storyId: string,
		chapters: Chapter[],
		onSubmit: (scene: Partial<Scene>) => Promise<void>,
		existingScenes: Scene[] = [],
		scene?: Scene,
		defaultChapterId?: string | null
	) {
		super(app);
		this.storyId = storyId;
		this.chapters = chapters;
		this.existingScenes = existingScenes;
		this.defaultChapterId = defaultChapterId;
		this.onSubmit = onSubmit;
		if (scene) {
			this.isEdit = true;
			this.scene = {
				story_id: scene.story_id,
				chapter_id: scene.chapter_id || null,
				order_num: scene.order_num,
				time_ref: scene.time_ref,
				goal: scene.goal,
				pov_character_id: scene.pov_character_id,
				location_id: scene.location_id,
			};
		} else {
			this.scene.story_id = storyId;
			// Pre-select chapter if provided
			if (defaultChapterId !== undefined) {
				this.scene.chapter_id = defaultChapterId;
			}
		}
	}

	onOpen() {
		const { contentEl } = this;
		contentEl.empty();

		contentEl.createEl("h2", {
			text: this.isEdit ? "Edit Scene" : "Create Scene",
		});

		new Setting(contentEl)
			.setName("Chapter")
			.setDesc("Select the chapter for this scene (optional)")
			.addDropdown((dropdown) => {
				dropdown.addOption("", "No Chapter");
				for (const chapter of this.chapters.sort((a, b) => a.number - b.number)) {
					dropdown.addOption(
						chapter.id,
						`Chapter ${chapter.number}: ${chapter.title}`
					);
				}
				dropdown.setValue(this.scene.chapter_id || "");
				dropdown.onChange((value) => {
					this.scene.chapter_id = value || null;
				});
			});

		if (this.isEdit) {
			new Setting(contentEl)
				.setName("Order Number")
				.setDesc("Scene order within chapter")
				.addText((text) =>
					text
						.setPlaceholder("1")
						.setValue(this.scene.order_num?.toString() || "1")
						.onChange((value) => {
							const num = parseInt(value);
							if (!isNaN(num) && num > 0) {
								this.scene.order_num = num;
							}
						})
				);
		}

		new Setting(contentEl)
			.setName("Goal")
			.setDesc("Scene goal or description")
			.addTextArea((text) =>
				text
					.setPlaceholder("What happens in this scene?")
					.setValue(this.scene.goal || "")
					.onChange((value) => {
						this.scene.goal = value;
					})
			);

		new Setting(contentEl)
			.setName("Time Reference")
			.setDesc("When does this scene take place?")
			.addText((text) =>
				text
					.setPlaceholder("Morning, Evening, etc.")
					.setValue(this.scene.time_ref || "")
					.onChange((value) => {
						this.scene.time_ref = value;
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

		const goalInput = contentEl.querySelector("textarea") as HTMLTextAreaElement | null;
		if (goalInput) {
			goalInput.focus();
		}
	}

	async submit() {
		// Auto-calculate order_num if creating
		if (!this.isEdit) {
			const chapterId = this.scene.chapter_id || null;
			const scenesInChapter = this.existingScenes.filter(s => 
				(s.chapter_id || null) === chapterId
			);
			const maxOrderNum = scenesInChapter.length > 0
				? Math.max(...scenesInChapter.map(s => s.order_num))
				: 0;
			this.scene.order_num = maxOrderNum + 1;
		} else {
			if (!this.scene.order_num || this.scene.order_num < 1) {
				new Notice("Order number must be greater than 0", 3000);
				return;
			}
		}

		try {
			await this.onSubmit(this.scene);
			this.close();
		} catch (err) {
			const errorMessage = err instanceof Error ? err.message : "Failed to save scene";
			new Notice(`Error: ${errorMessage}`, 5000);
		}
	}

	onClose() {
		const { contentEl } = this;
		contentEl.empty();
	}
}

