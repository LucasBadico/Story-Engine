import type {
	Artifact,
	Character,
	Chapter,
	ContentBlock,
	Faction,
	Lore,
	Location,
	Beat,
	Scene,
	Story,
	World,
	WorldEvent,
} from "../../../types";
import type { SyncContext } from "../../types/sync";
import { slugify } from "../../utils/slugify";
import { RelationsGenerator } from "../../generators/RelationsGenerator";
import { CitationsGenerator } from "../../generators/CitationsGenerator";
import {
	mapRelationsToGeneratorInput,
	mapCitationsToGeneratorInput,
} from "../../relations/mappers";
import type { RelationTargetResolver, CitationSourceResolver } from "../../relations/mappers";
import { RelationsPushHandler } from "../../push/RelationsPushHandler";
import { PathResolver } from "../../fileRenamer/PathResolver";
import { CharacterHandler } from "./CharacterHandler";
import { LocationHandler } from "./LocationHandler";
import { FactionHandler } from "./FactionHandler";
import { ArtifactHandler } from "./ArtifactHandler";
import { EventHandler } from "./EventHandler";
import { LoreHandler } from "./LoreHandler";

export class WorldHandler {
	readonly entityType = "world";

	constructor(
		private readonly now: () => string = () => new Date().toISOString(),
		private readonly relationsGenerator = new RelationsGenerator(),
		private readonly citationsGenerator = new CitationsGenerator(),
		private readonly relationsPushHandler = new RelationsPushHandler()
	) {}

	async pull(id: string, context: SyncContext): Promise<World> {
		const world = await context.apiClient.getWorld(id);
		const folderPath = context.fileManager.getWorldFolderPath(world.name);
		await context.fileManager.ensureFolderExists(folderPath);
		await context.fileManager.ensureFolderExists(`${folderPath}/characters`);
		await context.fileManager.ensureFolderExists(`${folderPath}/locations`);
		await context.fileManager.ensureFolderExists(`${folderPath}/factions`);
		await context.fileManager.ensureFolderExists(`${folderPath}/artifacts`);
		await context.fileManager.ensureFolderExists(`${folderPath}/events`);
		await context.fileManager.ensureFolderExists(`${folderPath}/lore`);

		await context.fileManager.writeWorldMetadata(world, folderPath);

		const [characters, locations, factions, artifacts, events, loreList] = await Promise.all([
			context.apiClient.getCharacters(world.id),
			context.apiClient.getLocations(world.id),
			context.apiClient.getFactions(world.id),
			context.apiClient.getArtifacts(world.id),
			context.apiClient.getEvents(world.id),
			context.apiClient.getLores(world.id),
		]);

		await context.fileManager.writeFile(
			`${folderPath}/world.outline.md`,
			this.renderOutline(world, { characters, locations, factions, artifacts, events, loreList }, folderPath)
		);
		await context.fileManager.writeFile(
			`${folderPath}/world.contents.md`,
			this.renderContents(world, { characters, locations, factions, artifacts, events, loreList }, folderPath)
		);
		await this.writeWorldEntityFiles(
			{ characters, locations, factions, artifacts, events, loreList },
			context
		);
		// Generate relations and citations files
		await this.generateRelations(world, folderPath, context);
		await this.generateCitations(world, folderPath, context);

		return world;
	}

	async push(entity: World, context: SyncContext): Promise<void> {
		// Push relations if relations file exists
		const folderPath = context.fileManager.getWorldFolderPath(entity.name);
		const relationsFilePath = `${folderPath}/world.relations.md`;

		try {
			await context.fileManager.readFile(relationsFilePath);
			// File exists, push relations
			const result = await this.relationsPushHandler.pushRelations(
				relationsFilePath,
				"world",
				entity.id,
				context,
				entity.id
			);

			if (result.warnings.length > 0) {
				result.warnings.forEach((warning) =>
					context.emitWarning?.({
						code: "relations_push_warning",
						message: warning,
						filePath: relationsFilePath,
					})
				);
			}
		} catch (error: any) {
			// File doesn't exist or error reading - skip push
			if (error?.message?.includes("missing") || error?.code === "ENOENT") {
				// File doesn't exist, that's fine - no relations to push
				return;
			}
			// Other error - log warning
			context.emitWarning?.({
				code: "relations_push_error",
				message: `Failed to push relations: ${error}`,
				filePath: relationsFilePath,
			});
		}
	}

	async delete(_id: string, _context: SyncContext): Promise<void> {
		// TODO: remove world directory
	}

	private renderOutline(
		world: World,
		opts: {
			characters: Character[];
			locations: Location[];
			factions: Faction[];
			artifacts: Artifact[];
			events: WorldEvent[];
			loreList: Lore[];
		},
		folderPath: string
	): string {
		const sections: Array<{ label: string; items: { name: string; extra?: string }[] }> = [
			{
				label: "Characters",
				items: opts.characters.map((character) => ({
					name: character.name,
					extra: `Lvl ${character.class_level}`,
				})),
			},
			{
				label: "Locations",
				items: opts.locations.map((location) => ({
					name: location.name,
					extra: location.type ?? undefined,
				})),
			},
			{
				label: "Factions",
				items: opts.factions.map((faction) => ({
					name: faction.name,
					extra: faction.type ?? undefined,
				})),
			},
			{
				label: "Artifacts",
				items: opts.artifacts.map((artifact) => ({
					name: artifact.name,
					extra: artifact.rarity,
				})),
			},
			{
				label: "Events",
				items: opts.events.map((event) => ({
					name: event.name,
					extra: event.timeline ?? undefined,
				})),
			},
			{
				label: "Lore",
				items: opts.loreList.map((lore) => ({
					name: lore.name,
					extra: lore.category ?? undefined,
				})),
			},
		];

		const lines = [
			`# ${world.name} - Outline`,
			"",
			`_Atualizado em ${this.now()}_`,
			"",
		];

		const linkFor = (name: string, folder: string) =>
			`[[${folderPath}/${folder}/${slugify(name)}.md|${name}]]`;

		for (const section of sections) {
			lines.push(`## ${section.label}`, "");
			if (!section.items.length) {
				lines.push("_Nenhum item sincronizado ainda._", "");
				continue;
			}
			section.items.forEach((item) => {
				const folder =
					section.label === "Characters"
						? "characters"
						: section.label === "Locations"
						? "locations"
						: section.label === "Factions"
						? "factions"
						: section.label === "Artifacts"
						? "artifacts"
						: section.label === "Events"
						? "events"
						: "lore";
				const link = linkFor(item.name, folder);
				lines.push(`- ${link}${item.extra ? ` (${item.extra})` : ""}`);
			});
			lines.push("");
		}

		return lines.join("\n").trimEnd() + "\n";
	}

	private async writeWorldEntityFiles(
		entities: {
			characters: Character[];
			locations: Location[];
			factions: Faction[];
			artifacts: Artifact[];
			events: WorldEvent[];
			loreList: Lore[];
		},
		context: SyncContext
	): Promise<void> {
		const characterHandler = new CharacterHandler();
		const locationHandler = new LocationHandler();
		const factionHandler = new FactionHandler();
		const artifactHandler = new ArtifactHandler();
		const eventHandler = new EventHandler();
		const loreHandler = new LoreHandler();

		await Promise.all([
			Promise.all(entities.characters.map((character) => characterHandler.pull(character.id, context))),
			Promise.all(entities.locations.map((location) => locationHandler.pull(location.id, context))),
			Promise.all(entities.factions.map((faction) => factionHandler.pull(faction.id, context))),
			Promise.all(entities.artifacts.map((artifact) => artifactHandler.pull(artifact.id, context))),
			Promise.all(entities.events.map((event) => eventHandler.pull(event.id, context))),
			Promise.all(entities.loreList.map((lore) => loreHandler.pull(lore.id, context))),
		]);
	}

	private renderContents(
		world: World,
		opts: {
			characters: Character[];
			locations: Location[];
			factions: Faction[];
			artifacts: Artifact[];
			events: WorldEvent[];
			loreList: Lore[];
		},
		folderPath: string
	): string {
		const lines: string[] = [
			`# ${world.name} - Contents`,
			"",
			"## Overview",
			world.description || "_Sem descrição._",
			"",
			"## Time Configuration",
			world.time_config ? "```json\n" + JSON.stringify(world.time_config, null, 2) + "\n```" : "_Não configurado._",
			"",
		];

		const addSection = (label: string, items: { name: string; path: string; extra?: string }[]) => {
			lines.push(`## ${label}`, "");
			if (!items.length) {
				lines.push("_Nenhum item sincronizado ainda._", "");
				return;
			}
			items.forEach((item) => {
				const link = `[[${item.path}|${item.name}]]`;
				lines.push(`- ${link}${item.extra ? ` (${item.extra})` : ""}`);
			});
			lines.push("");
		};

		addSection(
			"Characters",
			opts.characters.map((character) => ({
				name: character.name,
				path: `${folderPath}/characters/${slugify(character.name)}.md`,
				extra: `Lvl ${character.class_level}`,
			}))
		);
		addSection(
			"Locations",
			opts.locations.map((location) => ({
				name: location.name,
				path: `${folderPath}/locations/${slugify(location.name)}.md`,
				extra: location.type ?? undefined,
			}))
		);
		addSection(
			"Factions",
			opts.factions.map((faction) => ({
				name: faction.name,
				path: `${folderPath}/factions/${slugify(faction.name)}.md`,
				extra: faction.type ?? undefined,
			}))
		);
		addSection(
			"Artifacts",
			opts.artifacts.map((artifact) => ({
				name: artifact.name,
				path: `${folderPath}/artifacts/${slugify(artifact.name)}.md`,
				extra: artifact.rarity,
			}))
		);
		addSection(
			"Events",
			opts.events.map((event) => ({
				name: event.name,
				path: `${folderPath}/events/${slugify(event.name)}.md`,
				extra: event.timeline ?? undefined,
			}))
		);
		addSection(
			"Lore",
			opts.loreList.map((lore) => ({
				name: lore.name,
				path: `${folderPath}/lore/${slugify(lore.name)}.md`,
				extra: lore.category ?? undefined,
			}))
		);

		return lines.join("\n");
	}

	private renderRelationsPlaceholder(world: World): string {
		return [
			`# ${world.name} - Relations`,
			"",
			"_As relações entre characters, locations e stories serão preenchidas na Fase 8._",
			"",
		].join("\n");
	}

	private async generateRelations(
		world: World,
		folderPath: string,
		context: SyncContext
	): Promise<void> {
		try {
			// Fetch all relations for this world
			const relationsResponse = await context.apiClient.listRelationsByWorld({
				worldId: world.id,
			});

			// Build entity cache for efficient lookup
			const entityCache = new Map<string, { name: string; type: string }>();

			// Resolve all target names asynchronously
			const resolvedRelations = await Promise.all(
				relationsResponse.data.map(async (relation) => {
					try {
						let targetName = relation.target_id;
						let targetId = relation.target_id;
						let targetType = relation.target_type;

						const cacheKey = `${relation.target_type}:${relation.target_id}`;
						if (entityCache.has(cacheKey)) {
							const cached = entityCache.get(cacheKey)!;
							targetName = cached.name;
							targetId = relation.target_id;
							targetType = cached.type;
						} else {
							switch (relation.target_type) {
								case "character": {
									const char = await context.apiClient.getCharacter(relation.target_id);
									targetName = char.name;
									targetId = char.id;
									entityCache.set(cacheKey, { name: char.name, type: "character" });
									break;
								}
								case "location": {
									const loc = await context.apiClient.getLocation(relation.target_id);
									targetName = loc.name;
									targetId = loc.id;
									entityCache.set(cacheKey, { name: loc.name, type: "location" });
									break;
								}
								case "faction": {
									const faction = await context.apiClient.getFaction(relation.target_id);
									targetName = faction.name;
									targetId = faction.id;
									entityCache.set(cacheKey, { name: faction.name, type: "faction" });
									break;
								}
								case "artifact": {
									const artifact = await context.apiClient.getArtifact(relation.target_id);
									targetName = artifact.name;
									targetId = artifact.id;
									entityCache.set(cacheKey, { name: artifact.name, type: "artifact" });
									break;
								}
								case "event": {
									const event = await context.apiClient.getEvent(relation.target_id);
									targetName = event.name;
									targetId = event.id;
									entityCache.set(cacheKey, { name: event.name, type: "event" });
									break;
								}
								case "lore": {
									const lore = await context.apiClient.getLore(relation.target_id);
									targetName = lore.name;
									targetId = lore.id;
									entityCache.set(cacheKey, { name: lore.name, type: "lore" });
									break;
								}
								case "story": {
									const story = await context.apiClient.getStory(relation.target_id);
									targetName = story.title;
									targetId = story.id;
									entityCache.set(cacheKey, { name: story.title, type: "story" });
									break;
								}
							}
						}

						return {
							targetType,
							targetId,
							targetName,
							relationType: relation.relation_type,
							summary: relation.context,
						};
					} catch (error) {
						console.warn(`[Sync V2] Failed to resolve target for world relation`, {
							relation,
							error,
						});
						return {
							targetType: relation.target_type,
							targetId: relation.target_id,
							targetName: relation.target_id,
							relationType: relation.relation_type,
							summary: relation.context,
						};
					}
				})
			);

			// Build resolver that uses pre-resolved data
			const entityMap = new Map(
				resolvedRelations.map((r) => [`${r.targetType}:${r.targetId}`, r] as const)
			);
			const resolveTarget: RelationTargetResolver = (relation) => {
				const key = `${relation.target_type}:${relation.target_id}` as `${string}:${string}`;
				const resolved = entityMap.get(key);
				if (!resolved) return null;

				return {
					targetId: resolved.targetId,
					targetName: resolved.targetName,
					summary: resolved.summary,
				};
			};

			const input = mapRelationsToGeneratorInput({
				entity: {
					id: world.id,
					name: world.name,
					type: "world",
				},
				relations: relationsResponse.data,
				resolveTarget,
				options: {
					syncedAt: this.now(),
					showHelpBox: context.settings.showHelpBox,
					idField: context.settings.frontmatterIdField,
					worldFolderPath: folderPath,
				},
			});

			const relationsContent = this.relationsGenerator.generate(input);
			await context.fileManager.writeFile(`${folderPath}/world.relations.md`, relationsContent);
		} catch (error) {
			console.warn("[Sync V2] Failed to generate world relations file", { worldId: world.id, error });
			// Fallback to placeholder
			await context.fileManager.writeFile(
				`${folderPath}/world.relations.md`,
				this.renderRelationsPlaceholder(world)
			);
		}
	}

	private async generateCitations(
		world: World,
		folderPath: string,
		context: SyncContext
	): Promise<void> {
		try {
			// For world citations, we want relations where story elements (chapters, scenes, beats, content_blocks)
			// cite world entities. We fetch relations by world, then filter for story elements as sources.

			// First, get all relations for this world
			const relationsResponse = await context.apiClient.listRelationsByWorld({
				worldId: world.id,
			});

			// Filter relations where source is a story element
			const storyElementRelations = relationsResponse.data.filter((rel) =>
				["chapter", "scene", "beat", "content_block"].includes(rel.source_type)
			);

			if (storyElementRelations.length === 0) {
				// No citations, generate empty file
				const input = mapCitationsToGeneratorInput({
					entity: {
						id: world.id,
						name: world.name,
						type: "world",
					},
					relations: [],
					resolveSource: () => null,
					options: {
						syncedAt: this.now(),
						idField: context.settings.frontmatterIdField,
					},
				});
				const citationsContent = this.citationsGenerator.generate(input);
				await context.fileManager.writeFile(`${folderPath}/world.citations.md`, citationsContent);
				return;
			}

			// Group relations by source_id to reduce API calls
			const relationsBySource = new Map<string, typeof storyElementRelations>();
			for (const rel of storyElementRelations) {
				if (!relationsBySource.has(rel.source_id)) {
					relationsBySource.set(rel.source_id, []);
				}
				relationsBySource.get(rel.source_id)!.push(rel);
			}

		// Build story hierarchy cache for each story that has relations
		const storyCache = new Map<string, Story>();
		const chapterCache = new Map<string, Chapter>();
		const sceneCache = new Map<string, Scene>();
		const beatCache = new Map<string, Beat>();
		const contentBlockCache = new Map<string, ContentBlock>();
		const resolverCache = new Map<string, PathResolver>();

			// For chapters, scenes, and beats, we need to fetch their parent story
			// This is expensive, so we'll do it in batches and cache results
			const storyIdsToFetch = new Set<string>();
			for (const [sourceId, relations] of relationsBySource.entries()) {
				const firstRel = relations[0];
				if (firstRel.source_type === "chapter") {
					try {
						// Chapters have story_id in their data - we'd need to fetch the chapter first
						// For now, we'll skip this optimization and use a simpler approach
					} catch (error) {
						console.warn(`[Sync V2] Failed to resolve story for chapter`, { sourceId, error });
					}
				}
			}

		const getStory = async (storyId: string): Promise<Story> => {
			const cached = storyCache.get(storyId);
			if (cached) return cached;
			const story = await context.apiClient.getStory(storyId);
			storyCache.set(storyId, story);
			return story;
		};

		const getChapter = async (chapterId: string): Promise<Chapter> => {
			const cached = chapterCache.get(chapterId);
			if (cached) return cached;
			const chapter = await context.apiClient.getChapter(chapterId);
			chapterCache.set(chapterId, chapter);
			return chapter;
		};

		const getScene = async (sceneId: string): Promise<Scene> => {
			const cached = sceneCache.get(sceneId);
			if (cached) return cached;
			const scene = await context.apiClient.getScene(sceneId);
			sceneCache.set(sceneId, scene);
			return scene;
		};

		const getBeat = async (beatId: string): Promise<Beat> => {
			const cached = beatCache.get(beatId);
			if (cached) return cached;
			const beat = await context.apiClient.getBeat(beatId);
			beatCache.set(beatId, beat);
			return beat;
		};

		const getContentBlock = async (contentBlockId: string): Promise<ContentBlock> => {
			const cached = contentBlockCache.get(contentBlockId);
			if (cached) return cached;
			const contentBlock = await context.apiClient.getContentBlock(contentBlockId);
			contentBlockCache.set(contentBlockId, contentBlock);
			return contentBlock;
		};

		const getResolver = (story: Story): PathResolver => {
			const cached = resolverCache.get(story.id);
			if (cached) return cached;
			const storyFolder = context.fileManager.getStoryFolderPath(story.title);
			const resolver = new PathResolver(storyFolder);
			resolverCache.set(story.id, resolver);
			return resolver;
		};

		const resolveSource: CitationSourceResolver = (relation) => {
			if (!["chapter", "scene", "beat", "content_block"].includes(relation.source_type)) {
				return null;
			}

			return {
				storyId: "unknown",
				storyTitle: "Unknown Story",
				sourceTitle: relation.source_id,
				sourceType: relation.source_type as "chapter" | "scene" | "beat" | "content_block",
				summary: relation.context,
			};
		};

		const resolveSourceAsync = async (relation: typeof storyElementRelations[number]) => {
			if (relation.source_type === "chapter") {
				const chapter = await getChapter(relation.source_id);
				const story = await getStory(chapter.story_id);
				const resolver = getResolver(story);
				const storyFolder = context.fileManager.getStoryFolderPath(story.title);
				return {
					storyId: story.id,
					storyTitle: story.title,
					storyPath: `${storyFolder}/story.md`,
					sourceTitle: `Chapter ${chapter.number}: ${chapter.title}`,
					sourceType: "chapter" as const,
					sourcePath: resolver.getChapterPath(chapter),
					chapterTitle: chapter.title,
					summary: relation.context,
				};
			}

			if (relation.source_type === "scene") {
				const scene = await getScene(relation.source_id);
				const story = await getStory(scene.story_id);
				const resolver = getResolver(story);
				const storyFolder = context.fileManager.getStoryFolderPath(story.title);
				let chapterTitle: string | undefined;
				let chapterOrder = 0;
				if (scene.chapter_id) {
					const chapter = await getChapter(scene.chapter_id);
					chapterTitle = chapter.title;
					chapterOrder = chapter.number ?? 0;
				}
				return {
					storyId: story.id,
					storyTitle: story.title,
					storyPath: `${storyFolder}/story.md`,
					sourceTitle: `Scene ${scene.order_num ?? 0}: ${scene.goal || "Untitled"}`,
					sourceType: "scene" as const,
					sourcePath: resolver.getScenePath(scene, { chapterOrder }),
					chapterTitle,
					summary: relation.context,
				};
			}

			if (relation.source_type === "beat") {
				const beat = await getBeat(relation.source_id);
				const scene = await getScene(beat.scene_id);
				const story = await getStory(scene.story_id);
				const resolver = getResolver(story);
				const storyFolder = context.fileManager.getStoryFolderPath(story.title);
				let chapterTitle: string | undefined;
				let chapterOrder = 0;
				if (scene.chapter_id) {
					const chapter = await getChapter(scene.chapter_id);
					chapterTitle = chapter.title;
					chapterOrder = chapter.number ?? 0;
				}
				return {
					storyId: story.id,
					storyTitle: story.title,
					storyPath: `${storyFolder}/story.md`,
					sourceTitle: `Beat ${beat.order_num ?? 0}: ${beat.intent || "Untitled"}`,
					sourceType: "beat" as const,
					sourcePath: resolver.getBeatPath(beat, {
						chapterOrder,
						sceneOrder: scene.order_num ?? 0,
					}),
					chapterTitle,
					summary: relation.context,
				};
			}

			if (relation.source_type === "content_block") {
				const block = await getContentBlock(relation.source_id);
				if (!block.chapter_id) {
					return null;
				}
				const chapter = await getChapter(block.chapter_id);
				const story = await getStory(chapter.story_id);
				const resolver = getResolver(story);
				const storyFolder = context.fileManager.getStoryFolderPath(story.title);
				const sourceTitle =
					block.metadata?.title ??
					block.kind ??
					block.type ??
					"Content Block";
				return {
					storyId: story.id,
					storyTitle: story.title,
					storyPath: `${storyFolder}/story.md`,
					sourceTitle,
					sourceType: "content_block" as const,
					sourcePath: resolver.getContentBlockPath(block),
					chapterTitle: chapter.title,
					summary: relation.context,
				};
			}

			return null;
		};

			// For now, generate citations with limited information
			// This can be enhanced later with proper story hierarchy resolution
		const resolvedSources = await Promise.all(
			storyElementRelations.map(async (relation) => ({
				relation,
				source: await resolveSourceAsync(relation),
			}))
		);
		const input = mapCitationsToGeneratorInput({
			entity: {
				id: world.id,
				name: world.name,
				type: "world",
			},
			relations: storyElementRelations,
			resolveSource: (relation) =>
				resolvedSources.find((item) => item.relation === relation)?.source ?? resolveSource(relation),
			options: {
				syncedAt: this.now(),
				idField: context.settings.frontmatterIdField,
			},
		});

			const citationsContent = this.citationsGenerator.generate(input);
			await context.fileManager.writeFile(`${folderPath}/world.citations.md`, citationsContent);
		} catch (error) {
			console.warn("[Sync V2] Failed to generate world citations file", { worldId: world.id, error });
			// Fallback to placeholder
			await context.fileManager.writeFile(
				`${folderPath}/world.citations.md`,
				this.renderCitationsPlaceholder(world)
			);
		}
	}

	private renderCitationsPlaceholder(world: World): string {
		return [
			`# ${world.name} - Citations`,
			"",
			"_Citações para este world serão sincronizadas quando o Relations/Citations pipeline estiver ativo._",
			"",
		].join("\n");
	}
}

