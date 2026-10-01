package inventories

import (
	"fmt"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
)

var (
	ErrInvitationInvalid       = fmt.Errorf("%w: invalid invitation", apperrors.ErrNotFound)
	ErrInvitationEmailMismatch = fmt.Errorf("%w: invitation email mismatch", apperrors.ErrUnauthorized)
)
