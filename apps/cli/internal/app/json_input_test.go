package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func TestJSONInputDiagnosticsLocateErrorsWithoutExposingInput(t *testing.T) {
	for _, tt := range []struct{ name, body, detail string }{
		{"empty", " \n\t", "empty"},
		{"non-object", `["private-secret"]`, "object"},
		{"syntax", " \n{\"private-secret\":!}", "byte 21"},
		{"trailing-value", `{} {"private-secret":true}`, "byte 4"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := (Runner{InputFiles: requestInput(tt.body)}).prepareInput(context.Background(), Options{Command: []string{"assets", "update", "asset"}, InputPath: "-"})
			var failure *ports.Error
			if !errors.As(err, &failure) || failure.Category != "input" || !strings.Contains(failure.Message, tt.detail) {
				t.Fatalf("missing safe input diagnostic: %v", err)
			}
			if strings.Contains(failure.Message, "private-secret") {
				t.Fatal("input contents exposed")
			}
		})
	}
	body := " \n{\"customFields\":{\"serial\":9007199254740993},\"parentAssetId\":null} \n"
	got, err := (Runner{InputFiles: requestInput(body)}).prepareInput(context.Background(), Options{Command: []string{"assets", "update", "asset"}, InputPath: "-"})
	if err != nil || string(got.RequestBody) != body {
		t.Fatalf("valid input changed: %v", err)
	}
}
