ALTER TABLE provider_profiles
    DROP CONSTRAINT IF EXISTS chk_provider_profiles_prompt_template_capability;

ALTER TABLE provider_profiles
    DROP CONSTRAINT IF EXISTS chk_provider_profiles_prompt_template_length;

ALTER TABLE provider_profiles
    DROP COLUMN IF EXISTS prompt_template;
