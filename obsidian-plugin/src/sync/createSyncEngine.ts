import type { SyncEngine, SyncEngineDependencies } from "./engine";
import { SyncService } from "./syncService";
import { ModularSyncEngine } from "../sync-v2/core/ModularSyncEngine";

export function createSyncEngine(
	deps: SyncEngineDependencies,
	version: "v1" | "v2"
): SyncEngine {
	if (version === "v2") {
		return new ModularSyncEngine({
			app: deps.app,
			apiClient: deps.apiClient,
			fileManager: deps.fileManager,
			settings: deps.settings,
			timestamp: () => new Date().toISOString(),
			backupMode: deps.settings.backupMode ?? "snapshots",
		});
	}

	return new SyncService(deps.apiClient, deps.fileManager, deps.settings, deps.app);
}

