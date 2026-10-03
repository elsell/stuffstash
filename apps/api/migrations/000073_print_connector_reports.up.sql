ALTER TABLE print_connectors ADD COLUMN report_json BYTEA;
ALTER TABLE print_connectors ADD COLUMN report_received_at TIMESTAMPTZ;
