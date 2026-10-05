package httpapi

import (
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi/generated"
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func consumerError(err error) error {
	var failure *ports.Error
	if errors.As(err, &failure) {
		if failure.Category == "authentication" {
			return ports.Failure("authentication", "connector credential expired or revoked; pair this connector again")
		}
		if failure.Category == "forbidden" {
			return ports.Failure("authorization", "connector access was removed; review its printer assignments before pairing again")
		}
	}
	return err
}
func (c *Client) Claim(ctx context.Context, printerID string, control printing.AttemptControl) (*printing.Claim, error) {
	result, err := read[ports.Result[workerAttempt]](c.sdk.PostPrintConsumerClaims(ctx, nil, generated.PrintClaimRequest{PrinterId: printerID, AttemptId: control.AttemptID, SessionId: control.SessionID, ClaimToken: control.ClaimToken}))
	if err != nil {
		return nil, consumerError(err)
	}
	value := result.Data
	if value.AttemptID == "" && value.JobID == "" {
		return nil, nil
	}
	if value.Artifact == nil || value.Media == nil || !value.LeaseValid || value.Status != "claimed" || value.Revision <= 0 {
		return nil, ports.Failure("protocol", "invalid claimed print job")
	}
	artifact := value.Artifact
	control.AttemptID = value.AttemptID
	control.SessionID = value.SessionID
	control.Revision = value.Revision
	return &printing.Claim{Control: control, JobID: value.JobID, PrinterID: value.PrinterID, ContractVersion: int(value.ProtocolVersion), LeaseExpiresAt: value.LeaseExpiresAt, Media: consumerMedia(*value.Media), Copies: int(value.Copies), Artifact: printing.Artifact{SHA256: artifact.SHA256, ContentType: artifact.ContentType, ByteLength: artifact.ByteLength, Width: int(artifact.WidthPixels), Height: int(artifact.HeightPixels)}}, nil
}
func (c *Client) Start(ctx context.Context, control printing.AttemptControl) (printing.AttemptStatus, error) {
	body, err := workerRequest(proof(control))
	if err != nil {
		return printing.AttemptStatus{}, err
	}
	result, err := read[ports.Result[workerAttempt]](c.sdk.PostPrintConsumerClaimsByAttemptIdStartWithBody(ctx, control.AttemptID, nil, "application/json", body))
	if err != nil {
		return printing.AttemptStatus{}, consumerError(err)
	}
	return attemptStatus(result.Data)
}
func (c *Client) Renew(ctx context.Context, control printing.AttemptControl) (printing.AttemptStatus, error) {
	body, err := workerRequest(proof(control))
	if err != nil {
		return printing.AttemptStatus{}, err
	}
	result, err := read[ports.Result[workerAttempt]](c.sdk.PostPrintConsumerClaimsByAttemptIdRenewalWithBody(ctx, control.AttemptID, nil, "application/json", body))
	if err != nil {
		return printing.AttemptStatus{}, consumerError(err)
	}
	return attemptStatus(result.Data)
}
func (c *Client) Outcome(ctx context.Context, control printing.AttemptControl, evidence printing.Evidence) error {
	body, err := workerRequest(workerOutcome{workerProof: proof(control), Outcome: outcome(evidence)})
	if err != nil {
		return err
	}
	_, err = read[ports.Result[workerAttempt]](c.sdk.PostPrintConsumerClaimsByAttemptIdOutcomeWithBody(ctx, control.AttemptID, nil, "application/json", body))
	return consumerError(err)
}
func (c *Client) Attempt(ctx context.Context, id string) (printing.AttemptStatus, error) {
	result, err := read[ports.Result[workerAttempt]](c.sdk.GetPrintConsumerAttemptsByAttemptId(ctx, id, nil))
	var failure *ports.Error
	if errors.As(err, &failure) && failure.Category == "not_found" {
		return printing.AttemptStatus{}, ports.ErrAttemptNotFound
	}
	if err != nil {
		return printing.AttemptStatus{}, consumerError(err)
	}
	return attemptStatus(result.Data)
}
func (c *Client) Reconcile(ctx context.Context, id string, revision uint64, evidence printing.Evidence) error {
	body, err := workerRequest(workerReconciliation{Revision: revision, Outcome: outcome(evidence)})
	if err != nil {
		return err
	}
	_, err = read[ports.Result[workerAttempt]](c.sdk.PostPrintConsumerAttemptsByAttemptIdReconciliationWithBody(ctx, id, nil, "application/json", body))
	return consumerError(err)
}
func (c *Client) Unsettled(ctx context.Context, printerID string) ([]printing.AttemptStatus, error) {
	result := make([]printing.AttemptStatus, 0)
	cursor := ""
	limit := int64(100)
	status := generated.GetPrintConsumerAttemptsParamsStatus("unsettled")
	seen := map[string]bool{}
	for {
		response, err := read[ports.Result[[]workerAttempt]](c.sdk.GetPrintConsumerAttempts(ctx, &generated.GetPrintConsumerAttemptsParams{PrinterId: &printerID, Status: &status, Limit: &limit, Cursor: &cursor}))
		if err != nil {
			return nil, consumerError(err)
		}
		for _, value := range response.Data {
			if value.PrinterID != printerID {
				return nil, ports.Failure("protocol", "unexpected printer in recovery response")
			}
			attempt, err := attemptStatus(value)
			if err != nil {
				return nil, err
			}
			result = append(result, attempt)
		}
		var pagination *ports.Pagination
		if response.Meta != nil {
			pagination = response.Meta.Pagination
		}
		if pagination == nil || !pagination.HasMore {
			return result, nil
		}
		if pagination.NextCursor == nil || *pagination.NextCursor == "" || seen[*pagination.NextCursor] {
			return nil, ports.Failure("protocol", "invalid print recovery pagination")
		}
		cursor = *pagination.NextCursor
		seen[cursor] = true
	}
}
func (c *Client) Artifact(ctx context.Context, control printing.AttemptControl, maximum int64) ([]byte, string, error) {
	if maximum <= 0 || maximum >= 1<<30 {
		return nil, "", ports.Failure("configuration", "invalid print artifact byte limit")
	}
	response, err := c.sdk.ListPrintConsumerClaimsByAttemptIdContent(ctx, control.AttemptID, &generated.ListPrintConsumerClaimsByAttemptIdContentParams{XPrintSessionID: control.SessionID, XPrintClaimToken: control.ClaimToken, XPrintRevision: 0}, func(_ context.Context, request *http.Request) error {
		// The generated header parameter is signed; overwrite it before transport.
		request.Header.Set("X-Print-Revision", strconv.FormatUint(control.Revision, 10))
		return nil
	})
	if err != nil {
		return nil, "", ports.Failure("network", "could not fetch print artifact")
	}
	if response.StatusCode != http.StatusOK {
		_, err = read[ports.Result[workerAttempt]](response, nil)
		if err == nil {
			err = ports.Failure("protocol", "unexpected print artifact status")
		}
		return nil, "", consumerError(err)
	}
	defer response.Body.Close()
	contentType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || contentType != "image/png" || response.ContentLength > maximum {
		return nil, "", ports.Failure("protocol", "unsupported print artifact response")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maximum+1))
	if err != nil || int64(len(body)) > maximum {
		return nil, "", ports.Failure("protocol", "print artifact exceeds its byte limit or was interrupted")
	}
	return body, contentType, nil
}

func (c *Client) ConfirmIdle(ctx context.Context, id string, revision uint64) error {
	body, err := workerRequest(workerRevision{Revision: revision})
	if err != nil {
		return err
	}
	_, err = read[ports.Result[workerAttempt]](c.sdk.PostPrintConsumerAttemptsByAttemptIdIdleConfirmationWithBody(ctx, id, nil, "application/json", body))
	return consumerError(err)
}
