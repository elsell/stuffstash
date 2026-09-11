ALTER TABLE assets DROP CONSTRAINT chk_asset_expiration_precision;
ALTER TABLE assets DROP COLUMN expiration_precision;
ALTER TABLE assets DROP COLUMN expiration_date;
