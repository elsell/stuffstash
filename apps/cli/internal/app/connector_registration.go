package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
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
		return ports.Failure("usage", "Supply --name and a discovered printer to register the connector.")
	}
	if r.Receipts != nil && r.ReceiptDrainTimeout <= 0 {
		return ports.Failure("configuration", "Receipt drain timeout must be positive. Examine the CLI configuration.")
	}
	if r.PollInterval <= 0 {
		return ports.Failure("configuration", "pairing poll interval must be positive")
	}
	recovery := "connectors print register --name " + strconv.Quote(name)
	if target != nil {
		recovery = "connectors print rotate --connector " + strconv.Quote(target.ConnectorID)
	}
	recovery += " --server " + strconv.Quote(server)
	key, err := r.Keys.NewKey()
	if err != nil {
		return ports.Failure("configuration", "The CLI could not create a pairing key. Examine the system security configuration before you try again.")
	}
	challenge, err := r.API.Start(ctx, ports.PairingRequest{Rotation: target != nil, Name: name, PublicKey: key.PublicKey(), Candidates: candidates})
	if err != nil {
		return err
	}
	verification, err := url.Parse(challenge.VerificationURL)
	if err != nil || verification.Scheme != "https" || verification.Host == "" || verification.User != nil || verification.RawQuery != "" || verification.Fragment != "" || challenge.ID == "" || challenge.PollToken == "" || challenge.UserCode == "" || !challenge.ExpiresAt.After(r.Clock.Now()) {
		return ports.Failure("protocol", "The server returned an incorrect pairing challenge. Examine the server configuration before you start pairing again.")
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
				return ports.Failure("protocol", "The server returned an incorrect connector registration. Examine the connector state before you start pairing again.")
			}
			if target != nil && (registration.Server != target.Server || registration.TenantID != target.TenantID || registration.InventoryID != target.InventoryID || registration.ConnectorID != target.ConnectorID) {
				return ports.Failure("protocol", "The rotation approval belongs to a different connector. The CLI kept the existing credential. Examine the selected connector before you try again.")
			}
			if !registration.ActivationDeadline.After(r.Clock.Now()) {
				if target != nil {
					return ports.Failure("pairing", "The activation deadline passed. The stored credential was not changed. Run "+recovery+" to pair again.")
				}
				return ports.Failure("pairing", "The activation deadline passed. No credential was saved. Run "+recovery+" to pair again.")
			}
			if err := r.Credentials.Save(ctx, registration); err != nil {
				return err
			}
			if !registration.ActivationDeadline.After(r.Clock.Now()) {
				return ports.Failure("activation", "The registration was saved, but the activation deadline passed. Run connectors print rotate --connector "+strconv.Quote(registration.ConnectorID)+" --server "+strconv.Quote(server)+" to pair again.")
			}
			if err := r.API.Activate(ctx, registration, challenge.ID); err != nil {
				return ports.Failure("activation", "The registration was saved. Run connectors print run --connector "+strconv.Quote(registration.ConnectorID)+" --server "+strconv.Quote(server)+" to try activation again.")
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
			return ports.Failure("pairing", "The pairing is no longer available. Run "+recovery+" to start pairing again.")
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
	return ports.Failure("pairing", "The pairing expired. Run "+recovery+" to start pairing again.")
}
