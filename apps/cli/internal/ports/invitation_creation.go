package ports

import "context"

type CreatedInvitation struct {
	Invitation
	InviteURL string `json:"inviteUrl"`
}
type InvitationWriter interface {
	CreateInvitation(context.Context, Scope, []byte) (Result[CreatedInvitation], error)
}
