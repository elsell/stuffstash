DROP INDEX idx_authorization_outbox_events_claimed_until;
DROP INDEX idx_authorization_outbox_events_claim_id;

ALTER TABLE authorization_outbox_events
  DROP COLUMN claimed_until,
  DROP COLUMN claim_id;
