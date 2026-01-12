import type { ChapterWithContent, SceneWithBeats, StoryWithHierarchy } from "../../types";
import type { GeneratorMetadata } from "../types/generators";
import { getIdFieldName } from "../utils/frontmatterHelpers";

interface OutlineSection {
	type: "chapter" | "scene" | "beat";
	link: string;
	label: string;
	status: "+" | "-";
	depth: number;
}

export interface OutlineGeneratorOptions extends GeneratorMetadata {
	includePlaceholders?: boolean;
}

export class OutlineGenerator {
	constructor(private readonly now: () => string = () => new Date().toISOString()) {}

	generateStoryOutline(
		story: StoryWithHierarchy,
		options: OutlineGeneratorOptions = {}
	): string {
		const lines: string[] = [];
		const enableHelpBox = options.showHelpBox !== false;
		
		// Use configured ID field name, default to "id"
		const idField = getIdFieldName(options.idField);

		lines.push(
			"---",
			`${idField}: ${story.story.id}`,
			"type: story-outline",
			`synced_at: ${options.syncedAt ?? this.now()}`,
			"---",
			"",
			`# ${story.story.title}`,
			"",
			"## Hierarchy",
			""
		);

		if (enableHelpBox) {
			lines.push(
				"> [!tip] Como editar esta lista",
				"> - **Reordenar**: Arraste itens para mudar a ordem",
				"> - **Criar novo**: Edite a linha `_New..._` no final de cada seção",
				"> - **Indentação**: Tab define hierarquia (chapter → scene → beat)",
				"> - **Marcadores**: `+` tem conteúdo, `-` está vazio",
				""
			);
		}

		const sections = this.buildSections(story.chapters, options.includePlaceholders !== false);
		sections.forEach((section) => {
			const indent = "\t".repeat(section.depth);
			lines.push(`${indent}- ${section.link} ${section.status}`.trimEnd());
		});

		if (options.includePlaceholders !== false) {
			lines.push("- _New chapter: title_");
		}

		return lines.join("\n").trimEnd() + "\n";
	}

	private buildSections(
		chapters: ChapterWithContent[],
		includePlaceholders: boolean
	): OutlineSection[] {
		const sections: OutlineSection[] = [];

		chapters.forEach((chapter, chapterIdx) => {
			const chapterLink = this.buildLink(
				"ch",
				chapterIdx + 1,
				chapter.chapter.title,
				chapter.chapter.id,
				`Chapter ${chapterIdx + 1}: ${chapter.chapter.title}`
			);
			sections.push({
				type: "chapter",
				link: chapterLink,
				label: chapter.chapter.title,
				status: this.hasScenesWithContent(chapter.scenes) ? "+" : "-",
				depth: 0,
			});

			chapter.scenes.forEach((sceneWrapper, sceneIdx) => {
				const sceneLabel = this.composeSceneLabel(sceneWrapper);
				const sceneLink = this.buildLink(
					"sc",
					sceneIdx + 1,
					sceneWrapper.scene.goal || `Scene ${sceneIdx + 1}`,
					sceneWrapper.scene.id,
					`Scene ${sceneIdx + 1}: ${sceneLabel}`
				);
				sections.push({
					type: "scene",
					link: sceneLink,
					label: sceneLabel,
					status: sceneWrapper.beats.length ? "+" : "-",
					depth: 1,
				});

				sceneWrapper.beats.forEach((beat, beatIdx) => {
					sections.push({
						type: "beat",
						link: this.buildLink(
							"bt",
							beatIdx + 1,
							beat.intent || `Beat ${beatIdx + 1}`,
							beat.id,
							`Beat ${beatIdx + 1}: ${beat.intent || ""}`
						),
						label: beat.intent,
						status: beat.outcome ? "+" : "-",
						depth: 2,
					});
				});

				if (includePlaceholders) {
					sections.push({
						type: "beat",
						link: "_New beat: intent here..._",
						label: "",
						status: "-",
						depth: 2,
					});
				}
			});

			if (includePlaceholders) {
				sections.push({
					type: "scene",
					link: "_New scene: goal - time_",
					label: "",
					status: "-",
					depth: 1,
				});
			}
		});

		return sections;
	}

	private buildLink(
		prefix: string,
		order: number,
		title: string,
		id: string,
		display?: string
	): string {
		const slug = this.slugify(title);
		const label = display ?? title;
		return `[[${prefix}-${order.toString().padStart(2, "0")}-${slug}|${label}]]`;
	}

	private composeSceneLabel(scene: SceneWithBeats): string {
		const goal = scene.scene.goal || "Scene";
		const time = scene.scene.time_ref ? ` - ${scene.scene.time_ref}` : "";
		return `${goal}${time}`;
	}

	private hasScenesWithContent(scenes: SceneWithBeats[]): boolean {
		return scenes.some((scene) => scene.beats.length > 0);
	}

	generateChapterOutline(
		chapter: ChapterWithContent,
		options: OutlineGeneratorOptions = {}
	): string {
		const lines: string[] = [];
		const enableHelpBox = options.showHelpBox !== false;
		
		// Use configured ID field name, default to "id"
		const idField = getIdFieldName(options.idField);

		lines.push(
			"---",
			`${idField}: ${chapter.chapter.id}`,
			"type: chapter-outline",
			`synced_at: ${options.syncedAt ?? this.now()}`,
			"---",
			"",
			`# ${chapter.chapter.title}`,
			"",
			"## Hierarchy",
			""
		);

		if (enableHelpBox) {
			lines.push(
				"> [!tip] Como editar esta lista",
				"> - **Reordenar**: Arraste itens para mudar a ordem",
				"> - **Criar novo**: Edite a linha `_New..._` no final de cada seção",
				"> - **Indentação**: Tab define hierarquia (scene → beat)",
				"> - **Marcadores**: `+` tem conteúdo, `-` está vazio",
				""
			);
		}

		const sections = this.buildChapterSections(chapter.scenes, options.includePlaceholders !== false);
		sections.forEach((section) => {
			const indent = "\t".repeat(section.depth);
			lines.push(`${indent}- ${section.link} ${section.status}`.trimEnd());
		});

		if (options.includePlaceholders !== false) {
			lines.push("- _New scene: goal - time_");
		}

		return lines.join("\n").trimEnd() + "\n";
	}

	private buildChapterSections(
		scenes: SceneWithBeats[],
		includePlaceholders: boolean
	): OutlineSection[] {
		const sections: OutlineSection[] = [];

		scenes.forEach((sceneWrapper, sceneIdx) => {
			const sceneLabel = this.composeSceneLabel(sceneWrapper);
			const sceneLink = this.buildLink(
				"sc",
				sceneWrapper.scene.order_num ?? sceneIdx + 1,
				sceneWrapper.scene.goal || `Scene ${sceneIdx + 1}`,
				sceneWrapper.scene.id,
				`Scene ${sceneIdx + 1}: ${sceneLabel}`
			);
			sections.push({
				type: "scene",
				link: sceneLink,
				label: sceneLabel,
				status: sceneWrapper.beats.length ? "+" : "-",
				depth: 0,
			});

			sceneWrapper.beats.forEach((beat, beatIdx) => {
				sections.push({
					type: "beat",
					link: this.buildLink(
						"bt",
						beat.order_num ?? beatIdx + 1,
						beat.intent || `Beat ${beatIdx + 1}`,
						beat.id,
						`Beat ${beatIdx + 1}: ${beat.intent || ""}`
					),
					label: beat.intent,
					status: beat.outcome ? "+" : "-",
					depth: 1,
				});
			});

			if (includePlaceholders) {
				sections.push({
					type: "beat",
					link: "_New beat: intent here..._",
					label: "",
					status: "-",
					depth: 1,
				});
			}
		});

		return sections;
	}

	generateSceneOutline(
		scene: SceneWithBeats,
		options: OutlineGeneratorOptions = {}
	): string {
		const lines: string[] = [];
		const enableHelpBox = options.showHelpBox !== false;
		
		// Use configured ID field name, default to "id"
		const idField = getIdFieldName(options.idField);

		lines.push(
			"---",
			`${idField}: ${scene.scene.id}`,
			"type: scene-outline",
			`synced_at: ${options.syncedAt ?? this.now()}`,
			"---",
			"",
			`# ${scene.scene.goal || "Untitled Scene"}`,
			"",
			"## Hierarchy",
			""
		);

		if (enableHelpBox) {
			lines.push(
				"> [!tip] Como editar esta lista",
				"> - **Reordenar**: Arraste itens para mudar a ordem",
				"> - **Criar novo**: Edite a linha `_New..._` no final",
				"> - **Marcadores**: `+` tem conteúdo, `-` está vazio",
				""
			);
		}

		scene.beats.forEach((beat, beatIdx) => {
			const beatLink = this.buildLink(
				"bt",
				beat.order_num ?? beatIdx + 1,
				beat.intent || `Beat ${beatIdx + 1}`,
				beat.id,
				`Beat ${beat.order_num ?? beatIdx + 1}: ${beat.intent || ""}`
			);
			lines.push(`- ${beatLink} ${beat.outcome ? "+" : "-"}`);
		});

		if (options.includePlaceholders !== false) {
			lines.push("- _New beat: intent here..._");
		}

		return lines.join("\n").trimEnd() + "\n";
	}

	private slugify(value: string): string {
		return value
			.toLowerCase()
			.normalize("NFKD")
			.replace(/[^a-z0-9\s-]/g, "")
			.trim()
			.replace(/\s+/g, "-")
			.slice(0, 40);
	}
}

