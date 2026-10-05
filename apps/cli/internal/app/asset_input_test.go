package app

import (
	"context"
	"testing"
)

func TestAssetInputPreservesFullRequestAndRejectsMixedFields(t *testing.T) {
	body := `{"expiration":null,"parentAssetId":null,"tagIds":[],"customFields":{"serial":9007199254740993}}`
	r := Runner{InputFiles: requestInput(body)}
	o := Options{Command: []string{"assets", "update", "asset"}, InputPath: "-"}
	got, err := r.prepareInput(context.Background(), o)
	if err != nil || string(got.RequestBody) != body {
		t.Fatalf("request lost: %s %v", got.RequestBody, err)
	}
	o.Title = "Other"
	if _, err := r.prepareInput(context.Background(), o); err == nil {
		t.Fatal("mixed fields accepted")
	}
	for _, o := range []Options{{Command: []string{"assets", "create"}, JSON: true}, {Command: []string{"assets", "update", "asset"}, NoInput: true}} {
		if _, err := r.prepareInput(context.Background(), o); err == nil {
			t.Fatal("missing scripted fields accepted")
		}
	}
}

func TestAssetTitlesPromptButUpdateRetryKeysAreRejected(t *testing.T) {
	prompt := &namePrompt{}
	runner := Runner{TextInput: prompt, Picker: firstScopeChoice{}}
	o, err := runner.prepareInput(context.Background(), Options{Command: []string{"assets", "create"}})
	if err != nil || o.Title != "Garage" || o.Kind != "item" || prompt.calls != 1 {
		t.Fatalf("missing title prompt: %+v %v", o, err)
	}
	_, err = runner.prepareInput(context.Background(), Options{Command: []string{"assets", "update", "asset"}, Title: "New", IdempotencyKey: "retry"})
	if err == nil {
		t.Fatal("unsupported update retry key accepted")
	}
}
