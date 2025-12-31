import { Notice } from "obsidian";
import { StoryListModal } from "./views/StoryListModal";
import { StorySyncModal } from "./views/StorySyncModal";
import StoryEnginePlugin from "./main";

export function registerCommands(plugin: StoryEnginePlugin) {
	plugin.addCommand({
		id: "list-stories",
		name: "List Stories",
		callback: () => {
			new StoryListModal(plugin).open();
		},
	});

	plugin.addCommand({
		id: "create-story",
		name: "Create Story",
		callback: () => {
			plugin.createStoryCommand();
		},
	});

	plugin.addCommand({
		id: "sync-story-from-service",
		name: "Sync Story from Service",
		callback: () => {
			new StorySyncModal(plugin, "pull").open();
		},
	});

	plugin.addCommand({
		id: "push-story-to-service",
		name: "Push Story to Service",
		callback: () => {
			new StorySyncModal(plugin, "push").open();
		},
	});

	plugin.addCommand({
		id: "sync-all-stories",
		name: "Sync All Stories",
		callback: async () => {
			if (!plugin.settings.tenantId) {
				new Notice("Please configure Tenant ID in settings", 5000);
				return;
			}

			try {
				new Notice("Syncing all stories...");
				await plugin.syncService.pullAllStories();
			} catch (err) {
				const errorMessage =
					err instanceof Error ? err.message : "Failed to sync stories";
				new Notice(`Error: ${errorMessage}`, 5000);
			}
		},
	});
}

