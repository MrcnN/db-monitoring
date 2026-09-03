# Security Architecture

## Authentication Flow
1. User provides email/password.
2. Verified against bcrypt hash (cost=12).
3. JWT generated (HMAC-SHA256). Short-lived access token, long-lived refresh token.
4. Refresh tokens tracked in DB/Redis for explicit revocation.

## Credential Encryption
- All target DB passwords are encrypted before storage.
- Algorithm: AES-256-GCM.
- `ENCRYPTION_KEY` injected via environment.
- IV uniquely generated per secret.

## RBAC Roles
- **Admin**: Full access.
- **Editor**: Can add/modify databases, view metrics.
- **Viewer**: Read-only access to metrics.

## Rate Limiting
- Token bucket algorithm implemented with Redis.
- Defaults: 100 req/min, burst 200.

## Audit Logging
- Critical actions (login, credential change, DB delete) strictly logged.
- Immutable log append.

## Production Hardening Recommendations
1. Place API behind WAF.
2. Rotate `ENCRYPTION_KEY` via dual-key approach (future feature).
3. Enforce TLS for all target database connections.
