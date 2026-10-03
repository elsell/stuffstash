package printregistry

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"net/url"
	"strings"
	"time"
)

type ConnectorPolicy struct {
	PublicWebBaseURL                                                              string
	PairingLifetime, CredentialLifetime, ActivationLifetime, AuthorizationTimeout time.Duration
	ReportMaxAge                                                                  time.Duration
}

func (p ConnectorPolicy) Valid() bool {
	return p.PairingLifetime > 0 && p.CredentialLifetime > 0 && p.ActivationLifetime > 0 && p.AuthorizationTimeout > 0 && p.ActivationLifetime <= p.CredentialLifetime
}

type ConnectorService struct {
	Registry      Service
	Repository    ports.ConnectorRepository
	Authorization ports.PrintingAuthorization
	Secrets       ports.PairingSecrets
	Policy        ConnectorPolicy
}

type BeginPairing struct {
	Rotation   bool
	Name       string
	PublicKey  []byte
	Candidates []printing.PairingCandidate
}
type PairingStarted struct {
	Pairing                              printing.Pairing
	PollToken, UserCode, VerificationURL string
}
type PairingBinding struct {
	CandidateID string
	PrinterID   printing.PrinterID
}
type ApprovePairing struct {
	Actor     Actor
	PairingID printing.PairingID
	UserCode  string
	Bindings  []PairingBinding
}
type PairingCredential struct {
	Credential string
	Connector  printing.Connector
}

func connectorError(err error) error {
	if errors.Is(err, ports.ErrPrintDenied) {
		return apperrors.ErrUnauthenticated
	}
	return registryError(err)
}
func (s ConnectorService) Begin(ctx context.Context, input BeginPairing) (PairingStarted, error) {
	if !s.Policy.Valid() {
		return PairingStarted{}, errors.New("invalid connector policy")
	}
	input.Name = strings.TrimSpace(input.Name)
	if !validPrinterName(input.Name) || len(input.PublicKey) != 32 || (!input.Rotation && len(input.Candidates) < 1) || (input.Rotation && len(input.Candidates) != 0) || len(input.Candidates) > 16 {
		return PairingStarted{}, apperrors.ErrInvalidInput
	}
	ids := map[string]bool{}
	devices := map[string]bool{}
	for _, c := range input.Candidates {
		if c.ID == "" || len(c.ID) > 100 || ids[c.ID] || !validPrinterName(c.Name) || c.DeviceID == "" || len(c.DeviceID) > 1024 || devices[c.DeviceID] {
			return PairingStarted{}, apperrors.ErrInvalidInput
		}
		supported := false
		for _, profile := range s.Registry.Catalog.ListPrinterProfiles() {
			if profile.AdapterID == c.AdapterID {
				supported = true
				break
			}
		}
		if !supported {
			return PairingStarted{}, apperrors.ErrInvalidInput
		}
		ids[c.ID] = true
		devices[c.DeviceID] = true
	}
	token, err := s.Secrets.NewToken()
	if err != nil {
		return PairingStarted{}, err
	}
	code, err := s.Secrets.NewUserCode()
	if err != nil {
		return PairingStarted{}, err
	}
	now := s.Registry.Clock.Now()
	p := printing.Pairing{Rotation: input.Rotation, ID: printing.PairingID(s.Registry.IDs.NewID()), Name: input.Name, PublicKey: append([]byte(nil), input.PublicKey...), Candidates: append([]printing.PairingCandidate(nil), input.Candidates...), PollHash: s.Secrets.Digest(token), CodeHash: s.Secrets.Digest(code), State: printing.PairingPending, ExpiresAt: now.Add(s.Policy.PairingLifetime), CreatedAt: now}
	if err := s.Repository.CreatePrintPairing(ctx, p); err != nil {
		return PairingStarted{}, registryError(err)
	}
	return PairingStarted{Pairing: p, PollToken: token, UserCode: code, VerificationURL: strings.TrimRight(s.Policy.PublicWebBaseURL, "/") + "/print-connectors/pair/" + url.PathEscape(string(p.ID))}, nil
}
func (s ConnectorService) Poll(ctx context.Context, id printing.PairingID, token string) (printing.Pairing, error) {
	p, err := s.Repository.GetPrintPairing(ctx, id)
	if err != nil {
		return printing.Pairing{}, connectorError(err)
	}
	if !p.ExpiresAt.After(s.Registry.Clock.Now()) || !s.Secrets.Matches(token, p.PollHash) {
		return printing.Pairing{}, apperrors.ErrUnauthenticated
	}
	return p, nil
}
func (s ConnectorService) Exchange(ctx context.Context, id printing.PairingID, token string, signature []byte) (PairingCredential, error) {
	p, err := s.Poll(ctx, id, token)
	if err != nil {
		return PairingCredential{}, err
	}
	if !s.Secrets.Verify(p.PublicKey, id, token, signature) {
		return PairingCredential{}, apperrors.ErrUnauthenticated
	}
	if p.State != printing.PairingApproved {
		return PairingCredential{}, apperrors.ErrConflict
	}
	registration, err := s.Repository.GetPrintConnector(ctx, p.Scope, p.ConnectorID)
	if err != nil {
		return PairingCredential{}, connectorError(err)
	}
	if registration.Connector.Generation != registration.Connector.SyncedGeneration {
		return PairingCredential{}, apperrors.ErrConflict
	}
	credential, err := s.Secrets.NewToken()
	if err != nil {
		return PairingCredential{}, err
	}
	now := s.Registry.Clock.Now()
	record, err := s.machineAudit(registration.Connector, audit.ActionPrintConnectorCredentialIssued)
	if err != nil {
		return PairingCredential{}, err
	}
	c, err := s.Repository.ConsumePrintPairing(ctx, ports.PairingExchange{PairingID: id, Now: now, CredentialHash: s.Secrets.Digest(credential), CredentialExpiresAt: now.Add(s.Policy.CredentialLifetime), ActivationDeadline: now.Add(s.Policy.ActivationLifetime), Audit: record})
	if err != nil {
		return PairingCredential{}, connectorError(err)
	}
	return PairingCredential{Credential: credential, Connector: c}, nil
}
func (s ConnectorService) AuthenticateConsumer(ctx context.Context, credential string) (printing.Connector, error) {
	if credential == "" || len(credential) > 128 {
		return printing.Connector{}, apperrors.ErrUnauthenticated
	}
	c, err := s.Repository.FindPrintConnectorCredential(ctx, s.Secrets.Digest(credential))
	if err != nil {
		return printing.Connector{}, connectorError(err)
	}
	now := s.Registry.Clock.Now()
	if c.State != printing.ConnectorActive && c.State != printing.ConnectorAwaitingActivation || !c.CredentialExpiresAt.After(now) || c.Generation != c.SyncedGeneration || c.State == printing.ConnectorAwaitingActivation && !c.ActivationDeadline.After(now) {
		return printing.Connector{}, apperrors.ErrUnauthenticated
	}
	if err := s.Authorization.CheckPrintConnector(ctx, c.ServiceAccountID, c.ID); err != nil {
		return printing.Connector{}, err
	}
	return c, nil
}
func (s ConnectorService) AuthorizePrinter(ctx context.Context, c printing.Connector, id printing.PrinterID, permission ports.PrinterPermission) (printing.ConsumerAuthority, error) {
	if c.State != printing.ConnectorActive {
		return printing.ConsumerAuthority{}, apperrors.ErrUnauthorized
	}
	current, err := s.Repository.GetPrintConnector(ctx, c.Scope, c.ID)
	if err != nil {
		return printing.ConsumerAuthority{}, connectorError(err)
	}
	var binding printing.PrinterBinding
	found := false
	for _, b := range current.Bindings {
		if b.PrinterID == id {
			binding = b
			found = true
			break
		}
	}
	authority := printing.ConsumerAuthority{Scope: c.Scope, ConnectorID: c.ID, ServiceAccountID: c.ServiceAccountID, CredentialVersion: c.CredentialVersion, PrinterID: id, BindingGeneration: binding.Generation}
	if !found || !printing.AcceptsAuthority(current.Connector, binding, authority, s.Registry.Clock.Now()) {
		return printing.ConsumerAuthority{}, apperrors.ErrUnauthorized
	}
	if err := s.Authorization.CheckPrintConnector(ctx, c.ServiceAccountID, c.ID); err != nil {
		return printing.ConsumerAuthority{}, err
	}
	if err := s.Authorization.CheckPrinter(ctx, c.ServiceAccountID, id, permission); err != nil {
		return printing.ConsumerAuthority{}, err
	}
	return authority, nil
}
func (s ConnectorService) Heartbeat(ctx context.Context, c printing.Connector) (printing.Connector, error) {
	if err := s.Authorization.CheckPrintConnector(ctx, c.ServiceAccountID, c.ID); err != nil {
		return printing.Connector{}, err
	}
	result, err := s.Repository.HeartbeatPrintConnector(ctx, c, s.Registry.Clock.Now(), func(activated printing.Connector, rotated bool) (audit.Record, error) {
		action := audit.ActionPrintConnectorActivated
		if rotated {
			action = audit.ActionPrintConnectorCredentialRotated
		}
		return s.machineAudit(activated, action)
	})
	return result, connectorError(err)
}
func (s ConnectorService) Reconcile(ctx context.Context, scope printing.Scope, id printing.ConnectorID) error {
	return s.Repository.SynchronizePrintConnector(ctx, scope, id, func(ctx context.Context, r ports.ConnectorRegistration, printers []printing.Printer) error {
		bounded, cancel := context.WithTimeout(ctx, s.Policy.AuthorizationTimeout)
		defer cancel()
		for _, p := range printers {
			if err := s.Authorization.SyncPrinterInventory(bounded, p); err != nil {
				return err
			}
		}
		return s.Authorization.SyncPrintConnector(bounded, r.Connector, r.Bindings)
	})
}

func (s ConnectorService) DrainAuthorization(ctx context.Context, limit int) error {
	connectors, err := s.Repository.PendingPrintConnectorScopes(ctx, limit)
	if err != nil {
		return err
	}
	var failures []error
	for _, c := range connectors {
		if err := s.Reconcile(ctx, c.Scope, c.ID); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}
