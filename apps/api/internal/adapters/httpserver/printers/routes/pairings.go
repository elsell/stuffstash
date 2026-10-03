package routes

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/dto"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/mapper"
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/shared"
	"github.com/stuffstash/stuff-stash/internal/app"
	"github.com/stuffstash/stuff-stash/internal/app/printregistry"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"strings"
)

func registerPairings(api huma.API, application app.App) {
	huma.Post(api, "/print-connector-pairings", func(ctx context.Context, input *dto.BeginPairingInput) (*dto.PairingStartedOutput, error) {
		if !application.PrintConnectorsConfigured() || application.PrintConnectors().Policy.PublicWebBaseURL == "" {
			return nil, huma.Error503ServiceUnavailable("Print connectors are unavailable")
		}
		candidates := make([]printing.PairingCandidate, 0, len(input.Body.Candidates))
		for _, c := range input.Body.Candidates {
			candidates = append(candidates, printing.PairingCandidate{ID: c.ID, Name: c.Name, AdapterID: c.AdapterID, DeviceID: c.DeviceID})
		}
		result, err := application.PrintConnectors().Begin(ctx, printregistry.BeginPairing{Rotation: input.Body.Rotation, Name: input.Body.Name, PublicKey: input.Body.PublicKey, Candidates: candidates})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.PairingStartedOutput{CacheControl: "no-store", Body: shared.SuccessEnvelope[dto.PairingStarted]{Data: dto.PairingStarted{ID: string(result.Pairing.ID), PollToken: result.PollToken, UserCode: result.UserCode, VerificationURL: result.VerificationURL, ExpiresAt: result.Pairing.ExpiresAt}}}, nil
	}, huma.OperationTags("printing"), shared.CreatedOperation)
	huma.Get(api, "/print-connector-pairings/{pairingId}", func(ctx context.Context, input *dto.PairingInput) (*dto.PairingStatusOutput, error) {
		if !application.PrintConnectorsConfigured() {
			return nil, huma.Error503ServiceUnavailable("Print connectors are unavailable")
		}
		p, err := application.PrintConnectors().Poll(ctx, printing.PairingID(input.PairingID), input.PollToken)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.PairingStatusOutput{CacheControl: "no-store", Body: shared.SuccessEnvelope[dto.PairingStatus]{Data: dto.PairingStatus{ID: string(p.ID), State: string(p.State), ExpiresAt: p.ExpiresAt}}}, nil
	}, huma.OperationTags("printing"))
	huma.Post(api, "/print-connector-pairings/{pairingId}/approval", func(ctx context.Context, input *dto.ApprovePairingInput) (*dto.ConnectorOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		if !application.PrintConnectorsConfigured() {
			return nil, huma.Error503ServiceUnavailable("Print connectors are unavailable")
		}
		bindings := make([]printregistry.PairingBinding, 0, len(input.Body.Bindings))
		for _, b := range input.Body.Bindings {
			bindings = append(bindings, printregistry.PairingBinding{CandidateID: b.CandidateID, PrinterID: printing.PrinterID(b.PrinterID)})
		}
		r, err := application.PrintConnectors().Approve(ctx, printregistry.ApprovePairing{Actor: printregistry.Actor{Principal: principal, Scope: printing.Scope{TenantID: input.Body.TenantID, InventoryID: input.Body.InventoryID}, RequestID: input.RequestID}, PairingID: printing.PairingID(input.PairingID), UserCode: input.Body.UserCode, Bindings: bindings})
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ConnectorOutput{CacheControl: "no-store", Body: shared.SuccessEnvelope[dto.Connector]{Data: mapper.ConnectorRegistration(r)}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
	huma.Post(api, "/print-connector-pairings/{pairingId}/credential", func(ctx context.Context, input *dto.ExchangePairingInput) (*dto.PairingCredentialOutput, error) {
		if !application.PrintConnectorsConfigured() {
			return nil, huma.Error503ServiceUnavailable("Print connectors are unavailable")
		}
		result, err := application.PrintConnectors().Exchange(ctx, printing.PairingID(input.PairingID), input.PollToken, input.Body.Signature)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		c := result.Connector
		return &dto.PairingCredentialOutput{CacheControl: "no-store", Body: shared.SuccessEnvelope[dto.PairingCredential]{Data: dto.PairingCredential{Credential: result.Credential, ConnectorID: string(c.ID), TenantID: c.Scope.TenantID, InventoryID: c.Scope.InventoryID, ExpiresAt: c.CredentialExpiresAt, ActivationDeadline: c.ActivationDeadline}}}, nil
	}, huma.OperationTags("printing"))
	huma.Post(api, "/print-consumer/heartbeat", func(ctx context.Context, input *dto.HeartbeatInput) (*dto.ConnectorOutput, error) {
		if !application.PrintConnectorsConfigured() {
			return nil, huma.Error503ServiceUnavailable("Print connectors are unavailable")
		}
		scheme, token, ok := strings.Cut(input.Authorization, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") {
			return nil, huma.Error401Unauthorized("Invalid connector credential")
		}
		service := application.PrintConnectors()
		c, err := service.AuthenticateConsumer(ctx, token)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		c, err = service.Heartbeat(ctx, c)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		return &dto.ConnectorOutput{CacheControl: "no-store", Body: shared.SuccessEnvelope[dto.Connector]{Data: mapper.Connector(c)}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)
	huma.Post(api, "/print-connector-pairings/{pairingId}/review", func(ctx context.Context, input *dto.ReviewPairingInput) (*dto.PairingReviewOutput, error) {
		principal, err := shared.Authenticate(ctx, application, input.Authorization)
		if err != nil {
			return nil, err
		}
		if !application.PrintConnectorsConfigured() {
			return nil, huma.Error503ServiceUnavailable("Print connectors are unavailable")
		}
		review, err := application.PrintConnectors().Review(ctx, printregistry.Actor{Principal: principal, Scope: printing.Scope{TenantID: input.Body.TenantID, InventoryID: input.Body.InventoryID}, RequestID: input.RequestID}, printing.PairingID(input.PairingID), input.Body.UserCode)
		if err != nil {
			return nil, shared.ToHumaError(err)
		}
		candidates := make([]dto.PublicPairingCandidate, 0, len(review.Candidates))
		for _, c := range review.Candidates {
			candidates = append(candidates, dto.PublicPairingCandidate{ID: c.ID, Name: c.Name, AdapterID: c.AdapterID})
		}
		return &dto.PairingReviewOutput{CacheControl: "no-store", Body: shared.SuccessEnvelope[dto.PairingReview]{Data: dto.PairingReview{Rotation: review.Rotation, ID: string(review.ID), Name: review.Name, PublicKeyFingerprint: review.PublicKeyFingerprint, Candidates: candidates}}}, nil
	}, huma.OperationTags("printing"), shared.SecuredOperation)

}
