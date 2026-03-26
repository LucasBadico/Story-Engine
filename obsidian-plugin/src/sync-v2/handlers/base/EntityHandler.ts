import type { SyncContext } from "../../types/sync";

export abstract class EntityHandler<T> {
	abstract readonly entityType: string;

	abstract pull(id: string, context: SyncContext): Promise<T>;
	abstract push(entity: T, context: SyncContext): Promise<void>;
	abstract delete(id: string, context: SyncContext): Promise<void>;
	protected abstract generateFiles(entity: T, context: SyncContext): Promise<void>;

	protected async ensureBackup(context: SyncContext, reason: string): Promise<void> {
		if (context.backupMode === "off") return;
		const timestamp = context.timestamp();
		const message = `[StoryEngine Sync] ${reason} @ ${timestamp}`;
		await context.fileManager.createSnapshot(message);
	}
}

