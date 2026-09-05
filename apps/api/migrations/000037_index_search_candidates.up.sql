CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_assets_search_title ON assets USING gin ((translate(title, 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz') COLLATE "C") gin_trgm_ops);
CREATE INDEX idx_assets_search_description ON assets USING gin ((translate(description, 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz') COLLATE "C") gin_trgm_ops);
CREATE INDEX idx_attachments_search_filename ON attachments USING gin ((translate(file_name, 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz') COLLATE "C") gin_trgm_ops);
CREATE INDEX idx_attachments_search_content_type ON attachments USING gin ((translate(content_type, 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz') COLLATE "C") gin_trgm_ops);
CREATE INDEX idx_assets_search_cursor ON assets (tenant_id, ((inventory_id || ':' || id) COLLATE "C"));
