import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { ApiUpdateNotifierV2, type ApiUpdateEvent } from "../ApiUpdateNotifierV2";

describe("ApiUpdateNotifierV2", () => {
	let notifier: ApiUpdateNotifierV2;

	beforeEach(() => {
		vi.useFakeTimers();
		notifier = new ApiUpdateNotifierV2();
	});

	afterEach(() => {
		notifier.dispose();
		vi.useRealTimers();
	});

	describe("subscribe", () => {
		it("should add subscriber and return unsubscribe function", () => {
			const subscriber = vi.fn();
			const unsubscribe = notifier.subscribe(subscriber);

			expect(unsubscribe).toBeInstanceOf(Function);
		});

		it("should allow multiple subscribers", () => {
			const subscriber1 = vi.fn();
			const subscriber2 = vi.fn();

			notifier.subscribe(subscriber1);
			notifier.subscribe(subscriber2);

			// Subscribers will be called when events are processed
			// This will be tested in the notify tests
		});

		it("should remove subscriber when unsubscribe is called", async () => {
			const subscriber1 = vi.fn();
			const subscriber2 = vi.fn();

			const unsubscribe1 = notifier.subscribe(subscriber1);
			notifier.subscribe(subscriber2);

			unsubscribe1();

			const event: ApiUpdateEvent = {
				type: "story.updated",
				entityId: "story-1",
				entityType: "story",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);
			await vi.advanceTimersByTimeAsync(1000);

			// subscriber1 should not be called, subscriber2 should be called
			expect(subscriber1).not.toHaveBeenCalled();
			expect(subscriber2).toHaveBeenCalledWith(event);
		});
	});

	describe("notify", () => {
		it("should debounce events", async () => {
			const subscriber = vi.fn();
			notifier.subscribe(subscriber);

			const event1: ApiUpdateEvent = {
				type: "story.updated",
				entityId: "story-1",
				entityType: "story",
				timestamp: new Date().toISOString(),
			};

			const event2: ApiUpdateEvent = {
				type: "chapter.updated",
				entityId: "chapter-1",
				entityType: "chapter",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event1);
			await notifier.notify(event2);

			// Advance timers to trigger processing
			await vi.advanceTimersByTimeAsync(1000);

			// Both events should be processed (not deduplicated as they're different entities)
			expect(subscriber).toHaveBeenCalledTimes(2);
			expect(subscriber).toHaveBeenCalledWith(event1);
			expect(subscriber).toHaveBeenCalledWith(event2);
		});

		it("should process events after debounce delay", async () => {
			const subscriber = vi.fn();
			notifier.subscribe(subscriber);

			const event: ApiUpdateEvent = {
				type: "story.updated",
				entityId: "story-1",
				entityType: "story",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);

			// Subscriber should not be called immediately
			expect(subscriber).not.toHaveBeenCalled();

			// Advance timers to trigger processing
			await vi.advanceTimersByTimeAsync(1000);

			// Subscriber should be called with the event
			expect(subscriber).toHaveBeenCalledWith(event);
		});

		it("should deduplicate events (keep latest)", async () => {
			const subscriber = vi.fn();
			notifier.subscribe(subscriber);

			const event1: ApiUpdateEvent = {
				type: "story.updated",
				entityId: "story-1",
				entityType: "story",
				timestamp: new Date("2025-01-01T00:00:00Z").toISOString(),
			};

			const event2: ApiUpdateEvent = {
				type: "story.updated",
				entityId: "story-1",
				entityType: "story",
				timestamp: new Date("2025-01-01T01:00:00Z").toISOString(),
			};

			await notifier.notify(event1);
			await notifier.notify(event2);

			// Advance timers to trigger processing
			await vi.advanceTimersByTimeAsync(1000);

			// Only the latest event should be passed to subscriber
			expect(subscriber).toHaveBeenCalledTimes(1);
			expect(subscriber).toHaveBeenCalledWith(event2);
		});

		it("should limit queue size to MAX_QUEUE_SIZE", async () => {
			const subscriber = vi.fn();
			notifier.subscribe(subscriber);

			// Add more than MAX_QUEUE_SIZE events
			const events: ApiUpdateEvent[] = [];
			for (let i = 0; i < 150; i++) {
				events.push({
					type: "story.updated",
					entityId: `story-${i}`,
					entityType: "story",
					timestamp: new Date().toISOString(),
				});
			}

			for (const event of events) {
				await notifier.notify(event);
			}

			// Advance timers to trigger processing
			await vi.advanceTimersByTimeAsync(1000);

			// Should only process MAX_QUEUE_SIZE (100) events
			// The oldest events should be dropped
			expect(subscriber).toHaveBeenCalledTimes(100);
		});

		it("should handle subscriber errors gracefully", async () => {
			const subscriber1 = vi.fn(() => {
				throw new Error("Subscriber error");
			});
			const subscriber2 = vi.fn();

			notifier.subscribe(subscriber1);
			notifier.subscribe(subscriber2);

			const event: ApiUpdateEvent = {
				type: "story.updated",
				entityId: "story-1",
				entityType: "story",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);

			// Advance timers - should not throw even if one subscriber fails
			await expect(vi.advanceTimersByTimeAsync(1000)).resolves.not.toThrow();

			// Both subscribers should be called
			expect(subscriber1).toHaveBeenCalled();
			expect(subscriber2).toHaveBeenCalled();
		});
	});

	describe("clear", () => {
		it("should clear event queue and cancel pending timer", async () => {
			const subscriber = vi.fn();
			notifier.subscribe(subscriber);

			const event: ApiUpdateEvent = {
				type: "story.updated",
				entityId: "story-1",
				entityType: "story",
				timestamp: new Date().toISOString(),
			};

			await notifier.notify(event);

			// Clear should cancel the timer
			notifier.clear();

			// Advance timers - events should not be processed after clear
			await vi.advanceTimersByTimeAsync(1000);

			// Subscriber should not be called after clear
			expect(subscriber).not.toHaveBeenCalled();
		});
	});

	describe("dispose", () => {
		it("should clear queue and remove all subscribers", async () => {
			const subscriber = vi.fn();
			notifier.subscribe(subscriber);

			const event: ApiUpdateEvent = {
				type: "story.updated",
				entityId: "story-1",
				entityType: "story",
				timestamp: new Date().toISOString(),
			};

			notifier.dispose();

			await notifier.notify(event);

			// Advance timers - events should not be processed after dispose
			await vi.advanceTimersByTimeAsync(1000);

			// Subscriber should not be called after dispose
			expect(subscriber).not.toHaveBeenCalled();
		});
	});
});

