package parser

// TabsIndex holds export metadata
type TabsIndex struct {
	DocumentName string
	DocumentID   string
	ExportedAt   string
	Tabs         []string
}

// ParsedCanon holds all parsed story canon data
type ParsedCanon struct {
	Premise       string
	Themes        []string
	WorldRules    []string
	Characters    []ParsedCharacter
	Relationships []ParsedRelationship
	Locations     []ParsedLocation
	Artifacts     []ParsedArtifact
}

// ParsedCharacter from canon
type ParsedCharacter struct {
	Name        string // "Akiba, Haruka"
	Role        string // "protagonist"
	Age         string // "25 years" or ""
	Description string // full text body
}

// ParsedRelationship from canon
type ParsedRelationship struct {
	CharacterA  string // "Haruka"
	CharacterB  string // "Tetsuo"
	Description string // baseline description
}

// ParsedLocation from canon
type ParsedLocation struct {
	Name        string
	Description string // full text including sub-locations
}

// ParsedArtifact from canon
type ParsedArtifact struct {
	Name        string
	Description string
}

// ParsedChapterLog holds all chapter log data
type ParsedChapterLog struct {
	ChapterTitle       string
	CharacterStates    []ParsedCharacterState
	RelationshipStates []ParsedRelationshipState
	Scenes             []ParsedScene
	Secrets            []ParsedSecret
	ActiveConflicts    []ParsedConflict
	PlotThreads        []ParsedPlotThread
	Timeline           *ParsedTimeline
	ChapterLocations   []ParsedLocation
	ChapterArtifacts   []string
	ArcContext         string
}

// ParsedCharacterState from chapter log
type ParsedCharacterState struct {
	Name   string            // character name from ### header
	Fields map[string]string // "Emotion" -> "stunned...", "Goal" -> "obey..."
}

// ParsedRelationshipState from chapter log
type ParsedRelationshipState struct {
	CharacterA  string
	CharacterB  string
	Description string
}

// ParsedScene from chapter log
type ParsedScene struct {
	Number      int
	Title       string
	Description string
}

// ParsedSecret from chapter log
type ParsedSecret struct {
	EntityName string
	Secret     string
}

// ParsedConflict from chapter log
type ParsedConflict struct {
	Title       string
	Description string
}

// ParsedPlotThread from chapter log
type ParsedPlotThread struct {
	Title       string             // "Haruka's surrender"
	Type        string             // "main", "secondary", "minor"
	Action      string             // "open", "develop", "escalate", "close"
	Description string             // what happens
	Entities    []ParsedEntityLink // linked entities
}

// ParsedEntityLink represents a polymorphic entity reference
type ParsedEntityLink struct {
	Role       string // "subject", "catalyst", "location", "artifact"
	EntityType string // "character", "location", "artifact", "plot"
	Name       string // entity name for matching
}

// ParsedTimeline from chapter log
type ParsedTimeline struct {
	Day        *int   // nil if unparseable
	Confidence string // "explicit" when parsed from text
	RawText    string // full text of timeline section
}

// ParsedNote from notes tab
type ParsedNote struct {
	Date string
	Text string
}

// ParsedContentScene from content tab
type ParsedContentScene struct {
	Header  string // scene header text
	Content string // prose text
}
