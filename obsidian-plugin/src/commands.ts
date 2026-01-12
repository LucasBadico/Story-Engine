import { Notice } from "obsidian";
import { StorySyncModal } from "./views/StorySyncModal";
import StoryEnginePlugin from "./main";

export function registerCommands(plugin: StoryEnginePlugin) {
	plugin.addCommand({
		id: "list-stories",
		name: "List Stories",
		callback: () => {
			plugin.activateView();
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
			// Only validate tenant ID in remote mode
			if (plugin.settings.mode === "remote" && !plugin.settings.tenantId) {
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

	plugin.addCommand({
		id: "extract-entities-from-selection",
		name: "Extract Entities and Relations from Selection",
		editorCallback: (editor) => {
			const selection = editor.getSelection();
			if (!selection.trim()) {
				new Notice("Select text to extract entities and relations", 3000);
				return;
			}
			plugin.extractSelectionCommand(selection, true);
		},
	});

	plugin.addCommand({
		id: "extract-entities-only-from-selection",
		name: "Extract Entities Only from Selection",
		editorCallback: (editor) => {
			const selection = editor.getSelection();
			if (!selection.trim()) {
				new Notice("Select text to extract entities", 3000);
				return;
			}
			plugin.extractSelectionCommand(selection, false);
		},
	});
}
