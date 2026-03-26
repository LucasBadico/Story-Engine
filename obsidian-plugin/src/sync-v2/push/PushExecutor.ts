import type { StoryEngineClient } from "../../api/client";
import type { SyncError } from "../types/sync";
import type { PushAction } from "./PushPlanner";
import type { ContentCitationService } from "./ContentCitationService";

export interface PushExecutionSummary {
	applied: number;
	errors: SyncError[];
}

export class PushExecutor {
	constructor(
		private readonly apiClient: StoryEngineClient,
		private readonly citationService?: ContentCitationService
	) {}

	async execute(actions: PushAction[], options?: { worldId?: string }): Promise<PushExecutionSummary> {
		const summary: PushExecutionSummary = {
			applied: 0,
			errors: [],
		};

		for (const action of actions) {
			try {
				await this.applyAction(action, options?.worldId);
				summary.applied += 1;
			} catch (error) {
				summary.errors.push({
					code: `push_${action.type}`,
					message:
						error instanceof Error
							? error.message
							: "Unknown error while executing push action.",
					details: action,
					recoverable: true,
				});
			}
		}

		return summary;
	}

	private async applyAction(action: PushAction, worldId?: string): Promise<void> {
		switch (action.type) {
			case "chapter_reorder":
				await this.apiClient.updateChapter(action.chapterId, {
					number: action.newOrder,
				});
				return;
			case "scene_reorder":
				await this.apiClient.updateScene(action.sceneId, {
					order_num: action.newOrder,
				});
				return;
			case "scene_move":
				await this.apiClient.moveScene(action.sceneId, action.toChapterId ?? null);
				return;
			case "beat_reorder":
				await this.apiClient.updateBeat(action.beatId, {
					order_num: action.newOrder,
				});
				return;
			case "beat_move":
				await this.apiClient.moveBeat(action.beatId, action.toSceneId);
				return;
			case "content_update":
				await this.apiClient.updateContentBlock(action.contentBlockId, {
					content: action.newContent,
				});
				await this.citationService?.syncCitations(action.contentBlockId, action.newContent, worldId);
				return;
		}
	}
}

