import { App, Modal } from "obsidian";

export class PromptModal extends Modal {
	private value: string = "";
	private resolve: ((value: string | null) => void) | null = null;

	constructor(
		app: App,
		private promptText: string,
		private defaultValue: string = ""
	) {
		super(app);
		this.value = defaultValue;
	}

	static async prompt(
		app: App,
		promptText: string,
		defaultValue: string = ""
	): Promise<string | null> {
		return new Promise((resolve) => {
			const modal = new PromptModal(app, promptText, defaultValue);
			modal.resolve = resolve;
			modal.open();
		});
	}

	onOpen() {
		const { contentEl } = this;
		contentEl.empty();

		contentEl.createEl("h3", { text: this.promptText });

		const inputEl = contentEl.createEl("input", {
			type: "text",
			value: this.value,
		});
		inputEl.style.width = "100%";
		inputEl.style.marginBottom = "10px";

		inputEl.addEventListener("input", (e) => {
			this.value = (e.target as HTMLInputElement).value;
		});

		inputEl.addEventListener("keypress", (e) => {
			if (e.key === "Enter") {
				this.submit();
			}
		});

		const buttonContainer = contentEl.createEl("div", {
			cls: "modal-button-container",
		});

		const submitButton = buttonContainer.createEl("button", {
			text: "OK",
			cls: "mod-cta",
		});
		submitButton.onclick = () => this.submit();

		const cancelButton = buttonContainer.createEl("button", {
			text: "Cancel",
		});
		cancelButton.onclick = () => this.cancel();

		inputEl.focus();
	}

	submit() {
		if (this.resolve) {
			this.resolve(this.value || null);
		}
		this.close();
	}

	cancel() {
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

