package voice

import (
	"encoding/json"
	"io"
	"net/http"
)

const googleMaxErrorResponseBytes = 64 * 1024

// Classify only Google's structured reason; never retain arbitrary error text.
func googleBillingDisabled(response *http.Response) bool {
	if response.StatusCode != http.StatusForbidden {
		return false
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, googleMaxErrorResponseBytes+1))
	if err != nil || len(body) > googleMaxErrorResponseBytes {
		return false
	}
	var payload struct {
		Error struct {
			Details []struct {
				Type   string `json:"@type"`
				Reason string `json:"reason"`
				Domain string `json:"domain"`
			} `json:"details"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &payload) != nil {
		return false
	}
	for _, detail := range payload.Error.Details {
		if detail.Type == "type.googleapis.com/google.rpc.ErrorInfo" && detail.Domain == "googleapis.com" && detail.Reason == "BILLING_DISABLED" {
			return true
		}
	}
	return false
}
