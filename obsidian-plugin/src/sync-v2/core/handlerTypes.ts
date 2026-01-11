import type { SyncContext } from "../types/sync";
import type { Story } from "../../types";

export interface EntityHandler<T> {
	readonly entityType: string;
	pull(id: string, context: SyncContext): Promise<T>;
	push(entity: T, context: SyncContext): Promise<void>;
	delete(id: string, context: SyncContext): Promise<void>;
	generateFile(entity: T, context: SyncContext): Promise<void>;
}

export interface StoryHierarchy {
	story: Story;
}

