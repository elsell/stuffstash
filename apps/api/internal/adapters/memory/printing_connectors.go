package memory

import (
	"context"
	"crypto/subtle"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"sort"
	"time"
)

func (s *Store) printRegistrationLocked(c printing.Connector) ports.ConnectorRegistration {
	result := ports.ConnectorRegistration{Connector: clonePrintConnector(c), Bindings: []printing.PrinterBinding{}}
	for _, b := range s.printingBindings {
		if b.ConnectorID == c.ID && b.Scope == c.Scope {
			result.Bindings = append(result.Bindings, b)
		}
	}
	sort.Slice(result.Bindings, func(i, j int) bool { return result.Bindings[i].PrinterID < result.Bindings[j].PrinterID })
	return result
}
func (s *Store) GetPrintConnector(_ context.Context, scope printing.Scope, id printing.ConnectorID) (ports.ConnectorRegistration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.printingConnectors[id]
	if !ok || c.Scope != scope {
		return ports.ConnectorRegistration{}, ports.ErrPrintNotFound
	}
	return s.printRegistrationLocked(c), nil
}
func (s *Store) ListPrintConnectors(_ context.Context, scope printing.Scope, limit int, after string) ([]ports.ConnectorRegistration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []ports.ConnectorRegistration{}
	for _, c := range s.printingConnectors {
		if c.Scope == scope && string(c.ID) > after {
			result = append(result, s.printRegistrationLocked(c))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Connector.ID < result[j].Connector.ID })
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}
func (s *Store) FindPrintConnectorCredential(_ context.Context, hash string) (printing.Connector, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if hash != "" {
		for _, c := range s.printingConnectors {
			if c.PendingCredentialHash == hash && c.State != printing.ConnectorRevoked {
				return clonePrintConnector(c.PendingCredentialIdentity()), nil
			}
			if c.CredentialHash == hash {
				return clonePrintConnector(c), nil
			}
		}
	}
	return printing.Connector{}, ports.ErrPrintDenied
}
func (s *Store) SynchronizePrintConnector(ctx context.Context, scope printing.Scope, id printing.ConnectorID, sync ports.ConnectorAuthorizationSync) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.printingConnectors[id]
	if !ok || c.Scope != scope {
		return ports.ErrPrintNotFound
	}
	registration := s.printRegistrationLocked(c)
	printers := []printing.Printer{}
	for _, b := range registration.Bindings {
		p, ok := s.printingPrinters[b.PrinterID]
		if !ok || p.Scope != scope {
			return ports.ErrPrintDenied
		}
		printers = append(printers, clonePrintingPrinter(p))
	}
	if err := sync(ctx, registration, printers); err != nil {
		return err
	}
	c.SyncedGeneration = c.Generation
	s.printingConnectors[id] = c
	for _, b := range registration.Bindings {
		b.SyncedGeneration = b.Generation
		s.printingBindings[printingBindingKey(id, b.PrinterID)] = b
	}
	return nil
}
func (s *Store) PendingPrintConnectorScopes(_ context.Context, limit int) ([]printing.Connector, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []printing.Connector{}
	for _, c := range s.printingConnectors {
		if c.Generation != c.SyncedGeneration {
			out = append(out, clonePrintConnector(c))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
func (s *Store) HeartbeatPrintConnector(_ context.Context, authenticated printing.Connector, now time.Time, makeAudit ports.ConnectorActivationAudit) (printing.Connector, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.printingConnectors[authenticated.ID]
	if !ok || c.Scope != authenticated.Scope || c.ServiceAccountID != authenticated.ServiceAccountID || c.Generation != c.SyncedGeneration || c.State == printing.ConnectorRevoked {
		return printing.Connector{}, ports.ErrPrintDenied
	}
	activating := c.State == printing.ConnectorAwaitingActivation
	pending := c.PendingCredentialHash != "" && c.PendingCredentialVersion == authenticated.CredentialVersion && subtle.ConstantTimeCompare([]byte(c.PendingCredentialHash), []byte(authenticated.CredentialHash)) == 1
	if pending {
		if !c.PendingCredentialExpiresAt.After(now) || !c.PendingActivationDeadline.After(now) {
			return printing.Connector{}, ports.ErrPrintDenied
		}
		c.ActivatePending()
	} else {
		if c.CredentialVersion != authenticated.CredentialVersion || subtle.ConstantTimeCompare([]byte(c.CredentialHash), []byte(authenticated.CredentialHash)) != 1 || !c.CredentialExpiresAt.After(now) {
			return printing.Connector{}, ports.ErrPrintDenied
		}
		if c.State == printing.ConnectorAwaitingActivation && c.ActivationDeadline.After(now) {
			c.State = printing.ConnectorActive
		} else if c.State != printing.ConnectorActive {
			return printing.Connector{}, ports.ErrPrintDenied
		}
	}
	if pending || activating {
		if makeAudit == nil {
			return printing.Connector{}, ports.ErrPrintConflict
		}
		record, err := makeAudit(c, pending)
		if err != nil {
			return printing.Connector{}, err
		}
		if _, exists := s.auditRecords[record.ID]; exists {
			return printing.Connector{}, ports.ErrPrintConflict
		}
		s.auditRecords[record.ID] = record
	}
	c.LastSeenAt = &now
	c.UpdatedAt = now
	s.printingConnectors[c.ID] = clonePrintConnector(c)
	return clonePrintConnector(c), nil
}
func (s *Store) ReportPrintPrinter(_ context.Context, authority printing.ConsumerAuthority, report printing.PrinterReport, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.printingConsumerFenceLocked(authority, now); err != nil {
		return err
	}
	if report.Scope != authority.Scope || report.ConnectorID != authority.ConnectorID || report.PrinterID != authority.PrinterID {
		return ports.ErrPrintDenied
	}
	if s.printingReports == nil {
		s.printingReports = map[string]printing.PrinterReport{}
	}
	report.ReportedAt = now
	s.printingReports[printingBindingKey(report.ConnectorID, report.PrinterID)] = report
	return nil
}

var _ ports.ConnectorRepository = (*Store)(nil)

func (s *Store) UpdatePrintConnector(_ context.Context, scope printing.Scope, id printing.ConnectorID, generation uint64, change ports.ConnectorMutation, makeAudit ports.ConnectorAudit) (ports.ConnectorRegistration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.printingConnectors[id]
	if !ok || c.Scope != scope {
		return ports.ConnectorRegistration{}, ports.ErrPrintNotFound
	}
	if c.Generation != generation {
		return ports.ConnectorRegistration{}, ports.ErrPrintConflict
	}
	r := s.printRegistrationLocked(c)
	if err := change(&r); err != nil {
		return ports.ConnectorRegistration{}, err
	}
	if r.Connector.ID != id || r.Connector.Scope != scope || r.Connector.ServiceAccountID != c.ServiceAccountID {
		return ports.ConnectorRegistration{}, ports.ErrPrintConflict
	}
	record, err := makeAudit(r)
	if err != nil {
		return ports.ConnectorRegistration{}, err
	}
	if _, exists := s.auditRecords[record.ID]; exists {
		return ports.ConnectorRegistration{}, ports.ErrPrintConflict
	}
	s.printingConnectors[id] = clonePrintConnector(r.Connector)
	for _, b := range r.Bindings {
		s.printingBindings[printingBindingKey(id, b.PrinterID)] = b
	}
	s.auditRecords[record.ID] = record
	return s.printRegistrationLocked(r.Connector), nil
}

func (s *Store) ListPrintPrinterHealth(_ context.Context, scope printing.Scope, id printing.PrinterID) ([]ports.PrinterHealthReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []ports.PrinterHealthReport{}
	for _, report := range s.printingReports {
		if report.Scope != scope || report.PrinterID != id {
			continue
		}
		c, ok := s.printingConnectors[report.ConnectorID]
		if !ok || c.Scope != scope {
			continue
		}
		b, ok := s.printingBindings[printingBindingKey(c.ID, id)]
		if !ok || b.Scope != scope {
			continue
		}
		result = append(result, ports.PrinterHealthReport{Connector: clonePrintConnector(c), Binding: b, Report: report})
	}
	return result, nil
}
