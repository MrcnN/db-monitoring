CREATE TYPE alert_severity AS ENUM ('warning', 'critical');
CREATE TYPE alert_status AS ENUM ('active', 'resolved');

CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    database_id UUID NOT NULL REFERENCES monitored_databases(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    severity alert_severity NOT NULL,
    status alert_status NOT NULL DEFAULT 'active',
    metric_name VARCHAR(100),
    current_value VARCHAR(100),
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alerts_database_id ON alerts(database_id);
CREATE INDEX idx_alerts_status ON alerts(status);
CREATE INDEX idx_alerts_created_at ON alerts(created_at DESC);
