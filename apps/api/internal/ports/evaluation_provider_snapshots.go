package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

// SnapshotEvaluationProviders resolves only the supplied workflow's model steps.
// Identities cover effective non-secret configuration and credential version.
// Implementations must not return secrets, silently substitute profiles, or call a model.
type EvaluationProviderSnapshotResolver interface {
	SnapshotEvaluationProviders(context.Context, tenant.ID, agentmodel.WorkflowRevision) ([]agentmodel.EvaluationRunProvider, error)
}
