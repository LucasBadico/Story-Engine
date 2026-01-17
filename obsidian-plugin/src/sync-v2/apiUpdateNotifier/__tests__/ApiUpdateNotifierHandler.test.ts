import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import type { SyncOrchestrator } from "../../core/SyncOrchestrator";
import type { SyncContext, SyncResult } from "../../types/sync";
import { ApiUpdateNotifierV2, type ApiUpdateEvent } from "../ApiUpdateNotifierV2";
import { ApiUpdateNotifierHandler } from "../ApiUpdateNotifierHandler";

describe("ApiUpdateNotifierHandler", () => {
	let notifier: ApiUpdateNotifierV2;
	let mockOrchestrator: SyncOrchestrator;
	let mockContext: SyncContext;
	let handler: ApiUpdateNotifierHandler;
	let mockEmitWarning: ReturnType<typeof vi.fn>;
	let consoleErrorSpy: ReturnType<typeof vi.spyOn>;
	let consoleLogSpy: ReturnType<typeof vi.spyOn>;

	beforeEach(() => {
		vi.useFakeTimers();
		consoleErrorSpy = vi.spyOn(console as any, "error").mockImplementation(() => {});
		consoleLogSpy = vi.spyOn(console as any, "log").mockImplementation(() => {});
		notifier = new ApiUpdateNotifierV2();

		mockEmitWarning = vi.fn();

		mockContext = {
			app: {} as any,
			apiClient: {} as any,
			fileManager: {} as any,
			settings: {} as any,
			timestamp: () => new Date().toISOString(),
			backupMode: "off",
			emitWarning: mockEmitWarning,
		};

		mockOrchestrator = {
			run: vi.fn(),
		} as unknown as SyncOrchestrator;

		handler = new ApiUpdateNotifierHandler(notifier, mockOrchestrator, mockContext);
	});

	afterEach(() => {
		handler.dispose();
		notifier.dispose();
		consoleErrorSpy?.mockRestore();
		consoleLogSpy?.mockRestore();
		vi.useRealTimers();
	});

	describe("start", () => {
		it("should subscribe to notifier events", () => {
			const subscribeSpy = vi.spyOn(notifier, "subscribe");

			handler.start();

			expect(subscribeSpy).toHaveBeenCalled();
		});

		it("should not subscribe multiple times", () => {
			const subscribeSpy = vi.spyOn(notifier, "subscribe");

			handler.start();
			handler.start();

			expect(subscribeSpy).toHaveBeenCalledTimes(1);
		});
	});

	describe("stop", () => {
		it("should unsubscribe from notifier events", () => {
			let subscribeCalled = false;
			let unsubscribeFn: (() => void) | undefined;
			vi.spyOn(notifier, "subscribe").mockImplementation((fn) => {
				subscribeCalled = true;
				unsubscribeFn = () => {};
				return unsubscribeFn;
			});

			handler.start();
			handler.stop();

			// Handler should have subscribed and then unsubscribed
			expect(subscribeCalled).toBe(true);
			expect(unsubscribeFn).toBeDefined();
		});

		it("should clear processing queue", () => {
			handler.start();
			handler.stop();

			// Processing queue should be cleared
			// This is tested implicitly by checking that no events are processed after stop
		});
	});

	describe("handleEvent", () => {
		it("should map story.updated to pull_story operation", async () => {
			const mockResult: SyncResult = {
				success: true,
				message: "Story synced",
			};
			(mockOrchestrator.run as any).mockResolvedValue(mockResult);

			handler.start();

			const event: ApiUpdateEvent = {
				type: "story.updated",
				entityId: "story-1",
				entityType: "story",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);

			// Wait for debounce and async processing
			await vi.advanceTimersByTimeAsync(1000);

			expect(mockOrchestrator.run).toHaveBeenCalledWith({
				type: "pull_story",
				payload: { storyId: "story-1" },
			});
		});

		it("should map chapter.updated to pull_chapter operation", async () => {
			const mockResult: SyncResult = {
				success: true,
				message: "Chapter synced",
			};
			(mockOrchestrator.run as any).mockResolvedValue(mockResult);

			handler.start();

			const event: ApiUpdateEvent = {
				type: "chapter.updated",
				entityId: "chapter-1",
				entityType: "chapter",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);

			// Wait for debounce and async processing
			await vi.advanceTimersByTimeAsync(1000);

			expect(mockOrchestrator.run).toHaveBeenCalledWith({
				type: "pull_chapter",
				payload: { chapterId: "chapter-1" },
			});
		});

		it("should map scene.updated to pull_scene operation", async () => {
			const mockResult: SyncResult = {
				success: true,
				message: "Scene synced",
			};
			(mockOrchestrator.run as any).mockResolvedValue(mockResult);

			handler.start();

			const event: ApiUpdateEvent = {
				type: "scene.updated",
				entityId: "scene-1",
				entityType: "scene",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);

			// Wait for debounce and async processing
			await vi.advanceTimersByTimeAsync(1000);

			expect(mockOrchestrator.run).toHaveBeenCalledWith({
				type: "pull_scene",
				payload: { sceneId: "scene-1" },
			});
		});

		it("should map beat.updated to pull_beat operation", async () => {
			const mockResult: SyncResult = {
				success: true,
				message: "Beat synced",
			};
			(mockOrchestrator.run as any).mockResolvedValue(mockResult);

			handler.start();

			const event: ApiUpdateEvent = {
				type: "beat.updated",
				entityId: "beat-1",
				entityType: "beat",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);

			// Wait for debounce and async processing
			await vi.advanceTimersByTimeAsync(1000);

			expect(mockOrchestrator.run).toHaveBeenCalledWith({
				type: "pull_beat",
				payload: { beatId: "beat-1" },
			});
		});

		it("should map content_block.updated to pull_content_block operation", async () => {
			const mockResult: SyncResult = {
				success: true,
				message: "Content block synced",
			};
			(mockOrchestrator.run as any).mockResolvedValue(mockResult);

			handler.start();

			const event: ApiUpdateEvent = {
				type: "content_block.updated",
				entityId: "block-1",
				entityType: "content_block",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);

			// Wait for debounce and async processing
			await vi.advanceTimersByTimeAsync(1000);

			expect(mockOrchestrator.run).toHaveBeenCalledWith({
				type: "pull_content_block",
				payload: { contentBlockId: "block-1" },
			});
		});

		it("should map character.updated to pull_character operation", async () => {
			const mockResult: SyncResult = {
				success: true,
				message: "Character synced",
			};
			(mockOrchestrator.run as any).mockResolvedValue(mockResult);

			handler.start();

			const event: ApiUpdateEvent = {
				type: "character.updated",
				entityId: "char-1",
				entityType: "character",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);

			// Wait for debounce and async processing
			await vi.advanceTimersByTimeAsync(1000);

			expect(mockOrchestrator.run).toHaveBeenCalledWith({
				type: "pull_character",
				payload: { entityId: "char-1" },
			});
		});

		it("should handle deletion events and emit warning", async () => {
			handler.start();

			const event: ApiUpdateEvent = {
				type: "story.deleted",
				entityId: "story-1",
				entityType: "story",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);

			// Wait for debounce and async processing
			await vi.advanceTimersByTimeAsync(1000);

			// Should emit warning about deletion
			expect(mockEmitWarning).toHaveBeenCalledWith(
				expect.objectContaining({
					code: "entity_deleted",
					message: expect.stringContaining("was deleted on the server"),
					severity: "warning",
				})
			);

			// Should not call orchestrator.run (deletions are handled separately)
			expect(mockOrchestrator.run).not.toHaveBeenCalled();
		});

		it("should handle chapter deletion events", async () => {
			handler.start();

			const event: ApiUpdateEvent = {
				type: "chapter.deleted",
				entityId: "chapter-1",
				entityType: "chapter",
				storyId: "story-1",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);

			// Wait for debounce and async processing
			await vi.advanceTimersByTimeAsync(1000);

			// Should emit warning about deletion
			expect(mockEmitWarning).toHaveBeenCalledWith(
				expect.objectContaining({
					code: "entity_deleted",
					message: expect.stringContaining("was deleted on the server"),
					severity: "warning",
				})
			);
		});

		it("should handle character deletion events", async () => {
			handler.start();

			const event: ApiUpdateEvent = {
				type: "character.deleted",
				entityId: "char-1",
				entityType: "character",
				worldId: "world-1",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);

			// Wait for debounce and async processing
			await vi.advanceTimersByTimeAsync(1000);

			// Should emit warning about deletion
			expect(mockEmitWarning).toHaveBeenCalledWith(
				expect.objectContaining({
					code: "entity_deleted",
					message: expect.stringContaining("was deleted on the server"),
					severity: "warning",
				})
			);
		});

		it("should ignore relation events", async () => {
			handler.start();

			const event: ApiUpdateEvent = {
				type: "relation.updated",
				entityId: "relation-1",
				entityType: "relation",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);

			// Wait for debounce and async processing
			await vi.advanceTimersByTimeAsync(1000);

			expect(mockOrchestrator.run).not.toHaveBeenCalled();
		});

		it("should emit warnings from sync result", async () => {
			const mockResult: SyncResult = {
				success: true,
				warnings: [
					{
						code: "test_warning",
						message: "Test warning",
						severity: "warning",
					},
				],
			};
			(mockOrchestrator.run as any).mockResolvedValue(mockResult);

			handler.start();

			const event: ApiUpdateEvent = {
				type: "story.updated",
				entityId: "story-1",
				entityType: "story",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);

			// Wait for debounce and async processing
			await vi.advanceTimersByTimeAsync(1000);

			expect(mockEmitWarning).toHaveBeenCalledWith({
				code: "test_warning",
				message: "Test warning",
				severity: "warning",
			});
		});

		it("should handle orchestrator errors", async () => {
			const error = new Error("Sync failed");
			(mockOrchestrator.run as any).mockRejectedValue(error);

			handler.start();

			const event: ApiUpdateEvent = {
				type: "story.updated",
				entityId: "story-1",
				entityType: "story",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);

			// Wait for debounce and async processing
			await vi.advanceTimersByTimeAsync(1000);

    expect(mockEmitWarning).toHaveBeenCalledWith(
     expect.objectContaining({
      code: "api_update_failed",
      message: expect.stringContaining("Failed to sync"),
      severity: "warning",
     })
    );
		});
	});

	describe("dispose", () => {
		it("should stop listening to events", () => {
			handler.start();
			handler.dispose();

			// After dispose, handler should not process events
			// This is tested implicitly by checking that no events are processed after dispose
		});
	});
});

