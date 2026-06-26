CREATE TABLE provider_profiles (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    capability TEXT NOT NULL,
    provider_kind TEXT NOT NULL,
    display_name TEXT NOT NULL,
    endpoint_url TEXT NOT NULL DEFAULT '',
    model_name TEXT NOT NULL DEFAULT '',
    runtime_options_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    capability_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    credential_status TEXT NOT NULL,
    lifecycle_state TEXT NOT NULL,
    last_tested_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT chk_provider_profiles_capability CHECK (capability IN ('speech_to_text', 'language_inference', 'text_to_speech')),
    CONSTRAINT chk_provider_profiles_provider_kind CHECK (provider_kind IN ('gemini', 'openai_compatible', 'local_http')),
    CONSTRAINT chk_provider_profiles_credential_status CHECK (credential_status IN ('missing', 'configured')),
    CONSTRAINT chk_provider_profiles_lifecycle_state CHECK (lifecycle_state IN ('enabled', 'disabled', 'archived')),
    CONSTRAINT chk_provider_profiles_runtime_options_object CHECK (jsonb_typeof(runtime_options_json) = 'object'),
    CONSTRAINT chk_provider_profiles_capability_json_object CHECK (jsonb_typeof(capability_json) = 'object')
);

CREATE INDEX idx_provider_profiles_tenant_lifecycle
    ON provider_profiles (tenant_id, lifecycle_state, created_at, id);

CREATE INDEX idx_provider_profiles_tenant_capability
    ON provider_profiles (tenant_id, capability, provider_kind);
