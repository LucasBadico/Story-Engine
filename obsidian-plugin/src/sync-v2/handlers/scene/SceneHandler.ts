import type { SceneWithBeats, ContentBlock } from "../../../types";
import type { SyncContext } from "../../types/sync";
import { PathResolver } from "../../fileRenamer/PathResolver";

export class SceneHandler {
	readonly entityType = "scene";

	async pull(id: string, context: SyncContext): Promise<SceneWithBeats> {
		const scene = await context.apiClient.getScene(id);
		const beats = await context.apiClient.getBeats(scene.id);
		const sceneWithBeats: SceneWithBeats = { scene, beats };
		const story = await context.apiClient.getStory(scene.story_id);
		const folderPath = context.fileManager.getStoryFolderPath(story.title);
		const scenesFolder = `${folderPath}/01-scenes`;
		await context.fileManager.ensureFolderExists(scenesFolder);
		const pathResolver = new PathResolver(folderPath);
		const filePath = pathResolver.getScenePath(scene);
		const contentBlocks: ContentBlock[] = await context.apiClient.getContentBlocksByScene(scene.id);

		await context.fileManager.writeSceneFile(
			sceneWithBeats,
			filePath,
			story.title,
			contentBlocks,
			[]
		);

		return sceneWithBeats;
	}

	async push(entity: SceneWithBeats, context: SyncContext): Promise<void> {
		const scene = entity.scene;
		const story = await context.apiClient.getStory(scene.story_id);
		const folderPath = context.fileManager.getStoryFolderPath(story.title);
		const pathResolver = new PathResolver(folderPath);
		const filePath = pathResolver.getScenePath(scene);

		// Read current file to detect changes
		let fileContent: string;
		try {
			fileContent = await context.fileManager.readFile(filePath);
		} catch (error: any) {
			if (error?.message?.includes("missing") || error?.code === "ENOENT") {
				context.emitWarning?.({
					code: "scene_file_not_found",
					message: `Scene file not found: ${filePath}`,
					filePath,
				});
				return;
			}
			throw error;
		}

		// Parse frontmatter to get current values
		const frontmatter = context.fileManager.parseFrontmatter(fileContent);
		const currentPovCharacterId =
			frontmatter.pov_character_id && frontmatter.pov_character_id !== "null" && frontmatter.pov_character_id.trim() !== ""
				? frontmatter.pov_character_id.trim()
				: null;
		const currentLocationId =
			frontmatter.location_id && frontmatter.location_id !== "null" && frontmatter.location_id.trim() !== ""
				? frontmatter.location_id.trim()
				: null;

		// Get current scene state from API
		const currentScene = await context.apiClient.getScene(scene.id);
		const oldPovCharacterId = currentScene.pov_character_id || null;
		const oldLocationId = currentScene.location_id || null;

		// Update scene if needed
		const needsUpdate =
			currentPovCharacterId !== oldPovCharacterId ||
			currentLocationId !== oldLocationId ||
			frontmatter.goal !== currentScene.goal ||
			frontmatter.time_ref !== currentScene.time_ref ||
			parseInt(frontmatter.order_num) !== currentScene.order_num ||
			frontmatter.chapter_id !== (currentScene.chapter_id || null);

		if (needsUpdate) {
			await context.apiClient.updateScene(scene.id, {
				pov_character_id: currentPovCharacterId || undefined,
				location_id: currentLocationId || undefined,
				goal: frontmatter.goal || currentScene.goal,
				time_ref: frontmatter.time_ref || currentScene.time_ref,
				order_num: parseInt(frontmatter.order_num) || currentScene.order_num,
				chapter_id: frontmatter.chapter_id || null,
			});
		}

		// Create relations automatically when POV/Location changes
		await this.ensureSceneRelations(
			scene.id,
			story.id,
			story.world_id || undefined,
			currentPovCharacterId,
			oldPovCharacterId,
			currentLocationId,
			oldLocationId,
			context
		);
	}

	private async ensureSceneRelations(
		sceneId: string,
		storyId: string,
		worldId: string | undefined,
		currentPovCharacterId: string | null,
		oldPovCharacterId: string | null,
		currentLocationId: string | null,
		oldLocationId: string | null,
		context: SyncContext
	): Promise<void> {
		// Get existing relations for this scene
		const existingRelationsResponse = await context.apiClient.listRelationsBySource({
			sourceType: "scene",
			sourceId: sceneId,
		});
		const existingRelations = existingRelationsResponse.data;

		// Handle POV character relation
		if (currentPovCharacterId !== oldPovCharacterId) {
			// Delete old POV relation if it exists
			if (oldPovCharacterId) {
				const oldPovRelation = existingRelations.find(
					(rel) =>
						rel.target_type === "character" &&
						rel.target_id === oldPovCharacterId &&
						rel.relation_type === "pov"
				);
				if (oldPovRelation) {
					try {
						await context.apiClient.deleteRelation(oldPovRelation.id);
					} catch (error) {
						context.emitWarning?.({
							code: "relation_delete_error",
							message: `Failed to delete old POV relation: ${error}`,
						});
					}
				}
			}

			// Create new POV relation if character is set
			if (currentPovCharacterId) {
				// Check if relation already exists
				const existingPovRelation = existingRelations.find(
					(rel) =>
						rel.target_type === "character" &&
						rel.target_id === currentPovCharacterId &&
						rel.relation_type === "pov"
				);

				if (!existingPovRelation) {
					// Validate character exists before creating relation
					try {
						await context.apiClient.getCharacter(currentPovCharacterId);
					} catch (error) {
						context.emitWarning?.({
							code: "relation_validation_error",
							message: `Cannot create POV relation: character ${currentPovCharacterId} does not exist`,
						});
						return;
					}

					try {
						if (!worldId) {
							// Need world_id for relations, try to get from story
							const story = await context.apiClient.getStory(storyId);
							if (story.world_id) {
								worldId = story.world_id;
							} else {
								context.emitWarning?.({
									code: "relation_world_missing",
									message: `Cannot create POV relation: story ${storyId} has no world_id`,
								});
								return;
							}
						}

						await context.apiClient.createRelation({
							sourceType: "scene",
							sourceId: sceneId,
							targetType: "character",
							targetId: currentPovCharacterId,
							relationType: "pov",
						});
					} catch (error) {
						context.emitWarning?.({
							code: "relation_create_error",
							message: `Failed to create POV relation: ${error}`,
						});
					}
				}
			}
		}

		// Handle Location relation
		if (currentLocationId !== oldLocationId) {
			// Delete old Location relation if it exists
			if (oldLocationId) {
				const oldLocationRelation = existingRelations.find(
					(rel) =>
						rel.target_type === "location" &&
						rel.target_id === oldLocationId &&
						rel.relation_type === "setting"
				);
				if (oldLocationRelation) {
					try {
						await context.apiClient.deleteRelation(oldLocationRelation.id);
					} catch (error) {
						context.emitWarning?.({
							code: "relation_delete_error",
							message: `Failed to delete old Location relation: ${error}`,
						});
					}
				}
			}

			// Create new Location relation if location is set
			if (currentLocationId) {
				// Check if relation already exists
				const existingLocationRelation = existingRelations.find(
					(rel) =>
						rel.target_type === "location" &&
						rel.target_id === currentLocationId &&
						rel.relation_type === "setting"
				);

				if (!existingLocationRelation) {
					// Validate location exists before creating relation
					try {
						await context.apiClient.getLocation(currentLocationId);
					} catch (error) {
						context.emitWarning?.({
							code: "relation_validation_error",
							message: `Cannot create Location relation: location ${currentLocationId} does not exist`,
						});
						return;
					}

					try {
						if (!worldId) {
							// Need world_id for relations, try to get from story
							const story = await context.apiClient.getStory(storyId);
							if (story.world_id) {
								worldId = story.world_id;
							} else {
								context.emitWarning?.({
									code: "relation_world_missing",
									message: `Cannot create Location relation: story ${storyId} has no world_id`,
								});
								return;
							}
						}

						await context.apiClient.createRelation({
							sourceType: "scene",
							sourceId: sceneId,
							targetType: "location",
							targetId: currentLocationId,
							relationType: "setting",
						});
					} catch (error) {
						context.emitWarning?.({
							code: "relation_create_error",
							message: `Failed to create Location relation: ${error}`,
						});
					}
				}
			}
		}
	}

	async delete(id: string, context: SyncContext): Promise<void> {
		await context.apiClient.deleteScene(id);
	}
}

