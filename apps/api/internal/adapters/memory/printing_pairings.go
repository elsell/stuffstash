package memory

import (
	"context"
	"crypto/subtle"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func clonePrintPairing(p printing.Pairing) printing.Pairing {
	p.PublicKey = append([]byte(nil), p.PublicKey...)
	p.Candidates = append([]printing.PairingCandidate(nil), p.Candidates...)
	return p
}
func clonePrintConnector(c printing.Connector) printing.Connector {
	c.Report = c.Report.Clone()
	if c.ReportReceivedAt != nil {
		v := *c.ReportReceivedAt
		c.ReportReceivedAt = &v
	}
	c.PublicKey = append([]byte(nil), c.PublicKey...)
	c.PendingPublicKey = append([]byte(nil), c.PendingPublicKey...)
	if c.LastSeenAt != nil {
		value := *c.LastSeenAt
		c.LastSeenAt = &value
	}
	return c
}
func (s *Store) CreatePrintPairing(_ context.Context, p printing.Pairing) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.printingPairings[p.ID]; exists {
		return ports.ErrPrintConflict
	}
	for _, old := range s.printingPairings {
		if old.CodeHash == p.CodeHash {
			return ports.ErrPrintConflict
		}
	}
	if s.printingPairings == nil {
		s.printingPairings = map[printing.PairingID]printing.Pairing{}
	}
	s.printingPairings[p.ID] = clonePrintPairing(p)
	return nil
}
func (s *Store) GetPrintPairing(_ context.Context, id printing.PairingID) (printing.Pairing, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.printingPairings[id]
	if !ok {
		return printing.Pairing{}, ports.ErrPrintNotFound
	}
	return clonePrintPairing(p), nil
}
func (s *Store) ApprovePrintPairing(_ context.Context, input ports.PairingApproval) (ports.ConnectorRegistration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.printingPairings[input.PairingID]
	c := input.Registration.Connector
	if !ok || subtle.ConstantTimeCompare([]byte(p.CodeHash), []byte(input.CodeHash)) != 1 || !p.Approve(c.Scope, c.ID, input.Now) {
		return ports.ConnectorRegistration{}, ports.ErrPrintDenied
	}
	if _, exists := s.printingConnectors[c.ID]; exists {
		return ports.ConnectorRegistration{}, ports.ErrPrintConflict
	}
	if _, exists := s.auditRecords[input.Audit.ID]; exists {
		return ports.ConnectorRegistration{}, ports.ErrPrintConflict
	}
	devices := map[string]bool{}
	printers := map[printing.PrinterID]bool{}
	for _, b := range input.Registration.Bindings {
		printer, exists := s.printingPrinters[b.PrinterID]
		if !exists || printer.Scope != c.Scope || printer.Retired || b.Scope != c.Scope || b.ConnectorID != c.ID || b.DeviceID == "" || devices[b.DeviceID] || printers[b.PrinterID] {
			return ports.ConnectorRegistration{}, ports.ErrPrintConflict
		}
		devices[b.DeviceID] = true
		printers[b.PrinterID] = true
	}
	if s.printingConnectors == nil {
		s.printingConnectors = map[printing.ConnectorID]printing.Connector{}
	}
	if s.printingBindings == nil {
		s.printingBindings = map[string]printing.PrinterBinding{}
	}
	s.printingPairings[p.ID] = clonePrintPairing(p)
	s.printingConnectors[c.ID] = clonePrintConnector(c)
	for _, b := range input.Registration.Bindings {
		s.printingBindings[printingBindingKey(c.ID, b.PrinterID)] = b
	}
	s.auditRecords[input.Audit.ID] = input.Audit
	return s.printRegistrationLocked(c), nil
}
func (s *Store) ConsumePrintPairing(_ context.Context, input ports.PairingExchange) (printing.Connector, error) {
	id, now, credentialHash, expires, activation := input.PairingID, input.Now, input.CredentialHash, input.CredentialExpiresAt, input.ActivationDeadline
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.printingPairings[id]
	if !ok || !p.Consume(now) {
		return printing.Connector{}, ports.ErrPrintDenied
	}
	c, ok := s.printingConnectors[p.ConnectorID]
	if !ok || c.Scope != p.Scope || (!p.Rotation && c.State != printing.ConnectorPending) || (p.Rotation && (c.State == printing.ConnectorRevoked || c.State == printing.ConnectorPending)) || c.Generation != c.SyncedGeneration {
		return printing.Connector{}, ports.ErrPrintDenied
	}
	if credentialHash == "" || !expires.After(now) || !activation.After(now) {
		return printing.Connector{}, ports.ErrPrintConflict
	}
	if p.Rotation && (c.CredentialVersion != p.RotationVersion || (c.PendingCredentialHash != "" && c.PendingActivationDeadline.After(now))) {
		return printing.Connector{}, ports.ErrPrintConflict
	}
	if p.Rotation {
		c.PendingCredentialHash = credentialHash
		c.PendingCredentialVersion = c.CredentialVersion + 1
		c.PendingCredentialExpiresAt = expires
		c.PendingActivationDeadline = activation
		c.PendingPublicKey = append([]byte(nil), p.PublicKey...)
	} else {
		c.CredentialHash = credentialHash
		c.CredentialExpiresAt = expires
		c.ActivationDeadline = activation
		c.State = printing.ConnectorAwaitingActivation
	}
	c.UpdatedAt = now
	if _, exists := s.auditRecords[input.Audit.ID]; exists {
		return printing.Connector{}, ports.ErrPrintConflict
	}
	s.auditRecords[input.Audit.ID] = input.Audit
	s.printingPairings[id] = p
	s.printingConnectors[c.ID] = clonePrintConnector(c)
	if p.Rotation {
		return clonePrintConnector(c.PendingCredentialIdentity()), nil
	}
	return clonePrintConnector(c), nil
}

func (s *Store) ApprovePrintCredentialRotation(_ context.Context, input ports.RotationApproval) (printing.Pairing, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.printingPairings[input.PairingID]
	if !ok || subtle.ConstantTimeCompare([]byte(p.CodeHash), []byte(input.CodeHash)) != 1 || !p.Approve(input.Scope, input.ConnectorID, input.Now) {
		return printing.Pairing{}, ports.ErrPrintDenied
	}
	c, ok := s.printingConnectors[input.ConnectorID]
	if !ok || c.Scope != input.Scope {
		return printing.Pairing{}, ports.ErrPrintNotFound
	}
	if c.Generation != input.Generation || c.Generation != c.SyncedGeneration || c.State == printing.ConnectorRevoked || c.State == printing.ConnectorPending {
		return printing.Pairing{}, ports.ErrPrintConflict
	}
	if _, exists := s.auditRecords[input.Audit.ID]; exists {
		return printing.Pairing{}, ports.ErrPrintConflict
	}
	if c.PendingCredentialHash != "" && c.PendingActivationDeadline.After(input.Now) {
		return printing.Pairing{}, ports.ErrPrintConflict
	}
	p.RotationVersion = c.CredentialVersion
	p.Rotation = true
	s.printingPairings[p.ID] = clonePrintPairing(p)
	s.auditRecords[input.Audit.ID] = input.Audit
	return clonePrintPairing(p), nil
}
