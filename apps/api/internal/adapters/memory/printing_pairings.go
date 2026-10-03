package memory

import (
	"context"
	"crypto/subtle"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"time"
)

func clonePrintPairing(p printing.Pairing) printing.Pairing {
	p.PublicKey = append([]byte(nil), p.PublicKey...)
	p.Candidates = append([]printing.PairingCandidate(nil), p.Candidates...)
	return p
}
func clonePrintConnector(c printing.Connector) printing.Connector {
	c.PublicKey = append([]byte(nil), c.PublicKey...)
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
func (s *Store) ConsumePrintPairing(_ context.Context, id printing.PairingID, now time.Time, credentialHash string, expires, activation time.Time) (printing.Connector, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.printingPairings[id]
	if !ok || !p.Consume(now) {
		return printing.Connector{}, ports.ErrPrintDenied
	}
	c, ok := s.printingConnectors[p.ConnectorID]
	if !ok || c.Scope != p.Scope || c.State != printing.ConnectorPending || c.Generation != c.SyncedGeneration {
		return printing.Connector{}, ports.ErrPrintDenied
	}
	if credentialHash == "" || !expires.After(now) || !activation.After(now) {
		return printing.Connector{}, ports.ErrPrintConflict
	}
	c.CredentialHash = credentialHash
	c.CredentialExpiresAt = expires
	c.ActivationDeadline = activation
	c.State = printing.ConnectorAwaitingActivation
	c.UpdatedAt = now
	s.printingPairings[id] = p
	s.printingConnectors[c.ID] = clonePrintConnector(c)
	return clonePrintConnector(c), nil
}
