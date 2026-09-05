package agentmodel

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	domain "github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

type EvaluationCaseDependencies struct {
	Authorizer       ports.Authorizer
	Repository       ports.EvaluationCaseRepository
	Audit            ports.AuditRepository
	IDs              ports.IDGenerator
	Clock            ports.Clock
	Observer         ports.Observer
	DefaultPageLimit int
	MaxPageLimit     int
}
type EvaluationCaseService struct{ deps EvaluationCaseDependencies }

func NewEvaluationCaseService(deps EvaluationCaseDependencies) EvaluationCaseService {
	if deps.Observer == nil {
		deps.Observer = noopObserver{}
	}
	if deps.MaxPageLimit <= 0 || deps.MaxPageLimit > ports.MaxEvaluationCasePageLimit {
		deps.MaxPageLimit = ports.MaxEvaluationCasePageLimit
	}
	if deps.DefaultPageLimit <= 0 {
		deps.DefaultPageLimit = 50
	}
	deps.DefaultPageLimit = min(deps.DefaultPageLimit, deps.MaxPageLimit)
	return EvaluationCaseService{deps: deps}
}

type EvaluationCaseAccess struct {
	Principal identity.Principal
	TenantID  tenant.ID
	Source    audit.Source
	RequestID string
}
type SaveEvaluationCaseInput struct {
	EvaluationCaseAccess
	CaseID           domain.EvaluationCaseID
	ExpectedRevision int
	Definition       domain.EvaluationCaseDefinitionInput
}

func (s EvaluationCaseService) authorize(ctx context.Context, input EvaluationCaseAccess) error {
	if input.Principal.ID.String() == "" {
		return apperrors.ErrUnauthenticated
	}
	if s.deps.Authorizer == nil {
		return apperrors.ErrPrecondition
	}
	if err := s.deps.Authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		return err
	}
	if s.deps.Repository == nil || s.deps.IDs == nil || s.deps.Clock == nil {
		return apperrors.ErrPrecondition
	}
	if strings.TrimSpace(input.TenantID.String()) == "" {
		return apperrors.ErrValidation
	}
	return nil
}
func (s EvaluationCaseService) SaveRevision(ctx context.Context, input SaveEvaluationCaseInput) (domain.EvaluationCaseRevision, error) {
	if err := s.authorize(ctx, input.EvaluationCaseAccess); err != nil {
		return domain.EvaluationCaseRevision{}, err
	}
	if input.ExpectedRevision < 0 || (input.CaseID == "" && input.ExpectedRevision != 0) {
		return domain.EvaluationCaseRevision{}, apperrors.ErrValidation
	}
	definition, err := domain.NewEvaluationCaseDefinition(input.Definition)
	if err != nil {
		return domain.EvaluationCaseRevision{}, apperrors.ErrValidation
	}
	caseID := input.CaseID
	if caseID == "" {
		caseID = domain.EvaluationCaseID(s.deps.IDs.NewID())
	} else {
		head, found, err := s.deps.Repository.EvaluationCaseHead(ctx, input.TenantID, caseID)
		if err != nil {
			return domain.EvaluationCaseRevision{}, err
		}
		if !found {
			return domain.EvaluationCaseRevision{}, apperrors.ErrNotFound
		}
		if head.LatestRevision != input.ExpectedRevision {
			return domain.EvaluationCaseRevision{}, apperrors.ErrConflict
		}
	}
	revision, err := domain.NewEvaluationCaseRevision(domain.EvaluationCaseRevisionInput{ID: domain.EvaluationCaseRevisionID(s.deps.IDs.NewID()), CaseID: caseID, TenantID: domain.TenantID(input.TenantID), AuthorID: domain.EvaluationCaseAuthorID(input.Principal.ID), Number: input.ExpectedRevision + 1, Definition: definition, CreatedAt: s.deps.Clock.Now()})
	if err != nil {
		return domain.EvaluationCaseRevision{}, apperrors.ErrValidation
	}
	snapshot := revision.Snapshot()
	record, err := s.auditRecord(input.EvaluationCaseAccess, audit.ActionConversationEvaluationCaseRevisionCreated, string(caseID), map[string]string{"revision_id": string(snapshot.ID), "revision_number": strconv.Itoa(snapshot.Number)})
	if err != nil {
		return domain.EvaluationCaseRevision{}, err
	}
	if err := s.deps.Repository.AppendEvaluationCaseRevision(ctx, revision, input.ExpectedRevision, record); err != nil {
		if errors.Is(err, ports.ErrEvaluationCaseConflict) {
			err = apperrors.ErrConflict
		}
		return domain.EvaluationCaseRevision{}, err
	}
	s.deps.Observer.Record(ctx, ports.Event{Name: ports.EventConversationEvaluationCaseRevisionCreated, Message: "evaluation case revision created", Fields: map[string]string{"tenant_id": input.TenantID.String(), "case_id": string(caseID), "revision_id": string(snapshot.ID)}})
	return revision, nil
}
func (s EvaluationCaseService) auditRecord(input EvaluationCaseAccess, action audit.Action, targetID string, metadata map[string]string) (audit.Record, error) {
	record, ok := audit.NewRecord(audit.ID(s.deps.IDs.NewID()), audit.TenantID(input.TenantID), "", audit.PrincipalID(input.Principal.ID), action, input.Source, "conversation_evaluation_case", targetID, s.deps.Clock.Now(), input.RequestID, metadata)
	if !ok {
		return audit.Record{}, apperrors.ErrValidation
	}
	return record, nil
}
