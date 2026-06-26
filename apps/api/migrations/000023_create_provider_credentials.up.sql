CREATE TABLE provider_credentials (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    provider_profile_id TEXT NOT NULL,
    capability TEXT NOT NULL,
    provider_kind TEXT NOT NULL,
    purpose TEXT NOT NULL,
    key_id TEXT NOT NULL,
    algorithm TEXT NOT NULL,
    nonce BYTEA NOT NULL,
    ciphertext BYTEA NOT NULL,
    superseded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT chk_provider_credentials_capability CHECK (capability IN ('speech_to_text', 'language_inference', 'text_to_speech')),
    CONSTRAINT chk_provider_credentials_provider_kind CHECK (provider_kind IN ('gemini', 'openai_compatible', 'local_http')),
    CONSTRAINT chk_provider_credentials_purpose CHECK (purpose IN ('api_key', 'oauth_bearer')),
    CONSTRAINT chk_provider_credentials_algorithm CHECK (algorithm IN ('AES-256-GCM')),
    CONSTRAINT chk_provider_credentials_nonce CHECK (octet_length(nonce) = 12),
    CONSTRAINT chk_provider_credentials_ciphertext CHECK (octet_length(ciphertext) > 0)
);

CREATE UNIQUE INDEX idx_provider_credentials_one_active
    ON provider_credentials (tenant_id, provider_profile_id, capability, provider_kind, purpose)
    WHERE superseded_at IS NULL;

CREATE INDEX idx_provider_credentials_tenant_profile
    ON provider_credentials (tenant_id, provider_profile_id, created_at);
