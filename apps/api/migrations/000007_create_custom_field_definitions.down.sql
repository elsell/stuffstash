DROP TRIGGER IF EXISTS trg_custom_field_effective_key_unique ON custom_field_definitions;
DROP FUNCTION IF EXISTS stuffstash_custom_field_effective_key_unique();
DROP TABLE custom_field_definitions;
DROP FUNCTION IF EXISTS stuffstash_custom_field_enum_options_valid(jsonb);
