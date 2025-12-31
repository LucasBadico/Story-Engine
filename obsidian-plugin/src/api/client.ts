import { ErrorResponse, Story, Tenant } from "../types";

export class StoryEngineClient {
	constructor(
		private apiUrl: string,
		private apiKey: string
	) {}

	private async request<T>(
		method: string,
		endpoint: string,
		body?: unknown
	): Promise<T> {
		const url = `${this.apiUrl}${endpoint}`;
		const headers: Record<string, string> = {
			"Content-Type": "application/json",
		};

		if (this.apiKey) {
			headers["Authorization"] = `Bearer ${this.apiKey}`;
		}

		const options: RequestInit = {
			method,
			headers,
		};

		if (body) {
			options.body = JSON.stringify(body);
		}

		const response = await fetch(url, options);

		if (!response.ok) {
			let error: ErrorResponse;
			try {
				error = await response.json();
			} catch {
				error = {
					error: "unknown_error",
					message: `HTTP ${response.status}: ${response.statusText}`,
					code: "HTTP_ERROR",
				};
			}
			// Use the error message from the API, or fallback to status text
			const errorMessage = error.message || error.error || `HTTP ${response.status}: ${response.statusText}`;
			throw new Error(errorMessage);
		}

		return response.json();
	}

	async listStories(tenantId: string): Promise<Story[]> {
		// Ensure tenantId is trimmed and URL encoded
		const trimmedTenantId = encodeURIComponent(tenantId.trim());
		const response = await this.request<{ stories: Story[] }>(
			"GET",
			`/api/v1/stories?tenant_id=${trimmedTenantId}`
		);
		return response.stories || [];
	}

	async getStory(id: string): Promise<Story> {
		const response = await this.request<{ story: Story }>(
			"GET",
			`/api/v1/stories/${id}`
		);
		return response.story;
	}

	async createStory(tenantId: string, title: string): Promise<Story> {
		// Ensure tenantId is trimmed and valid
		const trimmedTenantId = tenantId.trim();
		if (!trimmedTenantId) {
			throw new Error("Tenant ID is required");
		}

		const response = await this.request<{ story: Story }>(
			"POST",
			"/api/v1/stories",
			{
				tenant_id: trimmedTenantId,
				title: title.trim(),
			}
		);
		return response.story;
	}

	async cloneStory(id: string): Promise<Story> {
		const response = await this.request<{ story: Story }>(
			"POST",
			`/api/v1/stories/${id}/clone`
		);
		return response.story;
	}

	async getTenant(id: string): Promise<Tenant> {
		const response = await this.request<{ tenant: Tenant }>(
			"GET",
			`/api/v1/tenants/${id}`
		);
		return response.tenant;
	}

	async testConnection(): Promise<boolean> {
		try {
			await this.request<{ status: string }>("GET", "/health");
			return true;
		} catch {
			return false;
		}
	}
}

