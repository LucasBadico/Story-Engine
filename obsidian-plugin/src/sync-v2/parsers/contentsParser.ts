export type FenceType = "chapter" | "scene" | "beat" | "content";

export interface ParsedFence {
	type: FenceType;
	id: string;
	order: number;
	name: string;
	content: string;
	innerText: string;
	startLine: number;
	endLine: number;
	positionInFile: number;
	parentId?: string;
	children: ParsedFence[];
}

export interface HierarchicalContent {
	chapters: ParsedFence[];
	orphanScenes: ParsedFence[];
	orphanBeats: ParsedFence[];
	orphanContents: ParsedFence[];
}

export interface FenceChange {
	id: string;
	type: FenceType;
	changeType: "created" | "updated" | "deleted" | "moved" | "reordered";
	oldOrder?: number;
	newOrder?: number;
	oldParentId?: string;
	newParentId?: string;
}

interface InternalFence extends ParsedFence {
	contentStart: number;
	contentEnd: number;
	startIndex: number;
	endIndex: number;
}

interface InternalHierarchy {
	chapters: InternalFence[];
	orphanScenes: InternalFence[];
	orphanBeats: InternalFence[];
	orphanContents: InternalFence[];
}

const FENCE_PATTERN =
	/<!--(chapter|scene|beat|content)-(start|end):(\d{4}):([a-z0-9-]+):([a-zA-Z0-9-]+)-->/gi;

export const PLACEHOLDER_DEFAULTS: Record<FenceType, string> = {
	chapter: "## _New Chapter Title_",
	scene: "### _New Scene Title_\n\n_Describe what happens in this scene..._",
	beat: "#### _New Beat Intent_",
	content: "_Write your content here..._",
};

export class ContentsParser {
	parseHierarchy(content: string): HierarchicalContent {
		const internal = this.parseInternal(content);
		return {
			chapters: internal.chapters.map((f) => this.stripInternalProps(f)),
			orphanScenes: internal.orphanScenes.map((f) => this.stripInternalProps(f)),
			orphanBeats: internal.orphanBeats.map((f) => this.stripInternalProps(f)),
			orphanContents: internal.orphanContents.map((f) => this.stripInternalProps(f)),
		};
	}

	parseFencesByType(content: string, type: FenceType): ParsedFence[] {
		return this.flattenInternal(this.parseInternal(content))
			.filter((f) => f.type === type)
			.map((f) => this.stripInternalProps(f));
	}

	detectChanges(oldContent: string, newContent: string): FenceChange[] {
		const oldMap = new Map(
			this.flattenInternal(this.parseInternal(oldContent)).map((f) => [f.id, f])
		);
		const newMap = new Map(
			this.flattenInternal(this.parseInternal(newContent)).map((f) => [f.id, f])
		);
		const changes: FenceChange[] = [];

		for (const [id, newFence] of newMap.entries()) {
			if (!oldMap.has(id)) {
				changes.push({
					id,
					type: newFence.type,
					changeType: "created",
					newOrder: newFence.order,
					newParentId: newFence.parentId,
				});
				continue;
			}

			const oldFence = oldMap.get(id)!;
			if (oldFence.parentId !== newFence.parentId) {
				changes.push({
					id,
					type: newFence.type,
					changeType: "moved",
					oldParentId: oldFence.parentId,
					newParentId: newFence.parentId,
				});
			} else if (oldFence.order !== newFence.order) {
				changes.push({
					id,
					type: newFence.type,
					changeType: "reordered",
					oldOrder: oldFence.order,
					newOrder: newFence.order,
				});
			} else if (oldFence.innerText !== newFence.innerText) {
				changes.push({
					id,
					type: newFence.type,
					changeType: "updated",
				});
			}
		}

		for (const [id, oldFence] of oldMap.entries()) {
			if (!newMap.has(id)) {
				changes.push({
					id,
					type: oldFence.type,
					changeType: "deleted",
					oldOrder: oldFence.order,
					oldParentId: oldFence.parentId,
				});
			}
		}

		return changes;
	}

	generateFenceStart(type: FenceType, order: number, name: string, id: string): string {
		return `<!--${type}-start:${order.toString().padStart(4, "0")}:${name}:${id}-->`;
	}

	generateFenceEnd(type: FenceType, order: number, name: string, id: string): string {
		return `<!--${type}-end:${order.toString().padStart(4, "0")}:${name}:${id}-->`;
	}

	generateFence(
		type: FenceType,
		order: number,
		name: string,
		id: string,
		innerContent: string
	): string {
		return `${this.generateFenceStart(type, order, name, id)}\n${innerContent}\n${this.generateFenceEnd(
			type,
			order,
			name,
			id
		)}`;
	}

	updateFenceContent(content: string, id: string, newInnerContent: string): string {
		const target = this.flattenInternal(this.parseInternal(content)).find((f) => f.id === id);
		if (!target) {
			return content;
		}

		const before = content.slice(0, target.contentStart);
		const after = content.slice(target.contentEnd);
		return `${before}${newInnerContent}${after}`;
	}

	updateFenceMeta(content: string, id: string, newOrder: number, newName: string): string {
		return content.replace(
			new RegExp(`(<!--(chapter|scene|beat|content)-(start|end):)\\d{4}:([a-z0-9-]+):(${id})-->`, "gi"),
			(_, prefix) => `${prefix}${newOrder.toString().padStart(4, "0")}:${newName}:${id}-->`
		);
	}

	removeFence(content: string, id: string): string {
		const matches = [...content.matchAll(FENCE_PATTERN)].filter((match) => match[5] === id);
		if (matches.length < 2) {
			return content;
		}
		const start = matches[0].index ?? 0;
		const end = (matches[matches.length - 1].index ?? 0) + matches[matches.length - 1][0].length;
		return content.slice(0, start) + content.slice(end);
	}

	recalculateOrders(content: string, type: FenceType): string {
		const fences = this.flattenInternal(this.parseInternal(content))
			.filter((f) => f.type === type)
			.sort((a, b) => a.positionInFile - b.positionInFile);
		let updated = content;
		fences.forEach((fence, index) => {
			const desiredOrder = index + 1;
			if (fence.order !== desiredOrder) {
				updated = this.updateFenceMeta(updated, fence.id, desiredOrder, fence.name);
			}
		});
		return updated;
	}

	sanitizeName(title: string): string {
		return title
			.toLowerCase()
			.replace(/[^a-z0-9\s-]/g, "")
			.trim()
			.replace(/\s+/g, "-")
			.slice(0, 30);
	}

	isPlaceholder(fence: ParsedFence): boolean {
		return fence.order === 0 || fence.name.startsWith("new-") || fence.id === "placeholder";
	}

	isModifiedPlaceholder(fence: ParsedFence): boolean {
		if (!this.isPlaceholder(fence)) {
			return false;
		}
		const defaultContent = PLACEHOLDER_DEFAULTS[fence.type];
		return fence.innerText !== defaultContent.trim();
	}

	generatePlaceholder(type: FenceType, parentId?: string): string {
		const name = `new-${type}${parentId ? `-${parentId.slice(0, 6)}` : ""}`;
		return this.generateFence(type, 0, name, "placeholder", PLACEHOLDER_DEFAULTS[type]);
	}

	ensurePlaceholders(content: string): string {
		let updated = content;
		for (const type of ["chapter", "scene", "beat", "content"] as FenceType[]) {
			const hasPlaceholder = this.parseFencesByType(updated, type).some((f) => this.isPlaceholder(f));
			if (!hasPlaceholder) {
				updated = `${updated.trimEnd()}\n${this.generatePlaceholder(type)}\n`;
			}
		}
		return updated;
	}

	replacePlaceholder(
		content: string,
		placeholderPosition: number,
		realFence: { type: FenceType; order: number; name: string; id: string; content: string }
	): string {
		const internal = this.parseInternal(content);
		const target = this.flattenInternal(internal).find(
			(f) => f.id === "placeholder" && placeholderPosition >= f.startIndex && placeholderPosition <= f.endIndex
		);
		if (!target) {
			return content;
		}
		const replacement = this.generateFence(
			realFence.type,
			realFence.order,
			realFence.name,
			realFence.id,
			realFence.content
		);
		return content.slice(0, target.startIndex) + replacement + content.slice(target.endIndex);
	}

	private parseInternal(content: string): InternalHierarchy {
		const normalizedContent = this.normalizeLegacyPlaceholders(content);
		const stack: InternalFence[] = [];
		const chapters: InternalFence[] = [];
		const orphanScenes: InternalFence[] = [];
		const orphanBeats: InternalFence[] = [];
		const orphanContents: InternalFence[] = [];

		const finalize = (fence: InternalFence) => {
			if (fence.type === "scene" && !fence.parentId) {
				orphanScenes.push(fence);
			} else if (fence.type === "beat" && !fence.parentId) {
				orphanBeats.push(fence);
			} else if (fence.type === "content" && !fence.parentId) {
				orphanContents.push(fence);
			} else if (fence.type === "chapter") {
				chapters.push(fence);
			}
		};

		for (const match of normalizedContent.matchAll(FENCE_PATTERN)) {
			const full = match[0];
			const type = match[1]?.toLowerCase() as FenceType;
			const action = match[2];
			const order = parseInt(match[3], 10);
			const name = match[4];
			const id = match[5];
			const index = match.index ?? 0;
			if (action === "start") {
				const fence: InternalFence = {
					type,
					id,
					order,
					name,
					content: "",
					innerText: "",
					startLine: this.getLineNumber(normalizedContent, index),
					endLine: this.getLineNumber(normalizedContent, index),
					positionInFile: index,
					children: [],
					contentStart: index + full.length,
					contentEnd: index + full.length,
					startIndex: index,
					endIndex: index + full.length,
					parentId: undefined,
				};

				if (stack.length > 0) {
					const parent = stack[stack.length - 1];
					parent.children.push(fence);
					fence.parentId = parent.id;
				}

				stack.push(fence);
			} else {
				const fence = stack.pop();
				if (!fence || fence.id !== id) {
					throw new Error(`Unmatched fence end for id "${id}".`);
				}
				fence.contentEnd = match.index ?? fence.contentStart;
				fence.endIndex = (match.index ?? 0) + full.length;
				fence.endLine = this.getLineNumber(normalizedContent, fence.endIndex);
				fence.content = normalizedContent.slice(fence.contentStart, match.index ?? 0);
				fence.innerText = this.stripFences(fence.content).trim();
				if (stack.length === 0) {
					finalize(fence);
				}
			}
		}

		return { chapters, orphanScenes, orphanBeats, orphanContents };
	}

	private flattenInternal(hierarchy: InternalHierarchy): InternalFence[] {
		const result: InternalFence[] = [];
		const visit = (fence: InternalFence) => {
			result.push(fence);
			fence.children.forEach((child) => visit(child as InternalFence));
		};
		hierarchy.chapters.forEach(visit);
		hierarchy.orphanScenes.forEach(visit);
		hierarchy.orphanBeats.forEach(visit);
		hierarchy.orphanContents.forEach(visit);
		return result;
	}

	private stripInternalProps(fence: InternalFence): ParsedFence {
		const { contentStart: _cS, contentEnd: _cE, startIndex: _sI, endIndex: _eI, ...rest } = fence;
		return {
			...rest,
			children: fence.children.map((child) => this.stripInternalProps(child as InternalFence)),
		};
	}

	private stripFences(innerContent: string): string {
		return innerContent.replace(FENCE_PATTERN, "").trim();
	}

	private getLineNumber(content: string, index: number): number {
		return content.slice(0, index).split("\n").length;
	}

	private normalizeLegacyPlaceholders(content: string): string {
		return content.replace(/<!--\s*new-(chapter|scene|beat|content)\s*-->/gi, (_match, type: string) => {
			const fenceType = type.toLowerCase() as FenceType;
			return `\n${this.generatePlaceholder(fenceType)}\n`;
		});
	}
}

