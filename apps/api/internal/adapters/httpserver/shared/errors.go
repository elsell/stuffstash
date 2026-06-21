package shared

import (
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stuffstash/stuff-stash/internal/app"
)

type ErrorEnvelope struct {
	status    int
	BodyError ErrorBody `json:"error"`
	Meta      Meta      `json:"meta"`
}

func (e *ErrorEnvelope) Error() string {
	return e.BodyError.Message
}

func (e *ErrorEnvelope) GetStatus() int {
	return e.status
}

type ErrorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

func NewErrorEnvelope(status int, msg string, errs ...error) huma.StatusError {
	return &ErrorEnvelope{
		status: status,
		BodyError: ErrorBody{
			Code:    errorCode(status),
			Message: safeErrorMessage(status, msg),
			Details: safeErrorDetails(status, errs),
		},
		Meta: Meta{},
	}
}

func ToHumaError(err error) error {
	switch {
	case errors.Is(err, app.ErrUnauthenticated):
		return huma.Error401Unauthorized("Authentication required.")
	case errors.Is(err, app.ErrUnauthorized):
		return huma.Error403Forbidden("Forbidden.")
	case errors.Is(err, app.ErrValidation), errors.Is(err, app.ErrInvalidInput):
		return huma.Error400BadRequest("Invalid request.")
	case errors.Is(err, app.ErrConflict):
		return huma.Error409Conflict("Conflict.")
	case errors.Is(err, app.ErrPrecondition):
		return huma.Error412PreconditionFailed("Precondition failed.")
	case errors.Is(err, app.ErrNotFound):
		return huma.Error404NotFound("Resource not found.")
	default:
		return huma.Error500InternalServerError("Internal server error.")
	}
}

func safeErrorDetails(status int, errs []error) []ErrorDetail {
	if status >= http.StatusInternalServerError {
		return []ErrorDetail{}
	}

	details := make([]ErrorDetail, 0, len(errs))
	for _, err := range errs {
		if err == nil {
			continue
		}
		details = append(details, ErrorDetail{Message: err.Error()})
	}
	return details
}

func errorCode(status int) string {
	switch status {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return "invalid_request"
	case http.StatusUnauthorized:
		return "authentication_required"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "resource_not_found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusPreconditionFailed:
		return "precondition_failed"
	case http.StatusRequestEntityTooLarge:
		return "payload_too_large"
	default:
		return "internal_error"
	}
}

func safeErrorMessage(status int, fallback string) string {
	switch status {
	case http.StatusUnauthorized:
		return "Authentication required."
	case http.StatusForbidden:
		return "Forbidden."
	case http.StatusNotFound:
		return "Resource not found."
	case http.StatusConflict:
		return "Conflict."
	case http.StatusPreconditionFailed:
		return "Precondition failed."
	case http.StatusRequestEntityTooLarge:
		return "Request body too large."
	case http.StatusInternalServerError:
		return "Internal server error."
	}
	if fallback == "" {
		return "Invalid request."
	}
	return fallback
}
