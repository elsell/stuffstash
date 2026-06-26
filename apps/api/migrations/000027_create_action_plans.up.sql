CREATE TABLE action_plans (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    inventory_id TEXT NOT NULL REFERENCES inventories(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    principal_id TEXT NOT NULL,
    source TEXT NOT NULL,
    realtime_session_id TEXT NOT NULL DEFAULT '',
    state TEXT NOT NULL CHECK (state IN ('proposed', 'approved', 'cancelled', 'executed', 'failed')),
    intent_summary TEXT NOT NULL DEFAULT '',
    model_interpretation_summary TEXT NOT NULL DEFAULT '',
    confirmation_summary TEXT NOT NULL,
    commands JSONB NOT NULL CHECK (jsonb_typeof(commands) = 'array' AND jsonb_array_length(commands) > 0),
    risks JSONB NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(risks) = 'array'),
    approved_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    executed_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CHECK (
        (state = 'proposed' AND approved_at IS NULL AND cancelled_at IS NULL AND executed_at IS NULL AND failed_at IS NULL)
        OR (state = 'approved' AND approved_at IS NOT NULL AND cancelled_at IS NULL AND executed_at IS NULL AND failed_at IS NULL)
        OR (state = 'cancelled' AND cancelled_at IS NOT NULL AND approved_at IS NULL AND executed_at IS NULL AND failed_at IS NULL)
        OR (state = 'executed' AND executed_at IS NOT NULL AND cancelled_at IS NULL AND failed_at IS NULL)
        OR (state = 'failed' AND failed_at IS NOT NULL AND cancelled_at IS NULL AND executed_at IS NULL)
    )
);

CREATE INDEX idx_action_plans_tenant_created ON action_plans (tenant_id, created_at);
CREATE INDEX idx_action_plans_inventory_created ON action_plans (tenant_id, inventory_id, created_at);
CREATE INDEX idx_action_plans_principal_created ON action_plans (tenant_id, principal_id, created_at);
CREATE INDEX idx_action_plans_state ON action_plans (state);
