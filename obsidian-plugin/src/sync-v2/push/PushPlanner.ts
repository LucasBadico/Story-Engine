import { DiffEngine } from "../diff/DiffEngine";
import type { DiffOperation } from "../diff/types";
import {
	ContentsParser,
	type HierarchicalContent,
	type ParsedFence,
} from "../parsers/contentsParser";
import type { SyncWarning } from "../types/sync";

export type PushAction =
	| {
			type: "chapter_reorder";
			chapterId: string;
			oldOrder?: number;
			newOrder: number;
	  }
	| {
			type: "scene_reorder";
			sceneId: string;
			chapterId?: string;
			oldOrder?: number;
			newOrder: number;
	  }
	| {
			type: "scene_move";
			sceneId: string;
			fromChapterId?: string | null;
			toChapterId?: string | null;
	  }
	| {
			type: "beat_reorder";
			beatId: string;
			sceneId?: string;
			oldOrder?: number;
			newOrder: number;
	  }
	| {
			type: "beat_move";
			beatId: string;
			fromSceneId?: string | null;
			toSceneId: string;
	  }
	| {
			type: "content_update";
			contentBlockId: string;
			newContent: string;
	  };

export interface PushPlan {
	actions: PushAction[];
	unsupportedOperations: DiffOperation[];
	untrackedSegments: string[];
	warnings: SyncWarning[];
}

export class PushPlanner {
	constructor(
		private readonly diffEngine = new DiffEngine(),
		private readonly contentsParser = new ContentsParser()
	) {}

	buildPlan(remoteContents: string, localContents: string): PushPlan {
		const diffRemoteVsLocal = this.diffEngine.diffContents(remoteContents, localContents);
		const diffLocalVsRemote = this.diffEngine.diffContents(localContents, remoteContents);
		const localMap = this.buildFenceMap(localContents);

		const actions: PushAction[] = [];
		const unsupported: DiffOperation[] = [];
		const warnings: SyncWarning[] = [];

		for (const op of diffRemoteVsLocal.operations) {
			switch (op.fenceType) {
				case "chapter":
					this.handleChapterOperation(op, actions, unsupported);
					break;
				case "scene":
					this.handleSceneOperation(op, actions, unsupported, localMap);
					break;
				case "beat":
					this.handleBeatOperation(op, actions, unsupported, localMap);
					break;
				case "content":
					this.handleContentOperation(op, actions, unsupported, localMap);
					break;
				default:
					unsupported.push(op);
					break;
			}
		}

		if (unsupported.length > 0) {
			warnings.push({
				code: "push_unsupported_operations",
				message: `Detectamos ${unsupported.length} mudanças locais ainda não suportadas pelo push automático.`,
				details: unsupported,
				severity: "warning",
			});
		}

		if (diffLocalVsRemote.untrackedSegments.length > 0) {
			warnings.push({
				code: "push_untracked_segments",
				message: `Há ${diffLocalVsRemote.untrackedSegments.length} trechos fora das fences; revise antes de tentar enviar.`,
				details: diffLocalVsRemote.untrackedSegments,
				severity: "warning",
			});
		}

		return {
			actions,
			unsupportedOperations: unsupported,
			untrackedSegments: diffLocalVsRemote.untrackedSegments,
			warnings,
		};
	}

	private handleChapterOperation(
		op: DiffOperation,
		actions: PushAction[],
		unsupported: DiffOperation[]
	): void {
		if (op.kind === "reordered" && op.metadata?.newOrder) {
			actions.push({
				type: "chapter_reorder",
				chapterId: op.fenceId,
				oldOrder: op.metadata.oldOrder,
				newOrder: op.metadata.newOrder,
			});
			return;
		}
		unsupported.push(op);
	}

	private handleSceneOperation(
		op: DiffOperation,
		actions: PushAction[],
		unsupported: DiffOperation[],
		localMap: Map<string, ParsedFence>
	): void {
		if (op.kind === "reordered" && op.metadata?.newOrder) {
			const scene = localMap.get(op.fenceId);
			actions.push({
				type: "scene_reorder",
				sceneId: op.fenceId,
				chapterId: scene?.parentId,
				oldOrder: op.metadata.oldOrder,
				newOrder: op.metadata.newOrder,
			});
			return;
		}

		if (op.kind === "moved") {
			actions.push({
				type: "scene_move",
				sceneId: op.fenceId,
				fromChapterId: op.metadata?.oldParentId ?? null,
				toChapterId: op.metadata?.newParentId ?? null,
			});
			return;
		}

		unsupported.push(op);
	}

	private handleBeatOperation(
		op: DiffOperation,
		actions: PushAction[],
		unsupported: DiffOperation[],
		localMap: Map<string, ParsedFence>
	): void {
		if (op.kind === "reordered" && op.metadata?.newOrder) {
			const beat = localMap.get(op.fenceId);
			actions.push({
				type: "beat_reorder",
				beatId: op.fenceId,
				sceneId: beat?.parentId,
				oldOrder: op.metadata.oldOrder,
				newOrder: op.metadata.newOrder,
			});
			return;
		}

		if (op.kind === "moved" && op.metadata?.newParentId) {
			actions.push({
				type: "beat_move",
				beatId: op.fenceId,
				fromSceneId: op.metadata.oldParentId ?? null,
				toSceneId: op.metadata.newParentId,
			});
			return;
		}

		unsupported.push(op);
	}

	private handleContentOperation(
		op: DiffOperation,
		actions: PushAction[],
		unsupported: DiffOperation[],
		localMap: Map<string, ParsedFence>
	): void {
		if (op.kind === "updated") {
			const fence = localMap.get(op.fenceId);
			if (fence && fence.innerText.trim().length > 0) {
				actions.push({
					type: "content_update",
					contentBlockId: op.fenceId,
					newContent: fence.innerText.trim(),
				});
				return;
			}
		}

		unsupported.push(op);
	}

	private buildFenceMap(content: string): Map<string, ParsedFence> {
		const hierarchy = this.contentsParser.parseHierarchy(content);
		const flattened = this.flattenHierarchy(hierarchy);
		return new Map(flattened.map((fence) => [fence.id, fence]));
	}

	private flattenHierarchy(hierarchy: HierarchicalContent): ParsedFence[] {
		const result: ParsedFence[] = [];
		const visit = (fence: ParsedFence) => {
			result.push(fence);
			fence.children.forEach(visit);
		};

		const buckets = [
			hierarchy.chapters,
			hierarchy.orphanScenes,
			hierarchy.orphanBeats,
			hierarchy.orphanContents,
		];

		buckets.forEach((bucket) => bucket.forEach(visit));

		return result;
	}
}

