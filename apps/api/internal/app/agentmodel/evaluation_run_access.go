package agentmodel

import (
	"context"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func authorizeEvaluationRunAccess(ctx context.Context, authorizer ports.Authorizer, input EvaluationRunAccess) error {
	if input.Principal.ID == "" {
		return apperrors.ErrUnauthenticated
	}
	if authorizer == nil {
		return apperrors.ErrPrecondition
	}
	if err := authorizer.CheckTenant(ctx, input.Principal, ports.TenantPermissionConfigure, input.TenantID); err != nil {
		return err
	}
	if strings.TrimSpace(input.TenantID.String()) == "" {
		return apperrors.ErrValidation
	}
	return nil
}
