Obsidian Plugin Auth Flow (Ideation)

Goal
- Plugin authenticates user against the online service.
- Session is persisted locally for future requests.
- Uses the same JWT-based auth as web clients.

High-Level Flow (Browser Login + Callback)
1) Plugin opens browser to Auth Service login URL.
2) User authenticates (email/password or magic link).
3) Auth Service redirects to a custom URL scheme:
   obsidian://story-engine/auth?token=JWT&refresh=REFRESH_TOKEN
4) Plugin captures the callback, validates token, saves session.
5) Plugin calls llm-gateway with Authorization: Bearer JWT.

Custom URL Scheme
- Scheme: obsidian://
- Path: /story-engine/auth
- Query:
  - token (JWT)
  - refresh (optional)
  - tenant (optional default tenant)

Session Storage
- Store in Obsidian plugin storage:
  - access_token
  - refresh_token (if supported)
  - expires_at
  - tenant_id (last used)

Token Refresh
- Before each request, check expiration.
- If expired or near expiry:
  - call Auth Service refresh endpoint
  - update access_token + expires_at

Logout
- Clear stored tokens.
- Optionally revoke refresh token server-side.

Failure Cases
- Missing/invalid callback token -> show login error UI.
- Token expired -> refresh flow.
- Tenant mismatch -> prompt tenant selection.

Security Notes
- Use short-lived access tokens.
- Use refresh tokens with rotation.
- Prefer PKCE for browser login if possible.
- Avoid logging tokens in plugin logs.

Open Questions
- Do we need multi-tenant selection UI in the plugin?
- Should the plugin allow multiple accounts?
