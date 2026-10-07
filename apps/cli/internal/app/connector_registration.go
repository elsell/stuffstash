package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// ConnectorRegistrar owns key-bound pairing and durable activation ordering.
type ConnectorRegistrar struct {
	Receipts            ports.ReceiptDrain
	ReceiptDrainTimeout time.Duration
	API                 ports.PairingAPI
	Credentials         ports.ConnectorCredentials
	Keys                ports.PairingKeys
	Clock               ports.Clock
	Waiter              ports.Waiter
	Output              ports.Output
	PollInterval        time.Duration
}

func (r ConnectorRegistrar) Register(ctx context.Context, server, name string, candidates []ports.PairingCandidate) error {
	return r.pair(ctx, server, name, candidates, nil)
}

func (r ConnectorRegistrar) pair(ctx context.Context, server, name string, candidates []ports.PairingCandidate, target *ports.ConnectorRegistration) error {
	if strings.TrimSpace(name) == "" || (target == nil && len(candidates) == 0) {
		return ports.Failure("usage", "registration needs --name and a discovered printer")
	}
	if r.Receipts != nil && r.ReceiptDrainTimeout <= 0 {
		return ports.Failure("configuration", "Receipt drain timeout must be positive. Examine the CLI configuration.")
	}
	if r.PollInterval <= 0 {
		return ports.Failure("configuration", "pairing poll interval must be positive")
	}
	key, err := r.Keys.NewKey()
	if err != nil {
		return ports.Failure("configuration", "could not create pairing key")
	}
	challenge, err := r.API.Start(ctx, ports.PairingRequest{Rotation: target != nil, Name: name, PublicKey: key.PublicKey(), Candidates: candidates})
	if err != nil {
		return err
	}
	verification, err := url.Parse(challenge.VerificationURL)
	if err != nil || verification.Scheme != "https" || verification.Host == "" || verification.User != nil || verification.RawQuery != "" || verification.Fragment != "" || challenge.ID == "" || challenge.PollToken == "" || challenge.UserCode == "" || !challenge.ExpiresAt.After(r.Clock.Now()) {
		return ports.Failure("protocol", "invalid pairing challenge")
	}
	if target != nil {
		query := verification.Query()
		query.Set("tenantId", target.TenantID)
		query.Set("inventoryId", target.InventoryID)
		query.Set("connectorId", target.ConnectorID)
		verification.RawQuery = query.Encode()
	}
	if err := r.Output.Notice(fmt.Sprintf("Open %s and enter code %s. Confirm the printer and label size in the browser.", verification.String(), challenge.UserCode)); err != nil {
		return err
	}
	for challenge.ExpiresAt.After(r.Clock.Now()) {
		state, err := r.API.Poll(ctx, challenge)
		if err != nil {
			return err
		}
		switch state {
		case ports.PairingPending:
		case ports.PairingApproved:
			registration, err := r.API.Exchange(ctx, challenge, key.Sign(challenge.ID, challenge.PollToken))
			if err != nil {
				var typed *ports.Error
				if errors.As(err, &typed) && typed.Category == "conflict" {
					break
				}
				return err
			}
			if registration.Server != server || registration.ConnectorID == "" || registration.TenantID == "" || registration.InventoryID == "" || registration.Credential == "" || registration.ActivationDeadline.IsZero() || !registration.ExpiresAt.After(r.Clock.Now()) {
				return ports.Failure("protocol", "invalid connector registration")
			}
			if target != nil && (registration.Server != target.Server || registration.TenantID != target.TenantID || registration.InventoryID != target.InventoryID || registration.ConnectorID != target.ConnectorID) {
				return ports.Failure("protocol", "rotation approval belongs to a different connector; existing credential preserved")
			}
			if !registration.ActivationDeadline.After(r.Clock.Now()) {
				if target != nil {
					return ports.Failure("pairing", fmt.Sprintf("The activation deadline passed. The stored credential was not changed. Run connectors print rotate --connector %q again.", target.ConnectorID))
				}
				return ports.Failure("pairing", "The activation deadline passed. No credential was saved. Run connectors print register again.")
			}
			if err := r.Credentials.Save(ctx, registration); err != nil {
				return err
			}
			if !registration.ActivationDeadline.After(r.Clock.Now()) {
				return ports.Failure("activation", fmt.Sprintf("The registration was saved, but the activation deadline passed. Run connectors print rotate --connector %q to pair again.", registration.ConnectorID))
			}
			if err := r.API.Activate(ctx, registration, challenge.ID); err != nil {
				return ports.Failure("activation", "registration saved; run connectors print run --connector "+registration.ConnectorID+" to retry activation")
			}
			if r.Receipts != nil {
				drainCtx, cancel := context.WithTimeout(context.Background(), r.ReceiptDrainTimeout)
				drained := r.Receipts.Close(drainCtx)
				cancel()
				if !drained {
					return nil
				}
			}
			return r.Output.Result(struct {
				ConnectorID string `json:"connectorId"`
				TenantID    string `json:"tenantId"`
				InventoryID string `json:"inventoryId"`
			}{registration.ConnectorID, registration.TenantID, registration.InventoryID})
		default:
			return ports.Failure("pairing", "pairing is no longer available; register again")
		}
		remaining := challenge.ExpiresAt.Sub(r.Clock.Now())
		if remaining <= 0 {
			break
		}
		wait := min(r.PollInterval, remaining)
		if err := r.Waiter.Wait(ctx, wait); err != nil {
			return err
		}
	}
	return ports.Failure("pairing", "pairing expired; register again")
}
