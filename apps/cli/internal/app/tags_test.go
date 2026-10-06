package app

import (
	"context"
	"testing"
)

func TestTagInputPreservesEmptyColorAndRejectsMixedSources(t *testing.T) {
	o, err := Parse([]string{"tags", "update", "tag-id", "--tag-color", ""}, func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	result, err := (Runner{}).prepareInput(context.Background(), o)
	if err != nil || string(result.RequestBody) != `{"color":""}` {
		t.Fatalf("empty color lost: %s %v", result.RequestBody, err)
	}
	if _, err := Parse([]string{"tags", "list", "--limit", "0"}, func(string) string { return "" }); err != nil {
		t.Fatal(err)
	}
	o.InputPath = "-"
	if _, err := (Runner{InputFiles: requestInput(`{}`)}).prepareInput(context.Background(), o); err == nil {
		t.Fatal("mixed body and field accepted")
	}
}

func TestTagCommandsRejectAssetOnlyPrintFlag(t *testing.T) {
	if err := validateCommandShape(Options{Command: []string{"tags", "create"}, PrintLabel: true}); err == nil {
		t.Fatal("unsupported print flag accepted")
	}
}

func TestUnhandledCommandCannotReachLegacyAssetExecutor(t *testing.T) {
	if _, err := execute(context.Background(), nil, Options{Command: []string{"tags", "update", "id"}}); err == nil {
		t.Fatal("unhandled domain accepted by asset executor")
	}
}
