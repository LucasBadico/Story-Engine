import { describe, it, expect } from "vitest";
import {
	getIdFieldName,
	getFrontmatterId,
	setFrontmatterId,
	buildFrontmatterFields,
} from "../frontmatterHelpers";

describe("frontmatterHelpers", () => {
	describe("getIdFieldName", () => {
		it("returns 'id' when idField is undefined", () => {
			expect(getIdFieldName(undefined)).toBe("id");
		});

		it("returns 'id' when idField is empty string", () => {
			expect(getIdFieldName("")).toBe("id");
		});

		it("returns custom field name when provided", () => {
			expect(getIdFieldName("story_engine_id")).toBe("story_engine_id");
		});
	});

	describe("getFrontmatterId", () => {
		it("reads ID from 'id' field by default", () => {
			const frontmatter = { id: "test-123" };
			expect(getFrontmatterId(frontmatter)).toBe("test-123");
		});

		it("reads ID from custom field name", () => {
			const frontmatter = { story_engine_id: "test-456" };
			expect(getFrontmatterId(frontmatter, "story_engine_id")).toBe("test-456");
		});

		it("returns undefined when ID field is missing", () => {
			const frontmatter = { name: "test" };
			expect(getFrontmatterId(frontmatter)).toBeUndefined();
		});

		it("returns undefined when ID field is null", () => {
			const frontmatter = { id: null };
			expect(getFrontmatterId(frontmatter)).toBeUndefined();
		});

		it("returns undefined when ID field is empty string", () => {
			const frontmatter = { id: "" };
			expect(getFrontmatterId(frontmatter)).toBeUndefined();
		});

		it("converts number ID to string", () => {
			const frontmatter = { id: 12345 };
			expect(getFrontmatterId(frontmatter)).toBe("12345");
		});

		it("handles ID from custom field with different types", () => {
			const frontmatter = { custom_id: 789 };
			expect(getFrontmatterId(frontmatter, "custom_id")).toBe("789");
		});
	});

	describe("setFrontmatterId", () => {
		it("sets ID in 'id' field by default", () => {
			const frontmatter: Record<string, unknown> = {};
			setFrontmatterId(frontmatter, "test-123");
			expect(frontmatter.id).toBe("test-123");
		});

		it("sets ID in custom field name", () => {
			const frontmatter: Record<string, unknown> = {};
			setFrontmatterId(frontmatter, "test-456", "story_engine_id");
			expect(frontmatter.story_engine_id).toBe("test-456");
			expect(frontmatter.id).toBeUndefined();
		});

		it("overwrites existing ID field", () => {
			const frontmatter: Record<string, unknown> = { id: "old-id" };
			setFrontmatterId(frontmatter, "new-id");
			expect(frontmatter.id).toBe("new-id");
		});
	});

	describe("buildFrontmatterFields", () => {
		it("builds fields with 'id' field by default", () => {
			const fields = buildFrontmatterFields("test-123", { name: "test" });
			expect(fields.id).toBe("test-123");
			expect(fields.name).toBe("test");
		});

		it("builds fields with custom ID field name", () => {
			const fields = buildFrontmatterFields("test-456", { name: "test" }, "story_engine_id");
			expect(fields.story_engine_id).toBe("test-456");
			expect(fields.name).toBe("test");
			expect(fields.id).toBeUndefined();
		});

		it("puts ID field first in object", () => {
			const fields = buildFrontmatterFields("test-789", { name: "test", age: 25 });
			const keys = Object.keys(fields);
			expect(keys[0]).toBe("id");
			expect(fields.id).toBe("test-789");
		});

		it("handles otherFields overriding ID", () => {
			const fields = buildFrontmatterFields("test-123", { id: "other-id" });
			// ID from buildFrontmatterFields should come first, but otherFields can override
			// Actually, we spread otherFields after, so it should override
			expect(fields.id).toBe("other-id");
		});
	});
});

