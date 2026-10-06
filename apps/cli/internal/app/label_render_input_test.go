package app

import (
	"context"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"testing"
)

func TestLabelRenderInputValidationBeforeCredentials(t *testing.T) {
	for _, body := range []string{`{}`, `{"format":"svg","media":{},"template":{}}`, `{"format":"png","media":false,"template":{}}`, `{"format":"png","media":{},"template":{"id":"x","version":4294967296,"options":{"show_reference":false}}}`, `{"unknown":"private"}`} {
		o, err := Parse([]string{"labels", "render", "asset", "--input", "-", "--output", "label.png", "--json"}, func(string) string { return "" })
		if err != nil {
			t.Fatal(err)
		}
		err = (Runner{InputFiles: requestInput(body)}).Run(context.Background(), o)
		var failure *ports.Error
		if !errors.As(err, &failure) || failure.Category != "usage" {
			t.Fatalf("invalid render reached credentials: %v", err)
		}
	}
	for _, flag := range []string{"format", "printer", "media-preset", "width-mm", "height-mm", "template", "template-version", "show-reference"} {
		value := "1"
		if flag == "format" {
			value = "png"
		}
		if flag == "show-reference" {
			value = "false"
		}
		if _, err := Parse([]string{"labels", "render", "asset", "--output", "label.png", "--input", "-", "--" + flag + "=" + value}, func(string) string { return "" }); err == nil {
			t.Fatalf("mixed selection accepted: %s", flag)
		}
	}
}
