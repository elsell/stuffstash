ALTER TABLE authorization_outbox_events
  ADD COLUMN claim_id varchar(26) NOT NULL DEFAULT '',
  ADD COLUMN claimed_until timestamptz;

CREATE INDEX idx_authorization_outbox_events_claim_id ON authorization_outbox_events(claim_id);
CREATE INDEX idx_authorization_outbox_events_claimed_until ON authorization_outbox_events(claimed_until);
