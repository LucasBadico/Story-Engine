var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// src/main.ts
var main_exports = {};
__export(main_exports, {
  default: () => StoryEnginePlugin
});
module.exports = __toCommonJS(main_exports);
var import_obsidian9 = require("obsidian");

// src/api/client.ts
var StoryEngineClient = class {
  constructor(apiUrl, apiKey) {
    this.apiUrl = apiUrl;
    this.apiKey = apiKey;
  }
  async request(method, endpoint, body) {
    const url = `${this.apiUrl}${endpoint}`;
    const headers = {
      "Content-Type": "application/json"
    };
    if (this.apiKey) {
      headers["Authorization"] = `Bearer ${this.apiKey}`;
    }
    const options = {
      method,
      headers
    };
    if (body) {
      options.body = JSON.stringify(body);
    }
    const response = await fetch(url, options);
    if (!response.ok) {
      let error;
      try {
        error = await response.json();
      } catch (e) {
        error = {
          error: "unknown_error",
          message: `HTTP ${response.status}: ${response.statusText}`,
          code: "HTTP_ERROR"
        };
      }
      const errorMessage = error.message || error.error || `HTTP ${response.status}: ${response.statusText}`;
      throw new Error(errorMessage);
    }
    return response.json();
  }
  async listStories(tenantId) {
    const trimmedTenantId = encodeURIComponent(tenantId.trim());
    const response = await this.request(
      "GET",
      `/api/v1/stories?tenant_id=${trimmedTenantId}`
    );
    return response.stories || [];
  }
  async getStory(id) {
    const response = await this.request(
      "GET",
      `/api/v1/stories/${id}`
    );
    return response.story;
  }
  async createStory(tenantId, title) {
    const trimmedTenantId = tenantId.trim();
    if (!trimmedTenantId) {
      throw new Error("Tenant ID is required");
    }
    const response = await this.request(
      "POST",
      "/api/v1/stories",
      {
        tenant_id: trimmedTenantId,
        title: title.trim()
      }
    );
    return response.story;
  }
  async cloneStory(id) {
    const response = await this.request(
      "POST",
      `/api/v1/stories/${id}/clone`
    );
    return response.story;
  }
  async getTenant(id) {
    const response = await this.request(
      "GET",
      `/api/v1/tenants/${id}`
    );
    return response.tenant;
  }
  async testConnection() {
    try {
      await this.request("GET", "/health");
      return true;
    } catch (e) {
      return false;
    }
  }
  // Story methods
  async updateStory(id, title, status) {
    const body = { title: title.trim() };
    if (status) {
      body.status = status;
    }
    const response = await this.request(
      "PUT",
      `/api/v1/stories/${id}`,
      body
    );
    return response.story;
  }
  async getStoryWithHierarchy(id) {
    const story = await this.getStory(id);
    const chapters = await this.getChapters(id);
    const chaptersWithContent = await Promise.all(
      chapters.map(async (chapter) => {
        const scenes = await this.getScenes(chapter.id);
        const scenesWithBeats = await Promise.all(
          scenes.map(async (scene) => {
            const beats = await this.getBeats(scene.id);
            return { scene, beats };
          })
        );
        return { chapter, scenes: scenesWithBeats };
      })
    );
    return {
      story,
      chapters: chaptersWithContent
    };
  }
  // Chapter methods
  async createChapter(storyId, chapter) {
    const response = await this.request(
      "POST",
      "/api/v1/chapters",
      {
        story_id: storyId,
        number: chapter.number,
        title: chapter.title,
        status: chapter.status
      }
    );
    return response.chapter;
  }
  async updateChapter(id, chapter) {
    const response = await this.request(
      "PUT",
      `/api/v1/chapters/${id}`,
      chapter
    );
    return response.chapter;
  }
  async getChapters(storyId) {
    const response = await this.request(
      "GET",
      `/api/v1/stories/${storyId}/chapters`
    );
    return response.chapters || [];
  }
  async getChapter(id) {
    const response = await this.request(
      "GET",
      `/api/v1/chapters/${id}`
    );
    return response.chapter;
  }
  async deleteChapter(id) {
    await this.request("DELETE", `/api/v1/chapters/${id}`);
  }
  // Scene methods
  async createScene(scene) {
    const response = await this.request(
      "POST",
      "/api/v1/scenes",
      scene
    );
    return response.scene;
  }
  async updateScene(id, scene) {
    const response = await this.request(
      "PUT",
      `/api/v1/scenes/${id}`,
      scene
    );
    return response.scene;
  }
  async getScenes(chapterId) {
    const response = await this.request(
      "GET",
      `/api/v1/chapters/${chapterId}/scenes`
    );
    return response.scenes || [];
  }
  async getScene(id) {
    const response = await this.request(
      "GET",
      `/api/v1/scenes/${id}`
    );
    return response.scene;
  }
  async deleteScene(id) {
    await this.request("DELETE", `/api/v1/scenes/${id}`);
  }
  // Beat methods
  async createBeat(beat) {
    const response = await this.request(
      "POST",
      "/api/v1/beats",
      beat
    );
    return response.beat;
  }
  async updateBeat(id, beat) {
    const response = await this.request(
      "PUT",
      `/api/v1/beats/${id}`,
      beat
    );
    return response.beat;
  }
  async getBeats(sceneId) {
    const response = await this.request(
      "GET",
      `/api/v1/scenes/${sceneId}/beats`
    );
    return response.beats || [];
  }
  async getBeat(id) {
    const response = await this.request(
      "GET",
      `/api/v1/beats/${id}`
    );
    return response.beat;
  }
  async deleteBeat(id) {
    await this.request("DELETE", `/api/v1/beats/${id}`);
  }
};

// src/settings.ts
var import_obsidian = require("obsidian");
var StoryEngineSettingTab = class extends import_obsidian.PluginSettingTab {
  constructor(app, plugin) {
    super(app, plugin);
    this.plugin = plugin;
  }
  display() {
    const { containerEl } = this;
    containerEl.empty();
    containerEl.createEl("h2", { text: "Story Engine Settings" });
    new import_obsidian.Setting(containerEl).setName("API URL").setDesc("The base URL of the Story Engine API").addText(
      (text) => text.setPlaceholder("http://localhost:8080").setValue(this.plugin.settings.apiUrl).onChange(async (value) => {
        this.plugin.settings.apiUrl = value;
        await this.plugin.saveSettings();
      })
    );
    new import_obsidian.Setting(containerEl).setName("API Key").setDesc("API key for authentication (optional for MVP)").addText((text) => {
      text.setPlaceholder("Enter API key").setValue(this.plugin.settings.apiKey).inputEl.type = "password";
      text.onChange(async (value) => {
        this.plugin.settings.apiKey = value;
        await this.plugin.saveSettings();
      });
    });
    new import_obsidian.Setting(containerEl).setName("Tenant ID").setDesc("Your workspace tenant ID (UUID format)").addText(
      (text) => text.setPlaceholder("00000000-0000-0000-0000-000000000000").setValue(this.plugin.settings.tenantId || "").onChange(async (value) => {
        this.plugin.settings.tenantId = value.trim();
        await this.plugin.saveSettings();
      })
    );
    new import_obsidian.Setting(containerEl).setName("Sync Folder Path").setDesc("Folder path where synced stories will be stored").addText(
      (text) => text.setPlaceholder("Stories").setValue(this.plugin.settings.syncFolderPath || "Stories").onChange(async (value) => {
        this.plugin.settings.syncFolderPath = value.trim() || "Stories";
        await this.plugin.saveSettings();
      })
    );
    new import_obsidian.Setting(containerEl).setName("Auto Version Snapshots").setDesc("Automatically create version snapshots when syncing").addToggle(
      (toggle) => {
        var _a;
        return toggle.setValue((_a = this.plugin.settings.autoVersionSnapshots) != null ? _a : true).onChange(async (value) => {
          this.plugin.settings.autoVersionSnapshots = value;
          await this.plugin.saveSettings();
        });
      }
    );
    new import_obsidian.Setting(containerEl).setName("Conflict Resolution").setDesc("How to resolve conflicts when both local and service have changes").addDropdown(
      (dropdown) => dropdown.addOption("service", "Service Wins").addOption("local", "Local Wins").addOption("manual", "Manual (Newer Wins)").setValue(this.plugin.settings.conflictResolution || "service").onChange(async (value) => {
        this.plugin.settings.conflictResolution = value;
        await this.plugin.saveSettings();
      })
    );
    new import_obsidian.Setting(containerEl).setName("Test Connection").setDesc("Test the connection to the Story Engine API").addButton(
      (button) => button.setButtonText("Test").onClick(async () => {
        button.setButtonText("Testing...");
        button.disabled = true;
        const success = await this.plugin.apiClient.testConnection();
        if (success) {
          button.setButtonText("\u2713 Connected");
          button.buttonEl.style.color = "green";
        } else {
          button.setButtonText("\u2717 Failed");
          button.buttonEl.style.color = "red";
        }
        setTimeout(() => {
          button.setButtonText("Test");
          button.disabled = false;
          button.buttonEl.style.color = "";
        }, 3e3);
      })
    );
  }
};

// src/commands.ts
var import_obsidian5 = require("obsidian");

// src/views/StoryListModal.ts
var import_obsidian3 = require("obsidian");

// src/views/StoryDetailsModal.ts
var import_obsidian2 = require("obsidian");
var StoryDetailsModal = class _StoryDetailsModal extends import_obsidian2.Modal {
  constructor(plugin, story) {
    super(plugin.app);
    this.plugin = plugin;
    this.story = story;
  }
  onOpen() {
    const { contentEl } = this;
    contentEl.empty();
    contentEl.createEl("h2", { text: this.story.title });
    const details = contentEl.createEl("div", { cls: "story-engine-details" });
    details.createEl("p", {
      text: `Status: ${this.story.status}`
    });
    details.createEl("p", {
      text: `Version: ${this.story.version_number}`
    });
    details.createEl("p", {
      text: `Created: ${new Date(this.story.created_at).toLocaleString()}`
    });
    details.createEl("p", {
      text: `Updated: ${new Date(this.story.updated_at).toLocaleString()}`
    });
    details.createEl("p", {
      text: `ID: ${this.story.id}`,
      cls: "story-engine-id"
    });
    const buttonContainer = contentEl.createEl("div", {
      cls: "story-engine-buttons"
    });
    const cloneButton = buttonContainer.createEl("button", {
      text: "Clone Story",
      cls: "mod-cta"
    });
    cloneButton.onclick = async () => {
      cloneButton.disabled = true;
      cloneButton.setText("Cloning...");
      try {
        const clonedStory = await this.plugin.apiClient.cloneStory(
          this.story.id
        );
        this.close();
        new _StoryDetailsModal(this.plugin, clonedStory).open();
      } catch (err) {
        cloneButton.setText(
          err instanceof Error ? err.message : "Clone failed"
        );
        setTimeout(() => {
          cloneButton.disabled = false;
          cloneButton.setText("Clone Story");
        }, 3e3);
      }
    };
    const copyIdButton = buttonContainer.createEl("button", {
      text: "Copy ID"
    });
    copyIdButton.onclick = () => {
      navigator.clipboard.writeText(this.story.id);
      copyIdButton.setText("Copied!");
      setTimeout(() => {
        copyIdButton.setText("Copy ID");
      }, 2e3);
    };
  }
  onClose() {
    const { contentEl } = this;
    contentEl.empty();
  }
};

// src/views/StoryListModal.ts
var StoryListModal = class extends import_obsidian3.Modal {
  constructor(plugin) {
    super(plugin.app);
    this.stories = [];
    this.loading = true;
    this.error = null;
    this.plugin = plugin;
  }
  async onOpen() {
    const { contentEl } = this;
    contentEl.empty();
    contentEl.createEl("h2", { text: "Stories" });
    await this.loadStories();
    if (this.loading) {
      contentEl.createEl("p", { text: "Loading stories..." });
      return;
    }
    if (this.error) {
      contentEl.createEl("p", {
        text: `Error: ${this.error}`,
        cls: "story-engine-error"
      });
      return;
    }
    if (this.stories.length === 0) {
      contentEl.createEl("p", { text: "No stories found." });
      const createButton2 = contentEl.createEl("button", {
        text: "Create Story"
      });
      createButton2.onclick = () => {
        this.close();
        this.plugin.createStoryCommand();
      };
      return;
    }
    const storiesList = contentEl.createEl("div", { cls: "story-engine-list" });
    for (const story of this.stories) {
      const storyItem = storiesList.createEl("div", {
        cls: "story-engine-item"
      });
      const title = storyItem.createEl("div", {
        cls: "story-engine-title",
        text: story.title
      });
      const meta = storyItem.createEl("div", {
        cls: "story-engine-meta"
      });
      meta.createEl("span", {
        text: `Version ${story.version_number}`
      });
      meta.createEl("span", {
        text: `Status: ${story.status}`
      });
      storyItem.onclick = () => {
        this.close();
        new StoryDetailsModal(this.plugin, story).open();
      };
    }
    const createButton = contentEl.createEl("button", {
      text: "Create New Story",
      cls: "mod-cta"
    });
    createButton.onclick = () => {
      this.close();
      this.plugin.createStoryCommand();
    };
  }
  async loadStories() {
    this.loading = true;
    this.error = null;
    try {
      if (!this.plugin.settings.tenantId) {
        this.error = "Tenant ID not configured";
        this.loading = false;
        return;
      }
      this.stories = await this.plugin.apiClient.listStories(
        this.plugin.settings.tenantId
      );
    } catch (err) {
      this.error = err instanceof Error ? err.message : "Unknown error";
    } finally {
      this.loading = false;
    }
  }
  onClose() {
    const { contentEl } = this;
    contentEl.empty();
  }
};

// src/views/StorySyncModal.ts
var import_obsidian4 = require("obsidian");
var StorySyncModal = class extends import_obsidian4.Modal {
  constructor(plugin, mode) {
    super(plugin.app);
    this.stories = [];
    this.loading = true;
    this.error = null;
    this.plugin = plugin;
    this.mode = mode;
  }
  async onOpen() {
    const { contentEl } = this;
    contentEl.empty();
    const title = this.mode === "pull" ? "Sync Story from Service" : "Push Story to Service";
    contentEl.createEl("h2", { text: title });
    await this.loadStories();
    if (this.loading) {
      contentEl.createEl("p", { text: "Loading stories..." });
      return;
    }
    if (this.error) {
      contentEl.createEl("p", {
        text: `Error: ${this.error}`,
        cls: "story-engine-error"
      });
      return;
    }
    if (this.stories.length === 0) {
      contentEl.createEl("p", { text: "No stories found." });
      return;
    }
    const storiesList = contentEl.createEl("div", { cls: "story-engine-list" });
    for (const story of this.stories) {
      const storyItem = storiesList.createEl("div", {
        cls: "story-engine-item"
      });
      const title2 = storyItem.createEl("div", {
        cls: "story-engine-title",
        text: story.title
      });
      const meta = storyItem.createEl("div", {
        cls: "story-engine-meta"
      });
      meta.createEl("span", {
        text: `Version ${story.version_number}`
      });
      meta.createEl("span", {
        text: `Status: ${story.status}`
      });
      storyItem.onclick = async () => {
        this.close();
        try {
          if (this.mode === "pull") {
            await this.plugin.syncService.pullStory(story.id);
          } else {
            const folderPath = this.plugin.fileManager.getStoryFolderPath(
              story.title
            );
            await this.plugin.syncService.pushStory(folderPath);
          }
        } catch (err) {
          const errorMessage = err instanceof Error ? err.message : "Failed to sync story";
          new import_obsidian4.Notice(`Error: ${errorMessage}`, 5e3);
        }
      };
    }
  }
  async loadStories() {
    this.loading = true;
    this.error = null;
    try {
      if (!this.plugin.settings.tenantId) {
        this.error = "Tenant ID not configured";
        this.loading = false;
        return;
      }
      this.stories = await this.plugin.apiClient.listStories(
        this.plugin.settings.tenantId
      );
    } catch (err) {
      this.error = err instanceof Error ? err.message : "Unknown error";
    } finally {
      this.loading = false;
    }
  }
  onClose() {
    const { contentEl } = this;
    contentEl.empty();
  }
};

// src/commands.ts
function registerCommands(plugin) {
  plugin.addCommand({
    id: "list-stories",
    name: "List Stories",
    callback: () => {
      new StoryListModal(plugin).open();
    }
  });
  plugin.addCommand({
    id: "create-story",
    name: "Create Story",
    callback: () => {
      plugin.createStoryCommand();
    }
  });
  plugin.addCommand({
    id: "sync-story-from-service",
    name: "Sync Story from Service",
    callback: () => {
      new StorySyncModal(plugin, "pull").open();
    }
  });
  plugin.addCommand({
    id: "push-story-to-service",
    name: "Push Story to Service",
    callback: () => {
      new StorySyncModal(plugin, "push").open();
    }
  });
  plugin.addCommand({
    id: "sync-all-stories",
    name: "Sync All Stories",
    callback: async () => {
      if (!plugin.settings.tenantId) {
        new import_obsidian5.Notice("Please configure Tenant ID in settings", 5e3);
        return;
      }
      try {
        new import_obsidian5.Notice("Syncing all stories...");
        await plugin.syncService.pullAllStories();
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : "Failed to sync stories";
        new import_obsidian5.Notice(`Error: ${errorMessage}`, 5e3);
      }
    }
  });
}

// src/views/CreateStoryModal.ts
var import_obsidian6 = require("obsidian");
var CreateStoryModal = class extends import_obsidian6.Modal {
  constructor(app, onSubmit) {
    super(app);
    this.title = "";
    this.shouldSync = true;
    this.onSubmit = onSubmit;
  }
  onOpen() {
    const { contentEl } = this;
    contentEl.createEl("h2", { text: "Create New Story" });
    new import_obsidian6.Setting(contentEl).setName("Story Title").setDesc("Enter the title for your new story").addText(
      (text) => text.setPlaceholder("My New Story").setValue(this.title).onChange((value) => {
        this.title = value;
      }).inputEl.addEventListener("keypress", (e) => {
        if (e.key === "Enter") {
          this.submit();
        }
      })
    );
    new import_obsidian6.Setting(contentEl).setName("Sync to Obsidian").setDesc("Automatically sync the story files to your vault after creation").addToggle(
      (toggle) => toggle.setValue(this.shouldSync).onChange((value) => {
        this.shouldSync = value;
      })
    );
    const buttonContainer = contentEl.createDiv({ cls: "modal-button-container" });
    const createButton = buttonContainer.createEl("button", {
      text: "Create",
      cls: "mod-cta"
    });
    createButton.addEventListener("click", () => this.submit());
    const cancelButton = buttonContainer.createEl("button", {
      text: "Cancel"
    });
    cancelButton.addEventListener("click", () => this.close());
    const titleInput = contentEl.querySelector("input");
    if (titleInput) {
      titleInput.focus();
    }
  }
  submit() {
    const trimmedTitle = this.title.trim();
    if (!trimmedTitle) {
      new import_obsidian6.Notice("Please enter a story title", 3e3);
      return;
    }
    this.close();
    this.onSubmit(trimmedTitle, this.shouldSync);
  }
  onClose() {
    const { contentEl } = this;
    contentEl.empty();
  }
};

// src/sync/fileManager.ts
var import_obsidian7 = require("obsidian");
var FileManager = class {
  constructor(vault, basePath = "Stories") {
    this.vault = vault;
    this.basePath = basePath;
  }
  getVault() {
    return this.vault;
  }
  // Parse frontmatter from markdown content
  parseFrontmatter(content) {
    const frontmatterRegex = /^---\s*\n([\s\S]*?)\n---\s*\n([\s\S]*)$/;
    const match = content.match(frontmatterRegex);
    if (!match) {
      return {
        frontmatter: {
          id: "",
          type: "story"
        },
        body: content
      };
    }
    const frontmatterText = match[1];
    const body = match[2];
    const frontmatter = {
      id: "",
      type: "story"
    };
    const lines = frontmatterText.split("\n");
    for (const line of lines) {
      const colonIndex = line.indexOf(":");
      if (colonIndex === -1)
        continue;
      const key = line.substring(0, colonIndex).trim();
      const value = line.substring(colonIndex + 1).trim().replace(/^["']|["']$/g, "");
      switch (key) {
        case "id":
          frontmatter.id = value;
          break;
        case "type":
          if (value === "story" || value === "chapter" || value === "scene" || value === "beat") {
            frontmatter.type = value;
          }
          break;
        case "story_id":
          frontmatter.story_id = value;
          break;
        case "chapter_id":
          frontmatter.chapter_id = value;
          break;
        case "scene_id":
          frontmatter.scene_id = value;
          break;
        case "number":
          frontmatter.number = parseInt(value, 10);
          break;
        case "version":
          frontmatter.version = parseInt(value, 10);
          break;
        case "synced_at":
          frontmatter.synced_at = value;
          break;
      }
    }
    return { frontmatter, body };
  }
  // Serialize frontmatter and body to markdown
  serializeFrontmatter(frontmatter, body) {
    const lines = ["---"];
    lines.push(`id: ${frontmatter.id}`);
    lines.push(`type: ${frontmatter.type}`);
    if (frontmatter.story_id) {
      lines.push(`story_id: ${frontmatter.story_id}`);
    }
    if (frontmatter.chapter_id) {
      lines.push(`chapter_id: ${frontmatter.chapter_id}`);
    }
    if (frontmatter.scene_id) {
      lines.push(`scene_id: ${frontmatter.scene_id}`);
    }
    if (frontmatter.number !== void 0) {
      lines.push(`number: ${frontmatter.number}`);
    }
    if (frontmatter.version !== void 0) {
      lines.push(`version: ${frontmatter.version}`);
    }
    if (frontmatter.synced_at) {
      lines.push(`synced_at: ${frontmatter.synced_at}`);
    }
    lines.push("---");
    lines.push("");
    return lines.join("\n") + body;
  }
  // Sanitize filename for filesystem
  sanitizeFilename(name) {
    return name.replace(/[<>:"/\\|?*]/g, "-").replace(/\s+/g, " ").trim();
  }
  // Get story folder path
  getStoryFolderPath(storyTitle) {
    const sanitizedTitle = this.sanitizeFilename(storyTitle);
    return `${this.basePath}/${sanitizedTitle}`;
  }
  // Ensure folder exists
  async ensureFolderExists(path) {
    const folders = path.split("/");
    let currentPath = "";
    for (const folder of folders) {
      if (folder === "")
        continue;
      currentPath = currentPath ? `${currentPath}/${folder}` : folder;
      const folderExists = await this.vault.adapter.exists(currentPath);
      if (!folderExists) {
        await this.vault.createFolder(currentPath);
      }
    }
  }
  // Write story metadata file
  async writeStoryMetadata(story, folderPath) {
    await this.ensureFolderExists(folderPath);
    const frontmatter = {
      id: story.id,
      type: "story",
      version: story.version_number,
      synced_at: (/* @__PURE__ */ new Date()).toISOString()
    };
    const body = `# ${story.title}

**Status**: ${story.status}
**Version**: ${story.version_number}
**Created**: ${new Date(story.created_at).toLocaleString()}
**Updated**: ${new Date(story.updated_at).toLocaleString()}

## Synopsis
(User can add content here)
`;
    const content = this.serializeFrontmatter(frontmatter, body);
    const filePath = `${folderPath}/metadata.md`;
    const existingFile = this.vault.getAbstractFileByPath(filePath);
    if (existingFile instanceof import_obsidian7.TFile) {
      await this.vault.modify(existingFile, content);
    } else {
      await this.vault.create(filePath, content);
    }
  }
  // Read story metadata file
  async readStoryMetadata(folderPath) {
    const filePath = `${folderPath}/metadata.md`;
    const file = this.vault.getAbstractFileByPath(filePath);
    if (!(file instanceof import_obsidian7.TFile)) {
      throw new Error(`Metadata file not found: ${filePath}`);
    }
    const content = await this.vault.read(file);
    const parsed = this.parseFrontmatter(content);
    return { frontmatter: parsed.frontmatter, content: parsed.body };
  }
  // Write chapter file with scenes and beats
  async writeChapterFile(chapter, filePath) {
    const frontmatter = {
      id: chapter.chapter.id,
      type: "chapter",
      story_id: chapter.chapter.story_id,
      number: chapter.chapter.number,
      synced_at: (/* @__PURE__ */ new Date()).toISOString()
    };
    let body = `# Chapter ${chapter.chapter.number}: ${chapter.chapter.title}

`;
    if (chapter.scenes.length === 0) {
      body += "(No scenes yet)\n";
    } else {
      body += "## Scenes\n\n";
      for (const sceneWithBeats of chapter.scenes) {
        const scene = sceneWithBeats.scene;
        body += `### Scene ${scene.order_num}`;
        if (scene.goal) {
          body += `: ${scene.goal}`;
        }
        body += `
`;
        body += `**ID**: ${scene.id}
`;
        body += `**Order**: ${scene.order_num}
`;
        if (scene.time_ref) {
          body += `**Time**: ${scene.time_ref}
`;
        }
        if (scene.goal) {
          body += `**Goal**: ${scene.goal}
`;
        }
        body += `
`;
        if (sceneWithBeats.beats.length > 0) {
          body += `#### Beats

`;
          for (const beat of sceneWithBeats.beats) {
            body += `**Beat ${beat.order_num}** (${beat.type})
`;
            if (beat.intent) {
              body += `- **Intent**: ${beat.intent}
`;
            }
            if (beat.outcome) {
              body += `- **Outcome**: ${beat.outcome}
`;
            }
            body += `
`;
          }
        }
        body += `---

`;
      }
    }
    const content = this.serializeFrontmatter(frontmatter, body);
    const existingFile = this.vault.getAbstractFileByPath(filePath);
    if (existingFile instanceof import_obsidian7.TFile) {
      await this.vault.modify(existingFile, content);
    } else {
      await this.vault.create(filePath, content);
    }
  }
  // Read chapter file
  async readChapterFile(filePath) {
    const file = this.vault.getAbstractFileByPath(filePath);
    if (!(file instanceof import_obsidian7.TFile)) {
      throw new Error(`Chapter file not found: ${filePath}`);
    }
    const content = await this.vault.read(file);
    const parsed = this.parseFrontmatter(content);
    return { frontmatter: parsed.frontmatter, content: parsed.body };
  }
  // Update frontmatter in a file
  async updateFrontmatter(filePath, updates) {
    const file = this.vault.getAbstractFileByPath(filePath);
    if (!(file instanceof import_obsidian7.TFile)) {
      throw new Error(`File not found: ${filePath}`);
    }
    const content = await this.vault.read(file);
    const { frontmatter, body } = this.parseFrontmatter(content);
    const updatedFrontmatter = {
      ...frontmatter,
      ...updates
    };
    const newContent = this.serializeFrontmatter(updatedFrontmatter, body);
    await this.vault.modify(file, newContent);
  }
};

// src/sync/syncService.ts
var import_obsidian8 = require("obsidian");

// src/sync/markdownParser.ts
var MarkdownParser = class {
  // Parse scenes and beats from chapter markdown content
  static parseChapterMarkdown(content) {
    var _a;
    const scenes = [];
    const sections = content.split(/^---$/m);
    for (const section of sections) {
      const trimmed = section.trim();
      if (!trimmed)
        continue;
      const sceneMatch = trimmed.match(
        /###\s*Scene\s*(\d+)(?::\s*(.+?))?\n([\s\S]*?)(?=####|###|---|$)/i
      );
      if (!sceneMatch)
        continue;
      const orderNum = parseInt(sceneMatch[1], 10);
      const sceneTitle = ((_a = sceneMatch[2]) == null ? void 0 : _a.trim()) || "";
      const sceneContent = sceneMatch[3] || "";
      const sceneIdMatch = sceneContent.match(/\*\*ID\*\*:\s*([a-f0-9-]+)/i);
      const timeRefMatch = sceneContent.match(/\*\*Time\*\*:\s*(.+?)(?:\n|$)/i);
      const goalMatch = sceneContent.match(/\*\*Goal\*\*:\s*(.+?)(?:\n|$)/i);
      const scene = {
        order_num: orderNum,
        goal: goalMatch ? goalMatch[1].trim() : sceneTitle || "",
        time_ref: timeRefMatch ? timeRefMatch[1].trim() : ""
      };
      if (sceneIdMatch) {
        scene.id = sceneIdMatch[1].trim();
      }
      const beats = [];
      const beatsSection = sceneContent.match(/####\s*Beats\s*\n([\s\S]*?)(?=---|$)/i);
      if (beatsSection) {
        const beatsText = beatsSection[1];
        const beatMatches = beatsText.matchAll(
          /\*\*Beat\s*(\d+)\*\*\s*\(([^)]+)\)\s*\n([\s\S]*?)(?=\*\*Beat|$)/gi
        );
        for (const beatMatch of beatMatches) {
          const beatOrderNum = parseInt(beatMatch[1], 10);
          const beatType = beatMatch[2].trim();
          const beatContent = beatMatch[3] || "";
          const intentMatch = beatContent.match(
            /-\s*\*\*Intent\*\*:\s*(.+?)(?:\n|$)/i
          );
          const outcomeMatch = beatContent.match(
            /-\s*\*\*Outcome\*\*:\s*(.+?)(?:\n|$)/i
          );
          const beat = {
            order_num: beatOrderNum,
            type: beatType,
            intent: intentMatch ? intentMatch[1].trim() : "",
            outcome: outcomeMatch ? outcomeMatch[1].trim() : ""
          };
          beats.push(beat);
        }
      }
      scenes.push({ scene, beats });
    }
    return scenes;
  }
  // Build chapter markdown from chapter data
  static buildChapterMarkdown(chapterTitle, chapterNumber, scenes) {
    let content = `# Chapter ${chapterNumber}: ${chapterTitle}

`;
    if (scenes.length === 0) {
      content += "(No scenes yet)\n";
    } else {
      content += "## Scenes\n\n";
      for (const sceneData of scenes) {
        const scene = sceneData.scene;
        content += `### Scene ${scene.order_num}`;
        if (scene.goal) {
          content += `: ${scene.goal}`;
        }
        content += `
`;
        if (scene.id) {
          content += `**ID**: ${scene.id}
`;
        }
        content += `**Order**: ${scene.order_num}
`;
        if (scene.time_ref) {
          content += `**Time**: ${scene.time_ref}
`;
        }
        if (scene.goal) {
          content += `**Goal**: ${scene.goal}
`;
        }
        content += `
`;
        if (sceneData.beats.length > 0) {
          content += `#### Beats

`;
          for (const beat of sceneData.beats) {
            content += `**Beat ${beat.order_num}** (${beat.type})
`;
            if (beat.intent) {
              content += `- **Intent**: ${beat.intent}
`;
            }
            if (beat.outcome) {
              content += `- **Outcome**: ${beat.outcome}
`;
            }
            content += `
`;
          }
        }
        content += `---

`;
      }
    }
    return content;
  }
};

// src/sync/syncService.ts
var SyncService = class {
  constructor(apiClient, fileManager, settings) {
    this.apiClient = apiClient;
    this.fileManager = fileManager;
    this.settings = settings;
  }
  // Pull story from service to Obsidian (Service → Obsidian)
  async pullStory(storyId) {
    try {
      const storyData = await this.apiClient.getStoryWithHierarchy(storyId);
      const folderPath = this.fileManager.getStoryFolderPath(
        storyData.story.title
      );
      await this.fileManager.writeStoryMetadata(
        storyData.story,
        folderPath
      );
      const chaptersFolderPath = `${folderPath}/chapters`;
      await this.fileManager.ensureFolderExists(chaptersFolderPath);
      for (const chapterWithContent of storyData.chapters) {
        const chapterFileName = `Chapter-${chapterWithContent.chapter.number}.md`;
        const chapterFilePath = `${chaptersFolderPath}/${chapterFileName}`;
        await this.fileManager.writeChapterFile(
          chapterWithContent,
          chapterFilePath
        );
      }
      const existingMetadata = await this.fileManager.readStoryMetadata(folderPath).catch(() => null);
      if (existingMetadata && existingMetadata.frontmatter.version !== void 0 && existingMetadata.frontmatter.version !== storyData.story.version_number) {
        await this.createVersionSnapshot(
          folderPath,
          existingMetadata.frontmatter.version
        );
      }
      new import_obsidian8.Notice(`Story "${storyData.story.title}" synced successfully`);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to sync story";
      new import_obsidian8.Notice(`Error syncing story: ${errorMessage}`, 5e3);
      throw err;
    }
  }
  // Pull all stories
  async pullAllStories() {
    if (!this.settings.tenantId) {
      throw new Error("Tenant ID is required");
    }
    const stories = await this.apiClient.listStories(this.settings.tenantId);
    for (const story of stories) {
      try {
        await this.pullStory(story.id);
      } catch (err) {
        console.error(`Failed to sync story ${story.id}:`, err);
      }
    }
    new import_obsidian8.Notice(`Synced ${stories.length} stories`);
  }
  // Push story from Obsidian to service (Obsidian → Service)
  async pushStory(folderPath) {
    try {
      const { frontmatter: storyFrontmatter, content: storyContent } = await this.fileManager.readStoryMetadata(folderPath);
      if (!storyFrontmatter.id) {
        throw new Error("Story metadata missing ID");
      }
      const storyId = storyFrontmatter.id;
      const titleMatch = storyContent.match(/^#\s+(.+)$/m);
      if (titleMatch) {
        const newTitle = titleMatch[1].trim();
        await this.apiClient.updateStory(storyId, newTitle);
      }
      const vault = this.fileManager.getVault();
      const chaptersFolder = vault.getAbstractFileByPath(
        `${folderPath}/chapters`
      );
      if (chaptersFolder && chaptersFolder instanceof import_obsidian8.TFolder) {
        const chapterFiles = chaptersFolder.children.filter(
          (file) => file instanceof import_obsidian8.TFile && file.extension === "md"
        );
        for (const chapterFile of chapterFiles) {
          await this.syncChapterFromFile(chapterFile.path, storyId);
        }
      }
      new import_obsidian8.Notice("Story pushed to service successfully");
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to push story";
      new import_obsidian8.Notice(`Error pushing story: ${errorMessage}`, 5e3);
      throw err;
    }
  }
  // Sync chapter from markdown file
  async syncChapterFromFile(filePath, storyId) {
    const { frontmatter, content } = await this.fileManager.readChapterFile(
      filePath
    );
    const titleMatch = content.match(/^#\s+Chapter\s+\d+:\s*(.+)$/m);
    const numberMatch = content.match(/^#\s+Chapter\s+(\d+):/m);
    if (!titleMatch || !numberMatch) {
      throw new Error(`Invalid chapter format in ${filePath}`);
    }
    const chapterTitle = titleMatch[1].trim();
    const chapterNumber = parseInt(numberMatch[1], 10);
    let chapterId;
    if (frontmatter.id) {
      chapterId = frontmatter.id;
      await this.apiClient.updateChapter(frontmatter.id, {
        title: chapterTitle,
        number: chapterNumber
      });
    } else {
      const newChapter = await this.apiClient.createChapter(storyId, {
        title: chapterTitle,
        number: chapterNumber
      });
      chapterId = newChapter.id;
      await this.fileManager.updateFrontmatter(filePath, {
        id: newChapter.id
      });
    }
    const parsedScenes = MarkdownParser.parseChapterMarkdown(content);
    for (const parsedScene of parsedScenes) {
      const scene = parsedScene.scene;
      if (!scene.order_num)
        continue;
      let sceneId;
      if (scene.id) {
        sceneId = scene.id;
        await this.apiClient.updateScene(sceneId, {
          order_num: scene.order_num,
          goal: scene.goal || "",
          time_ref: scene.time_ref || ""
        });
      } else {
        const newScene = await this.apiClient.createScene({
          story_id: storyId,
          chapter_id: chapterId,
          order_num: scene.order_num,
          goal: scene.goal || "",
          time_ref: scene.time_ref || ""
        });
        sceneId = newScene.id;
      }
      for (const beat of parsedScene.beats) {
        if (!beat.order_num || !beat.type)
          continue;
        if (beat.id) {
          await this.apiClient.updateBeat(beat.id, {
            order_num: beat.order_num,
            type: beat.type,
            intent: beat.intent || "",
            outcome: beat.outcome || ""
          });
        } else {
          await this.apiClient.createBeat({
            scene_id: sceneId,
            order_num: beat.order_num,
            type: beat.type,
            intent: beat.intent || "",
            outcome: beat.outcome || ""
          });
        }
      }
    }
  }
  // Bidirectional sync (pull then push)
  async syncStory(storyId) {
    await this.pullStory(storyId);
    const story = await this.apiClient.getStory(storyId);
    const folderPath = this.fileManager.getStoryFolderPath(story.title);
    await this.pushStory(folderPath);
  }
  // Create version snapshot
  async createVersionSnapshot(storyFolder, versionNumber) {
    const versionsFolder = `${storyFolder}/versions`;
    await this.fileManager.ensureFolderExists(versionsFolder);
    const versionFolder = `${versionsFolder}/v${versionNumber}`;
    await this.fileManager.ensureFolderExists(versionFolder);
    const metadataSource = `${storyFolder}/metadata.md`;
    const metadataDest = `${versionFolder}/metadata.md`;
    const vault = this.fileManager.getVault();
    const metadataFile = vault.getAbstractFileByPath(metadataSource);
    if (metadataFile instanceof import_obsidian8.TFile) {
      const content = await vault.read(metadataFile);
      await vault.create(metadataDest, content);
    }
    const chaptersSource = `${storyFolder}/chapters`;
    const chaptersDest = `${versionFolder}/chapters`;
    const chaptersFolder = vault.getAbstractFileByPath(chaptersSource);
    if (chaptersFolder && chaptersFolder instanceof import_obsidian8.TFolder) {
      await this.fileManager.ensureFolderExists(chaptersDest);
      for (const file of chaptersFolder.children) {
        if (file instanceof import_obsidian8.TFile) {
          const content = await vault.read(file);
          const destPath = `${chaptersDest}/${file.name}`;
          await vault.create(destPath, content);
        }
      }
    }
  }
};

// src/main.ts
var DEFAULT_SETTINGS = {
  apiUrl: "http://localhost:8080",
  apiKey: "",
  tenantId: "",
  tenantName: "",
  syncFolderPath: "Stories",
  autoVersionSnapshots: true,
  conflictResolution: "service"
};
var StoryEnginePlugin = class extends import_obsidian9.Plugin {
  async onload() {
    await this.loadSettings();
    this.apiClient = new StoryEngineClient(
      this.settings.apiUrl,
      this.settings.apiKey
    );
    this.fileManager = new FileManager(
      this.app.vault,
      this.settings.syncFolderPath || "Stories"
    );
    this.syncService = new SyncService(
      this.apiClient,
      this.fileManager,
      this.settings
    );
    this.addSettingTab(new StoryEngineSettingTab(this.app, this));
    registerCommands(this);
  }
  async onunload() {
  }
  async loadSettings() {
    this.settings = Object.assign(
      {},
      DEFAULT_SETTINGS,
      await this.loadData()
    );
  }
  async saveSettings() {
    await this.saveData(this.settings);
    this.apiClient = new StoryEngineClient(
      this.settings.apiUrl,
      this.settings.apiKey
    );
    this.fileManager = new FileManager(
      this.app.vault,
      this.settings.syncFolderPath || "Stories"
    );
    this.syncService = new SyncService(
      this.apiClient,
      this.fileManager,
      this.settings
    );
  }
  async createStoryCommand() {
    var _a;
    const tenantId = (_a = this.settings.tenantId) == null ? void 0 : _a.trim();
    if (!tenantId) {
      new import_obsidian9.Notice("Please configure Tenant ID in settings", 5e3);
      return;
    }
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    if (!uuidRegex.test(tenantId)) {
      new import_obsidian9.Notice("Invalid Tenant ID format. Please check your settings.", 5e3);
      return;
    }
    new CreateStoryModal(this.app, async (title, shouldSync) => {
      try {
        new import_obsidian9.Notice(`Creating story "${title}"...`);
        const story = await this.apiClient.createStory(tenantId, title);
        new import_obsidian9.Notice(`Story "${title}" created successfully`);
        if (shouldSync) {
          try {
            new import_obsidian9.Notice(`Syncing story to Obsidian...`);
            await this.syncService.pullStory(story.id);
            new import_obsidian9.Notice(`Story synced to your vault!`);
          } catch (syncErr) {
            const syncErrorMessage = syncErr instanceof Error ? syncErr.message : "Failed to sync story";
            new import_obsidian9.Notice(`Story created but sync failed: ${syncErrorMessage}`, 5e3);
          }
        } else {
          new StoryDetailsModal(this, story).open();
        }
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : "Failed to create story";
        new import_obsidian9.Notice(`Error: ${errorMessage}`, 5e3);
      }
    }).open();
  }
};
