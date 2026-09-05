package ports

import (
	"context"

	"github.com/stuffstash/stuff-stash/internal/domain/tenant"
)

// Explicit workflow profile resolution never selects an alternative profile.
type WorkflowLanguageProviderResolutionInput struct {
	TenantID  tenant.ID
	ProfileID string
}

type WorkflowLanguageProviderBinding struct {
	ProfileID      string
	PromptTemplate string
	Provider       RealtimeLanguageProvider
}

type WorkflowLanguageProviderResolver interface {
	ResolveWorkflowLanguageProvider(context.Context, WorkflowLanguageProviderResolutionInput) (WorkflowLanguageProviderBinding, error)
}
