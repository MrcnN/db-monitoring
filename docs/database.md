# Database Schema Documentation

## Core Tables

### 1. `users`
- **id** (UUID, PK): Unique identifier.
- **email** (VARCHAR, UNIQUE): Login email.
- **password_hash** (VARCHAR): Bcrypt hash.
- **role** (VARCHAR): Enum for RBAC (e.g., admin, viewer).
- **created_at** / **updated_at** (TIMESTAMPTZ).

### 2. `user_sessions`
- **id** (UUID, PK): Session ID.
- **user_id** (UUID, FK): References `users`.
- **token_signature** (VARCHAR): Hashed JWT signature.
- **expires_at** (TIMESTAMPTZ): Expiration timestamp.

### 3. `monitored_databases`
- **id** (UUID, PK): DB identifier.
- **name** (VARCHAR): Human-readable name.
- **engine** (VARCHAR): `postgres` or `mysql`.
- **host** / **port** / **database_name** (VARCHAR/INT).
- **status** (VARCHAR): Current connection status.

### 4. `database_credentials`
- **database_id** (UUID, PK, FK): References `monitored_databases`.
- **username** (VARCHAR): Connection user.
- **encrypted_password** (BYTEA): AES-256 encrypted payload.
- **encryption_iv** (BYTEA): Initialization vector.

### 5. `audit_logs`
- **id** (UUID, PK): Log ID.
- **user_id** (UUID, FK): Actor.
- **action** (VARCHAR): Event type (e.g., `DB_CREATED`).
- **resource_type** (VARCHAR): e.g., `DATABASE`.
- **resource_id** (UUID): Target resource.
- **details** (JSONB): Contextual metadata.

## Index Strategy
- **users**: B-Tree on `email`.
- **audit_logs**: B-Tree on `user_id` and `created_at` for fast querying.
- **monitored_databases**: B-Tree on `engine` and `status`.

## Data Retention Plan
- Audit logs: Retained for 90 days.
- User sessions: Pruned when `expires_at` is past.

## Migration Strategy
Using `golang-migrate` for versioned schema files (up/down).
