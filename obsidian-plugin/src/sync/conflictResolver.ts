import { Story, Chapter, Scene, Beat } from "../types";

export type ConflictResolutionStrategy = "service" | "local" | "manual";

export interface ConflictInfo {
	entityType: "story" | "chapter" | "scene" | "beat";
	localUpdatedAt: string;
	serviceUpdatedAt: string;
}

export class ConflictResolver {
	constructor(private strategy: ConflictResolutionStrategy) {}

	// Determine which version wins based on strategy
	resolveConflict(conflict: ConflictInfo): "service" | "local" {
		switch (this.strategy) {
			case "service":
				return "service";
			case "local":
				return "local";
			case "manual":
				// For manual, default to newer timestamp
				const localTime = new Date(conflict.localUpdatedAt).getTime();
				const serviceTime = new Date(conflict.serviceUpdatedAt).getTime();
				return serviceTime > localTime ? "service" : "local";
			default:
				return "service";
		}
	}

	// Compare timestamps to detect conflicts
	hasConflict(
		localUpdatedAt: string,
		serviceUpdatedAt: string
	): boolean {
		// Consider it a conflict if times differ by more than 1 second
		// (to account for sync timing)
		const localTime = new Date(localUpdatedAt).getTime();
		const serviceTime = new Date(serviceUpdatedAt).getTime();
		const diff = Math.abs(localTime - serviceTime);
		return diff > 1000; // More than 1 second difference
	}
}

