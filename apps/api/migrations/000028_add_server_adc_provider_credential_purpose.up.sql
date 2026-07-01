ALTER TABLE provider_credentials DROP CONSTRAINT chk_provider_credentials_purpose;

ALTER TABLE provider_credentials
    ADD CONSTRAINT chk_provider_credentials_purpose CHECK (purpose IN ('api_key', 'oauth_bearer', 'server_adc'));
