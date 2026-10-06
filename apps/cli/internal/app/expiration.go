package app

import (
	"flag"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"time"
	"unicode/utf8"
)

func expirationFlags(f *flag.FlagSet, o *Options) {
	q := &o.Expiration
	f.StringVar(&q.Mode, "mode", "", "expiration mode: all, soon, or expired")
	f.StringVar(&q.CheckoutState, "checkout-state", "", "any, available, or checked_out")
	f.StringVar(&q.Query, "query", "", "search expiration items")
	f.StringVar(&q.TypeID, "type-id", "", "custom asset type ID")
	f.StringVar(&q.LocationID, "location-id", "", "location ID")
	f.StringVar(&q.FromDate, "from-date", "", "first expiration date, YYYY-MM-DD")
	f.StringVar(&q.ThroughDate, "through-date", "", "last expiration date, YYYY-MM-DD")
	f.Func("tag-id", "required tag ID; repeat for more tags", func(v string) error {
		if v == "" {
			return ports.Failure("usage", "Supply a tag ID.")
		}
		q.TagIDs = append(q.TagIDs, v)
		return nil
	})
}
func validateExpiration(o Options) error {
	q := o.Expiration
	selected := len(o.Command) == 2 && o.Command[0] == "assets" && o.Command[1] == "expiration"
	has := q.Mode != "" || q.CheckoutState != "" || q.Query != "" || q.TypeID != "" || q.LocationID != "" || q.FromDate != "" || q.ThroughDate != "" || len(q.TagIDs) > 0
	if !selected {
		if has {
			return ports.Failure("usage", "Use expiration filters only with assets expiration.")
		}
		return nil
	}
	if q.Mode != "" && q.Mode != "all" && q.Mode != "soon" && q.Mode != "expired" {
		return ports.Failure("usage", "Use --mode all, soon, or expired.")
	}
	if o.Kind != "" && o.Kind != "item" && o.Kind != "container" && o.Kind != "location" {
		return ports.Failure("usage", "Use --kind item, container, or location.")
	}
	if q.CheckoutState != "" && q.CheckoutState != "any" && q.CheckoutState != "available" && q.CheckoutState != "checked_out" {
		return ports.Failure("usage", "Use --checkout-state any, available, or checked_out.")
	}
	if o.Page.Limit < 1 || o.Page.Limit > 100 {
		return ports.Failure("usage", "Use --limit from 1 to 100 for expiration items.")
	}
	if utf8.RuneCountInString(q.Query) > 120 {
		return ports.Failure("usage", "Keep --query to 120 characters or fewer.")
	}
	if utf8.RuneCountInString(q.TypeID) > 128 || utf8.RuneCountInString(q.LocationID) > 128 {
		return ports.Failure("usage", "Keep type and location IDs to 128 characters or fewer.")
	}
	for _, v := range []string{q.FromDate, q.ThroughDate} {
		if v != "" {
			if _, err := time.Parse("2006-01-02", v); err != nil {
				return ports.Failure("usage", "Use a valid date in YYYY-MM-DD format.")
			}
		}
	}
	if q.FromDate != "" && q.ThroughDate != "" && q.FromDate > q.ThroughDate {
		return ports.Failure("usage", "Set --through-date on or after --from-date.")
	}
	return nil
}
