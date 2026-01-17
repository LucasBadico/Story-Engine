import { describe, expect, it } from "vitest";
import { FrontmatterGenerator } from "../FrontmatterGenerator";

describe("FrontmatterGenerator", () => {
	it("generates basic frontmatter with fields only", () => {
		const generator = new FrontmatterGenerator();
		const result = generator.generate({
			id: "test-123",
			name: "Test Entity",
			created_at: "2025-01-01T00:00:00Z",
		});

		expect(result).toContain("---");
		expect(result).toContain("id: test-123");
		expect(result).toContain("name: Test Entity"); // Strings simples não têm aspas
		expect(result).toContain('created_at: "2025-01-01T00:00:00Z"'); // Strings com `:` têm aspas
		expect(result).toContain("---");
	});

	it("generates frontmatter with tags for entity type", () => {
		const generator = new FrontmatterGenerator();
		const result = generator.generate(
			{
				id: "story-1",
				title: "Test Story",
			},
			undefined,
			{
				entityType: "story",
			}
		);

		expect(result).toContain("tags:");
		expect(result).toContain("  - story-engine/story");
	});

	it("generates frontmatter with story and world tags", () => {
		const generator = new FrontmatterGenerator();
		const result = generator.generate(
			{
				id: "character-1",
				name: "Aria Moon",
			},
			undefined,
			{
				entityType: "character",
				storyName: "The Great Adventure",
				worldName: "Eldoria",
			}
		);

		expect(result).toContain("tags:");
		expect(result).toContain("  - story-engine/character");
		expect(result).toContain("  - story/the-great-adventure");
		expect(result).toContain("  - world/eldoria");
	});

	it("generates frontmatter with date tag", () => {
		const generator = new FrontmatterGenerator();
		const result = generator.generate(
			{
				id: "event-1",
				name: "The Great War",
			},
			undefined,
			{
				entityType: "event",
				date: "2025-01-15T10:30:00Z",
			}
		);

		expect(result).toContain("tags:");
		expect(result).toContain("  - date/2025/01/15");
	});

	it("handles null values correctly", () => {
		const generator = new FrontmatterGenerator();
		const result = generator.generate({
			id: "test-1",
			name: "Test",
			chapter_id: null,
			story_id: "story-1",
		});

		expect(result).toContain("chapter_id: null");
		expect(result).toContain("story_id: story-1");
	});

	it("handles extra fields correctly", () => {
		const generator = new FrontmatterGenerator();
		const result = generator.generate(
			{
				id: "scene-1",
				story_id: "story-1",
			},
			{
				pov_character_id: "char-1",
				location_id: "loc-1",
			},
			{
				entityType: "scene",
			}
		);

		expect(result).toContain("id: scene-1");
		expect(result).toContain("story_id: story-1");
		expect(result).toContain("pov_character_id: char-1");
		expect(result).toContain("location_id: loc-1");
	});

	it("escapes special characters in string values", () => {
		const generator = new FrontmatterGenerator();
		const result = generator.generate({
			id: "test-1",
			title: 'Story with "quotes" and: colons',
			description: "Multi\nline\ndescription",
		});

		expect(result).toContain('title: "Story with \\"quotes\\" and: colons"'); // Aspas e dois pontos requerem aspas
		// Strings com \n são renderizadas entre aspas no YAML (pode ser multi-linha ou com \n)
		expect(result).toContain("description:");
		// Verifica que todas as palavras da descrição estão presentes
		expect(result).toContain("Multi");
		expect(result).toContain("line");
		expect(result).toContain("description");
	});

	it("handles numeric values correctly", () => {
		const generator = new FrontmatterGenerator();
		const result = generator.generate({
			id: "chapter-1",
			number: 1,
			order_num: 5,
		});

		expect(result).toContain("number: 1");
		expect(result).toContain("order_num: 5");
	});

	it("sanitizes story and world names for tags", () => {
		const generator = new FrontmatterGenerator();
		const result = generator.generate(
			{
				id: "test-1",
			},
			undefined,
			{
				entityType: "character",
				storyName: "The Great Adventure!",
				worldName: "Eldoria - Realm of Magic",
			}
		);

		expect(result).toContain("  - story/the-great-adventure");
		expect(result).toContain("  - world/eldoria-realm-of-magic");
	});

	it("uses custom ID field name when provided", () => {
		const generator = new FrontmatterGenerator();
		const result = generator.generate(
			{
				id: "test-123",
				name: "Test Entity",
			},
			undefined,
			{
				entityType: "character",
				idField: "story_engine_id",
			}
		);

		expect(result).toContain("story_engine_id: test-123");
		// Verify that the old 'id' field is not present in the result
		// We check that there's no "id: test-123" as a standalone line
		const lines = result.split("\n");
		const idLine = lines.find((line) => line.trim().startsWith("id:"));
		expect(idLine).toBeUndefined();
		expect(result).toContain("name: Test Entity");
	});

	it("defaults to 'id' field when idField is not provided", () => {
		const generator = new FrontmatterGenerator();
		const result = generator.generate(
			{
				id: "test-456",
				name: "Test Entity",
			},
			undefined,
			{
				entityType: "character",
			}
		);

		expect(result).toContain("id: test-456");
		expect(result).not.toContain("story_engine_id:");
	});

	it("uses 'id' field when idField is 'id' explicitly", () => {
		const generator = new FrontmatterGenerator();
		const result = generator.generate(
			{
				id: "test-789",
				name: "Test Entity",
			},
			undefined,
			{
				entityType: "character",
				idField: "id",
			}
		);

		expect(result).toContain("id: test-789");
	});
});

