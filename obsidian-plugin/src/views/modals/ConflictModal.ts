import { App, Modal, Setting } from "obsidian";
import { ProseBlock } from "../../types";

export type ConflictResolution = "local" | "remote" | "manual";

export interface ConflictResolutionResult {
	resolution: ConflictResolution;
	mergedContent?: string; // Only if resolution is "manual"
}

export class ConflictModal extends Modal {
	localProseBlock: ProseBlock;
	remoteProseBlock: ProseBlock;
	resolution: ConflictResolutionResult | null = null;
	onResolve: (result: ConflictResolutionResult) => Promise<void>;

	constructor(
		app: App,
		localProseBlock: ProseBlock,
		remoteProseBlock: ProseBlock,
		onResolve: (result: ConflictResolutionResult) => Promise<void>
	) {
		super(app);
		this.localProseBlock = localProseBlock;
		this.remoteProseBlock = remoteProseBlock;
		this.onResolve = onResolve;
	}

	onOpen() {
		const { contentEl } = this;
		contentEl.empty();

		contentEl.createEl("h2", {
			text: "Prose Block Conflict",
		});

		contentEl.createEl("p", {
			text: "This prose block has been modified both locally and remotely. Choose how to resolve the conflict:",
		});

		// Show diff
		const diffContainer = contentEl.createDiv("conflict-diff-container");
		
		// Local version
		const localDiv = diffContainer.createDiv("conflict-local");
		localDiv.createEl("h3", { text: "Local Version" });
		const localContent = localDiv.createEl("pre", {
			text: this.localProseBlock.content,
			cls: "conflict-content",
		});
		localContent.style.whiteSpace = "pre-wrap";
		localContent.style.maxHeight = "200px";
		localContent.style.overflow = "auto";
		localContent.style.border = "1px solid var(--background-modifier-border)";
		localContent.style.padding = "10px";
		localContent.style.borderRadius = "4px";

		// Remote version
		const remoteDiv = diffContainer.createDiv("conflict-remote");
		remoteDiv.createEl("h3", { text: "Remote Version" });
		const remoteContent = remoteDiv.createEl("pre", {
			text: this.remoteProseBlock.content,
			cls: "conflict-content",
		});
		remoteContent.style.whiteSpace = "pre-wrap";
		remoteContent.style.maxHeight = "200px";
		remoteContent.style.overflow = "auto";
		remoteContent.style.border = "1px solid var(--background-modifier-border)";
		remoteContent.style.padding = "10px";
		remoteContent.style.borderRadius = "4px";

		// Manual merge option
		const manualDiv = contentEl.createDiv("conflict-manual");
		manualDiv.createEl("h3", { text: "Manual Merge (Optional)" });
		const manualTextarea = manualDiv.createEl("textarea", {
			text: this.localProseBlock.content,
			cls: "conflict-manual-input",
		});
		manualTextarea.style.width = "100%";
		manualTextarea.style.minHeight = "150px";
		manualTextarea.style.padding = "10px";
		manualTextarea.style.border = "1px solid var(--background-modifier-border)";
		manualTextarea.style.borderRadius = "4px";
		manualTextarea.style.fontFamily = "var(--font-monospace)";

		// Resolution buttons
		const buttonContainer = contentEl.createDiv("conflict-buttons");
		buttonContainer.style.marginTop = "20px";
		buttonContainer.style.display = "flex";
		buttonContainer.style.gap = "10px";

		// Use Local button
		const useLocalBtn = buttonContainer.createEl("button", {
			text: "Use Local",
			cls: "mod-cta",
		});
		useLocalBtn.onclick = async () => {
			this.resolution = { resolution: "local" };
			await this.onResolve(this.resolution);
			this.close();
		};

		// Use Remote button
		const useRemoteBtn = buttonContainer.createEl("button", {
			text: "Use Remote",
		});
		useRemoteBtn.onclick = async () => {
			this.resolution = { resolution: "remote" };
			await this.onResolve(this.resolution);
			this.close();
		};

		// Use Manual button
		const useManualBtn = buttonContainer.createEl("button", {
			text: "Use Manual Merge",
			cls: "mod-primary",
		});
		useManualBtn.onclick = async () => {
			this.resolution = {
				resolution: "manual",
				mergedContent: manualTextarea.value,
			};
			await this.onResolve(this.resolution);
			this.close();
		};
	}

	onClose() {
		const { contentEl } = this;
		contentEl.empty();
	}
}

