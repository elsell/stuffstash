package inventories

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/netip"
	"net/url"
	"strings"

	"github.com/stuffstash/stuff-stash/internal/app/apperrors"
	"github.com/stuffstash/stuff-stash/internal/app/appsupport"
	"github.com/stuffstash/stuff-stash/internal/domain/audit"
	"github.com/stuffstash/stuff-stash/internal/domain/identity"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a Service) CreateInventoryAccessInvitation(ctx context.Context, input CreateInventoryAccessInvitationInput) (CreateInventoryAccessInvitationResult, error) {
	if err := a.EnsureActiveInventoryAccess(ctx, input.Principal, input.TenantID, input.InventoryID, ports.InventoryPermissionShare); err != nil {
		return CreateInventoryAccessInvitationResult{}, err
	}
	email, ok := identity.NewEmail(input.Email)
	if !ok {
		return CreateInventoryAccessInvitationResult{}, apperrors.ErrInvalidInput
	}
	relationship, ok := inventoryAccessRelationship(input.Relationship)
	if !ok {
		return CreateInventoryAccessInvitationResult{}, apperrors.ErrInvalidInput
	}
	acceptanceToken, err := newInventoryInvitationToken()
	if err != nil {
		return CreateInventoryAccessInvitationResult{}, err
	}

	invitation := ports.InventoryAccessInvitation{
		ID:                 a.ids.NewID(),
		TenantID:           input.TenantID,
		InventoryID:        input.InventoryID,
		Email:              email,
		TokenHash:          HashInventoryInvitationToken(acceptanceToken),
		Relationship:       relationship,
		Status:             ports.InventoryAccessInvitationPending,
		InviterPrincipalID: input.Principal.ID,
		ExpiresAt:          a.clock.Now().Add(a.invitationTTL),
	}
	inviteURL, err := BuildInventoryInvitationURL(a.invitationPublicBaseURL, invitation, acceptanceToken, a.invitationAllowInsecureHTTP)
	if err != nil {
		return CreateInventoryAccessInvitationResult{}, err
	}
	auditRecord, err := appsupport.NewAuditRecord(a.ids, a.clock, appsupport.AuditRecordInput{
		Principal:   input.Principal,
		TenantID:    input.TenantID,
		InventoryID: input.InventoryID,
		Source:      input.Source,
		RequestID:   input.RequestID,
		Action:      audit.ActionInventoryInvitationCreated,
		TargetType:  audit.TargetInventoryInvitation,
		TargetID:    invitation.ID,
		Metadata: map[string]string{
			"relationship": string(relationship),
		},
	})
	if err != nil {
		return CreateInventoryAccessInvitationResult{}, err
	}

	saved, err := a.inventoryAccessUnitOfWork.SaveInventoryAccessInvitation(ctx, invitation, auditRecord)
	if err != nil {
		if errors.Is(err, ports.ErrForbidden) || errors.Is(err, ports.ErrConflict) {
			return CreateInventoryAccessInvitationResult{}, apperrors.ErrInvalidInput
		}
		return CreateInventoryAccessInvitationResult{}, err
	}
	a.observer.Record(ctx, ports.Event{
		Name:    ports.EventInventoryInvitationCreated,
		Message: "inventory invitation created",
		Fields: map[string]string{
			"tenant_id":    input.TenantID.String(),
			"inventory_id": input.InventoryID.String(),
			"principal_id": input.Principal.ID.String(),
			"relationship": string(relationship),
			"status":       string(saved.Status),
		},
	})
	return CreateInventoryAccessInvitationResult{
		Invitation: saved,
		InviteURL:  inviteURL,
	}, nil
}

func NormalizeInvitationPublicBaseURL(value string) string {
	return strings.TrimSpace(value)
}

func BuildInventoryInvitationURL(baseURL string, invitation ports.InventoryAccessInvitation, acceptanceToken string, allowInsecureLocalHTTP bool) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("invalid invitation public base URL")
	}
	if parsed.Scheme != "https" && !(parsed.Scheme == "http" && allowInsecureLocalHTTP && isLocalInvitationHost(parsed.Hostname())) {
		return "", errors.New("invitation public base URL must use HTTPS")
	}
	parsed.Path = "/invitations/accept"
	parsed.RawPath = ""
	query := parsed.Query()
	query.Set("tenant", invitation.TenantID.String())
	query.Set("inventory", invitation.InventoryID.String())
	query.Set("invitation", invitation.ID)
	parsed.RawQuery = query.Encode()
	parsed.Fragment = "token=" + acceptanceToken
	return parsed.String(), nil
}

func isLocalInvitationHost(host string) bool {
	if host == "localhost" {
		return true
	}
	address, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	if address.IsLoopback() {
		return true
	}
	if !address.Is4() {
		return false
	}
	octets := address.As4()
	return octets[0] == 10 ||
		(octets[0] == 172 && octets[1] >= 16 && octets[1] <= 31) ||
		(octets[0] == 192 && octets[1] == 168)
}

func newInventoryInvitationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func HashInventoryInvitationToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
