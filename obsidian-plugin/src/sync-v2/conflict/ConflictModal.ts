import type { App } from "obsidian";
import type { Conflict } from "./types";

export type ConflictModalChoice = "local" | "remote" | "merge";

export class ConflictModal {
	private static nextChoice: ConflictModalChoice | null = null;

	constructor(private readonly app: App, private readonly conflict: Conflict) {
		void this.app;
		void this.conflict;
	}

	static setNextChoice(choice: ConflictModalChoice | null): void {
		ConflictModal.nextChoice = choice;
	}

	async open(): Promise<ConflictModalChoice> {
		if (ConflictModal.nextChoice) {
			const choice = ConflictModal.nextChoice;
			ConflictModal.nextChoice = null;
			return choice;
		}

		// TODO: Implement Obsidian modal UI with diff + buttons.
		return "local";
	}
}
