import { App, Modal, Notice, Setting } from "obsidian";
import { Beat, Scene, Chapter } from "../../types";

export class BeatModal extends Modal {
	beat: Partial<Beat> = {
		type: "setup",
		intent: "",
		outcome: "",
	};
	isEdit: boolean = false;
	scenes: Scene[] = [];
	chapters: Chapter[] = [];
	existingBeats: Beat[] = [];
	storyId: string;
	defaultSceneId?: string;
	onSubmit: (beat: Partial<Beat>) => Promise<void>;

	constructor(
		app: App,
		storyId: string,
		scenes: Scene[],
		onSubmit: (beat: Partial<Beat>) => Promise<void>,
		existingBeats: Beat[] = [],
		beat?: Beat,
		chapters: Chapter[] = [],
		defaultSceneId?: string
	) {
		super(app);
		this.storyId = storyId;
		this.scenes = scenes;
		this.chapters = chapters;
		this.existingBeats = existingBeats;
		this.defaultSceneId = defaultSceneId;
		this.onSubmit = onSubmit;
		if (beat) {
			this.isEdit = true;
			this.beat = {
				scene_id: beat.scene_id,
				order_num: beat.order_num,
				type: beat.type,
				intent: beat.intent,
				outcome: beat.outcome,
			};
		} else if (defaultSceneId) {
			this.beat.scene_id = defaultSceneId;
		}
	}

	onOpen() {
		const { contentEl } = this;
		contentEl.empty();

		contentEl.createEl("h2", {
			text: this.isEdit ? "Edit Beat" : "Create Beat",
		});

		new Setting(contentEl)
			.setName("Scene")
			.setDesc("Select the scene for this beat")
			.addDropdown((dropdown) => {
				// Group scenes by chapter for better UX
				const scenesByChapter = new Map<string | null, Scene[]>();
				for (const scene of this.scenes) {
					const chapterId = scene.chapter_id || null;
					if (!scenesByChapter.has(chapterId)) {
						scenesByChapter.set(chapterId, []);
					}
					scenesByChapter.get(chapterId)!.push(scene);
				}

				// Helper to get chapter label
				const getChapterLabel = (chapterId: string | null): string => {
					if (!chapterId) return "No Chapter";
					const chapter = this.chapters.find(c => c.id === chapterId);
					return chapter ? `Chapter ${chapter.number}: ${chapter.title}` : "No Chapter";
				};

				// Add scenes grouped
				for (const [chapterId, chapterScenes] of scenesByChapter.entries()) {
					const label = getChapterLabel(chapterId);
					for (const scene of chapterScenes.sort((a, b) => a.order_num - b.order_num)) {
						dropdown.addOption(
							scene.id,
							`${label} > Scene ${scene.order_num}: ${scene.goal || "Untitled"}`
						);
					}
				}

				dropdown.setValue(this.beat.scene_id || "");
				dropdown.onChange((value) => {
					this.beat.scene_id = value;
				});
			});

		if (this.isEdit) {
			new Setting(contentEl)
				.setName("Order Number")
				.setDesc("Beat order within scene")
				.addText((text) =>
					text
						.setPlaceholder("1")
						.setValue(this.beat.order_num?.toString() || "1")
						.onChange((value) => {
							const num = parseInt(value);
							if (!isNaN(num) && num > 0) {
								this.beat.order_num = num;
							}
						})
				);
		}

		new Setting(contentEl)
			.setName("Type")
			.setDesc("Beat type")
			.addDropdown((dropdown) =>
				dropdown
					.addOption("setup", "Setup")
					.addOption("turn", "Turn")
					.addOption("reveal", "Reveal")
					.addOption("conflict", "Conflict")
					.addOption("climax", "Climax")
					.addOption("resolution", "Resolution")
					.addOption("hook", "Hook")
					.addOption("transition", "Transition")
					.setValue(this.beat.type || "setup")
					.onChange((value) => {
						this.beat.type = value;
					})
			);

		new Setting(contentEl)
			.setName("Intent")
			.setDesc("What is the intent of this beat?")
			.addTextArea((text) =>
				text
					.setPlaceholder("What does the character want?")
					.setValue(this.beat.intent || "")
					.onChange((value) => {
						this.beat.intent = value;
					})
			);

		new Setting(contentEl)
			.setName("Outcome")
			.setDesc("What is the outcome of this beat?")
			.addTextArea((text) =>
				text
					.setPlaceholder("What happens as a result?")
					.setValue(this.beat.outcome || "")
					.onChange((value) => {
						this.beat.outcome = value;
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

		const intentInput = contentEl.querySelector("textarea") as HTMLTextAreaElement | null;
		if (intentInput) {
			intentInput.focus();
		}
	}

	async submit() {
		if (!this.beat.scene_id) {
			new Notice("Please select a scene", 3000);
			return;
		}

		if (!this.beat.type) {
			new Notice("Please select a beat type", 3000);
			return;
		}

		// Auto-calculate order_num if creating
		if (!this.isEdit) {
			const beatsInScene = this.existingBeats.filter(b => b.scene_id === this.beat.scene_id);
			const maxOrderNum = beatsInScene.length > 0
				? Math.max(...beatsInScene.map(b => b.order_num))
				: 0;
			this.beat.order_num = maxOrderNum + 1;
		} else {
			if (!this.beat.order_num || this.beat.order_num < 1) {
				new Notice("Order number must be greater than 0", 3000);
				return;
			}
		}

		try {
			await this.onSubmit(this.beat);
			this.close();
		} catch (err) {
			const errorMessage = err instanceof Error ? err.message : "Failed to save beat";
			new Notice(`Error: ${errorMessage}`, 5000);
		}
	}

	onClose() {
		const { contentEl } = this;
		contentEl.empty();
	}
}

