CREATE TABLE realtime_sessions (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    inventory_id TEXT NOT NULL REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    principal_id TEXT NOT NULL,
    source TEXT NOT NULL,
    state TEXT NOT NULL,
    speech_to_text_profile_id TEXT NOT NULL,
    language_inference_profile_id TEXT NOT NULL,
    text_to_speech_profile_id TEXT NOT NULL,
    safe_failure_code TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ NOT NULL,
    last_activity_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT chk_realtime_sessions_state CHECK (state IN ('started', 'completed', 'failed', 'cancelled')),
    CONSTRAINT chk_realtime_sessions_started_at CHECK (last_activity_at >= started_at),
    CONSTRAINT chk_realtime_sessions_ended_at CHECK (ended_at IS NULL OR ended_at >= started_at),
    CONSTRAINT chk_realtime_sessions_failure_code CHECK ((state = 'failed' AND safe_failure_code <> '') OR (state <> 'failed' AND safe_failure_code = ''))
);

CREATE INDEX idx_realtime_sessions_tenant_started
    ON realtime_sessions (tenant_id, started_at);

CREATE INDEX idx_realtime_sessions_inventory_started
    ON realtime_sessions (tenant_id, inventory_id, started_at);

CREATE INDEX idx_realtime_sessions_principal_started
    ON realtime_sessions (tenant_id, principal_id, started_at);
