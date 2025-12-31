import { Scene, Beat } from "../types";

export interface ParsedScene {
	scene: Partial<Scene>;
	beats: Partial<Beat>[];
}

export class MarkdownParser {
	// Parse scenes and beats from chapter markdown content
	static parseChapterMarkdown(content: string): ParsedScene[] {
		const scenes: ParsedScene[] = [];

		// Split content by scene separators (---)
		const sections = content.split(/^---$/m);

		for (const section of sections) {
			const trimmed = section.trim();
			if (!trimmed) continue;

			// Try to extract scene information
			const sceneMatch = trimmed.match(
				/###\s*Scene\s*(\d+)(?::\s*(.+?))?\n([\s\S]*?)(?=####|###|---|$)/i
			);

			if (!sceneMatch) continue;

			const orderNum = parseInt(sceneMatch[1], 10);
			const sceneTitle = sceneMatch[2]?.trim() || "";
			const sceneContent = sceneMatch[3] || "";

			// Extract scene metadata
			const sceneIdMatch = sceneContent.match(/\*\*ID\*\*:\s*([a-f0-9-]+)/i);
			const timeRefMatch = sceneContent.match(/\*\*Time\*\*:\s*(.+?)(?:\n|$)/i);
			const goalMatch = sceneContent.match(/\*\*Goal\*\*:\s*(.+?)(?:\n|$)/i);

			const scene: Partial<Scene> = {
				order_num: orderNum,
				goal: goalMatch ? goalMatch[1].trim() : sceneTitle || "",
				time_ref: timeRefMatch ? timeRefMatch[1].trim() : "",
			};

			if (sceneIdMatch) {
				scene.id = sceneIdMatch[1].trim();
			}

			// Extract beats
			const beats: Partial<Beat>[] = [];
			const beatsSection = sceneContent.match(/####\s*Beats\s*\n([\s\S]*?)(?=---|$)/i);

			if (beatsSection) {
				const beatsText = beatsSection[1];
				const beatMatches = beatsText.matchAll(
					/\*\*Beat\s*(\d+)\*\*\s*\(([^)]+)\)\s*\n([\s\S]*?)(?=\*\*Beat|$)/gi
				);

				for (const beatMatch of beatMatches) {
					const beatOrderNum = parseInt(beatMatch[1], 10);
					const beatType = beatMatch[2].trim();
					const beatContent = beatMatch[3] || "";

					const intentMatch = beatContent.match(
					/-\s*\*\*Intent\*\*:\s*(.+?)(?:\n|$)/i
					);
					const outcomeMatch = beatContent.match(
						/-\s*\*\*Outcome\*\*:\s*(.+?)(?:\n|$)/i
					);

					const beat: Partial<Beat> = {
						order_num: beatOrderNum,
						type: beatType,
						intent: intentMatch ? intentMatch[1].trim() : "",
						outcome: outcomeMatch ? outcomeMatch[1].trim() : "",
					};

					beats.push(beat);
				}
			}

			scenes.push({ scene, beats });
		}

		return scenes;
	}

	// Build chapter markdown from chapter data
	static buildChapterMarkdown(
		chapterTitle: string,
		chapterNumber: number,
		scenes: ParsedScene[]
	): string {
		let content = `# Chapter ${chapterNumber}: ${chapterTitle}\n\n`;

		if (scenes.length === 0) {
			content += "(No scenes yet)\n";
		} else {
			content += "## Scenes\n\n";

			for (const sceneData of scenes) {
				const scene = sceneData.scene;
				content += `### Scene ${scene.order_num}`;
				if (scene.goal) {
					content += `: ${scene.goal}`;
				}
				content += `\n`;

				if (scene.id) {
					content += `**ID**: ${scene.id}\n`;
				}
				content += `**Order**: ${scene.order_num}\n`;
				if (scene.time_ref) {
					content += `**Time**: ${scene.time_ref}\n`;
				}
				if (scene.goal) {
					content += `**Goal**: ${scene.goal}\n`;
				}
				content += `\n`;

				if (sceneData.beats.length > 0) {
					content += `#### Beats\n\n`;

					for (const beat of sceneData.beats) {
						content += `**Beat ${beat.order_num}** (${beat.type})\n`;
						if (beat.intent) {
							content += `- **Intent**: ${beat.intent}\n`;
						}
						if (beat.outcome) {
							content += `- **Outcome**: ${beat.outcome}\n`;
						}
						content += `\n`;
					}
				}

				content += `---\n\n`;
			}
		}

		return content;
	}
}

