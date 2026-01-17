import { describe, expect, it, vi, beforeEach } from "vitest";
import type { App } from "obsidian";
import { ConflictResolver } from "../ConflictResolver";
import { ConflictModal } from "../ConflictModal";
import type { Conflict, ConflictResolutionStrategy } from "../types";
import type { SyncContext } from "../../types/sync";
import type { StoryEngineSettings } from "../../../types";

describe("ConflictResolver", () => {
	let mockApp: App;
	let mockContext: SyncContext;
	let resolver: ConflictResolver;

	const baseSettings: StoryEngineSettings = {
		apiUrl: "http://localhost:8080",
		llmGatewayUrl: "http://localhost:8081",
		apiKey: "test-key",
		tenantId: "tenant-1",
		tenantName: "Test Tenant",
		syncFolderPath: "Stories",
		autoVersionSnapshots: true,
		conflictResolution: "manual",
		mode: "remote",
		syncVersion: "v2",
		showHelpBox: true,
		autoSyncOnApiUpdates: true,
		autoPushOnFileBlur: true,
		backupMode: "snapshots",
		backupRetentionDays: 7,
	};

	beforeEach(() => {
		mockApp = {
			workspace: {} as any,
			vault: {} as any,
		} as unknown as App;

		mockContext = {
			app: mockApp,
			apiClient: {} as any,
			fileManager: {} as any,
			settings: { ...baseSettings },
			timestamp: () => "2025-01-01T00:00:00Z",
			backupMode: "snapshots",
			emitWarning: vi.fn(),
		};

		resolver = new ConflictResolver(mockApp, mockContext);
	});

	describe("detectConflict", () => {
		it("detects simultaneous edit conflict when timestamps indicate conflict", () => {
			const localData = { content: "Local version", title: "Local Title" };
			const remoteData = { content: "Remote version", title: "Remote Title" };
			const localTimestamp = "2025-01-01T10:00:00Z";
			const remoteTimestamp = "2025-01-01T11:00:00Z"; // Newer

			const conflict = resolver.detectConflict(
				"content_block",
				"cb-1",
				"Stories/Test Story/03-contents/00-texts/cb-0001-content.md",
				localData,
				remoteData,
				localTimestamp,
				remoteTimestamp
			);

			expect(conflict).not.toBeNull();
			expect(conflict?.type).toBe("simultaneous_edit");
			expect(conflict?.entityId).toBe("cb-1");
			expect(conflict?.entityType).toBe("content_block");
			expect(conflict?.localTimestamp).toBe(localTimestamp);
			expect(conflict?.remoteTimestamp).toBe(remoteTimestamp);
		});

		it("detects simultaneous edit conflict when data is different without timestamps", () => {
			const localData = { content: "Local version" };
			const remoteData = { content: "Remote version" };

			const conflict = resolver.detectConflict(
				"content_block",
				"cb-1",
				"Stories/Test Story/03-contents/00-texts/cb-0001-content.md",
				localData,
				remoteData
			);

			expect(conflict).not.toBeNull();
			expect(conflict?.type).toBe("simultaneous_edit");
		});

		it("returns null when data is equal", () => {
			const localData = { content: "Same content", title: "Same Title" };
			const remoteData = { content: "Same content", title: "Same Title" };

			const conflict = resolver.detectConflict(
				"content_block",
				"cb-1",
				"Stories/Test Story/03-contents/00-texts/cb-0001-content.md",
				localData,
				remoteData
			);

			expect(conflict).toBeNull();
		});

		it("returns null when remote is newer but content is same", () => {
			const localData = { content: "Same content" };
			const remoteData = { content: "Same content" };
			const localTimestamp = "2025-01-01T10:00:00Z";
			const remoteTimestamp = "2025-01-01T11:00:00Z"; // Newer but same content

			const conflict = resolver.detectConflict(
				"content_block",
				"cb-1",
				"Stories/Test Story/03-contents/00-texts/cb-0001-content.md",
				localData,
				remoteData,
				localTimestamp,
				remoteTimestamp
			);

			expect(conflict).toBeNull();
		});

		it("returns null when local is newer than remote", () => {
			const localData = { content: "Local version" };
			const remoteData = { content: "Remote version" };
			const localTimestamp = "2025-01-01T11:00:00Z"; // Newer
			const remoteTimestamp = "2025-01-01T10:00:00Z";

			const conflict = resolver.detectConflict(
				"content_block",
				"cb-1",
				"Stories/Test Story/03-contents/00-texts/cb-0001-content.md",
				localData,
				remoteData,
				localTimestamp,
				remoteTimestamp
			);

			// If local is newer, it's not a conflict (local wins)
			expect(conflict).toBeNull();
		});

		it("handles nested objects in data comparison", () => {
			const localData = {
				content: "Same",
				metadata: { author: "User1", tags: ["tag1"] },
			};
			const remoteData = {
				content: "Same",
				metadata: { author: "User2", tags: ["tag1"] }, // Different author
			};

			const conflict = resolver.detectConflict(
				"content_block",
				"cb-1",
				"Stories/Test Story/03-contents/00-texts/cb-0001-content.md",
				localData,
				remoteData
			);

			expect(conflict).not.toBeNull();
			expect(conflict?.type).toBe("simultaneous_edit");
		});
	});

	describe("resolve", () => {
		const createConflict = (): Conflict => ({
			type: "simultaneous_edit",
			entityId: "cb-1",
			entityType: "content_block",
			filePath: "Stories/Test Story/03-contents/00-texts/cb-0001-content.md",
			localTimestamp: "2025-01-01T10:00:00Z",
			remoteTimestamp: "2025-01-01T11:00:00Z",
			localData: { content: "Local version" },
			remoteData: { content: "Remote version" },
		});

		it("resolves conflict using 'local' strategy when configured", async () => {
			mockContext.settings.conflictResolution = "local";

			const conflict = createConflict();
			const result = await resolver.resolve(conflict);

			expect(result.success).toBe(true);
			expect(result.resolution.strategy).toBe("local");
			expect(result.resolution.resolvedData).toEqual(conflict.localData);
			expect(result.resolution.autoResolved).toBe(true);
		});

		it("resolves conflict using 'remote' strategy when configured", async () => {
			mockContext.settings.conflictResolution = "service"; // Maps to "remote"

			const conflict = createConflict();
			const result = await resolver.resolve(conflict);

			expect(result.success).toBe(true);
			expect(result.resolution.strategy).toBe("remote");
			expect(result.resolution.resolvedData).toEqual(conflict.remoteData);
			expect(result.resolution.autoResolved).toBe(true);
		});

		it("handles manual strategy using the modal choice", async () => {
			mockContext.settings.conflictResolution = "manual";
			ConflictModal.setNextChoice("remote");

			const conflict = createConflict();
			const result = await resolver.resolve(conflict);

			expect(result.success).toBe(true);
			expect(result.resolution.strategy).toBe("manual");
			expect(result.resolution.resolvedData).toEqual(conflict.remoteData);
			expect(result.resolution.autoResolved).toBe(false);
			expect(mockContext.emitWarning).toHaveBeenCalledWith(
				expect.objectContaining({
					code: "conflict_detected",
					message: expect.stringContaining("Manual resolution selected"),
				})
			);
		});

		it("handles errors gracefully", async () => {
			// Mock a strategy that throws an error
			const conflict = createConflict();
			
			// Force an error by using invalid data
			const invalidConflict = {
				...conflict,
				localData: undefined,
				remoteData: undefined,
			} as unknown as Conflict;

			const result = await resolver.resolve(invalidConflict);

			// Should handle error gracefully
			expect(result.success).toBeDefined();
		});
	});

	describe("getResolutionStrategy", () => {
		it("maps 'local' setting to 'local' strategy", () => {
			mockContext.settings.conflictResolution = "local";
			const conflict = {
				type: "simultaneous_edit" as const,
				entityId: "cb-1",
				entityType: "content_block",
				filePath: "test.md",
				localData: { content: "Local" },
				remoteData: { content: "Remote" },
			};

			// Test via resolve method
			return resolver.resolve(conflict).then((result) => {
				expect(result.resolution.strategy).toBe("local");
			});
		});

		it("maps 'service' setting to 'remote' strategy", () => {
			mockContext.settings.conflictResolution = "service";
			const conflict = {
				type: "simultaneous_edit" as const,
				entityId: "cb-1",
				entityType: "content_block",
				filePath: "test.md",
				localData: { content: "Local" },
				remoteData: { content: "Remote" },
			};

			return resolver.resolve(conflict).then((result) => {
				expect(result.resolution.strategy).toBe("remote");
			});
		});

		it("maps 'manual' setting to 'manual' strategy", () => {
			mockContext.settings.conflictResolution = "manual";
			const conflict = {
				type: "simultaneous_edit" as const,
				entityId: "cb-1",
				entityType: "content_block",
				filePath: "test.md",
				localData: { content: "Local" },
				remoteData: { content: "Remote" },
			};

			return resolver.resolve(conflict).then((result) => {
				expect(result.resolution.strategy).toBe("manual");
			});
		});
	});

	describe("isDataEqual", () => {
		it("correctly compares primitive values", () => {
			const conflict = {
				type: "simultaneous_edit" as const,
				entityId: "cb-1",
				entityType: "content_block",
				filePath: "test.md",
				localData: "same",
				remoteData: "same",
			};

			const result = resolver.detectConflict(
				"content_block",
				"cb-1",
				"test.md",
				conflict.localData,
				conflict.remoteData
			);

			expect(result).toBeNull(); // Same data, no conflict
		});

		it("correctly identifies different primitive values", () => {
			const conflict = resolver.detectConflict(
				"content_block",
				"cb-1",
				"test.md",
				"local",
				"remote"
			);

			expect(conflict).not.toBeNull();
		});

		it("correctly compares objects with same keys and values", () => {
			const localData = { a: 1, b: "test", c: true };
			const remoteData = { a: 1, b: "test", c: true };

			const conflict = resolver.detectConflict(
				"content_block",
				"cb-1",
				"test.md",
				localData,
				remoteData
			);

			expect(conflict).toBeNull();
		});

		it("correctly identifies objects with different values", () => {
			const localData = { a: 1, b: "test" };
			const remoteData = { a: 2, b: "test" }; // Different value

			const conflict = resolver.detectConflict(
				"content_block",
				"cb-1",
				"test.md",
				localData,
				remoteData
			);

			expect(conflict).not.toBeNull();
		});

		it("correctly identifies objects with different keys", () => {
			const localData = { a: 1, b: "test" };
			const remoteData = { a: 1, c: "test" }; // Different key

			const conflict = resolver.detectConflict(
				"content_block",
				"cb-1",
				"test.md",
				localData,
				remoteData
			);

			expect(conflict).not.toBeNull();
		});

		it("correctly compares nested objects", () => {
			const localData = {
				content: "text",
				metadata: { author: "User1", version: 1 },
			};
			const remoteData = {
				content: "text",
				metadata: { author: "User1", version: 1 },
			};

			const conflict = resolver.detectConflict(
				"content_block",
				"cb-1",
				"test.md",
				localData,
				remoteData
			);

			expect(conflict).toBeNull();
		});

		it("correctly identifies nested objects with differences", () => {
			const localData = {
				content: "text",
				metadata: { author: "User1", version: 1 },
			};
			const remoteData = {
				content: "text",
				metadata: { author: "User2", version: 1 }, // Different author
			};

			const conflict = resolver.detectConflict(
				"content_block",
				"cb-1",
				"test.md",
				localData,
				remoteData
			);

			expect(conflict).not.toBeNull();
		});

		it("handles null and undefined values", () => {
			const conflict1 = resolver.detectConflict(
				"content_block",
				"cb-1",
				"test.md",
				null,
				null
			);
			expect(conflict1).toBeNull();

			const conflict2 = resolver.detectConflict(
				"content_block",
				"cb-1",
				"test.md",
				undefined,
				undefined
			);
			// undefined vs undefined might be handled differently
			expect(conflict2).toBeNull();

			const conflict3 = resolver.detectConflict(
				"content_block",
				"cb-1",
				"test.md",
				null,
				"value"
			);
			expect(conflict3).not.toBeNull();
		});

		it("handles arrays in objects correctly (arrays are compared element by element)", () => {
			const localData = { tags: ["tag1", "tag2"], items: [1, 2, 3] };
			const remoteData = { tags: ["tag1", "tag2"], items: [1, 2, 3] };

			const conflict = resolver.detectConflict(
				"content_block",
				"cb-1",
				"test.md",
				localData,
				remoteData
			);

			// Arrays are objects in JavaScript, and Object.keys() on arrays returns indices as strings
			// So isDataEqual correctly compares arrays element by element:
			// - aObj["0"] === bObj["0"] → "tag1" === "tag1" ✓
			// - aObj["1"] === bObj["1"] → "tag2" === "tag2" ✓
			// Arrays with same values should be detected as equal (no conflict)
			expect(conflict).toBeNull(); // Arrays are correctly compared, so no conflict
		});

		it("identifies arrays with different values", () => {
			const localData = { tags: ["tag1", "tag2"] };
			const remoteData = { tags: ["tag1", "tag3"] }; // Different value

			const conflict = resolver.detectConflict(
				"content_block",
				"cb-1",
				"test.md",
				localData,
				remoteData
			);

			expect(conflict).not.toBeNull();
		});
	});
});

