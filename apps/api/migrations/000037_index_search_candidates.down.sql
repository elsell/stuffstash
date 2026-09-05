DROP INDEX IF EXISTS idx_assets_search_cursor;
DROP INDEX IF EXISTS idx_attachments_search_content_type;
DROP INDEX IF EXISTS idx_attachments_search_filename;
DROP INDEX IF EXISTS idx_assets_search_description;
DROP INDEX IF EXISTS idx_assets_search_title;
-- pg_trgm may be shared by other indexes or applications; retain the extension.
