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
	API          ports.PairingAPI
	Credentials  ports.ConnectorCredentials
	Keys         ports.PairingKeys
	Clock        ports.Clock
	Waiter       ports.Waiter
	Output       ports.Output
	PollInterval time.Duration
}

func (r ConnectorRegistrar) Register(ctx context.Context, server, name string, candidates []ports.PairingCandidate) error {
	return r.pair(ctx, server, name, candidates, nil)
}

func (r ConnectorRegistrar) pair(ctx context.Context, server, name string, candidates []ports.PairingCandidate, target *ports.ConnectorRegistration) error {
	if strings.TrimSpace(name) == "" || (target == nil && len(candidates) == 0) {
		return ports.Failure("usage", "registration needs --name and a discovered printer")
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
			if registration.Server != server || registration.ConnectorID == "" || registration.TenantID == "" || registration.InventoryID == "" || registration.Credential == "" || !registration.ExpiresAt.After(r.Clock.Now()) {
				return ports.Failure("protocol", "invalid connector registration")
			}
			if target != nil && (registration.Server != target.Server || registration.TenantID != target.TenantID || registration.InventoryID != target.InventoryID || registration.ConnectorID != target.ConnectorID) {
				return ports.Failure("protocol", "rotation approval belongs to a different connector; existing credential preserved")
			}
			if err := r.Credentials.Save(ctx, registration); err != nil {
				return err
			}
			if err := r.API.Activate(ctx, registration, challenge.ID); err != nil {
				return ports.Failure("activation", "registration saved; run connectors print run --connector "+registration.ConnectorID+" to retry activation")
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
