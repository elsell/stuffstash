DROP INDEX idx_authorization_outbox_events_dead_lettered_at;

ALTER TABLE authorization_outbox_events
  DROP COLUMN dead_letter_reason,
  DROP COLUMN dead_lettered_at;
