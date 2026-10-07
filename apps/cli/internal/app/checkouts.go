package app

import (
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isCheckoutWrite(o Options) bool {
	return len(o.Command) > 1 && o.Command[0] == "assets" && (o.Command[1] == "checkout" || o.Command[1] == "return" || o.Command[1] == "return-details")
}
func prepareCheckoutInput(o Options) (Options, error) {
	if o.Details == nil && o.Command[1] == "return-details" {
		return o, ports.Failure("usage", "Supply --details TEXT or --input FILE. Use an empty details value to clear the notes.")
	}
	fields := map[string]string{}
	if o.Details != nil {
		fields["details"] = *o.Details
	}
	o.RequestBody, _ = json.Marshal(fields)
	return o, nil
}
func checkoutFailure(err error) error {
	var failure *ports.Error
	if errors.As(err, &failure) {
		switch failure.Category {
		case "network", "protocol", "unavailable", "api":
			return ports.Failure(failure.Category, "The checkout result is unknown. Run assets checkouts ASSET_ID before you try again.")
		}
	}
	return err
}
