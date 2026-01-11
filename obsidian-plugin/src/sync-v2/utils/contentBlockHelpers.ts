import type { Chapter, Scene, Beat, ContentAnchor, Story } from "../../types";
import type { StoryEngineClient } from "../../api/client";

/**
 * Resolved hierarchy for a ContentBlock
 * Represents the parent chain (Beat > Scene > Chapter)
 */
export interface ContentBlockHierarchy {
	contentBlockId: string;
	beat?: { id: string; intent?: string; order_num: number };
	scene?: { id: string; goal?: string; order_num: number };
	chapter?: { id: string; title: string; number: number };
	story: { id: string; title: string };
	worldId?: string;
}

/**
 * Resolve ContentBlock hierarchy by fetching ContentAnchors and parent entities
 * Returns the most specific hierarchy (Beat > Scene > Chapter)
 */
export async function resolveContentBlockHierarchy(
	contentBlockId: string,
	apiClient: StoryEngineClient
): Promise<ContentBlockHierarchy | null> {
	try {
		// Get ContentAnchors to find parent entities
		const anchors = await apiClient.getContentAnchors(contentBlockId);
		if (anchors.length === 0) {
			// ContentBlock might be directly attached to chapter
			// Try to get from ContentBlock itself
			const contentBlock = await apiClient.getContentBlock(contentBlockId);
			if (contentBlock.chapter_id) {
				const chapter = await apiClient.getChapter(contentBlock.chapter_id);
				const story = await apiClient.getStory(chapter.story_id);
				return {
					contentBlockId,
					chapter: {
						id: chapter.id,
						title: chapter.title,
						number: chapter.number ?? 0,
					},
					story: {
						id: story.id,
						title: story.title,
					},
					worldId: story.world_id ?? undefined,
				};
			}
			return null;
		}

		// Find anchors by entity type
		const beatAnchor = anchors.find((a) => a.entity_type === "beat");
		const sceneAnchor = anchors.find((a) => a.entity_type === "scene");
		const chapterAnchor = anchors.find((a) => a.entity_type === "chapter");

		// Determine hierarchy level - Beat is most specific, then Scene, then Chapter
		let beat: Beat | null = null;
		let scene: Scene | null = null;
		let chapter: Chapter | null = null;
		let story: Story | null = null;

		if (beatAnchor) {
			// Beat level - fetch beat, then scene, then chapter, then story
			beat = await apiClient.getBeat(beatAnchor.entity_id);
			scene = await apiClient.getScene(beat.scene_id);
			if (scene.chapter_id) {
				chapter = await apiClient.getChapter(scene.chapter_id);
			}
			story = await apiClient.getStory(scene.story_id);
		} else if (sceneAnchor) {
			// Scene level - fetch scene, then chapter, then story
			scene = await apiClient.getScene(sceneAnchor.entity_id);
			if (scene.chapter_id) {
				chapter = await apiClient.getChapter(scene.chapter_id);
			}
			story = await apiClient.getStory(scene.story_id);
		} else if (chapterAnchor) {
			// Chapter level - fetch chapter, then story
			chapter = await apiClient.getChapter(chapterAnchor.entity_id);
			story = await apiClient.getStory(chapter.story_id);
		} else {
			// No recognized anchor type
			return null;
		}

		if (!story) {
			return null;
		}

		return {
			contentBlockId,
			beat: beat
				? {
						id: beat.id,
						intent: beat.intent ?? undefined,
						order_num: beat.order_num ?? 0,
					}
				: undefined,
			scene: scene
				? {
						id: scene.id,
						goal: scene.goal ?? undefined,
						order_num: scene.order_num ?? 0,
					}
				: undefined,
			chapter: chapter
				? {
						id: chapter.id,
						title: chapter.title,
						number: chapter.number ?? 0,
					}
				: undefined,
			story: {
				id: story.id,
				title: story.title,
			},
			worldId: story.world_id ?? undefined,
		};
	} catch (error) {
		console.warn("[Sync V2] Failed to resolve ContentBlock hierarchy", { contentBlockId, error });
		return null;
	}
}

/**
 * Build a human-readable context string from hierarchy
 * Format: "Chapter 1: Introduction > Scene 2: The Meeting > Beat 3: Confrontation"
 */
export function buildHierarchyContext(hierarchy: ContentBlockHierarchy): string {
	const parts: string[] = [];

		if (hierarchy.chapter) {
			const chapterNum =
				hierarchy.chapter.number > 0 ? String(hierarchy.chapter.number) : "?";
			parts.push(`Chapter ${chapterNum}: ${hierarchy.chapter.title}`);
		}

	if (hierarchy.scene) {
		const sceneNum = hierarchy.scene.order_num > 0 ? String(hierarchy.scene.order_num) : "?";
		const sceneTitle = hierarchy.scene.goal || "Untitled Scene";
		parts.push(`Scene ${sceneNum}: ${sceneTitle}`);
	}

	if (hierarchy.beat) {
		const beatNum = hierarchy.beat.order_num > 0 ? String(hierarchy.beat.order_num) : "?";
		const beatTitle = hierarchy.beat.intent || "Untitled Beat";
		parts.push(`Beat ${beatNum}: ${beatTitle}`);
	}

	return parts.join(" > ") || hierarchy.story.title;
}

/**
 * Create citation relations for detected entity mentions
 * Validates target entities exist and creates relations at the appropriate level
 */
export async function createCitationRelations(
	mentions: Array<{ entityId: string; entityType: string; worldId?: string }>,
	hierarchy: ContentBlockHierarchy,
	sourceContentBlockId: string,
	apiClient: StoryEngineClient,
	contextString: string
): Promise<{ created: number; errors: string[] }> {
	const result = { created: 0, errors: [] as string[] };

	// Determine source type and ID based on hierarchy level
	// Priority: Beat > Scene > Chapter > ContentBlock
	let sourceType: "beat" | "scene" | "chapter" | "content_block";
	let sourceId: string;

	if (hierarchy.beat) {
		sourceType = "beat";
		sourceId = hierarchy.beat.id;
	} else if (hierarchy.scene) {
		sourceType = "scene";
		sourceId = hierarchy.scene.id;
	} else if (hierarchy.chapter) {
		sourceType = "chapter";
		sourceId = hierarchy.chapter.id;
	} else {
		sourceType = "content_block";
		sourceId = sourceContentBlockId;
	}

	// Get existing relations for this source to avoid duplicates
	const existingRelationsResponse = await apiClient.listRelationsBySource({
		sourceType,
		sourceId,
	});
	const existingRelations = existingRelationsResponse.data;

	// Group existing relations by target_id for quick lookup
	const existingRelationsByTarget = new Map<string, typeof existingRelations>();
	for (const rel of existingRelations) {
		if (!existingRelationsByTarget.has(rel.target_id)) {
			existingRelationsByTarget.set(rel.target_id, []);
		}
		existingRelationsByTarget.get(rel.target_id)!.push(rel);
	}

	// Create citation relations for each mention
	for (const mention of mentions) {
		try {
			// Check if citation relation already exists
			const existingForTarget = existingRelationsByTarget.get(mention.entityId) || [];
			const existingCitation = existingForTarget.find(
				(rel) =>
					rel.target_type === mention.entityType &&
					rel.target_id === mention.entityId &&
					rel.relation_type === "citation" &&
					rel.source_type === sourceType &&
					rel.source_id === sourceId
			);

			if (existingCitation) {
				// Relation already exists, skip
				continue;
			}

			// Validate that target entity exists (by trying to fetch it)
			// We'll validate based on entity type
			let targetExists = false;
			try {
				switch (mention.entityType) {
					case "character":
						await apiClient.getCharacter(mention.entityId);
						targetExists = true;
						break;
					case "location":
						await apiClient.getLocation(mention.entityId);
						targetExists = true;
						break;
					case "faction":
						await apiClient.getFaction(mention.entityId);
						targetExists = true;
						break;
					case "artifact":
						await apiClient.getArtifact(mention.entityId);
						targetExists = true;
						break;
					case "event":
						await apiClient.getEvent(mention.entityId);
						targetExists = true;
						break;
					case "lore":
						await apiClient.getLore(mention.entityId);
						targetExists = true;
						break;
					default:
						result.errors.push(
							`Unsupported entity type for citation: ${mention.entityType}`
						);
						continue;
				}
			} catch (error) {
				result.errors.push(
					`Target entity ${mention.entityType}:${mention.entityId} does not exist: ${error}`
				);
				continue;
			}

			if (!targetExists) {
				result.errors.push(`Failed to validate target entity ${mention.entityType}:${mention.entityId}`);
				continue;
			}

			// Create citation relation
			// Include context string and source content block ID in the context field
			const contextWithMetadata = `${contextString} (Source: ContentBlock ${sourceContentBlockId})`;
			await apiClient.createRelation({
				sourceType,
				sourceId,
				targetType: mention.entityType,
				targetId: mention.entityId,
				relationType: "citation",
				context: contextWithMetadata,
			});

			result.created++;
		} catch (error) {
			result.errors.push(
				`Failed to create citation for ${mention.entityType}:${mention.entityId}: ${error instanceof Error ? error.message : String(error)}`
			);
		}
	}

	return result;
}

