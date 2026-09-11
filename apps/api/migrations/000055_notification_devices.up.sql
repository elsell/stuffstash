CREATE TABLE notification_device_token_locks (
 key VARCHAR(64) PRIMARY KEY
);
CREATE TABLE notification_devices (
 id VARCHAR(26) PRIMARY KEY,
 tenant_id VARCHAR(26) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
 inventory_id VARCHAR(26) NOT NULL REFERENCES inventories(id) ON DELETE CASCADE,
 principal_id VARCHAR(128) NOT NULL,
 installation_id VARCHAR(128) NOT NULL,
 transport VARCHAR(8) NOT NULL CHECK (transport IN ('apns','fcm')),
 token VARCHAR(4096) NOT NULL,
 token_key VARCHAR(64) NOT NULL,
 active BOOLEAN NOT NULL,
 revision BIGINT NOT NULL CHECK (revision > 0),
 created_at TIMESTAMPTZ NOT NULL,
 updated_at TIMESTAMPTZ NOT NULL,
 CONSTRAINT idx_notification_device_installation UNIQUE (tenant_id, inventory_id, principal_id, installation_id)
);
CREATE INDEX idx_notification_device_scope ON notification_devices (tenant_id, inventory_id, principal_id, id);
CREATE INDEX idx_notification_device_token ON notification_devices (token_key, active);
