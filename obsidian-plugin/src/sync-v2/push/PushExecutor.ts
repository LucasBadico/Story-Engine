import type { StoryEngineClient } from "../../api/client";
import type { SyncError } from "../types/sync";
import type { PushAction } from "./PushPlanner";

export interface PushExecutionSummary {
	applied: number;
	errors: SyncError[];
}

export class PushExecutor {
	constructor(private readonly apiClient: StoryEngineClient) {}

	async execute(actions: PushAction[]): Promise<PushExecutionSummary> {
		const summary: PushExecutionSummary = {
			applied: 0,
			errors: [],
		};

		for (const action of actions) {
			try {
				await this.applyAction(action);
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

	private async applyAction(action: PushAction): Promise<void> {
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
		}
	}
}

