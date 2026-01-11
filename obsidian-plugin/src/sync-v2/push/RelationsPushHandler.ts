import type { SyncContext } from "../types/sync";
import type { EntityRelation, CreateRelationParams, UpdateRelationParams } from "../types/relations";
import { RelationsParser, type RelationsFile, type RelationEntry } from "../parsers/relationsParser";
import { getFrontmatterId } from "../utils/frontmatterHelpers";

export interface RelationsPushResult {
	created: number;
	updated: number;
	deleted: number;
	warnings: string[];
}

const TARGET_TYPE_MAP: Record<string, string> = {
	"Main Characters": "character",
	"Key Locations": "location",
	"Referenced Factions": "faction",
	"Timeline Events": "event",
	"Artifacts": "artifact",
	"Lore References": "lore",
};

export class RelationsPushHandler {
	constructor(private readonly parser = new RelationsParser()) {}

	async pushRelations(
		relationsFilePath: string,
		sourceEntityType: string,
		sourceEntityId: string,
		context: SyncContext,
		worldId?: string
	): Promise<RelationsPushResult> {
		const result: RelationsPushResult = {
			created: 0,
			updated: 0,
			deleted: 0,
			warnings: [],
		};

		try {
			// Read current relations file
			const currentContent = await context.fileManager.readFile(relationsFilePath);
			const parsed: RelationsFile = this.parser.parse(currentContent);

			// Validate frontmatter using configured ID field
			const idField = context.settings.frontmatterIdField;
			const frontmatterId = getFrontmatterId(parsed.frontmatter, idField);
			if (!frontmatterId || frontmatterId !== sourceEntityId) {
				result.warnings.push(
					`Frontmatter ID mismatch: expected ${sourceEntityId}, found ${frontmatterId ?? "none"} (using field: ${idField || "id"})`
				);
				return result;
			}

			// Fetch existing relations from API
			let existingRelations: EntityRelation[];
			if (worldId) {
				// For worlds, fetch all relations in the world
				const existingRelationsResponse = await context.apiClient.listRelationsByWorld({
					worldId,
				});
				existingRelations = existingRelationsResponse.data;
			} else {
				// For stories and other entities, fetch relations where entity is target
				const existingRelationsResponse = await context.apiClient.listRelationsByTarget({
					targetType: sourceEntityType,
					targetId: sourceEntityId,
				});
				existingRelations = existingRelationsResponse.data;
			}

			// Build maps for comparison
			// For stories: relations are character→story, so we key by source_type:source_id:relation_type
			//   When we fetch relationsByTarget for story, we get relations where:
			//     - rel.source_type = "character" (or location, etc.)
			//     - rel.source_id = "char-1" (or loc-1, etc.)
			//     - rel.target_type = "story"
			//     - rel.target_id = "story-1"
			//   So we key by source_type:source_id:relation_type to match entries in file
			// For worlds: relations are grouped by target type, so we key by target_type:target_id:relation_type
			//   When we fetch relationsByWorld, we get all relations in the world
			//   We key by target_type:target_id:relation_type to match entries in file (which show targets)
			const existingRelationsMap = new Map<string, EntityRelation[]>();
			existingRelations.forEach((rel) => {
				if (worldId) {
					// For world relations: group by target_type:target_id:relation_type
					// Multiple sources can relate to same target, so we store an array
					const key = `${rel.target_type}:${rel.target_id}:${rel.relation_type}`;
					if (!existingRelationsMap.has(key)) {
						existingRelationsMap.set(key, []);
					}
					existingRelationsMap.get(key)!.push(rel);
				} else {
					// For story relations: key by source_type:source_id:relation_type
					// The file shows sources (characters, locations, etc.), so we match by source
					const key = `${rel.source_type}:${rel.source_id}:${rel.relation_type}`;
					if (!existingRelationsMap.has(key)) {
						existingRelationsMap.set(key, []);
					}
					existingRelationsMap.get(key)!.push(rel);
				}
			});

			const fileRelationsMap = new Map<string, { entry: RelationEntry; sectionName: string }>();

			// Parse relations from file
			// For stories: relations are character→story, so sections are "Main Characters", "Key Locations", etc.
			//   Each entry is a character/location/etc (source) that relates to the story (target)
			//   In the file: `[[char-1|John]]` means source=char-1, target=story
			// For worlds: relations are grouped by target type, so sections are "Main Characters", "Key Locations", etc.
			//   Each entry is a character/location/etc (target) that is related to by various sources
			//   In the file: `[[char-1|John]]` means target=char-1, source is not explicit

			for (const section of parsed.sections) {
				if (section.name === "World") continue; // Skip World section

				const entityType = TARGET_TYPE_MAP[section.name];
				if (!entityType) {
					result.warnings.push(`Unknown section type: ${section.name}`);
					continue;
				}

				for (const entry of section.entries) {
					if (entry.placeholder) continue; // Skip placeholders

					if (!entry.link) {
						result.warnings.push(`Entry without link in section ${section.name}: ${entry.displayText}`);
						continue;
					}

					// Resolve entity ID from link
					const entityId = await this.resolveEntityId(entry.link, entityType, context);
					if (!entityId) {
						result.warnings.push(`Could not resolve entity ID for link: ${entry.link}`);
						continue;
					}

					// Determine relation type from section or use default
					const relationType = this.inferRelationType(entityType, section.name);

					// Build key for comparison
					if (worldId) {
						// For world: key by target_type:target_id:relation_type (entity is target)
						const key = `${entityType}:${entityId}:${relationType}`;
						if (!fileRelationsMap.has(key)) {
							fileRelationsMap.set(key, { entry, sectionName: section.name });
						}
					} else {
						// For story: key by source_type:source_id:relation_type (entity is source)
						const key = `${entityType}:${entityId}:${relationType}`;
						fileRelationsMap.set(key, { entry, sectionName: section.name });
					}
				}
			}

			if (worldId) {
				// For worlds: handle relations grouped by target type
				// Each entry in the file represents a target entity, but source is not explicit
				// We'll update existing relations where target matches, or show warning if not found
				for (const [key, { entry, sectionName }] of fileRelationsMap.entries()) {
					if (!entry.link) continue;

					const targetType = TARGET_TYPE_MAP[sectionName];
					if (!targetType) continue;

					const targetId = await this.resolveEntityId(entry.link, targetType, context);
					if (!targetId) continue;

					const relationType = this.inferRelationType(targetType, sectionName);

					const existingRelations = existingRelationsMap.get(key) ?? [];

					if (existingRelations.length === 0) {
						// No existing relation found - cannot infer source for world relations
						result.warnings.push(
							`No existing relation found for target ${targetType}:${targetId} with type ${relationType}. Cannot create new relation without source.`
						);
						continue;
					}

					// Update all existing relations with matching target (or just the first one for now)
					// In the future, we might need to handle multiple sources better
					const relationToUpdate = existingRelations[0];
					if (relationToUpdate.context !== entry.description) {
						try {
							await context.apiClient.updateRelation({
								id: relationToUpdate.id,
								context: entry.description,
							});
							result.updated++;
						} catch (error) {
							result.warnings.push(`Failed to update relation ${relationToUpdate.id}: ${error}`);
						}
					}
				}

				// Delete relations that exist in API but not in file (for world relations, be more conservative)
				for (const [key, relations] of existingRelationsMap.entries()) {
					if (!fileRelationsMap.has(key)) {
						// Only delete if we're sure it should be deleted
						// For now, skip deletion for world relations to be safe
						// TODO: Implement smarter deletion logic for world relations
					}
				}
			} else {
				// For stories: handle relations where source is explicit in file
				// Delete relations that exist in API but not in file
				for (const [key, relations] of existingRelationsMap.entries()) {
					if (!fileRelationsMap.has(key)) {
						for (const relation of relations) {
							try {
								await context.apiClient.deleteRelation(relation.id);
								result.deleted++;
							} catch (error) {
								result.warnings.push(`Failed to delete relation ${relation.id}: ${error}`);
							}
						}
					}
				}

				// Create or update relations from file
				// For stories: entity from file is source (character/location/etc), target is story
				for (const [key, { entry, sectionName }] of fileRelationsMap.entries()) {
					if (!entry.link) continue;

					const sourceType = TARGET_TYPE_MAP[sectionName];
					if (!sourceType) continue;

					const sourceId = await this.resolveEntityId(entry.link, sourceType, context);
					if (!sourceId) continue;

					const relationType = this.inferRelationType(sourceType, sectionName);

					const existingRelations = existingRelationsMap.get(key) ?? [];
					const existingRelation = existingRelations[0];

					if (existingRelation) {
						// Update if description changed
						if (existingRelation.context !== entry.description) {
							try {
								await context.apiClient.updateRelation({
									id: existingRelation.id,
									context: entry.description,
								});
								result.updated++;
							} catch (error) {
								result.warnings.push(`Failed to update relation ${existingRelation.id}: ${error}`);
							}
						}
					} else {
						// Create new relation: source is character/location/etc (from file), target is story
						// Validate that source entity exists before creating relation
						try {
							const sourceExists = await this.validateEntityExists(sourceType, sourceId, context);
							if (!sourceExists) {
								result.warnings.push(`Source entity ${sourceType}:${sourceId} does not exist`);
								continue;
							}

							// Validate target entity exists
							const targetExists = await this.validateEntityExists(sourceEntityType, sourceEntityId, context);
							if (!targetExists) {
								result.warnings.push(`Target entity ${sourceEntityType}:${sourceEntityId} does not exist`);
								continue;
							}

							await context.apiClient.createRelation({
								sourceType,
								sourceId,
								targetType: sourceEntityType,
								targetId: sourceEntityId,
								relationType,
								context: entry.description,
							});
							result.created++;
						} catch (error) {
							result.warnings.push(
								`Failed to create relation from ${sourceType}:${sourceId} to ${sourceEntityType}:${sourceEntityId}: ${error}`
							);
						}
					}
				}
			}
		} catch (error) {
			result.warnings.push(`Failed to push relations: ${error}`);
		}

		return result;
	}

	private async resolveEntityId(link: string, entityType: string, context: SyncContext): Promise<string | null> {
		// If link is already a valid UUID, return it
		const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
		if (uuidRegex.test(link)) {
			return link;
		}

		// Otherwise, try to resolve by name/slug
		// This is a simplified approach - in production, we'd need a proper entity lookup
		try {
			switch (entityType) {
				case "character": {
					// Try to get by ID first (in case it's a slug-like ID)
					try {
						const char = await context.apiClient.getCharacter(link);
						return char.id;
					} catch {
						// If not found, we'd need to search by name
						// For now, return null - this should be handled by entity name resolution
						return null;
					}
				}
				case "location": {
					try {
						const loc = await context.apiClient.getLocation(link);
						return loc.id;
					} catch {
						return null;
					}
				}
				case "faction": {
					try {
						const faction = await context.apiClient.getFaction(link);
						return faction.id;
					} catch {
						return null;
					}
				}
				case "artifact": {
					try {
						const artifact = await context.apiClient.getArtifact(link);
						return artifact.id;
					} catch {
						return null;
					}
				}
				case "event": {
					try {
						const event = await context.apiClient.getEvent(link);
						return event.id;
					} catch {
						return null;
					}
				}
				case "lore": {
					try {
						const lore = await context.apiClient.getLore(link);
						return lore.id;
					} catch {
						return null;
					}
				}
				default:
					return null;
			}
		} catch (error) {
			console.warn(`[RelationsPushHandler] Failed to resolve entity ID for ${entityType}:${link}`, error);
			return null;
		}
	}

	private inferRelationType(targetType: string, sectionName: string): string {
		// Map section names to relation types
		switch (sectionName) {
			case "Main Characters":
				return "pov";
			case "Key Locations":
				return "setting";
			case "Referenced Factions":
				return "faction_reference";
			case "Timeline Events":
				return "timeline_event";
			case "Artifacts":
				return "artifact_reference";
			case "Lore References":
				return "lore_reference";
			default:
				return "reference";
		}
	}

	private async validateEntityExists(entityType: string, entityId: string, context: SyncContext): Promise<boolean> {
		try {
			switch (entityType) {
				case "character":
					await context.apiClient.getCharacter(entityId);
					return true;
				case "location":
					await context.apiClient.getLocation(entityId);
					return true;
				case "faction":
					await context.apiClient.getFaction(entityId);
					return true;
				case "artifact":
					await context.apiClient.getArtifact(entityId);
					return true;
				case "event":
					await context.apiClient.getEvent(entityId);
					return true;
				case "lore":
					await context.apiClient.getLore(entityId);
					return true;
				case "story":
					await context.apiClient.getStory(entityId);
					return true;
				case "world":
					await context.apiClient.getWorld(entityId);
					return true;
				default:
					// For unknown types, assume exists (let API handle validation)
					return true;
			}
		} catch (error) {
			return false;
		}
	}
}

