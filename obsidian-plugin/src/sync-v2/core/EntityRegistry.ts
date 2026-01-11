export class EntityRegistry {
	private handlers = new Map<string, unknown>();

	register(key: string, handler: unknown): void {
		this.handlers.set(key, handler);
	}

	get<T>(key: string): T | undefined {
		return this.handlers.get(key) as T | undefined;
	}
}

