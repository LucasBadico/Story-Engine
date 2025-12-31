import { StoryListModal } from "./views/StoryListModal";
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
}

