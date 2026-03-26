package parser

import (
	"strings"
	"testing"
)

const fixtureGranularCharacter = `# Steps to Surrender
## Haruka's Destiny series \- Book 1
## Characters
### Akiba, Haruka \- protagonist \- 25 years

Main protagonist. Once proudly self-reliant after losing parents young. Now lives inside increasingly brittle "perfect wife" performance.
Core fracture \= she still believes she can fix the marriage if she just polishes hard enough.
`

const fixtureGranularRelationship = `# Steps to Surrender
## Haruka's Destiny series \- Book 1
## Relationships
### Haruka ↔ Tetsuo

Married. He carried her over the honeymoon threshold four months ago.
`

const fixtureGranularArtifact = `# Steps to Surrender
## Haruka's Destiny series \- Book 1
## Artifacts
### Mirrors

**(especially Suite 2\)** Self-confrontation / shame multiplication / inescapable visibility to self
`

func TestParseGranularCanonFile_Character(t *testing.T) {
	entityType, parsed, err := ParseGranularCanonFile("04_character_akiba_haruka.md", fixtureGranularCharacter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entityType != "character" {
		t.Errorf("entityType = %q, want %q", entityType, "character")
	}

	ch, ok := parsed.(ParsedCharacter)
	if !ok {
		t.Fatalf("parsed is %T, want ParsedCharacter", parsed)
	}

	if !strings.Contains(ch.Name, "Haruka") {
		t.Errorf("Name = %q, want containing %q", ch.Name, "Haruka")
	}
	if ch.Role != "protagonist" {
		t.Errorf("Role = %q, want %q", ch.Role, "protagonist")
	}
	if ch.Age != "25 years" {
		t.Errorf("Age = %q, want %q", ch.Age, "25 years")
	}
	if ch.Description == "" {
		t.Error("Description is empty")
	}
}

func TestParseGranularCanonFile_Relationship(t *testing.T) {
	entityType, parsed, err := ParseGranularCanonFile("11_relationship_haruka_tetsuo.md", fixtureGranularRelationship)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entityType != "relationship" {
		t.Errorf("entityType = %q, want %q", entityType, "relationship")
	}

	rel, ok := parsed.(ParsedRelationship)
	if !ok {
		t.Fatalf("parsed is %T, want ParsedRelationship", parsed)
	}

	if !strings.Contains(rel.CharacterA, "Haruka") {
		t.Errorf("CharacterA = %q, want containing %q", rel.CharacterA, "Haruka")
	}
	if !strings.Contains(rel.CharacterB, "Tetsuo") {
		t.Errorf("CharacterB = %q, want containing %q", rel.CharacterB, "Tetsuo")
	}
	if rel.Description == "" {
		t.Error("Description is empty")
	}
}

func TestParseGranularCanonFile_Artifact(t *testing.T) {
	entityType, parsed, err := ParseGranularCanonFile("15_artifact_mirrors.md", fixtureGranularArtifact)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entityType != "artifact" {
		t.Errorf("entityType = %q, want %q", entityType, "artifact")
	}

	art, ok := parsed.(ParsedArtifact)
	if !ok {
		t.Fatalf("parsed is %T, want ParsedArtifact", parsed)
	}

	if art.Name != "Mirrors" {
		t.Errorf("Name = %q, want %q", art.Name, "Mirrors")
	}
	if art.Description == "" {
		t.Error("Description is empty")
	}
}
