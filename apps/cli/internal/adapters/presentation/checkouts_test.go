package presentation

import (
	"bytes"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strings"
	"testing"
)

func TestHumanHistoryShowsDetailsNeededForRecovery(t *testing.T) {
	var out bytes.Buffer
	empty, notes, date := "", "With charger", "2026-10-05"
	err := (Output{Stdout: &out}).Result(ports.Result[[]ports.Checkout]{Data: []ports.Checkout{{ID: "checkout", State: "returned", CheckoutDetails: &notes, ReturnDetails: &empty, ReturnedAt: &date}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"With charger", "Return notes:", `""`, date} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %s in %s", want, &out)
		}
	}
}
