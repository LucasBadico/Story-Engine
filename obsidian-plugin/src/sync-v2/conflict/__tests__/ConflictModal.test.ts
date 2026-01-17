import { describe, expect, it } from "vitest";
import type { App } from "obsidian";
import type { Conflict } from "../types";
import { ConflictModal } from "../ConflictModal";

describe("ConflictModal", () => {
	it("returns configured next choice and resets it", async () => {
		const conflict: Conflict = {
			type: "simultaneous_edit",
			entityId: "cb-1",
			entityType: "content_block",
			filePath: "test.md",
			localData: { content: "Local" },
			remoteData: { content: "Remote" },
		};

		ConflictModal.setNextChoice("remote");

		const modal = new ConflictModal({} as App, conflict);
		const choice = await modal.open();
		const fallbackChoice = await modal.open();

		expect(choice).toBe("remote");
		expect(fallbackChoice).toBe("local");
	});
});
