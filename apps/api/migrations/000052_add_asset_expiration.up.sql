ALTER TABLE assets ADD COLUMN expiration_date VARCHAR(10) NOT NULL DEFAULT '';
ALTER TABLE assets ADD COLUMN expiration_precision VARCHAR(5) NOT NULL DEFAULT '';
ALTER TABLE assets ADD CONSTRAINT chk_asset_expiration_precision CHECK ((expiration_date = '' AND expiration_precision = '') OR (expiration_date <> '' AND expiration_precision IN ('day', 'month')));
