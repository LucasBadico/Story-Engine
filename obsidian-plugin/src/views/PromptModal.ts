import { Modal } from "obsidian";

export class PromptModal extends Modal {
	result: string;
	placeholder: string;
	promptText: string;
	resolve: (value: string | null) => void;

	constructor(
		app: any,
		promptText: string,
		placeholder: string = ""
	) {
		super(app);
		this.promptText = promptText;
		this.placeholder = placeholder;
		this.result = "";
	}

	onOpen() {
		const { contentEl } = this;
		contentEl.empty();

		contentEl.createEl("h2", { text: this.promptText });

		const input = contentEl.createEl("input", {
			type: "text",
			placeholder: this.placeholder,
			cls: "prompt-input",
		});

		input.focus();
		input.select();

		const buttonContainer = contentEl.createEl("div", {
			cls: "prompt-button-container",
		});

		const confirmButton = buttonContainer.createEl("button", {
			text: "OK",
			cls: "mod-cta",
		});

		const cancelButton = buttonContainer.createEl("button", {
			text: "Cancel",
		});

		const handleConfirm = () => {
			this.result = input.value.trim();
			this.close();
			this.resolve(this.result || null);
		};

		const handleCancel = () => {
			this.close();
			this.resolve(null);
		};

		confirmButton.onclick = handleConfirm;
		cancelButton.onclick = handleCancel;

		input.onkeydown = (e: KeyboardEvent) => {
			if (e.key === "Enter") {
				e.preventDefault();
				handleConfirm();
			} else if (e.key === "Escape") {
				e.preventDefault();
				handleCancel();
			}
		};
	}

	onClose() {
		const { contentEl } = this;
		contentEl.empty();
	}

	static async prompt(
		app: any,
		promptText: string,
		placeholder: string = ""
	): Promise<string | null> {
		return new Promise((resolve) => {
			const modal = new PromptModal(app, promptText, placeholder);
			modal.resolve = resolve;
			modal.open();
		});
	}
}

