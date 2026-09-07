CREATE TABLE media_blob_keys (storage_key VARCHAR(512) PRIMARY KEY);
INSERT INTO media_blob_keys (storage_key) SELECT storage_key FROM attachments ON CONFLICT DO NOTHING;
INSERT INTO media_blob_keys (storage_key) SELECT storage_key FROM blob_deletion_events ON CONFLICT DO NOTHING;
