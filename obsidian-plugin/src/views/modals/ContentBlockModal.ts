import { App, Modal, Notice, Setting } from "obsidian";
import { ContentBlock } from "../../types";
import { UnsplashClient, UnsplashPhoto } from "../../api/unsplash";
import StoryEnginePlugin from "../../main";

export class ContentBlockModal extends Modal {
	contentBlock: Partial<ContentBlock> = {
		type: "text",
		kind: "final",
		content: "",
		metadata: {},
	};
	isEdit: boolean = false;
	onSubmit: (contentBlock: Partial<ContentBlock>) => Promise<void>;
	unsplashClient: UnsplashClient;
	unsplashResults: UnsplashPhoto[] = [];
	unsplashSearchQuery: string = "";
	unsplashSearching: boolean = false;
	selectedImageUrl: string = "";
	plugin: StoryEnginePlugin;
	currentImageSourceTab: "unsplash" | "internet link" | "local" = "internet link";

	constructor(
		app: App,
		onSubmit: (contentBlock: Partial<ContentBlock>) => Promise<void>,
		contentBlock?: ContentBlock,
		plugin?: StoryEnginePlugin
	) {
		super(app);
		this.onSubmit = onSubmit;
		this.plugin = plugin as StoryEnginePlugin;
		const unsplashAccessKey = this.plugin?.settings?.unsplashAccessKey || "";
		const unsplashSecretKey = this.plugin?.settings?.unsplashSecretKey || "";
		this.unsplashClient = new UnsplashClient(unsplashAccessKey, unsplashSecretKey);
		if (contentBlock) {
			this.isEdit = true;
			this.contentBlock = {
				type: contentBlock.type,
				kind: contentBlock.kind || "final",
				content: contentBlock.content,
				metadata: contentBlock.metadata || {},
			};
			if (contentBlock.type === "image" && contentBlock.content) {
				this.selectedImageUrl = contentBlock.content;
			}
		}
	}

	onOpen() {
		const { contentEl } = this;
		contentEl.empty();

		contentEl.createEl("h2", {
			text: this.isEdit ? "Edit Content Block" : "Create Content Block",
		});

		// Only show type selector if editing and type is not already set
		if (this.isEdit && !this.contentBlock.type) {
			new Setting(contentEl)
				.setName("Type")
				.setDesc("Select the content type")
				.addDropdown((dropdown) =>
					dropdown
						.addOption("text", "Text")
						.addOption("image", "Image")
						.setValue(this.contentBlock.type || "text")
						.onChange((value) => {
							this.contentBlock.type = value as ContentBlock["type"];
							this.renderContentFields();
						})
				);
		}

		// Render content fields based on type
		this.renderContentFields();

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
	}

	renderContentFields() {
		const { contentEl } = this;
		
		// Remove existing content fields
		const existingFields = contentEl.querySelectorAll(".content-block-field");
		existingFields.forEach((el) => el.remove());

		if (this.contentBlock.type === "text") {
			const textField = contentEl.createDiv({ cls: "content-block-field" });
			new Setting(textField)
				.setName("Content")
				.setDesc("Enter the text content")
				.addTextArea((text) =>
					text
						.setPlaceholder("Enter text...")
						.setValue(this.contentBlock.content || "")
						.onChange((value) => {
							this.contentBlock.content = value;
							// Update word count in metadata
							if (!this.contentBlock.metadata) {
								this.contentBlock.metadata = {};
							}
							this.contentBlock.metadata.word_count = value.trim().split(/\s+/).filter(w => w.length > 0).length;
						})
				);
		} else if (this.contentBlock.type === "image") {
			const imageField = contentEl.createDiv({ cls: "content-block-field" });
			
			// Determine current source tab from metadata or default
			const sourceValue = this.contentBlock.metadata?.source || "internet link";
			this.currentImageSourceTab = sourceValue as "unsplash" | "internet link" | "local";

			// Source tabs
			const tabsContainer = imageField.createDiv({ cls: "content-block-source-tabs" });
			const unsplashTab = tabsContainer.createEl("button", {
				text: "Unsplash",
				cls: `content-block-source-tab ${this.currentImageSourceTab === "unsplash" ? "is-active" : ""}`,
			});
			const internetTab = tabsContainer.createEl("button", {
				text: "Internet Link",
				cls: `content-block-source-tab ${this.currentImageSourceTab === "internet link" ? "is-active" : ""}`,
			});
			const localTab = tabsContainer.createEl("button", {
				text: "Local Upload",
				cls: `content-block-source-tab ${this.currentImageSourceTab === "local" ? "is-active" : ""}`,
			});

			// Tab content container
			const tabContent = imageField.createDiv({ cls: "content-block-source-tab-content" });

			// Tab click handlers
			unsplashTab.onclick = () => {
				this.currentImageSourceTab = "unsplash";
				if (!this.contentBlock.metadata) {
					this.contentBlock.metadata = {};
				}
				this.contentBlock.metadata.source = "unsplash";
				this.renderContentFields();
			};

			internetTab.onclick = () => {
				this.currentImageSourceTab = "internet link";
				if (!this.contentBlock.metadata) {
					this.contentBlock.metadata = {};
				}
				this.contentBlock.metadata.source = "internet link";
				this.renderContentFields();
			};

			localTab.onclick = () => {
				this.currentImageSourceTab = "local";
				if (!this.contentBlock.metadata) {
					this.contentBlock.metadata = {};
				}
				this.contentBlock.metadata.source = "local";
				this.renderContentFields();
			};

			// Render content based on selected tab
			if (this.currentImageSourceTab === "unsplash") {
				// Unsplash tab content
				const unsplashAccessKey = this.plugin?.settings?.unsplashAccessKey || "";
				if (unsplashAccessKey && unsplashAccessKey !== "YOUR_UNSPLASH_ACCESS_KEY") {
					const unsplashSetting = new Setting(tabContent);
					unsplashSetting.setName("Search Unsplash");
					unsplashSetting.setDesc("Search for free images");
					unsplashSetting.addButton((button) => {
						button.setButtonText("Search");
						button.onClick(() => {
							this.showUnsplashSearch();
						});
					});
				} else {
					const unsplashSetting = new Setting(tabContent);
					unsplashSetting.setName("Search Unsplash");
					unsplashSetting.setDesc("Configure Unsplash Access Key and Secret Key in plugin settings to enable image search");
					unsplashSetting.addButton((button) => {
						button.setButtonText("Search");
						button.setDisabled(true);
					});
				}

				// Show image URL if already selected
				if (this.selectedImageUrl) {
					new Setting(tabContent)
						.setName("Selected Image URL")
						.setDesc("Image URL from Unsplash")
						.addText((text) =>
							text
								.setValue(this.selectedImageUrl)
								.setDisabled(true)
						);
				}
			} else if (this.currentImageSourceTab === "internet link") {
				// Internet Link tab content
				new Setting(tabContent)
					.setName("Image URL")
					.setDesc("Enter image URL")
					.addText((text) =>
						text
							.setPlaceholder("https://example.com/image.jpg")
							.setValue(this.selectedImageUrl)
							.onChange((value) => {
								this.selectedImageUrl = value;
								this.contentBlock.content = value;
								if (!this.contentBlock.metadata) {
									this.contentBlock.metadata = {};
								}
								this.contentBlock.metadata.source = "internet link";
							})
					);
			} else if (this.currentImageSourceTab === "local") {
				// Local Upload tab content
				new Setting(tabContent)
					.setName("Upload from Computer")
					.setDesc("Upload image from your computer (coming soon)")
					.addButton((button) => {
						button.setButtonText("Choose File");
						button.setDisabled(true);
					});
			}

			// Common fields (shown for all tabs)
			// Alt text
			const altTextValue = this.contentBlock.metadata?.alt_text || "";
			const altTextSetting = new Setting(imageField);
			altTextSetting.setName("Alt Text");
			altTextSetting.setDesc("Alt text for accessibility");
			altTextSetting.addText((text) =>
				text
					.setPlaceholder("Describe the image for accessibility")
					.setValue(altTextValue)
					.onChange((value) => {
						if (!this.contentBlock.metadata) {
							this.contentBlock.metadata = {};
						}
						this.contentBlock.metadata.alt_text = value || undefined;
					})
			);

			// Author Name
			const authorNameValue = this.contentBlock.metadata?.author_name || "";
			new Setting(imageField)
				.setName("Author Name")
				.setDesc("Name of the image author/photographer")
				.addText((text) =>
					text
						.setPlaceholder("Author name")
						.setValue(authorNameValue)
						.onChange((value) => {
							if (!this.contentBlock.metadata) {
								this.contentBlock.metadata = {};
							}
							this.contentBlock.metadata.author_name = value || undefined;
						})
				);

			// Attribution
			const attributionValue = this.contentBlock.metadata?.attribution || "";
			new Setting(imageField)
				.setName("Attribution")
				.setDesc("Attribution text (e.g., 'Photo by John Doe on Unsplash')")
				.addText((text) =>
					text
						.setPlaceholder("Photo by Author Name on Source")
						.setValue(attributionValue)
						.onChange((value) => {
							if (!this.contentBlock.metadata) {
								this.contentBlock.metadata = {};
							}
							this.contentBlock.metadata.attribution = value || undefined;
						})
				);

			// Attribution URL
			const attributionUrlValue = this.contentBlock.metadata?.attribution_url || "";
			new Setting(imageField)
				.setName("Attribution URL")
				.setDesc("Link to the original image or author page")
				.addText((text) =>
					text
						.setPlaceholder("https://example.com/photo")
						.setValue(attributionUrlValue)
						.onChange((value) => {
							if (!this.contentBlock.metadata) {
								this.contentBlock.metadata = {};
							}
							this.contentBlock.metadata.attribution_url = value || undefined;
						})
				);
		}
	}

	showUnsplashSearch() {
		const { contentEl } = this;
		
		// Create search modal
		const searchModal = new Modal(this.app);
		searchModal.titleEl.setText("Search Unsplash");

		const searchContent = searchModal.contentEl;
		
		// Search input
		const searchInput = searchContent.createEl("input", {
			type: "text",
			placeholder: "Search for images...",
			cls: "unsplash-search-input",
		});
		searchInput.value = this.unsplashSearchQuery;
		searchInput.style.width = "100%";
		searchInput.style.padding = "0.5rem";
		searchInput.style.marginBottom = "1rem";

		// Search button
		const searchButton = searchContent.createEl("button", {
			text: "Search",
			cls: "mod-cta",
		});
		searchButton.style.marginBottom = "1rem";

		// Results container
		const resultsContainer = searchContent.createDiv({ cls: "unsplash-results-grid" });

		const performSearch = async () => {
			const query = searchInput.value.trim();
			if (!query) {
				new Notice("Please enter a search query", 3000);
				return;
			}

			this.unsplashSearching = true;
			searchButton.disabled = true;
			searchButton.setText("Searching...");
			resultsContainer.empty();
			resultsContainer.createEl("p", { text: "Searching..." });

			try {
				const response = await this.unsplashClient.searchImages(query, 1, 20);
				this.unsplashResults = response.results;

				resultsContainer.empty();
				if (this.unsplashResults.length === 0) {
					resultsContainer.createEl("p", { text: "No results found." });
				} else {
					for (const photo of this.unsplashResults) {
						const photoItem = resultsContainer.createDiv({ cls: "unsplash-photo-item" });
						
						const img = photoItem.createEl("img", {
							attr: {
								src: this.unsplashClient.getImageUrl(photo, "thumb"),
								alt: photo.alt_description || photo.description || "Unsplash photo",
							},
						});
						img.style.width = "100%";
						img.style.height = "150px";
						img.style.objectFit = "cover";
						img.style.borderRadius = "4px";
						img.style.cursor = "pointer";

						const photoInfo = photoItem.createDiv({ cls: "unsplash-photo-info" });
						photoInfo.createEl("p", {
							text: photo.alt_description || photo.description || "Untitled",
							cls: "unsplash-photo-title",
						});
						photoInfo.createEl("p", {
							text: this.unsplashClient.getAttributionText(photo),
							cls: "unsplash-photo-attribution",
						});

						photoItem.onclick = () => {
							const imageUrl = this.unsplashClient.getImageUrl(photo, "regular");
							this.selectedImageUrl = imageUrl;
							this.contentBlock.content = imageUrl;
							if (!this.contentBlock.metadata) {
								this.contentBlock.metadata = {};
							}
							const altText = photo.alt_description || photo.description || "";
							this.contentBlock.metadata.alt_text = altText;
							// Add attribution as required by Unsplash
							this.contentBlock.metadata.attribution = this.unsplashClient.getAttributionText(photo);
							this.contentBlock.metadata.attribution_url = this.unsplashClient.getAttributionUrl(photo);
							this.contentBlock.metadata.author_name = photo.user.name;
							this.contentBlock.metadata.source = "unsplash";
							searchModal.close();
							this.renderContentFields();
							new Notice("Image selected");
						};
					}
				}
			} catch (err) {
				const errorMessage = err instanceof Error ? err.message : "Failed to search Unsplash";
				new Notice(`Error: ${errorMessage}`, 5000);
				resultsContainer.empty();
				resultsContainer.createEl("p", {
					text: `Error: ${errorMessage}`,
					cls: "story-engine-error",
				});
			} finally {
				this.unsplashSearching = false;
				searchButton.disabled = false;
				searchButton.setText("Search");
			}
		};

		searchButton.onclick = performSearch;
		searchInput.addEventListener("keypress", (e) => {
			if (e.key === "Enter") {
				performSearch();
			}
		});

		// Load initial results if query exists
		if (this.unsplashSearchQuery) {
			performSearch();
		}

		searchModal.open();
	}

	async submit() {
		// Type should always be set (either from edit or from button click)
		if (!this.contentBlock.type) {
			this.contentBlock.type = "text"; // Default fallback
		}

		if (this.contentBlock.type === "text" && !this.contentBlock.content?.trim()) {
			new Notice("Please enter text content", 3000);
			return;
		}

		if (this.contentBlock.type === "image" && !this.contentBlock.content?.trim()) {
			new Notice("Please enter an image URL or select an image", 3000);
			return;
		}

		try {
			await this.onSubmit(this.contentBlock);
			this.close();
		} catch (err) {
			const errorMessage = err instanceof Error ? err.message : "Failed to save content block";
			new Notice(`Error: ${errorMessage}`, 5000);
		}
	}

	onClose() {
		const { contentEl } = this;
		contentEl.empty();
	}
}

