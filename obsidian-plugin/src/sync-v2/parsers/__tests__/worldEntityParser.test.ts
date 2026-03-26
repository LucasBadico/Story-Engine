import { describe, expect, it } from "vitest";
import { parseWorldEntityFile } from "../worldEntityParser";

describe("parseWorldEntityFile", () => {
	it("parses character file with description", () => {
		const content = `---
id: char-123
world_id: world-456
---

# John Smith

## Description
A brave warrior from the north.

## Metadata
- Tenant: tenant-1
`;
		const result = parseWorldEntityFile(content);
		expect(result.id).toBe("char-123");
		expect(result.name).toBe("John Smith");
		expect(result.description).toBe("A brave warrior from the north.");
	});

	it("returns null description for placeholder text", () => {
		const content = `---
id: char-123
---

# Jane Doe

## Description
_No description yet._
`;
		const result = parseWorldEntityFile(content);
		expect(result.description).toBeNull();
	});
});
