CREATE TABLE notification_deliveries (
 id VARCHAR(26) PRIMARY KEY,
 tenant_id VARCHAR(26) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
 inventory_id VARCHAR(26) NOT NULL REFERENCES inventories(id) ON DELETE CASCADE,
 principal_id VARCHAR(128) NOT NULL,
 notification_id VARCHAR(26) NOT NULL REFERENCES notification_inbox(id) ON DELETE CASCADE,
 device_id VARCHAR(26) NOT NULL REFERENCES notification_devices(id) ON DELETE CASCADE,
 device_revision BIGINT NOT NULL CHECK (device_revision > 0),
 created_at TIMESTAMPTZ NOT NULL,
 status VARCHAR(16) NOT NULL CHECK (status IN ('pending','leased','accepted','cancelled','failed')),
 attempts INTEGER NOT NULL CHECK (attempts >= 0),
 next_attempt_at TIMESTAMPTZ,
 lease_until TIMESTAMPTZ,
 fence VARCHAR(128) NOT NULL,
 CONSTRAINT idx_notification_delivery_device UNIQUE (notification_id, device_id)
);
CREATE INDEX idx_notification_delivery_due ON notification_deliveries (status, next_attempt_at, lease_until, id);
