export class RelationsReconciler {
	reconcile(localContent: string | null, generatedContent: string): string {
		if (!localContent) {
			return generatedContent;
		}

		const segments = localContent.split("\n");
		const generatedSegments = generatedContent.split("\n");
		const merged = new Set(segments);
		for (const line of generatedSegments) {
			if (!merged.has(line)) {
				merged.add(line);
			}
		}

		return Array.from(merged).join("\n");
	}
}

