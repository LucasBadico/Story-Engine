package parser

import (
	"strings"
	"testing"
)

const fixtureChapterLog = `<!-- generated: split from Google Doc export -->
<!-- source document: 1_0_arc_build_up_prologue -->
<!-- tab path: Chapter Log -->

# Chapter Log

# Steps to Surrender — Prologue: Kneeling

## Character States

### Akiba, Haruka \- protagonist

**Emotion**: stunned, grieving the death of the bride-self, trembling with shame-arousal cocktail, final threads of resistance snapping
**Goal**: obey the seven instructions / reach the moment of kneeling and waiting
**Secrets**: she has already been communicating with M-sama for months; this night is the physical culmination of a long, secret psychological courtship
**Clothes**: arrives in tight black silk dress → strips completely → puts on her own wedding-night white lace bra & panties (sheer, barely-there) \+ tiny frilled white honeymoon apron → ends the scene collared, kneeling, in nothing else
**Physical**: trembling, sweating, soaked through lace

### Tsubana, Shouta \- childhood friend

**State**: not present in this chapter

### M-sama

**Emotion**: calm, absolute control, velvet menace
**State**: physically arrives at the very end; voice only until the final line
**Known to Haruka as**: the one who already owns her thoughts; tonight he claims her body

### Akiba, Tetsuo \- husband

**State**: absent. Not mentioned as actively present in her thoughts during the instructions; only remembered in contrast (the man who once carried her over this threshold)
**Emotion**: already cold / irrelevant to this night

## Relationship States

### Haruka ↔ Tetsuo

Marriage emotionally cold / functionally over. She is here alone because the husband no longer claims the honeymoon suite or her body.
Tetsuo unaware of Haruka's descent (and his unawareness is the sharpest knife).

### Haruka ↔ Shouta

Childhood trust. Shouta secretly worried about Haruka.

### Haruka ↔ M-sama

Full surrender. Extreme power imbalance. She has been communicating with him for months — this is the physical culmination.

## Scenes

### 1\. Elevator ascent

Establishing collision of past and present. Mirrored walls. Black silk dress. Small black box.

### 2\. Entering the unchanged suite

Crimson carpet. Vanilla-rose scent. Sensory memory trigger.

### 3\. Opening the box

Black leather collar on top. Wedding-night lingerie beneath. Devastating symbolic violence.

### 4\. Grief and arousal war

Internal fracture moment. "No… not these… not like this…"

### 5\. Reading the seven instructions

Close curtains. Strip. Put on honeymoon clothes. Turn off lights. Open door a crack. Kneel holding collar. Wait.

### 6\. Kneeling and waiting

Center of carpet. Thighs parted. Collar held as offering. Calves burn. Thoughts race.

### 7\. Arrival and surrender

Soft footsteps. "Good girl." Collar locked. Day Zero begins.

## Secrets

### M-sama

True identity unknown.

### Haruka

Has not told anyone about the messages or about M-sama.

### Tetsuo

Financial problems hidden from Haruka.

## Active Conflicts

### Haruka vs herself

Last shred of "hopeful wife" identity vs emerging need to be owned / controlled / named. Already lost — she chose to come here.

## Plot Threads

### Haruka's surrender — main — open

Haruka returns alone to the honeymoon suite to submit to M-sama's ritual. The bride identity dies.
**Entities**: subject: character: Haruka, catalyst: character: M-sama, location: location: Rouge Seasons Hotel

### The collar ritual — secondary — open

Seven ritual instructions leading to collaring.
**Entities**: subject: character: Haruka, artifact: artifact: slave collar

### Marriage dissolution — main — open

Tetsuo exists only as a memory contrast. The marriage is functionally over.
**Entities**: subject: character: Haruka, subject: character: Tetsuo

## Timeline

Day 0: Sunday late evening – Rouge Seasons Hotel – physical surrender / collaring

Timeline starts at the prologue. Day 0\. The first 63 days of the story are negative in the counting (the story goes back to tell how she got here).

## Chapter Locations

### Hotel Rouge Seasons

* Elevator – main lobby to 25th floor (mirrored walls)
* Honeymoon Suite – 25th floor (crimson carpet, king-size bed with posts, heavy velvet curtains, vanilla-rose scent)

## Chapter Artifacts

* Small black gift box wrapped with thin red ribbon
* White note tucked under ribbon: "Open when in your Honeymoon Room"
* Black leather slave collar (thick, sturdy, silver O-ring)
* Haruka's honeymoon treasure: white lace bra & panties (wedding night set) \+ tiny frilled white honeymoon apron
* Thin crimson lace ribbon on right wrist
* Phone – not present / not used during the instructions (this night is analog / ritual)

## Phone Messages \- recent history

None — this night is analog / ritual.

## Arc Context

# Arc 1: Build up \+ Prologue

The opening movement of the book (Prologue through Chapter 6\) functions as **the seduction-before-the-fall arc** — or more precisely, **the long, quiet implosion of the perfect-wife fantasy before the five-day experiment even begins**.

It is **not** an origin story in the conventional sense.

It is the portrait of a woman who has already been quietly losing ground for months, even years, and who only now — in the mirror of other people's eyes or the silence of her own home — begins to notice how wide the cracks have become.

### Prologue: Kneeling

On the night that should have belonged to memory and nostalgia, Haruka returns alone to the Rouge Seasons honeymoon suite, opens a black gift box containing her own wedding-night lingerie topped with a heavy slave collar, follows seven ritual instructions in mounting shame and arousal, kneels in darkness holding the collar like an offering, and surrenders completely the instant an unseen man steps through the cracked door and says "Good girl."

That is the prologue in nine clean beats: memory → violation → ritual → exposure → waiting → naming → collapse.
`

func TestParseChapterLog(t *testing.T) {
	result, err := ParseChapterLog(fixtureChapterLog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Character States
	if len(result.CharacterStates) != 4 {
		t.Fatalf("len(CharacterStates) = %d, want 4", len(result.CharacterStates))
	}

	charNames := []string{"Haruka", "Shouta", "M-sama", "Tetsuo"}
	for _, expected := range charNames {
		found := false
		for _, cs := range result.CharacterStates {
			if strings.Contains(cs.Name, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected character state containing %q not found", expected)
		}
	}

	// Haruka's state has Emotion, Goal, Secrets, Clothes, Physical fields
	for _, cs := range result.CharacterStates {
		if strings.Contains(cs.Name, "Haruka") {
			requiredFields := []string{"Emotion", "Goal", "Secrets", "Clothes", "Physical"}
			for _, field := range requiredFields {
				if _, ok := cs.Fields[field]; !ok {
					t.Errorf("Haruka state missing field %q, has: %v", field, cs.Fields)
				}
			}
			break
		}
	}

	// Relationship States
	if len(result.RelationshipStates) != 3 {
		t.Fatalf("len(RelationshipStates) = %d, want 3", len(result.RelationshipStates))
	}

	relPairs := [][2]string{
		{"Haruka", "Tetsuo"},
		{"Haruka", "Shouta"},
		{"Haruka", "M-sama"},
	}
	for _, pair := range relPairs {
		found := false
		for _, rs := range result.RelationshipStates {
			if strings.Contains(rs.CharacterA, pair[0]) && strings.Contains(rs.CharacterB, pair[1]) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected relationship state %s <-> %s not found", pair[0], pair[1])
		}
	}

	// Scenes
	if len(result.Scenes) != 7 {
		t.Fatalf("len(Scenes) = %d, want 7", len(result.Scenes))
	}
	for i, scene := range result.Scenes {
		expectedNum := i + 1
		if scene.Number != expectedNum {
			t.Errorf("Scene[%d].Number = %d, want %d", i, scene.Number, expectedNum)
		}
	}

	// Secrets
	if len(result.Secrets) != 3 {
		t.Fatalf("len(Secrets) = %d, want 3", len(result.Secrets))
	}
	secretNames := []string{"M-sama", "Haruka", "Tetsuo"}
	for i, expected := range secretNames {
		if !strings.Contains(result.Secrets[i].EntityName, expected) {
			t.Errorf("Secrets[%d].EntityName = %q, want containing %q", i, result.Secrets[i].EntityName, expected)
		}
	}

	// Active Conflicts
	if len(result.ActiveConflicts) != 1 {
		t.Fatalf("len(ActiveConflicts) = %d, want 1", len(result.ActiveConflicts))
	}

	// Timeline
	if result.Timeline == nil {
		t.Fatal("Timeline is nil")
	}
	if result.Timeline.Day == nil {
		t.Fatal("Timeline.Day is nil")
	}
	if *result.Timeline.Day != 0 {
		t.Errorf("Timeline.Day = %d, want 0", *result.Timeline.Day)
	}
	if result.Timeline.Confidence != "explicit" {
		t.Errorf("Timeline.Confidence = %q, want %q", result.Timeline.Confidence, "explicit")
	}

	// Plot Threads
	if len(result.PlotThreads) != 3 {
		t.Fatalf("len(PlotThreads) = %d, want 3", len(result.PlotThreads))
	}

	// First plot thread
	pt0 := result.PlotThreads[0]
	if pt0.Title != "Haruka's surrender" {
		t.Errorf("PlotThreads[0].Title = %q, want %q", pt0.Title, "Haruka's surrender")
	}
	if pt0.Type != "main" {
		t.Errorf("PlotThreads[0].Type = %q, want %q", pt0.Type, "main")
	}
	if pt0.Action != "open" {
		t.Errorf("PlotThreads[0].Action = %q, want %q", pt0.Action, "open")
	}
	if !strings.Contains(pt0.Description, "honeymoon suite") {
		t.Errorf("PlotThreads[0].Description should contain 'honeymoon suite', got %q", pt0.Description)
	}
	if len(pt0.Entities) < 2 {
		t.Fatalf("PlotThreads[0].Entities len = %d, want >= 2", len(pt0.Entities))
	}
	// Check first entity: subject: character: Haruka
	if pt0.Entities[0].Role != "subject" {
		t.Errorf("PlotThreads[0].Entities[0].Role = %q, want %q", pt0.Entities[0].Role, "subject")
	}
	if pt0.Entities[0].EntityType != "character" {
		t.Errorf("PlotThreads[0].Entities[0].EntityType = %q, want %q", pt0.Entities[0].EntityType, "character")
	}
	if pt0.Entities[0].Name != "Haruka" {
		t.Errorf("PlotThreads[0].Entities[0].Name = %q, want %q", pt0.Entities[0].Name, "Haruka")
	}

	// Second plot thread: secondary type
	if result.PlotThreads[1].Type != "secondary" {
		t.Errorf("PlotThreads[1].Type = %q, want %q", result.PlotThreads[1].Type, "secondary")
	}

	// Third plot thread: main type
	if result.PlotThreads[2].Type != "main" {
		t.Errorf("PlotThreads[2].Type = %q, want %q", result.PlotThreads[2].Type, "main")
	}

	// Arc Context
	if result.ArcContext == "" {
		t.Error("ArcContext is empty")
	}
}
