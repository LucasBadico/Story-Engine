import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import type { App, TFile, Vault, MetadataCache } from "obsidian";
import type { SyncContext } from "../../types/sync";
import { detectEntityMentions, resolveEntityMention, type DetectedEntityMention } from "../detectEntityMentions";

describe("detectEntityMentions", () => {
	it("detects official format link without label", () => {
		const content = "Check out [[worlds/eldoria/characters/aria-moon]] for more info.";
		const mentions = detectEntityMentions(content);
		
		expect(mentions).toHaveLength(1);
		expect(mentions[0]).toEqual({
			linkText: "[[worlds/eldoria/characters/aria-moon]]",
			filenamePath: "worlds/eldoria/characters/aria-moon",
			format: "official",
		});
	});

	it("detects official format link with label", () => {
		const content = "Check out [[worlds/eldoria/characters/aria-moon|Aria Moon]] for more info.";
		const mentions = detectEntityMentions(content);
		
		expect(mentions).toHaveLength(1);
		expect(mentions[0]).toEqual({
			linkText: "[[worlds/eldoria/characters/aria-moon|Aria Moon]]",
			filenamePath: "worlds/eldoria/characters/aria-moon",
			displayLabel: "Aria Moon",
			format: "official",
		});
	});

	it("detects Obsidian format link without label", () => {
		const content = "Check out [[aria-moon]] for more info.";
		const mentions = detectEntityMentions(content);
		
		expect(mentions).toHaveLength(1);
		expect(mentions[0]).toEqual({
			linkText: "[[aria-moon]]",
			filenamePath: "aria-moon",
			format: "obsidian",
		});
	});

	it("detects Obsidian format link with label", () => {
		const content = "Check out [[aria-moon|Aria]] for more info.";
		const mentions = detectEntityMentions(content);
		
		expect(mentions).toHaveLength(1);
		expect(mentions[0]).toEqual({
			linkText: "[[aria-moon|Aria]]",
			filenamePath: "aria-moon",
			displayLabel: "Aria",
			format: "obsidian",
		});
	});

	it("detects multiple links in content", () => {
		const content = `
		The hero [[aria-moon|Aria]] traveled to [[worlds/eldoria/locations/crystal-cave|Crystal Cave]].
		There she met [[john-smith|John]] who told her about [[worlds/eldoria/events/the-great-war]].
		`;
		const mentions = detectEntityMentions(content);
		
		expect(mentions).toHaveLength(4);
		expect(mentions[0].filenamePath).toBe("aria-moon");
		expect(mentions[0].format).toBe("obsidian");
		expect(mentions[1].filenamePath).toBe("worlds/eldoria/locations/crystal-cave");
		expect(mentions[1].format).toBe("official");
		expect(mentions[2].filenamePath).toBe("john-smith");
		expect(mentions[2].format).toBe("obsidian");
		expect(mentions[3].filenamePath).toBe("worlds/eldoria/events/the-great-war");
		expect(mentions[3].format).toBe("official");
	});

	it("handles links with nested paths", () => {
		const content = "Check [[worlds/eldoria/characters/_archetypes/warrior]] and [[worlds/eldoria/characters/_traits/bravery]].";
		const mentions = detectEntityMentions(content);
		
		expect(mentions).toHaveLength(2);
		expect(mentions[0].filenamePath).toBe("worlds/eldoria/characters/_archetypes/warrior");
		expect(mentions[0].format).toBe("official");
		expect(mentions[1].filenamePath).toBe("worlds/eldoria/characters/_traits/bravery");
		expect(mentions[1].format).toBe("official");
	});

	it("handles links with special characters in label", () => {
		const content = "Check [[aria-moon|Aria's Story]] and [[john-smith|John & Jane]].";
		const mentions = detectEntityMentions(content);
		
		expect(mentions).toHaveLength(2);
		expect(mentions[0].displayLabel).toBe("Aria's Story");
		expect(mentions[1].displayLabel).toBe("John & Jane");
	});

	it("returns empty array for content without links", () => {
		const content = "This is plain text without any links.";
		const mentions = detectEntityMentions(content);
		
		expect(mentions).toHaveLength(0);
	});

	it("handles malformed links gracefully", () => {
		const content = "This has [[unclosed link and [[another link]] and [not a link].";
		const mentions = detectEntityMentions(content);
		
		// The regex /\[\[([^\]]+)\]\]/g will match the first [[...]] pattern it finds
		// In "[[unclosed link and [[another link]]", it will match:
		// - First [[: captures everything until first ]], resulting in "unclosed link and [[another link"
		// Then it continues from after that ]], so it will also find "another link" as a separate match
		// Actually, let me verify: the regex finds "[[unclosed link and [[another link]]" as one match
		// because [^]]+ means "any character except ]", so it captures "unclosed link and [[another link"
		// Then it continues and finds the remaining "]] and [not a link]" but there's no [[ before the next ]]
		// So actually it only finds one match: "unclosed link and [[another link"
		
		// We should find at least one link (the malformed one)
		expect(mentions.length).toBeGreaterThanOrEqual(1);
		
		// The regex behavior is: it will match the first [[...]] it finds, capturing everything until the first ]]
		// So "[[unclosed link and [[another link]]" will be matched as one link with content "unclosed link and [[another link"
		const firstMatch = mentions[0];
		expect(firstMatch.filenamePath).toContain("unclosed link");
	});
});

describe("resolveEntityMention", () => {
	let mockApp: App;
	let mockVault: Vault;
	let mockMetadataCache: MetadataCache;
	let mockContext: SyncContext;
	let mockFile: TFile;
	let consoleWarnSpy: ReturnType<typeof vi.spyOn>;

	beforeEach(() => {
		vi.clearAllMocks();
		consoleWarnSpy = vi.spyOn(console as any, "warn").mockImplementation(() => {});
		
		mockFile = {
			path: "worlds/eldoria/characters/aria-moon.md",
			name: "aria-moon.md",
			basename: "aria-moon",
			extension: "md",
			stat: {
				ctime: 0,
				mtime: 0,
				size: 100,
			},
		} as TFile;

		mockVault = {
			getAbstractFileByPath: vi.fn(),
			getMarkdownFiles: vi.fn().mockReturnValue([mockFile]),
			read: vi.fn(),
		} as unknown as Vault;

		mockMetadataCache = {
			getFirstLinkpathDest: vi.fn(),
		} as unknown as MetadataCache;

		mockApp = {
			vault: mockVault,
			metadataCache: mockMetadataCache,
		} as unknown as App;

		mockContext = {
			app: mockApp,
			apiClient: {} as any,
			fileManager: {} as any,
			settings: {} as any,
			timestamp: () => "2025-01-01T00:00:00Z",
			backupMode: "snapshots",
		};
	});

	afterEach(() => {
		consoleWarnSpy?.mockRestore();
	});

	it("resolves official format link successfully", async () => {
		const mention: DetectedEntityMention = {
			linkText: "[[worlds/eldoria/characters/aria-moon]]",
			filenamePath: "worlds/eldoria/characters/aria-moon",
			format: "official",
		};

		const fileContent = `---
id: char-123
world_id: world-456
tags:
  - story-engine/character
  - world/eldoria
---

# Aria Moon
`;

		const testFile = {
			path: "worlds/eldoria/characters/aria-moon.md",
			name: "aria-moon.md",
			basename: "aria-moon",
			extension: "md",
			stat: {
				ctime: 0,
				mtime: 0,
				size: 100,
			},
		} as TFile;

		vi.mocked(mockVault.getAbstractFileByPath).mockReturnValue(testFile);
		vi.mocked(mockVault.read).mockResolvedValue(fileContent);

		const result = await resolveEntityMention(mention, mockContext);

		expect(result).not.toBeNull();
		expect(result?.entityId).toBe("char-123");
		expect(result?.entityType).toBe("character");
		expect(result?.worldId).toBe("world-456");
		expect(mockVault.getAbstractFileByPath).toHaveBeenCalledWith("worlds/eldoria/characters/aria-moon.md");
		expect(mockVault.read).toHaveBeenCalledWith(testFile);
	});

	it("resolves Obsidian format link by basename", async () => {
		const mention: DetectedEntityMention = {
			linkText: "[[aria-moon]]",
			filenamePath: "aria-moon",
			format: "obsidian",
		};

		const fileContent = `---
id: char-123
world_id: world-456
tags:
  - story-engine/character
  - world/eldoria
---

# Aria Moon
`;

		vi.mocked(mockVault.getMarkdownFiles).mockReturnValue([mockFile]);
		vi.mocked(mockVault.read).mockResolvedValue(fileContent);

		const result = await resolveEntityMention(mention, mockContext);

		expect(result).not.toBeNull();
		expect(result?.entityId).toBe("char-123");
		expect(result?.entityType).toBe("character");
		expect(result?.worldId).toBe("world-456");
		expect(mockVault.getMarkdownFiles).toHaveBeenCalled();
		expect(mockVault.read).toHaveBeenCalledWith(mockFile);
	});

	it("resolves Obsidian format link via metadataCache", async () => {
		const mention: DetectedEntityMention = {
			linkText: "[[aria-moon]]",
			filenamePath: "aria-moon",
			format: "obsidian",
		};

		const fileContent = `---
id: char-123
tags:
  - story-engine/character
---

# Aria Moon
`;

		const testFile = {
			path: "worlds/eldoria/characters/aria-moon.md",
			name: "aria-moon.md",
			basename: "aria-moon",
			extension: "md",
			stat: {
				ctime: 0,
				mtime: 0,
				size: 100,
			},
		} as TFile;

		// First strategy (basename) fails, so try metadataCache
		vi.mocked(mockVault.getMarkdownFiles).mockReturnValue([]);
		vi.mocked(mockMetadataCache.getFirstLinkpathDest).mockReturnValue(testFile);
		vi.mocked(mockVault.read).mockResolvedValue(fileContent);

		const result = await resolveEntityMention(mention, mockContext);

		expect(result).not.toBeNull();
		expect(result?.entityId).toBe("char-123");
		expect(result?.entityType).toBe("character");
		expect(mockMetadataCache.getFirstLinkpathDest).toHaveBeenCalledWith("aria-moon", "");
		expect(mockVault.read).toHaveBeenCalledWith(testFile);
	});

	it("infers entity type from file path", async () => {
		const locationFile = {
			path: "worlds/eldoria/locations/crystal-cave.md",
			name: "crystal-cave.md",
			basename: "crystal-cave",
			extension: "md",
			stat: {
				ctime: 0,
				mtime: 0,
				size: 100,
			},
		} as TFile;

		const mention: DetectedEntityMention = {
			linkText: "[[worlds/eldoria/locations/crystal-cave]]",
			filenamePath: "worlds/eldoria/locations/crystal-cave",
			format: "official",
		};

		const fileContent = `---
id: loc-789
world_id: world-456
---

# Crystal Cave
`;

		vi.mocked(mockVault.getAbstractFileByPath).mockReturnValue(locationFile);
		vi.mocked(mockVault.read).mockResolvedValue(fileContent);

		const result = await resolveEntityMention(mention, mockContext);

		expect(result).not.toBeNull();
		expect(result?.entityType).toBe("location");
		expect(result?.entityId).toBe("loc-789");
	});

	it("infers entity type from frontmatter tags (priority over path)", async () => {
		const mention: DetectedEntityMention = {
			linkText: "[[aria-moon]]",
			filenamePath: "aria-moon",
			format: "obsidian",
		};

		const fileContent = `---
id: loc-123
tags:
  - story-engine/location
  - world/eldoria
---

# Aria Moon
`;

		// File is in characters folder but tag says location - tag should win
		vi.mocked(mockVault.getMarkdownFiles).mockReturnValue([mockFile]);
		vi.mocked(mockVault.read).mockResolvedValue(fileContent);

		const result = await resolveEntityMention(mention, mockContext);

		expect(result).not.toBeNull();
		expect(result?.entityType).toBe("location");
		expect(result?.entityId).toBe("loc-123");
	});

	it("infers entity type from frontmatter entity_type field (highest priority)", async () => {
		const mention: DetectedEntityMention = {
			linkText: "[[some-entity]]",
			filenamePath: "some-entity",
			format: "obsidian",
		};

		const fileContent = `---
id: entity-123
entity_type: artifact
tags:
  - story-engine/character
  - other-tag
---

# Some Entity
`;

		const otherFile = {
			...mockFile,
			basename: "some-entity",
			name: "some-entity.md",
			path: "worlds/eldoria/characters/some-entity.md", // Path says character
		} as TFile;

		vi.mocked(mockVault.getMarkdownFiles).mockReturnValue([otherFile]);
		vi.mocked(mockVault.read).mockResolvedValue(fileContent);

		const result = await resolveEntityMention(mention, mockContext);

		expect(result).not.toBeNull();
		// entity_type field should win over tags and path
		expect(result?.entityType).toBe("artifact");
		expect(result?.entityId).toBe("entity-123");
	});

	it("returns null when file not found (official format)", async () => {
		const mention: DetectedEntityMention = {
			linkText: "[[worlds/eldoria/characters/non-existent]]",
			filenamePath: "worlds/eldoria/characters/non-existent",
			format: "official",
		};

		vi.mocked(mockVault.getAbstractFileByPath).mockReturnValue(null);

		const result = await resolveEntityMention(mention, mockContext);

		expect(result).toBeNull();
	});

	it("returns null when file not found (Obsidian format)", async () => {
		const mention: DetectedEntityMention = {
			linkText: "[[non-existent]]",
			filenamePath: "non-existent",
			format: "obsidian",
		};

		vi.mocked(mockVault.getMarkdownFiles).mockReturnValue([]);
		vi.mocked(mockMetadataCache.getFirstLinkpathDest).mockReturnValue(null);

		const result = await resolveEntityMention(mention, mockContext);

		expect(result).toBeNull();
	});

	it("returns null when frontmatter has no id", async () => {
		const mention: DetectedEntityMention = {
			linkText: "[[aria-moon]]",
			filenamePath: "aria-moon",
			format: "obsidian",
		};

		const fileContent = `---
tags:
  - story-engine/character
---

# Aria Moon
`;

		vi.mocked(mockVault.getMarkdownFiles).mockReturnValue([mockFile]);
		vi.mocked(mockVault.read).mockResolvedValue(fileContent);

		const result = await resolveEntityMention(mention, mockContext);

		expect(result).toBeNull();
	});

	it("returns null when entity type cannot be inferred", async () => {
		const mention: DetectedEntityMention = {
			linkText: "[[unknown-file]]",
			filenamePath: "unknown-file",
			format: "obsidian",
		};

		const fileContent = `---
id: entity-123
---

# Unknown File
`;

		const otherFile = {
			...mockFile,
			basename: "unknown-file",
			name: "unknown-file.md",
			path: "random/path/unknown-file.md",
		} as TFile;

		vi.mocked(mockVault.getMarkdownFiles).mockReturnValue([otherFile]);
		vi.mocked(mockVault.read).mockResolvedValue(fileContent);

		const result = await resolveEntityMention(mention, mockContext);

		expect(result).toBeNull();
	});

	it("handles files with archetype path pattern", async () => {
		const archetypeFile = {
			path: "worlds/eldoria/characters/_archetypes/warrior.md",
			name: "warrior.md",
			basename: "warrior",
			extension: "md",
			stat: {
				ctime: 0,
				mtime: 0,
				size: 100,
			},
		} as TFile;

		const mention: DetectedEntityMention = {
			linkText: "[[worlds/eldoria/characters/_archetypes/warrior]]",
			filenamePath: "worlds/eldoria/characters/_archetypes/warrior",
			format: "official",
		};

		const fileContent = `---
id: arch-123
tags:
  - story-engine/archetype
---

# Warrior
`;

		vi.mocked(mockVault.getAbstractFileByPath).mockReturnValue(archetypeFile);
		vi.mocked(mockVault.read).mockResolvedValue(fileContent);

		const result = await resolveEntityMention(mention, mockContext);

		expect(result).not.toBeNull();
		expect(result?.entityType).toBe("archetype");
		expect(result?.entityId).toBe("arch-123");
	});

	it("handles files with trait path pattern", async () => {
		const traitFile = {
			path: "worlds/eldoria/characters/_traits/bravery.md",
			name: "bravery.md",
			basename: "bravery",
			extension: "md",
			stat: {
				ctime: 0,
				mtime: 0,
				size: 100,
			},
		} as TFile;

		const mention: DetectedEntityMention = {
			linkText: "[[worlds/eldoria/characters/_traits/bravery]]",
			filenamePath: "worlds/eldoria/characters/_traits/bravery",
			format: "official",
		};

		const fileContent = `---
id: trait-123
category: virtue
tags:
  - story-engine/trait
---

# Bravery
`;

		vi.mocked(mockVault.getAbstractFileByPath).mockReturnValue(traitFile);
		vi.mocked(mockVault.read).mockResolvedValue(fileContent);

		const result = await resolveEntityMention(mention, mockContext);

		expect(result).not.toBeNull();
		expect(result?.entityType).toBe("trait");
		expect(result?.entityId).toBe("trait-123");
	});

	it("handles error gracefully and returns null", async () => {
		const mention: DetectedEntityMention = {
			linkText: "[[aria-moon]]",
			filenamePath: "aria-moon",
			format: "obsidian",
		};

		// Reset mocks first - getMarkdownFiles() is synchronous, so we need to throw an error synchronously
		vi.mocked(mockVault.getMarkdownFiles).mockReset();
		vi.mocked(mockVault.getMarkdownFiles).mockImplementation(() => {
			throw new Error("Vault error");
		});

		const result = await resolveEntityMention(mention, mockContext);

		expect(result).toBeNull();
	});

	it("handles frontmatter with null values (converts to undefined)", async () => {
		const mention: DetectedEntityMention = {
			linkText: "[[aria-moon]]",
			filenamePath: "aria-moon",
			format: "obsidian",
		};

		const fileContent = `---
id: char-123
world_id: null
archetype_id: null
tags:
  - story-engine/character
---

# Aria Moon
`;

		vi.mocked(mockVault.getMarkdownFiles).mockReturnValue([mockFile]);
		vi.mocked(mockVault.read).mockResolvedValue(fileContent);

		const result = await resolveEntityMention(mention, mockContext);

		expect(result).not.toBeNull();
		expect(result?.entityId).toBe("char-123");
		expect(result?.entityType).toBe("character");
		// null values should be converted to undefined for optional fields
		expect(result?.worldId).toBeUndefined();
	});

	it("handles frontmatter with quoted strings", async () => {
		const mention: DetectedEntityMention = {
			linkText: "[[aria-moon]]",
			filenamePath: "aria-moon",
			format: "obsidian",
		};

		const fileContent = `---
id: "char-123"
world_id: 'world-456'
tags:
  - story-engine/character
---

# Aria Moon
`;

		vi.mocked(mockVault.getMarkdownFiles).mockReturnValue([mockFile]);
		vi.mocked(mockVault.read).mockResolvedValue(fileContent);

		const result = await resolveEntityMention(mention, mockContext);

		expect(result).not.toBeNull();
		expect(result?.entityId).toBe("char-123");
		expect(result?.worldId).toBe("world-456");
	});

	it("reads entity ID from custom ID field name", async () => {
		const mention: DetectedEntityMention = {
			linkText: "[[aria-moon]]",
			filenamePath: "aria-moon",
			format: "obsidian",
		};

		const fileContent = `---
story_engine_id: char-789
world_id: world-456
tags:
  - story-engine/character
---

# Aria Moon
`;

		vi.mocked(mockVault.getMarkdownFiles).mockReturnValue([mockFile]);
		vi.mocked(mockVault.read).mockResolvedValue(fileContent);

		// Update context to use custom ID field
		const customContext = {
			...mockContext,
			settings: {
				...mockContext.settings,
				frontmatterIdField: "story_engine_id",
			},
		};

		const result = await resolveEntityMention(mention, customContext);

		expect(result).not.toBeNull();
		expect(result?.entityId).toBe("char-789");
		expect(result?.entityType).toBe("character");
	});

	it("returns null when custom ID field is missing", async () => {
		const mention: DetectedEntityMention = {
			linkText: "[[aria-moon]]",
			filenamePath: "aria-moon",
			format: "obsidian",
		};

		const fileContent = `---
name: Aria Moon
world_id: world-456
tags:
  - story-engine/character
---

# Aria Moon
`;

		vi.mocked(mockVault.getMarkdownFiles).mockReturnValue([mockFile]);
		vi.mocked(mockVault.read).mockResolvedValue(fileContent);

		// Update context to use custom ID field that doesn't exist
		const customContext = {
			...mockContext,
			settings: {
				...mockContext.settings,
				frontmatterIdField: "story_engine_id",
			},
		};

		const result = await resolveEntityMention(mention, customContext);

		expect(result).toBeNull();
	});

	it("uses default 'id' field when custom ID field is undefined", async () => {
		const mention: DetectedEntityMention = {
			linkText: "[[aria-moon]]",
			filenamePath: "aria-moon",
			format: "obsidian",
		};

		const fileContent = `---
id: char-999
world_id: world-456
tags:
  - story-engine/character
---

# Aria Moon
`;

		vi.mocked(mockVault.getMarkdownFiles).mockReturnValue([mockFile]);
		vi.mocked(mockVault.read).mockResolvedValue(fileContent);

		// Context with undefined frontmatterIdField (defaults to "id")
		const defaultContext = {
			...mockContext,
			settings: {
				...mockContext.settings,
				frontmatterIdField: undefined,
			},
		};

		const result = await resolveEntityMention(mention, defaultContext);

		expect(result).not.toBeNull();
		expect(result?.entityId).toBe("char-999");
	});
});

