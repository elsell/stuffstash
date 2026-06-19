ALTER TABLE authorization_outbox_events
  ADD COLUMN dead_lettered_at timestamptz,
  ADD COLUMN dead_letter_reason text NOT NULL DEFAULT '';

CREATE INDEX idx_authorization_outbox_events_dead_lettered_at
  ON authorization_outbox_events(dead_lettered_at);
