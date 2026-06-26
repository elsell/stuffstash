ALTER TABLE provider_profiles
    ADD COLUMN prompt_template TEXT NOT NULL DEFAULT '';

ALTER TABLE provider_profiles
    ADD CONSTRAINT chk_provider_profiles_prompt_template_length CHECK (char_length(prompt_template) <= 8192);

ALTER TABLE provider_profiles
    ADD CONSTRAINT chk_provider_profiles_prompt_template_capability CHECK (capability = 'language_inference' OR prompt_template = '');
