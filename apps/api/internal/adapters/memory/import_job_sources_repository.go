package memory

import (
	"context"
	"time"

	"github.com/stuffstash/stuff-stash/internal/domain/importjob"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (s *Store) ReplaceImportJobSource(_ context.Context, source ports.ImportJobSourceRecord) error {
	if err := validateImportJobSource(source); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.importJobSources[importJobSourceKey(source.Scope)] = cloneImportJobSource(source)
	return nil
}

func (s *Store) ImportJobSource(_ context.Context, scope ports.ImportJobSourceScope) (ports.ImportJobSourceRecord, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	source, ok := s.importJobSources[importJobSourceKey(scope)]
	if !ok {
		return ports.ImportJobSourceRecord{}, false, nil
	}
	return cloneImportJobSource(source), true, nil
}

func (s *Store) DeleteImportJobSource(_ context.Context, scope ports.ImportJobSourceScope) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := importJobSourceKey(scope)
	_, found := s.importJobSources[key]
	delete(s.importJobSources, key)
	return found, nil
}

func (s *Store) DeleteExpiredImportJobSources(_ context.Context, now time.Time) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	deleted := 0
	for key, source := range s.importJobSources {
		if !source.ExpiresAt.After(now) {
			delete(s.importJobSources, key)
			deleted++
		}
	}
	return deleted, nil
}

func (s *Store) DeleteVacuumableImportJobSources(_ context.Context, terminalStatuses []importjob.Status, now time.Time) ([]ports.ImportJobSourceScope, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	terminal := map[importjob.Status]struct{}{}
	for _, status := range terminalStatuses {
		terminal[status] = struct{}{}
	}
	deleted := []ports.ImportJobSourceScope{}
	for key, source := range s.importJobSources {
		job, jobFound := s.importJobs[source.Scope.JobID.String()]
		if !source.ExpiresAt.After(now) || (jobFound && job.TenantID == source.Scope.TenantID && job.InventoryID == source.Scope.InventoryID && statusInSet(job.Status, terminal)) {
			delete(s.importJobSources, key)
			deleted = append(deleted, source.Scope)
		}
	}
	return deleted, nil
}

func statusInSet(status importjob.Status, statuses map[importjob.Status]struct{}) bool {
	_, ok := statuses[status]
	return ok
}

func validateImportJobSource(source ports.ImportJobSourceRecord) error {
	if source.Scope.TenantID.String() == "" ||
		source.Scope.InventoryID.String() == "" ||
		source.Scope.JobID.String() == "" ||
		source.Sealed.KeyID == "" ||
		source.Sealed.Algorithm != ports.ProviderCredentialAlgorithmAES256GCM ||
		len(source.Sealed.Nonce) != ports.ProviderCredentialAESGCMNonceBytes ||
		len(source.Sealed.Ciphertext) == 0 ||
		source.ExpiresAt.IsZero() ||
		source.CreatedAt.IsZero() ||
		source.UpdatedAt.IsZero() {
		return ports.ErrInvalidProviderInput
	}
	return nil
}

func importJobSourceKey(scope ports.ImportJobSourceScope) string {
	return scope.TenantID.String() + "/" + scope.InventoryID.String() + "/" + scope.JobID.String()
}

func cloneImportJobSource(source ports.ImportJobSourceRecord) ports.ImportJobSourceRecord {
	source.Sealed.Nonce = append([]byte{}, source.Sealed.Nonce...)
	source.Sealed.Ciphertext = append([]byte{}, source.Sealed.Ciphertext...)
	return source
}
