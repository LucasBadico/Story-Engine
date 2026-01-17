import type { Chapter, ContentBlock, Story } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { PathResolver } from "../../fileRenamer/PathResolver";
import { EntityFileParser } from "../../parsers/entityFileParser";
import { getFrontmatterId } from "../../utils/frontmatterHelpers";
import { detectEntityMentions, resolveEntityMention } from "../../utils/detectEntityMentions";
import {
	resolveContentBlockHierarchy,
	buildHierarchyContext,
	createCitationRelations,
} from "../../utils/contentBlockHelpers";

export class ContentBlockHandler {
	readonly entityType = "content_block";
	private readonly parser = new EntityFileParser();

	async pull(id: string, context: SyncContext): Promise<ContentBlock> {
		const contentBlock = await context.apiClient.getContentBlock(id);
		let chapter: Chapter | null = null;
		if (contentBlock.chapter_id) {
			chapter = await context.apiClient.getChapter(contentBlock.chapter_id);
		}
		const storyId = chapter?.story_id;
		if (!storyId) {
			return contentBlock;
		}
		const story = await context.apiClient.getStory(storyId);
		const folderPath = context.fileManager.getStoryFolderPath(story.title);
		const pathResolver = new PathResolver(folderPath);
		const filePath = pathResolver.getContentBlockPath(contentBlock);
		const directory = filePath.split("/").slice(0, -1).join("/");
		await context.fileManager.ensureFolderExists(directory);
		await context.fileManager.writeContentBlockFile(contentBlock, filePath, story.title);
		return contentBlock;
	}

	async push(entity: ContentBlock, context: SyncContext): Promise<void> {
		try {
			// Resolve story path to find the content block file
			const story = await this.resolveStoryForContentBlock(entity, context);
			if (!story) {
				throw new Error(`Could not resolve story for content block ${entity.id}`);
			}

			const folderPath = context.fileManager.getStoryFolderPath(story.title);
			const pathResolver = new PathResolver(folderPath);
			const filePath = pathResolver.getContentBlockPath(entity);

			// Read current file content
			let fileContent: string;
			try {
				fileContent = await context.fileManager.readFile(filePath);
			} catch (error) {
				context.emitWarning?.({
					code: "content_block_file_not_found",
					message: `Content block file not found at ${filePath}. Skipping push.`,
					filePath,
					details: error,
				});
				return;
			}

			// Parse frontmatter and extract body content
			const parsed = this.parser.parse(fileContent);
			const idField = context.settings.frontmatterIdField;
			const contentBlockId = getFrontmatterId(parsed.frontmatter, idField);

			if (!contentBlockId || contentBlockId !== entity.id) {
				context.emitWarning?.({
					code: "content_block_id_mismatch",
					message: `Content block ID mismatch: expected ${entity.id}, found ${contentBlockId ?? "none"}`,
					filePath,
				});
				return;
			}

			// Extract content from body (content after frontmatter)
			const content = parsed.body.trim();

			// Parse frontmatter fields for updates
			const frontmatter = parsed.frontmatter;
			const updates: Partial<ContentBlock> = {
				content,
			};

			// Update order_num if changed
			if (frontmatter.order_num !== undefined) {
				const orderNum = parseInt(frontmatter.order_num);
				if (!isNaN(orderNum) && orderNum !== entity.order_num) {
					updates.order_num = orderNum;
				}
			}

			// Update chapter_id if changed
			if (frontmatter.chapter_id !== undefined) {
				const chapterId = frontmatter.chapter_id === "null" || frontmatter.chapter_id.trim() === "" 
					? null 
					: frontmatter.chapter_id.trim();
				if (chapterId !== entity.chapter_id) {
					updates.chapter_id = chapterId;
				}
			}

			// Only update if there are actual changes
			if (Object.keys(updates).length > 1 || (updates.content && updates.content !== entity.content)) {
				await context.apiClient.updateContentBlock(entity.id, updates);
			}

			// Detect entity mentions and create citation relations
			if (content && content.trim().length > 0) {
				await this.processEntityMentions(entity.id, content, context);
			}
		} catch (error) {
			context.emitWarning?.({
				code: "content_block_push_error",
				message: `Failed to push content block ${entity.id}: ${error instanceof Error ? error.message : String(error)}`,
				details: error,
			});
			throw error;
		}
	}

	private async processEntityMentions(
		contentBlockId: string,
		content: string,
		context: SyncContext
	): Promise<void> {
		try {
			// Detect entity mentions in content
			const mentions = detectEntityMentions(content);
			if (mentions.length === 0) {
				return; // No mentions found
			}

			// Resolve mentions to entity IDs
			const resolvedMentions: Array<{ entityId: string; entityType: string; worldId?: string }> = [];
			for (const mention of mentions) {
				const resolved = await resolveEntityMention(mention, context);
				if (resolved) {
					resolvedMentions.push({
						entityId: resolved.entityId,
						entityType: resolved.entityType,
						worldId: resolved.worldId,
					});
				}
			}

			if (resolvedMentions.length === 0) {
				return; // No valid mentions found
			}

			// Resolve ContentBlock hierarchy
			const hierarchy = await resolveContentBlockHierarchy(contentBlockId, context.apiClient);
			if (!hierarchy) {
				context.emitWarning?.({
					code: "hierarchy_resolution_failed",
					message: `Could not resolve hierarchy for content block ${contentBlockId}. Citations will not be created.`,
				});
				return;
			}

			// Build context string
			const contextString = buildHierarchyContext(hierarchy);

			// Create citation relations
			const result = await createCitationRelations(
				resolvedMentions,
				hierarchy,
				contentBlockId,
				context.apiClient,
				contextString
			);

			if (result.created > 0) {
				console.log(
					`[Sync V2] Created ${result.created} citation relation(s) for content block ${contentBlockId}`
				);
			}

			if (result.errors.length > 0) {
				for (const error of result.errors) {
					context.emitWarning?.({
						code: "citation_creation_error",
						message: error,
					});
				}
			}
		} catch (error) {
			context.emitWarning?.({
				code: "entity_mention_processing_error",
				message: `Failed to process entity mentions for content block ${contentBlockId}: ${error instanceof Error ? error.message : String(error)}`,
				details: error,
			});
			// Don't throw - citations are non-critical, push can succeed even if citations fail
		}
	}

	private async resolveStoryForContentBlock(
		contentBlock: ContentBlock,
		context: SyncContext
	): Promise<Story | null> {
		// Try to resolve via chapter_id first (most common case)
		if (contentBlock.chapter_id) {
			try {
				const chapter = await context.apiClient.getChapter(contentBlock.chapter_id);
				if (chapter.story_id) {
					return await context.apiClient.getStory(chapter.story_id);
				}
			} catch (error) {
				console.warn(`Failed to resolve story via chapter_id for content block ${contentBlock.id}`, error);
			}
		}

		// Fallback: try to resolve via ContentAnchors
		try {
			const anchors = await context.apiClient.getContentAnchors(contentBlock.id);
			// Try scene anchor first (more specific)
			const sceneAnchor = anchors.find((anchor) => anchor.entity_type === "scene");
			if (sceneAnchor) {
				const scene = await context.apiClient.getScene(sceneAnchor.entity_id);
				if (scene.story_id) {
					return await context.apiClient.getStory(scene.story_id);
				}
			}

			// Try beat anchor
			const beatAnchor = anchors.find((anchor) => anchor.entity_type === "beat");
			if (beatAnchor) {
				const beat = await context.apiClient.getBeat(beatAnchor.entity_id);
				const scene = await context.apiClient.getScene(beat.scene_id);
				if (scene.story_id) {
					return await context.apiClient.getStory(scene.story_id);
				}
			}

			// Try chapter anchor as last resort
			const chapterAnchor = anchors.find((anchor) => anchor.entity_type === "chapter");
			if (chapterAnchor) {
				const chapter = await context.apiClient.getChapter(chapterAnchor.entity_id);
				if (chapter.story_id) {
					return await context.apiClient.getStory(chapter.story_id);
				}
			}
		} catch (error) {
			console.warn(`Failed to resolve story via ContentAnchors for content block ${contentBlock.id}`, error);
		}

		return null;
	}

	async delete(id: string, context: SyncContext): Promise<void> {
		await context.apiClient.deleteContentBlock(id);
	}
}

