import { App, Modal } from "obsidian";
import { World, Story } from "../types";

export class WorldDetailsModal extends Modal {
	world: World;
	stories: Story[];

	constructor(app: App, world: World, stories: Story[]) {
		super(app);
		this.world = world;
		this.stories = stories;
	}

	onOpen() {
		const { contentEl } = this;
		contentEl.empty();

		contentEl.createEl("h2", { text: this.world.name });

		// Description
		if (this.world.description) {
			const descSection = contentEl.createDiv({ cls: "world-details-section" });
			descSection.createEl("h3", { text: "Description" });
			descSection.createEl("p", { text: this.world.description });
		}

		// Genre
		const genreSection = contentEl.createDiv({ cls: "world-details-section" });
		genreSection.createEl("h3", { text: "Genre" });
		genreSection.createEl("p", { text: this.world.genre });

		// Stories
		const storiesSection = contentEl.createDiv({ cls: "world-details-section" });
		storiesSection.createEl("h3", { text: "Stories" });
		if (this.stories.length === 0) {
			storiesSection.createEl("p", { text: "No stories in this world." });
		} else {
			const storiesList = storiesSection.createEl("ul");
			for (const story of this.stories) {
				const item = storiesList.createEl("li");
				item.createEl("strong", { text: story.title });
				item.appendText(` (${story.status}, v${story.version_number})`);
			}
		}

		// Timestamps
		const metaSection = contentEl.createDiv({ cls: "world-details-section" });
		metaSection.createEl("h3", { text: "Metadata" });
		const metaList = metaSection.createEl("ul");
		metaList.createEl("li", { text: `Created: ${new Date(this.world.created_at).toLocaleString()}` });
		metaList.createEl("li", { text: `Updated: ${new Date(this.world.updated_at).toLocaleString()}` });
	}

	onClose() {
		const { contentEl } = this;
		contentEl.empty();
	}
}

