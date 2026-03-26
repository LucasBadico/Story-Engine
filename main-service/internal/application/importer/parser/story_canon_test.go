package parser

import (
	"strings"
	"testing"
)

const fixtureStoryCanon = `<!-- generated: split from Google Doc export -->
<!-- source document: 1_0_arc_build_up_prologue -->
<!-- tab path: Story Canon -->

# Story Canon

# Steps to Surrender

## Haruka's Destiny series \- Book 1

## Premise

Haruka Akiba, 25, newlywed, once fiercely independent, now full-time housewife in a luxurious but emotionally hollow duplex, begins to fracture under the triple weight of:

* a rapidly cooling marriage
* the reappearance of a warm, attentive childhood friend (Shouta)
* the appearance of a faceless, piercingly perceptive dominant voice (M-sama) who seems to know exactly which parts of her are already broken

What begins as curiosity about her husband's increasing distance slowly mutates into a five-day "experiment" of escalating羞恥 (shame) tasks that Haruka tells herself are purely informational — a means to understand why Tetsuo keeps disappearing emotionally and physically.

She is wrong.

Every task is designed to make her name, feel, and eventually crave the very identity she most denies: a woman who is addicted to being *seen* in her weakness, especially while remaining completely invisible to the man she married.

The story is **not** primarily about sex.

It is about **identity fracture through shame as the central addictive mechanism**, and about how a person can be slowly trained to prefer the pain of being truthfully named over the comfort of being politely ignored.

## Themes

* Shame as the strongest drug (stronger than pain, stronger than pleasure)
* Visibility vs Invisibility dialectic (hyper-visible to M-sama / completely invisible to husband & normal life)
* The unbearable lightness of being "good wife" when no one is watching
* Praise / degradation polarity kept deliberately unstable ("good girl" always arrives dangerous)
* Childhood nickname used in adult degrading context \= nuclear shame trigger
* Slow corrosion of self-narrative via micro-betrayals & looping self-justification
* Contrast between three gravitational male poles: cold absence (Tetsuo), warm safety (Shouta), burning-cold truth (M-sama)
* The fantasy that " one more task / one more answer" will allow her to stop

## World Rules

1. The experiment structure is sacred (started \~chapter 8). Daily rules must be clear, escalating, and tied to punishment economy.
2. "Answers about Tetsuo" remain the official, stated goal Haruka clings to.
3. No sudden personality transplant. Haruka stays conflicted, self-justifying, comparison-making until very late. Body can obey long before mind stops negotiating.
4. Shame density must remain high. Pleasure without shame component feels narratively flat.
5. M-sama never appears physically before major psychological preparation. Mystery preserved.
6. Shouta and Tetsuo never directly witness the marks / tasks / phone until story logic demands maximum fracture.
7. Physical descriptions of Haruka's body appear sparsely (≈1 full catalog every 2–3 chapters), always in high-shame / high-visibility moments, never gratuitous.
8. Mirrors are almost mandatory in suite scenes — they multiply shame geometrically.

## Characters

### Akiba, Haruka \- protagonist \- 25 years

Main protagonist. Once proudly self-reliant after losing parents young. Now lives inside increasingly brittle "perfect wife" performance.
Core fracture \= she still believes she can fix the marriage if she just polishes hard enough.

### Tsubana, Shouta \- childhood friend \- 20 years

Warm gravitational pole. Childhood friend, Sayaka's younger brother.
Tall, steady, observant, kind without being weak. Remembers small things. Uses "Harukie" naturally.

### M-sama

Burning-cold gravitational pole.
Voice only. Sees everything Haruka tries to hide from herself. Uses childhood nickname \+ surgically precise degradation.

Never purely cruel; always mixes cold analysis with occasional velvet praise — keeps the carrot perpetually poisoned.

### Akiba, Tetsuo \- husband \- late 20s–early 30s

Cold gravitational pole.
Successful, busy, emotionally absent. Does not notice ink, scent, trembling, tears, new dresses, honeymoon apron attempts — nothing.

Function \= original wound / living proof that "good wife" performance is not being received.

### Tsubana, Sayaka \- childhood best friend \- 22 years

Childhood best friend. Haruka's cheerleader / external validator of "couple goals" narrative.

## Relationships

### Haruka ↔ Tetsuo

Married. He carried her over the honeymoon threshold four months ago.

### Haruka ↔ Shouta

Childhood trust. Shouta is Sayaka's younger brother. He uses "Harukie" naturally.

### Haruka ↔ M-sama

Faceless dominant voice. Communicates via phone messages. Psychological leash via shame \+ promise of truth.

### Haruka ↔ Sayaka

Childhood best friends. Sayaka is Shouta's older sister.

## Locations

### Rouge Seasons Hotel – Honeymoon Suite (25th floor)

Original site of Day Zero. Re-visited in flash-forward prologue.

* Symbol \= fall from bride to slave.

### Akiba Duplex (main stage)

* Open kitchen \+ living room (performance space)
  * Living-room balcony (rare air, occasional Shouta scenes)
* Upstairs
  * Master bedroom (+ master balcony) → Tetsuo's cold absence
  * Suite 1 – dark / black room → late-night desperation space
  * Suite 2 – fully mirrored room → ultimate shame-multiplication chamber
    * Haruka chose it as private room after being left alone for 2 weeks

## Artifacts

### Mirrors

**(especially Suite 2\)** Self-confrontation / shame multiplication / inescapable visibility to self

### Phone messages from M-sama

Invisible collar / proof of being seen when no one else looks

### Vanilla-rose diffuser scent

False domestic comfort / slowly turning cloying

### Childhood nickname "Harukie"

Most reliable shame trigger when used by M-sama

### Wedding-night lace set \+ frilled honeymoon apron

Corrupted purity symbols (appear in punishment / regression scenes)

### King-size bed (empty side)

Monument to abandonment

### Thin crimson lace ribbon

**(≈20 cm satin, two-fingers wide, grosgrain texture) \- uses in prologue \- will show up in chapter 8**
Self-applied early marker / voluntary first "collar"
Worn around right wrist starting in prologue
Symbol of pre-emptive surrender: she tied it on herself before the ritual even began
Recurring tactile reminder (trailing ends brushing inner wrist, warming to her pulse)
`

func TestParseStoryCanon(t *testing.T) {
	result, err := ParseStoryCanon(fixtureStoryCanon)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Premise
	if result.Premise == "" {
		t.Fatal("Premise is empty")
	}
	if !strings.Contains(result.Premise, "Haruka Akiba") {
		t.Errorf("Premise should contain 'Haruka Akiba', got: %s", result.Premise[:100])
	}

	// Themes
	if len(result.Themes) != 8 {
		t.Errorf("len(Themes) = %d, want 8", len(result.Themes))
	}

	// World Rules
	if len(result.WorldRules) != 8 {
		t.Errorf("len(WorldRules) = %d, want 8", len(result.WorldRules))
	}

	// Characters
	if len(result.Characters) != 5 {
		t.Fatalf("len(Characters) = %d, want 5", len(result.Characters))
	}

	expectedNames := []string{"Haruka", "Shouta", "M-sama", "Tetsuo", "Sayaka"}
	for _, expected := range expectedNames {
		found := false
		for _, ch := range result.Characters {
			if strings.Contains(ch.Name, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected character name containing %q not found", expected)
		}
	}

	// Character roles
	expectedRoles := map[string]string{
		"Haruka": "protagonist",
		"Shouta": "childhood friend",
	}
	for nameSubstr, expectedRole := range expectedRoles {
		for _, ch := range result.Characters {
			if strings.Contains(ch.Name, nameSubstr) {
				if ch.Role != expectedRole {
					t.Errorf("character %q: Role = %q, want %q", ch.Name, ch.Role, expectedRole)
				}
				break
			}
		}
	}

	// Haruka age
	for _, ch := range result.Characters {
		if strings.Contains(ch.Name, "Haruka") {
			if ch.Age != "25 years" {
				t.Errorf("Haruka Age = %q, want %q", ch.Age, "25 years")
			}
			break
		}
	}

	// Relationships
	if len(result.Relationships) != 4 {
		t.Fatalf("len(Relationships) = %d, want 4", len(result.Relationships))
	}

	relPairs := [][2]string{
		{"Haruka", "Tetsuo"},
		{"Haruka", "Shouta"},
		{"Haruka", "M-sama"},
		{"Haruka", "Sayaka"},
	}
	for _, pair := range relPairs {
		found := false
		for _, rel := range result.Relationships {
			if strings.Contains(rel.CharacterA, pair[0]) && strings.Contains(rel.CharacterB, pair[1]) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected relationship %s <-> %s not found", pair[0], pair[1])
		}
	}

	// Locations
	if len(result.Locations) != 2 {
		t.Fatalf("len(Locations) = %d, want 2", len(result.Locations))
	}

	locationNames := []string{"Rouge Seasons Hotel", "Akiba Duplex"}
	for _, expected := range locationNames {
		found := false
		for _, loc := range result.Locations {
			if strings.Contains(loc.Name, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected location containing %q not found", expected)
		}
	}

	// Artifacts
	if len(result.Artifacts) != 7 {
		t.Fatalf("len(Artifacts) = %d, want 7", len(result.Artifacts))
	}

	artifactNames := []string{"Mirrors", "Thin crimson lace ribbon"}
	for _, expected := range artifactNames {
		found := false
		for _, art := range result.Artifacts {
			if strings.Contains(art.Name, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected artifact containing %q not found", expected)
		}
	}
}
