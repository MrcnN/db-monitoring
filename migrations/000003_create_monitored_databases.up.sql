CREATE TYPE database_type AS ENUM ('postgresql', 'mysql');
CREATE TYPE database_status AS ENUM ('active', 'inactive', 'error');
CREATE TYPE ssl_mode AS ENUM ('disable', 'require', 'verify-ca', 'verify-full');

CREATE TABLE monitored_databases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type database_type NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL,
    database_name VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    ssl_mode ssl_mode NOT NULL DEFAULT 'disable',
    monitoring_interval INTEGER NOT NULL DEFAULT 15,
    status database_status NOT NULL DEFAULT 'active',
    is_monitoring_enabled BOOLEAN NOT NULL DEFAULT true,
    last_checked_at TIMESTAMPTZ,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE database_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    database_id UUID NOT NULL UNIQUE REFERENCES monitored_databases(id) ON DELETE CASCADE,
    encrypted_password TEXT NOT NULL,
    encryption_key_id VARCHAR(50) NOT NULL DEFAULT 'v1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_monitored_databases_status ON monitored_databases(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_monitored_databases_type ON monitored_databases(type) WHERE deleted_at IS NULL;
