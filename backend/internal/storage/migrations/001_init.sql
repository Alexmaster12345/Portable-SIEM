-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Events table (core log storage)
CREATE TABLE IF NOT EXISTS events (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp   TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    host        TEXT NOT NULL,
    source      TEXT NOT NULL,
    event_type  TEXT NOT NULL,
    severity    TEXT NOT NULL DEFAULT 'info',
    message     TEXT,
    raw         TEXT,
    fields      JSONB DEFAULT '{}',
    tags        TEXT[] DEFAULT '{}',
    mitre_ids   TEXT[] DEFAULT '{}'
);

CREATE INDEX idx_events_timestamp   ON events (timestamp DESC);
CREATE INDEX idx_events_host        ON events (host);
CREATE INDEX idx_events_source      ON events (source);
CREATE INDEX idx_events_event_type  ON events (event_type);
CREATE INDEX idx_events_severity    ON events (severity);
CREATE INDEX idx_events_fields      ON events USING GIN (fields);
CREATE INDEX idx_events_message_fts ON events USING GIN (to_tsvector('english', message));

-- Alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rule_id      TEXT NOT NULL,
    rule_name    TEXT NOT NULL,
    severity     TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'open',
    title        TEXT NOT NULL,
    description  TEXT,
    host         TEXT,
    event_ids    UUID[] DEFAULT '{}',
    mitre_ids    TEXT[] DEFAULT '{}',
    assigned_to  TEXT,
    notes        TEXT,
    incident_id  UUID,
    fields       JSONB DEFAULT '{}'
);

CREATE INDEX idx_alerts_created_at ON alerts (created_at DESC);
CREATE INDEX idx_alerts_severity   ON alerts (severity);
CREATE INDEX idx_alerts_status     ON alerts (status);
CREATE INDEX idx_alerts_rule_id    ON alerts (rule_id);

-- Incidents table
CREATE TABLE IF NOT EXISTS incidents (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    title       TEXT NOT NULL,
    description TEXT,
    severity    TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'open',
    assigned_to TEXT,
    alert_ids   UUID[] DEFAULT '{}',
    tags        TEXT[] DEFAULT '{}'
);

-- Timeline entries for incidents
CREATE TABLE IF NOT EXISTS timeline_entries (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    timestamp   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    author      TEXT NOT NULL,
    type        TEXT NOT NULL,
    content     TEXT NOT NULL
);

CREATE INDEX idx_timeline_incident_id ON timeline_entries (incident_id);
CREATE INDEX idx_timeline_timestamp   ON timeline_entries (timestamp DESC);

-- IOC (Indicators of Compromise) table
CREATE TABLE IF NOT EXISTS iocs (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    type       TEXT NOT NULL,
    value      TEXT NOT NULL,
    confidence INTEGER NOT NULL DEFAULT 80,
    source     TEXT NOT NULL,
    tags       TEXT[] DEFAULT '{}',
    expires_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_iocs_type_value ON iocs (type, value);
CREATE INDEX idx_iocs_value ON iocs (value);

-- Rules table (persisted custom rules)
CREATE TABLE IF NOT EXISTS rules (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT,
    type        TEXT NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    severity    TEXT NOT NULL,
    source      TEXT,
    event_type  TEXT,
    conditions  JSONB DEFAULT '[]',
    threshold   INTEGER DEFAULT 1,
    window_secs INTEGER DEFAULT 300,
    group_by    TEXT[] DEFAULT '{}',
    actions     TEXT[] DEFAULT '{alert}',
    mitre_ids   TEXT[] DEFAULT '{}',
    tags        TEXT[] DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Asset inventory
CREATE TABLE IF NOT EXISTS assets (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    hostname    TEXT NOT NULL UNIQUE,
    ip_address  TEXT,
    os          TEXT,
    type        TEXT, -- server, workstation, vm, network
    owner       TEXT,
    tags        TEXT[] DEFAULT '{}',
    last_seen   TIMESTAMPTZ,
    metadata    JSONB DEFAULT '{}'
);

-- Saved searches
CREATE TABLE IF NOT EXISTS saved_searches (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    name       TEXT NOT NULL,
    query      TEXT NOT NULL,
    filters    JSONB DEFAULT '{}',
    owner      TEXT
);
