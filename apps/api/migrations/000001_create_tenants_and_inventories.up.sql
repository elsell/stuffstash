CREATE TABLE tenants (
  id varchar(26) PRIMARY KEY,
  name varchar(120) NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE TABLE inventories (
  id varchar(26) PRIMARY KEY,
  tenant_id varchar(26) NOT NULL REFERENCES tenants(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  name varchar(120) NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE INDEX idx_inventories_tenant_id ON inventories(tenant_id);
