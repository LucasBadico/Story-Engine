export interface Tenant {
	id: string;
	name: string;
	status: string;
	created_at: string;
	updated_at: string;
}

export interface Story {
	id: string;
	tenant_id: string;
	title: string;
	status: string;
	version_number: number;
	root_story_id: string;
	previous_story_id?: string;
	created_at: string;
	updated_at: string;
}

export interface StoryEngineSettings {
	apiUrl: string;
	apiKey: string;
	tenantId: string;
	tenantName: string;
}

export interface ErrorResponse {
	error: string;
	message: string;
	code: string;
	details?: Record<string, string>;
}

