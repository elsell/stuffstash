package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func assetPrintRequested(o Options) bool {
	if o.PrintLabel {
		return true
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(o.RequestBody, &fields) != nil {
		return false
	}
	value := bytes.TrimSpace(fields["printLabel"])
	return len(value) > 0 && !bytes.Equal(value, []byte("null"))
}
func validateAssetCreateKey(o Options) error {
	if isAssetWrite(o) && o.Command[1] == "create" && o.IdempotencyKey != "" && !assetPrintRequested(o) {
		return ports.Failure("usage", "Ordinary asset creation does not support retry keys. Remove --idempotency-key.")
	}
	return nil
}
func assetCreateFailure(o Options, err error) error {
	var failure *ports.Error
	if !errors.As(err, &failure) {
		return err
	}
	switch failure.Category {
	case "network", "protocol", "unavailable", "api":
		if assetPrintRequested(o) {
			return ports.Failure(failure.Category, "The create result is unknown. Retry with the same request key and unchanged print selection and asset fields.")
		}
		return ports.Failure(failure.Category, "The create result is unknown. Run assets list before you retry.")
	default:
		return err
	}
}
