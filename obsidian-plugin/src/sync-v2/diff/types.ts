import type { FenceType } from "../parsers/contentsParser";

export type DiffOperationKind = "created" | "updated" | "deleted" | "moved" | "reordered";

export interface DiffOperation {
	kind: DiffOperationKind;
	fenceId: string;
	fenceType: FenceType;
	metadata?: {
		oldOrder?: number;
		newOrder?: number;
		oldParentId?: string;
		newParentId?: string;
	};
}

export interface DiffResult {
	operations: DiffOperation[];
	untrackedSegments: string[];
}

