package app

import (
	"context"
	"testing"
)

func TestCheckoutDetailsAreExplicitAndPreserved(t *testing.T) {
	empty := ""
	for _, test := range []struct {
		o    Options
		body string
		fail bool
	}{
		{Options{Command: []string{"assets", "checkout", "asset"}}, `{}`, false},
		{Options{Command: []string{"assets", "return", "asset"}, Details: &empty}, `{"details":""}`, false},
		{Options{Command: []string{"assets", "return-details", "asset", "checkout"}}, "", true},
		{Options{Command: []string{"assets", "checkout", "asset"}, InputPath: "-", Details: &empty}, "", true},
	} {
		got, err := (Runner{InputFiles: requestInput(`{"details":null}`)}).prepareInput(context.Background(), test.o)
		if (err != nil) != test.fail || !test.fail && string(got.RequestBody) != test.body {
			t.Fatalf("input: %s %v", got.RequestBody, err)
		}
	}
}
