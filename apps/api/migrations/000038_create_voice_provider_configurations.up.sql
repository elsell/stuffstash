CREATE TABLE voice_provider_configurations (
    tenant_id TEXT PRIMARY KEY REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    speech_to_text_profile_id TEXT NOT NULL DEFAULT '',
    language_inference_profile_id TEXT NOT NULL DEFAULT '',
    text_to_speech_profile_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
