Auth + Tenant Roles Architecture (Ideation)

Goal
- First-party auth service.
- JWT carries tenant list + roles per tenant.
- llm-gateway exposes public endpoints and validates JWT.
- Internal gRPC remains tenant_id-based (no JWT).

Key Decisions
- Auth is owned in-house (email + password or magic link).
- JWT includes all tenant memberships + roles.
- Roles: reader, writer, llm-user, admin.
- Tenant context is selected per request via X-Tenant-ID.
- gRPC internal calls only require tenant_id metadata.

JWT Claims (Proposed)
{
  "iss": "story-engine-auth",
  "sub": "user_uuid",
  "exp": 1710000000,
  "iat": 1709990000,
  "tenants": [
    { "id": "tenant_uuid_1", "roles": ["admin", "llm-user"] },
    { "id": "tenant_uuid_2", "roles": ["reader"] }
  ],
  "default_tenant": "tenant_uuid_1",
  "version": 1
}

Request Flow (Public HTTP)
1) Client sends Authorization: Bearer <JWT>
2) Middleware validates JWT signature/exp
3) Middleware selects tenant:
   - X-Tenant-ID header (preferred)
   - fallback to default_tenant
4) Middleware checks tenant exists in token + required role
5) Injects tenant_id into request context

llm-gateway Public Endpoints
- Must validate JWT
- Must check tenant + roles
- Must forward tenant_id to main-service gRPC

Internal gRPC Flow
- No JWT validation on gRPC
- Only tenant_id in metadata for consistency
- Trust boundary is between services inside the cluster

Role Semantics (Initial)
- reader: read-only access
- writer: create/update/delete domain content
- llm-user: allowed to use LLM endpoints
- admin: tenant admin (manage members, billing, settings)

Suggested Middleware Contract (HTTP)
Inputs:
- Authorization: Bearer <JWT>
- X-Tenant-ID: <tenant_uuid>
Outputs:
- context.user_id
- context.tenant_id
- context.roles

Open Questions
- Do we allow "tenant list" to be large (pagination)?
- Should tokens embed role scopes per product feature?
- Token refresh policy and rotation?
