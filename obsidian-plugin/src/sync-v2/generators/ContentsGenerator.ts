import type {
	ChapterWithContent,
	ContentBlock,
	SceneWithBeats,
	Story,
} from "../../types";
import { ContentsParser } from "../parsers/contentsParser";
import type { StoryContentsInput } from "../types/generators";
import { getIdFieldName } from "../utils/frontmatterHelpers";

export class ContentsGenerator {
	private readonly parser = new ContentsParser();

	constructor(private readonly now: () => string = () => new Date().toISOString()) {}

	generateStoryContents(input: StoryContentsInput): string {
		const { story, chapters } = input;
		const lines: string[] = [];

		// Use configured ID field name, default to "id"
		const idField = getIdFieldName(input.options?.idField);

		lines.push(
			"---",
			`${idField}: ${story.id}`,
			"type: story-contents",
			`synced_at: ${input.options?.syncedAt ?? this.now()}`,
			"---",
			"",
			`# ${story.title} - Contents`,
			""
		);

		chapters.forEach((chapter, chapterIdx) => {
			const chapterContent = this.buildChapterContent(chapter, chapterIdx, input);
			lines.push(chapterContent);
		});

		return lines.join("\n").trimEnd() + "\n";
	}

	private buildChapterContent(
		chapter: ChapterWithContent,
		index: number,
		input: StoryContentsInput
	): string {
		const chapterName = this.slugify(chapter.chapter.title);
		const order = index + 1;
		const contentLines: string[] = [];

		contentLines.push(`## Chapter ${order}: ${chapter.chapter.title}`, "");

		const chapterBlocks =
			input.chapterContentBlocks?.get(chapter.chapter.id) ?? [];
		chapterBlocks
			.sort((a, b) => (a.order_num ?? 0) - (b.order_num ?? 0))
			.forEach((block) => {
				contentLines.push(this.renderContentBlock(block));
			});

		chapter.scenes.forEach((scene, sceneIdx) => {
			contentLines.push(
				this.buildSceneSection(scene, sceneIdx, chapter.chapter.id, input)
			);
		});

		contentLines.push(this.parser.generatePlaceholder("scene"));

		return this.parser.generateFence(
			"chapter",
			order,
			chapterName,
			chapter.chapter.id,
			contentLines.join("\n").trimEnd()
		);
	}

	private buildSceneSection(
		scene: SceneWithBeats,
		index: number,
		chapterId: string,
		input: StoryContentsInput
	): string {
		const lines: string[] = [];
		const order = index + 1;
		const sceneName = this.slugify(scene.scene.goal || `scene-${order}`);
		lines.push(`### Scene ${order}: ${scene.scene.goal || "Untitled"}`, "");

		const sceneBlocks =
			input.sceneContentBlocks?.get(scene.scene.id) ?? [];
		if (sceneBlocks.length === 0) {
			lines.push(this.parser.generatePlaceholder("content"));
		} else {
			sceneBlocks
				.sort((a, b) => (a.order_num ?? 0) - (b.order_num ?? 0))
				.forEach((block) => {
					lines.push(this.renderContentBlock(block));
				});
		}

		scene.beats.forEach((beat, beatIdx) => {
			lines.push(this.buildBeatSection(beat, beatIdx, input));
		});

		lines.push(this.parser.generatePlaceholder("beat"));

		return this.parser.generateFence(
			"scene",
			order,
			sceneName,
			scene.scene.id,
			lines.join("\n").trimEnd()
		);
	}

	private buildBeatSection(
		beat: SceneWithBeats["beats"][number],
		index: number,
		input: StoryContentsInput
	): string {
		const lines: string[] = [];
		const order = index + 1;
		const beatName = this.slugify(beat.intent || `beat-${order}`);
		lines.push(`#### Beat ${order}: ${beat.intent || "Untitled"}`, "");

		const beatBlocks = input.beatContentBlocks?.get(beat.id) ?? [];
		if (beatBlocks.length === 0) {
			lines.push(this.parser.generatePlaceholder("content"));
		} else {
			beatBlocks
				.sort((a, b) => (a.order_num ?? 0) - (b.order_num ?? 0))
				.forEach((block) => {
					lines.push(this.renderContentBlock(block));
				});
		}

		return this.parser.generateFence(
			"beat",
			order,
			beatName,
			beat.id,
			lines.join("\n").trimEnd()
		);
	}

	generateSceneContents(
		scene: SceneWithBeats,
		sceneContentBlocks: Map<string, ContentBlock[]>,
		beatContentBlocks: Map<string, ContentBlock[]>,
		options: { syncedAt?: string; idField?: string } = {}
	): string {
		const lines: string[] = [];

		// Use configured ID field name, default to "id"
		const idField = getIdFieldName(options.idField);

		lines.push(
			"---",
			`${idField}: ${scene.scene.id}`,
			"type: scene-contents",
			`synced_at: ${options.syncedAt ?? this.now()}`,
			"---",
			"",
			`# ${scene.scene.goal || "Untitled Scene"} - Contents`,
			""
		);

		const sceneBlocks = sceneContentBlocks.get(scene.scene.id) ?? [];
		if (sceneBlocks.length === 0) {
			lines.push(this.parser.generatePlaceholder("content"));
		} else {
			sceneBlocks
				.sort((a, b) => (a.order_num ?? 0) - (b.order_num ?? 0))
				.forEach((block) => {
					lines.push(this.renderContentBlock(block));
				});
		}

		scene.beats.forEach((beat, beatIdx) => {
			lines.push(this.buildBeatSection(beat, beatIdx, {
				story: {
					id: scene.scene.story_id,
					tenant_id: "",
					title: "",
					status: "draft",
					version_number: 1,
					root_story_id: scene.scene.story_id,
					previous_story_id: null,
					created_by_user_id: "",
					world_id: null,
					created_at: "",
					updated_at: "",
				},
				chapters: [],
				chapterContentBlocks: new Map(),
				sceneContentBlocks,
				beatContentBlocks,
				options,
			}));
		});

		lines.push(this.parser.generatePlaceholder("beat"));

		return lines.join("\n").trimEnd() + "\n";
	}

	private renderContentBlock(block: ContentBlock): string {
		const name = this.slugify(block.metadata?.title || block.kind || `content-${block.id}`);
		const fence = this.parser.generateFence(
			"content",
			block.order_num ?? 0,
			name,
			block.id,
			block.content.trim() || "*No content yet*"
		);
		return fence;
	}

	generateChapterContents(
		chapter: ChapterWithContent,
		chapterContentBlocks: Map<string, ContentBlock[]>,
		sceneContentBlocks: Map<string, ContentBlock[]>,
		beatContentBlocks: Map<string, ContentBlock[]>,
		options: { syncedAt?: string; idField?: string } = {}
	): string {
		const lines: string[] = [];

		// Use configured ID field name, default to "id"
		const idField = getIdFieldName(options.idField);

		lines.push(
			"---",
			`${idField}: ${chapter.chapter.id}`,
			"type: chapter-contents",
			`synced_at: ${options.syncedAt ?? this.now()}`,
			"---",
			"",
			`# ${chapter.chapter.title} - Contents`,
			""
		);

		const chapterBlocks = chapterContentBlocks.get(chapter.chapter.id) ?? [];
		chapterBlocks
			.sort((a, b) => (a.order_num ?? 0) - (b.order_num ?? 0))
			.forEach((block) => {
				lines.push(this.renderContentBlock(block));
			});

		chapter.scenes.forEach((scene, sceneIdx) => {
			lines.push(
				this.buildSceneSection(scene, sceneIdx, chapter.chapter.id, {
					story: {
						id: chapter.chapter.story_id,
						tenant_id: "",
						title: "",
						status: "draft",
						version_number: 1,
						root_story_id: chapter.chapter.story_id,
						previous_story_id: null,
						created_by_user_id: "",
						world_id: null,
						created_at: "",
						updated_at: "",
					},
					chapters: [],
					chapterContentBlocks,
					sceneContentBlocks,
					beatContentBlocks,
					options,
				})
			);
		});

		lines.push(this.parser.generatePlaceholder("scene"));

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

