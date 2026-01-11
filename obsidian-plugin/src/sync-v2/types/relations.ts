export type RelationDirection = "source" | "target";

export interface EntityRelation {
	id: string;
	tenant_id: string;
	source_type: string;
	source_id: string;
	target_type: string;
	target_id: string;
	relation_type: string;
	context?: string;
	created_at: string;
	updated_at: string;
	direction: RelationDirection;
}

export interface EntityCitation {
	id: string;
	source_type: string;
	source_name?: string;
	source_id: string;
	relation_type: string;
	context?: string;
	created_at: string;
}

export interface RelationsPagination {
	has_more: boolean;
	next_cursor?: string;
}

export interface RelationsListResponse {
	data: EntityRelation[];
	pagination?: RelationsPagination;
}

interface BaseListRelationsParams {
	relationType?: string;
	cursor?: string;
	limit?: number;
	orderBy?: string;
	orderDir?: "asc" | "desc";
	excludeMirrors?: boolean;
}

export interface ListRelationsBySourceParams extends BaseListRelationsParams {
	sourceType: string;
	sourceId: string;
}

export interface ListRelationsByTargetParams extends BaseListRelationsParams {
	targetType: string;
	targetId: string;
}

export interface ListRelationsByWorldParams extends BaseListRelationsParams {
	worldId: string;
}

export interface CreateRelationParams {
	sourceType: string;
	sourceId: string;
	targetType: string;
	targetId: string;
	relationType: string;
	context?: string;
}

export interface UpdateRelationParams {
	id: string;
	relationType?: string;
	context?: string;
}

