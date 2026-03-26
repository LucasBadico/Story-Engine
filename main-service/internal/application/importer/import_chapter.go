package importer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/story-engine/main-service/internal/application/importer/parser"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/core/world"
	"github.com/story-engine/main-service/internal/platform/logger"
)

// ImportChapterUseCase orchestrates importing a chapter from a Google Drive export folder.
// It reads files, parses them, deduplicates entities, and creates them via the main service API.
type ImportChapterUseCase struct {
	api           MainServiceAPI
	gatewayClient *GatewayClient
	logger        logger.Logger
}

// NewImportChapterUseCase creates a new ImportChapterUseCase
func NewImportChapterUseCase(api MainServiceAPI, logger logger.Logger) *ImportChapterUseCase {
	return &ImportChapterUseCase{
		api:    api,
		logger: logger,
	}
}

// SetGatewayClient sets the optional gateway client for LLM inference
func (uc *ImportChapterUseCase) SetGatewayClient(client *GatewayClient) {
	uc.gatewayClient = client
}

// exportFiles holds the raw file contents read from the export folder
type exportFiles struct {
	tabsIndex     string
	storyCanon    string
	chapterLog    string
	content       string // optional
	notes         string // optional
	granularDir   string // path to story_canon/ subdirectory, empty if not present
	granularFiles map[string]string // filename -> content
}

// Execute imports a chapter from an export folder.
// It reads all files, parses them, deduplicates entities by name,
// and creates new entities via the main service API. Errors on individual entities
// are logged and added to the report, not propagated.
func (uc *ImportChapterUseCase) Execute(ctx context.Context, input ImportChapterInput) (*ImportReport, error) {
	report := NewImportReport()

	// 1. Read all files from the export folder
	files, err := readExportFolder(input.FolderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read export folder: %w", err)
	}

	// 2. Parse tabs_index
	_, err = parser.ParseTabsIndex(files.tabsIndex)
	if err != nil {
		uc.logger.Warn("failed to parse tabs_index, continuing", "error", err)
	}

	// 3. Parse story canon — try granular files first, fall back to combined
	var canon *parser.ParsedCanon
	if len(files.granularFiles) > 0 {
		canon = uc.parseGranularCanon(files.granularFiles)
	}
	if canon == nil || (len(canon.Characters) == 0 && len(canon.Locations) == 0 && len(canon.Artifacts) == 0) {
		if files.storyCanon != "" {
			parsed, err := parser.ParseStoryCanon(files.storyCanon)
			if err != nil {
				uc.logger.Warn("failed to parse story_canon", "error", err)
			} else {
				canon = parsed
			}
		}
	}

	// 4. Parse chapter log
	var chapterLog *parser.ParsedChapterLog
	if files.chapterLog != "" {
		parsed, err := parser.ParseChapterLog(files.chapterLog)
		if err != nil {
			uc.logger.Warn("failed to parse chapter_log", "error", err)
		} else {
			chapterLog = parsed
		}
	}

	// 5. Load existing entities for dedup
	existingCharacters, err := uc.api.ListCharacters(ctx, input.WorldID)
	if err != nil {
		return nil, fmt.Errorf("failed to list existing characters: %w", err)
	}
	existingLocations, err := uc.api.ListLocations(ctx, input.WorldID)
	if err != nil {
		return nil, fmt.Errorf("failed to list existing locations: %w", err)
	}
	existingArtifacts, err := uc.api.ListArtifacts(ctx, input.WorldID)
	if err != nil {
		return nil, fmt.Errorf("failed to list existing artifacts: %w", err)
	}

	// Character name -> ID map for later use (character states, relationship states)
	characterNameToID := make(map[string]uuid.UUID)

	// Populate map with existing characters
	for _, c := range existingCharacters {
		characterNameToID[strings.ToLower(c.Name)] = c.ID
	}

	// 6. Import characters (dedup by name)
	if canon != nil {
		for _, pc := range canon.Characters {
			if pc.Name == "" {
				continue
			}
			existing := findCharacterByName(existingCharacters, pc.Name)
			if existing != nil {
				report.AddSkipped("character", pc.Name, existing.ID)
				characterNameToID[strings.ToLower(pc.Name)] = existing.ID
				continue
			}

			character, err := uc.api.CreateCharacter(ctx, APICreateCharacterInput{
				WorldID:     input.WorldID,
				Name:        pc.Name,
				Description: pc.Description,
			})
			if err != nil {
				uc.logger.Error("failed to create character", "error", err, "name", pc.Name)
				report.AddFailed("character", pc.Name, err.Error())
				continue
			}
			report.AddCreated("character", pc.Name, character.ID)
			characterNameToID[strings.ToLower(pc.Name)] = character.ID
			existingCharacters = append(existingCharacters, character)
		}
	}

	// 7. Import locations (dedup by name)
	if canon != nil {
		for _, pl := range canon.Locations {
			if pl.Name == "" {
				continue
			}
			existing := findLocationByName(existingLocations, pl.Name)
			if existing != nil {
				report.AddSkipped("location", pl.Name, existing.ID)
				continue
			}

			location, err := uc.api.CreateLocation(ctx, APICreateLocationInput{
				WorldID:     input.WorldID,
				Name:        pl.Name,
				Description: pl.Description,
			})
			if err != nil {
				uc.logger.Error("failed to create location", "error", err, "name", pl.Name)
				report.AddFailed("location", pl.Name, err.Error())
				continue
			}
			report.AddCreated("location", pl.Name, location.ID)
			existingLocations = append(existingLocations, location)
		}
	}

	// 8. Import artifacts (dedup by name)
	if canon != nil {
		for _, pa := range canon.Artifacts {
			if pa.Name == "" {
				continue
			}
			existing := findArtifactByName(existingArtifacts, pa.Name)
			if existing != nil {
				report.AddSkipped("artifact", pa.Name, existing.ID)
				continue
			}

			artifact, err := uc.api.CreateArtifact(ctx, APICreateArtifactInput{
				WorldID:     input.WorldID,
				Name:        pa.Name,
				Description: pa.Description,
			})
			if err != nil {
				uc.logger.Error("failed to create artifact", "error", err, "name", pa.Name)
				report.AddFailed("artifact", pa.Name, err.Error())
				continue
			}
			report.AddCreated("artifact", pa.Name, artifact.ID)
			existingArtifacts = append(existingArtifacts, artifact)
		}
	}

	// 9. Resolve or create chapter
	var chapterID *uuid.UUID
	if input.ChapterID != nil {
		chapterID = input.ChapterID
		uc.logger.Info("using provided chapter ID", "chapter_id", *chapterID)
	} else {
		existingChapters, err := uc.api.ListChapters(ctx, input.StoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to list existing chapters: %w", err)
		}
		var found *uuid.UUID
		for _, ch := range existingChapters {
			if ch.Number == input.ChapterNumber {
				found = &ch.ID
				report.AddSkipped("chapter", ch.Title, ch.ID)
				break
			}
		}
		if found != nil {
			chapterID = found
			uc.logger.Info("found existing chapter by number", "chapter_id", *chapterID, "chapter_number", input.ChapterNumber)
		} else {
			chapterTitle := fmt.Sprintf("Chapter %d", input.ChapterNumber)
			if chapterLog != nil && chapterLog.ChapterTitle != "" {
				chapterTitle = chapterLog.ChapterTitle
			}
			chapter, err := uc.api.CreateChapter(ctx, APICreateChapterInput{
				StoryID: input.StoryID,
				Number:  input.ChapterNumber,
				Title:   chapterTitle,
				Status:  "draft",
			})
			if err != nil {
				uc.logger.Error("failed to create chapter", "error", err)
				report.AddFailed("chapter", chapterTitle, err.Error())
			} else {
				chapterID = &chapter.ID
				report.AddCreated("chapter", chapterTitle, chapter.ID)
				uc.logger.Info("created chapter", "chapter_id", *chapterID, "title", chapterTitle)
			}
		}
	}

	// 10. Import scenes from chapter log and build scene number -> ID map
	sceneNumberToID := make(map[int]uuid.UUID)
	if chapterLog != nil && chapterID != nil {
		for _, ps := range chapterLog.Scenes {
			goal := ps.Description
			if goal == "" {
				goal = ps.Title
			}
			orderNum := ps.Number
			if orderNum < 1 {
				orderNum = 1
			}
			scene, err := uc.api.CreateScene(ctx, APICreateSceneInput{
				StoryID:   input.StoryID,
				ChapterID: chapterID,
				OrderNum:  orderNum,
				Goal:      goal,
			})
			if err != nil {
				uc.logger.Error("failed to create scene", "error", err, "scene_number", ps.Number, "title", ps.Title)
				report.AddFailed("scene", ps.Title, err.Error())
				continue
			}
			report.AddCreated("scene", ps.Title, scene.ID)
			sceneNumberToID[ps.Number] = scene.ID
		}
	}

	// 11. Infer chapter structure and create beats + content blocks with ContentAnchors
	inferenceUsed := false
	if uc.gatewayClient != nil && files.content != "" && chapterID != nil && len(sceneNumberToID) > 0 && chapterLog != nil {
		inferScenes := make([]InferenceSceneInput, 0, len(chapterLog.Scenes))
		for _, s := range chapterLog.Scenes {
			inferScenes = append(inferScenes, InferenceSceneInput{
				Number:      s.Number,
				Title:       s.Title,
				Description: s.Description,
			})
		}

		contentScenes, parseErr := parser.ParseContent(files.content)
		var rawProse string
		if parseErr == nil {
			var parts []string
			for _, cs := range contentScenes {
				if strings.TrimSpace(cs.Content) != "" {
					parts = append(parts, cs.Content)
				}
			}
			rawProse = strings.Join(parts, "\n\n")
		}

		if len(inferScenes) > 0 && rawProse != "" {
			uc.logger.Info("calling LLM gateway for chapter structure inference",
				"scene_count", len(inferScenes),
				"prose_length", len(rawProse),
			)

			inferResult, err := uc.gatewayClient.InferChapterStructure(ctx, inferScenes, rawProse)
			if err != nil {
				uc.logger.Warn("chapter structure inference failed, falling back to basic content blocks", "error", err)
			} else {
				inferenceUsed = true
				paragraphs := splitParagraphs(rawProse)

				for _, inferredScene := range inferResult.Scenes {
					sceneID, ok := sceneNumberToID[inferredScene.SceneNumber]
					if !ok {
						uc.logger.Warn("inferred scene not found in created scenes", "scene_number", inferredScene.SceneNumber)
						continue
					}

					beatOrder := 0
					for _, inferredBeat := range inferredScene.Beats {
						beatOrder++

						beatType := parseBeatType(inferredBeat.Type)
						beat, err := uc.api.CreateBeat(ctx, APICreateBeatInput{
							SceneID:  sceneID,
							OrderNum: beatOrder,
							Type:     beatType,
							Intent:   inferredBeat.Intent,
						})
						if err != nil {
							uc.logger.Error("failed to create beat", "error", err, "scene", inferredScene.SceneNumber, "beat", beatOrder)
							report.AddFailed("beat", fmt.Sprintf("scene_%d_beat_%d", inferredScene.SceneNumber, beatOrder), err.Error())
							continue
						}
						report.AddCreated("beat", fmt.Sprintf("%s: %s", inferredBeat.Type, inferredBeat.Intent), beat.ID)

						beatText := joinParagraphs(paragraphs, inferredBeat.StartParagraph, inferredBeat.EndParagraph)
						if strings.TrimSpace(beatText) == "" {
							continue
						}

						orderNum := (inferredScene.SceneNumber * 100) + beatOrder
						cb, err := uc.api.CreateContentBlock(ctx, APICreateContentBlockInput{
							ChapterID: chapterID,
							OrderNum:  &orderNum,
							Type:      story.ContentTypeText,
							Kind:      story.ContentKindFinal,
							Content:   beatText,
						})
						if err != nil {
							uc.logger.Error("failed to create content block for beat", "error", err)
							report.AddFailed("content_block", fmt.Sprintf("beat_%s", inferredBeat.Type), err.Error())
							continue
						}
						report.AddCreated("content_block", fmt.Sprintf("beat_%s", inferredBeat.Type), cb.ID)

						// Create content anchors linking content block to beat and scene
						if _, err := uc.api.CreateContentAnchor(ctx, cb.ID, string(story.EntityTypeBeat), beat.ID); err != nil {
							uc.logger.Error("failed to create beat anchor", "error", err)
						}
						if _, err := uc.api.CreateContentAnchor(ctx, cb.ID, string(story.EntityTypeScene), sceneID); err != nil {
							uc.logger.Error("failed to create scene anchor", "error", err)
						}
					}
				}
			}
		}
	}

	// Fallback: import content blocks without inference (original behavior)
	if !inferenceUsed && files.content != "" && chapterID != nil {
		uc.createBasicContentBlocks(ctx, files.content, chapterID, report)
	}

	// 12. Import character states (always create — one per character per chapter)
	if chapterLog != nil {
		for _, pcs := range chapterLog.CharacterStates {
			if pcs.Name == "" {
				continue
			}
			charID, ok := resolveCharacterID(characterNameToID, pcs.Name)
			if !ok {
				uc.logger.Warn("character not found for state, skipping", "character_name", pcs.Name)
				report.AddFailed("character_state", pcs.Name, "character not found by name")
				continue
			}

			var timelineDay *int
			var timelineConfidence *string
			if chapterLog.Timeline != nil && chapterLog.Timeline.Day != nil {
				timelineDay = chapterLog.Timeline.Day
				conf := chapterLog.Timeline.Confidence
				timelineConfidence = &conf
			}

			var secrets []string
			for _, s := range chapterLog.Secrets {
				if strings.EqualFold(s.EntityName, pcs.Name) {
					secrets = append(secrets, s.Secret)
				}
			}

			cs, err := uc.api.CreateCharacterState(ctx, APICreateCharacterStateInput{
				CharacterID:        charID,
				ChapterID:          chapterID,
				ChapterNumber:      input.ChapterNumber,
				Emotion:            pcs.Fields["Emotion"],
				Goal:               pcs.Fields["Goal"],
				Secrets:            secrets,
				Clothes:            pcs.Fields["Clothes"],
				PhysicalState:      pcs.Fields["Physical State"],
				TimelineDay:        timelineDay,
				TimelineConfidence: timelineConfidence,
			})
			if err != nil {
				uc.logger.Error("failed to create character state", "error", err, "character_name", pcs.Name)
				report.AddFailed("character_state", pcs.Name, err.Error())
				continue
			}
			report.AddCreated("character_state", pcs.Name, cs.ID)
		}
	}

	// 13. Import relationship states
	if chapterLog != nil && chapterID != nil {
		for _, prs := range chapterLog.RelationshipStates {
			if prs.CharacterA == "" || prs.CharacterB == "" {
				continue
			}
			sourceID, okA := resolveCharacterID(characterNameToID, prs.CharacterA)
			targetID, okB := resolveCharacterID(characterNameToID, prs.CharacterB)
			if !okA || !okB {
				name := prs.CharacterA + " <-> " + prs.CharacterB
				uc.logger.Warn("character not found for relationship state", "characterA", prs.CharacterA, "characterB", prs.CharacterB)
				report.AddFailed("relationship_state", name, "one or both characters not found by name")
				continue
			}

			var timelineDay *int
			var timelineConfidence *string
			if chapterLog.Timeline != nil && chapterLog.Timeline.Day != nil {
				timelineDay = chapterLog.Timeline.Day
				conf := chapterLog.Timeline.Confidence
				timelineConfidence = &conf
			}

			name := prs.CharacterA + " <-> " + prs.CharacterB
			rs, err := uc.api.CreateRelationshipState(ctx, APICreateRelationshipStateInput{
				ChapterID:          *chapterID,
				ChapterNumber:      input.ChapterNumber,
				TensionDescription: prs.Description,
				TimelineDay:        timelineDay,
				TimelineConfidence: timelineConfidence,
				Entities: []APIEntityLink{
					{Role: "source", EntityType: "character", EntityID: sourceID},
					{Role: "target", EntityType: "character", EntityID: targetID},
				},
			})
			if err != nil {
				uc.logger.Error("failed to create relationship state", "error", err, "relationship", name)
				report.AddFailed("relationship_state", name, err.Error())
				continue
			}
			report.AddCreated("relationship_state", name, rs.ID)
		}
	}

	// 14. Import timeline events
	if chapterLog != nil && chapterLog.Timeline != nil && chapterLog.Timeline.RawText != "" && chapterID != nil {
		var timelineDay *int
		var timelineConfidence *string
		if chapterLog.Timeline.Day != nil {
			timelineDay = chapterLog.Timeline.Day
			conf := chapterLog.Timeline.Confidence
			timelineConfidence = &conf
		}

		te, err := uc.api.CreateTimelineEvent(ctx, APICreateTimelineEventInput{
			StoryID:            input.StoryID,
			ChapterID:          *chapterID,
			ChapterNumber:      input.ChapterNumber,
			TimelineDay:        timelineDay,
			TimelineConfidence: timelineConfidence,
			Description:        chapterLog.Timeline.RawText,
		})
		if err != nil {
			uc.logger.Error("failed to create timeline event", "error", err)
			report.AddFailed("timeline_event", fmt.Sprintf("chapter_%d", input.ChapterNumber), err.Error())
		} else {
			report.AddCreated("timeline_event", fmt.Sprintf("chapter_%d", input.ChapterNumber), te.ID)
		}
	}

	// 15. Import plot threads from chapter log
	if chapterLog != nil && len(chapterLog.PlotThreads) > 0 {
		for _, pt := range chapterLog.PlotThreads {
			if pt.Title == "" {
				continue
			}

			plotType := pt.Type
			if plotType == "" {
				plotType = story.PlotTypeSecondary
			}

			plot, err := uc.api.CreatePlot(ctx, APICreatePlotInput{
				StoryID:     input.StoryID,
				Title:       pt.Title,
				Type:        plotType,
				Description: pt.Description,
			})
			if err != nil {
				uc.logger.Error("failed to create plot", "error", err, "title", pt.Title)
				report.AddFailed("plot", pt.Title, err.Error())
				continue
			}
			report.AddCreated("plot", pt.Title, plot.ID)

			// Create the initial development (the action for this chapter)
			if pt.Action != "" && chapterID != nil {
				dev, err := uc.api.CreatePlotDevelopment(ctx, APICreatePlotDevelopmentInput{
					PlotID:        plot.ID,
					ChapterNumber: input.ChapterNumber,
					Action:        pt.Action,
					Description:   pt.Description,
				})
				if err != nil {
					uc.logger.Error("failed to create plot development", "error", err, "plot", pt.Title)
					report.AddFailed("plot_development", pt.Title, err.Error())
				} else {
					report.AddCreated("plot_development", pt.Title+" ("+pt.Action+")", dev.ID)
				}
			}
		}
	}

	uc.logger.Info("chapter import complete",
		"chapter_number", input.ChapterNumber,
		"created", len(report.Created),
		"skipped", len(report.Skipped),
		"failed", len(report.Failed),
	)

	return report, nil
}

// createBasicContentBlocks creates content blocks without inference (fallback path)
func (uc *ImportChapterUseCase) createBasicContentBlocks(ctx context.Context, content string, chapterID *uuid.UUID, report *ImportReport) {
	contentScenes, err := parser.ParseContent(content)
	if err != nil {
		uc.logger.Warn("failed to parse content file", "error", err)
		return
	}
	for i, cs := range contentScenes {
		if strings.TrimSpace(cs.Content) == "" {
			continue
		}
		orderNum := i + 1
		cb, err := uc.api.CreateContentBlock(ctx, APICreateContentBlockInput{
			ChapterID: chapterID,
			OrderNum:  &orderNum,
			Type:      story.ContentTypeText,
			Kind:      story.ContentKindFinal,
			Content:   cs.Content,
		})
		if err != nil {
			name := cs.Header
			if name == "" {
				name = fmt.Sprintf("block_%d", orderNum)
			}
			uc.logger.Error("failed to create content block", "error", err, "order_num", orderNum)
			report.AddFailed("content_block", name, err.Error())
			continue
		}
		name := cs.Header
		if name == "" {
			name = fmt.Sprintf("block_%d", orderNum)
		}
		report.AddCreated("content_block", name, cb.ID)
	}
}

// splitParagraphs splits prose by double newlines, trims whitespace, and filters empty entries.
func splitParagraphs(prose string) []string {
	raw := strings.Split(prose, "\n\n")
	result := make([]string, 0, len(raw))
	for _, p := range raw {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// joinParagraphs joins paragraphs from start to end (1-based, inclusive) with double newlines.
func joinParagraphs(paragraphs []string, start, end int) string {
	startIdx := start - 1
	endIdx := end - 1

	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx >= len(paragraphs) {
		endIdx = len(paragraphs) - 1
	}
	if startIdx > endIdx || startIdx >= len(paragraphs) {
		return ""
	}

	return strings.Join(paragraphs[startIdx:endIdx+1], "\n\n")
}

// parseBeatType maps an inferred beat type string to the story.BeatType enum
func parseBeatType(t string) story.BeatType {
	switch strings.ToLower(t) {
	case "setup":
		return story.BeatTypeSetup
	case "turn":
		return story.BeatTypeTurn
	case "reveal":
		return story.BeatTypeReveal
	case "conflict":
		return story.BeatTypeConflict
	case "climax":
		return story.BeatTypeClimax
	case "resolution":
		return story.BeatTypeResolution
	case "hook":
		return story.BeatTypeHook
	case "transition":
		return story.BeatTypeTransition
	default:
		return story.BeatTypeSetup
	}
}

// parseGranularCanon parses individual canon files from the story_canon/ subdirectory
func (uc *ImportChapterUseCase) parseGranularCanon(granularFiles map[string]string) *ParsedCanonFromGranular {
	canon := &ParsedCanonFromGranular{}

	for filename, content := range granularFiles {
		entityType, parsed, err := parser.ParseGranularCanonFile(filename, content)
		if err != nil {
			uc.logger.Warn("failed to parse granular canon file", "filename", filename, "error", err)
			continue
		}

		switch entityType {
		case "character":
			if ch, ok := parsed.(parser.ParsedCharacter); ok {
				canon.Characters = append(canon.Characters, ch)
			}
		case "location":
			if loc, ok := parsed.(parser.ParsedLocation); ok {
				canon.Locations = append(canon.Locations, loc)
			}
		case "artifact":
			if art, ok := parsed.(parser.ParsedArtifact); ok {
				canon.Artifacts = append(canon.Artifacts, art)
			}
		case "relationship":
			if rel, ok := parsed.(parser.ParsedRelationship); ok {
				canon.Relationships = append(canon.Relationships, rel)
			}
		case "premise":
			if p, ok := parsed.(string); ok {
				canon.Premise = p
			}
		case "themes":
			if t, ok := parsed.([]string); ok {
				canon.Themes = t
			}
		case "world_rules":
			if wr, ok := parsed.([]string); ok {
				canon.WorldRules = wr
			}
		}
	}

	return canon
}

// ParsedCanonFromGranular mirrors parser.ParsedCanon but is built from granular files
type ParsedCanonFromGranular = parser.ParsedCanon

// readExportFolder reads all expected files from the export folder
func readExportFolder(folderPath string) (*exportFiles, error) {
	files := &exportFiles{
		granularFiles: make(map[string]string),
	}

	tabsIndex, err := os.ReadFile(filepath.Join(folderPath, "tabs_index.txt"))
	if err != nil {
		return nil, fmt.Errorf("failed to read tabs_index.txt: %w", err)
	}
	files.tabsIndex = string(tabsIndex)

	storyCanon, err := os.ReadFile(filepath.Join(folderPath, "01_story_canon.md"))
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read 01_story_canon.md: %w", err)
	}
	if err == nil {
		files.storyCanon = string(storyCanon)
	}

	chapterLog, err := os.ReadFile(filepath.Join(folderPath, "02_chapter_log.md"))
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read 02_chapter_log.md: %w", err)
	}
	if err == nil {
		files.chapterLog = string(chapterLog)
	}

	content, err := os.ReadFile(filepath.Join(folderPath, "03_content.md"))
	if err == nil {
		files.content = string(content)
	}

	notes, err := os.ReadFile(filepath.Join(folderPath, "04_notes.md"))
	if err == nil {
		files.notes = string(notes)
	}

	granularDir := filepath.Join(folderPath, "story_canon")
	info, err := os.Stat(granularDir)
	if err == nil && info.IsDir() {
		files.granularDir = granularDir
		entries, err := os.ReadDir(granularDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read story_canon/ directory: %w", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(granularDir, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("failed to read granular file %s: %w", entry.Name(), err)
			}
			files.granularFiles[entry.Name()] = string(data)
		}
	}

	return files, nil
}

// findCharacterByName finds a character by case-insensitive name match
func findCharacterByName(characters []*world.Character, name string) *world.Character {
	for _, c := range characters {
		if strings.EqualFold(c.Name, name) {
			return c
		}
	}
	return nil
}

// findLocationByName finds a location by case-insensitive name match
func findLocationByName(locations []*world.Location, name string) *world.Location {
	for _, l := range locations {
		if strings.EqualFold(l.Name, name) {
			return l
		}
	}
	return nil
}

// findArtifactByName finds an artifact by case-insensitive name match
func findArtifactByName(artifacts []*world.Artifact, name string) *world.Artifact {
	for _, a := range artifacts {
		if strings.EqualFold(a.Name, name) {
			return a
		}
	}
	return nil
}

// resolveCharacterID looks up a character ID by name in the name-to-ID map.
func resolveCharacterID(nameToID map[string]uuid.UUID, name string) (uuid.UUID, bool) {
	if id, ok := nameToID[strings.ToLower(name)]; ok {
		return id, true
	}

	lowerName := strings.ToLower(name)
	for storedName, id := range nameToID {
		if strings.Contains(storedName, lowerName) {
			return id, true
		}
		if strings.Contains(lowerName, storedName) {
			return id, true
		}
		parts := strings.SplitN(storedName, ", ", 2)
		if len(parts) == 2 {
			if strings.EqualFold(parts[1], name) || strings.EqualFold(parts[0], name) {
				return id, true
			}
		}
	}

	return uuid.UUID{}, false
}
