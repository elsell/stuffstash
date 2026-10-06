package app

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/adapters/httpapi"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const completePrintRequest = `{"printerId":"printer","expectedMediaFingerprint":"old-media","templateId":"asset","templateVersion":4294967295,"templateOptions":{"showReference":false},"copies":2,"previewFingerprint":null}`

func TestPrintInputValidationBeforeAuthentication(t *testing.T) {
	valid, err := Parse([]string{"labels", "print", "asset", "--input", "-"}, func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := (Runner{InputFiles: requestInput(completePrintRequest)}).prepareInput(context.Background(), valid)
	if err != nil || string(prepared.RequestBody) != completePrintRequest {
		t.Fatalf("complete input changed: %v", err)
	}
	for _, tc := range []struct{ old, new string }{{`"printerId":"printer"`, `"printerId":""`}, {`"expectedMediaFingerprint":"old-media"`, `"expectedMediaFingerprint":null`}, {`"templateVersion":4294967295`, `"templateVersion":4294967296`}, {`"templateVersion":4294967295`, `"templateVersion":1.5`}, {`"copies":2`, `"copies":0`}, {`"copies":2`, `"copies":9223372036854775808`}, {`"showReference":false`, `"showReference":"false"`}, {`"templateOptions":{"showReference":false}`, `"templateOptions":null`}} {
		body := strings.Replace(completePrintRequest, tc.old, tc.new, 1)
		o, err := Parse([]string{"labels", "print", "asset", "--input", "-"}, func(string) string { return "" })
		if err != nil {
			t.Fatal(err)
		}
		// Missing credential port ensures malformed selections stop before login.
		if err = (Runner{InputFiles: requestInput(body)}).Run(context.Background(), o); err == nil {
			t.Fatalf("invalid selection accepted: %s", body)
		}
	}
	for _, flag := range []string{"--printer", "--template", "--template-version", "--copies", "--show-reference", "--expected-media-fingerprint", "--preview-fingerprint"} {
		value := "1"
		if flag == "--show-reference" {
			value = "false"
		}
		if _, err := Parse([]string{"labels", "print", "asset", "--input", "-", flag + "=" + value}, func(string) string { return "" }); err == nil {
			t.Fatalf("mixed input: %s", flag)
		}
	}
	o, err := Parse([]string{"printers", "test", "different", "--input", "-"}, func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	if err = (Runner{InputFiles: requestInput(completePrintRequest)}).Run(context.Background(), o); err == nil {
		t.Fatal("printer target mismatch accepted")
	}
	for _, args := range [][]string{{"labels", "print", "asset", "--limit", "2"}, {"printers", "test", "printer", "--printer", "other"}, {"labels", "render", "asset", "--preview-fingerprint", "preview"}} {
		if _, err := Parse(args, func(string) string { return "" }); err == nil {
			t.Fatalf("irrelevant or conflicting flag accepted: %v", args)
		}
	}
}

func TestPrintSubmissionFlagsPreserveReviewedSelection(t *testing.T) {
	for _, command := range []string{"labels print asset", "printers test printer", "print-jobs reprint prior"} {
		t.Run(command, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.Method == "GET" {
					if r.URL.Path != "/tenants/home/inventories/garage/print-settings" {
						t.Error("reviewed media fingerprint triggered a printer lookup")
					}
					io.WriteString(w, `{"data":{"defaultPrinterId":"default","template":{"id":"default-template","version":1,"options":{"showReference":true}}},"meta":{}}`)
					return
				}
				var input struct {
					Version  uint32 `json:"templateVersion"`
					Copies   int64  `json:"copies"`
					Expected string `json:"expectedMediaFingerprint"`
					Preview  string `json:"previewFingerprint"`
					Options  struct {
						Show bool `json:"showReference"`
					} `json:"templateOptions"`
				}
				if json.NewDecoder(r.Body).Decode(&input) != nil || input.Version != 4294967295 || input.Copies != 2 || input.Expected != "reviewed-media" || input.Preview != "reviewed-preview" || input.Options.Show {
					t.Errorf("selection narrowed or replaced: %+v", input)
				}
				io.WriteString(w, `{"data":{"id":"job"},"meta":{}}`)
			}))
			defer server.Close()
			api, err := httpapi.New(server.URL, "owner", server.Client())
			if err != nil {
				t.Fatal(err)
			}
			args := append(strings.Fields(command), "--tenant", "home", "--inventory", "garage", "--printer", "printer", "--template", "asset", "--template-version", "4294967295", "--copies", "2", "--show-reference=false", "--expected-media-fingerprint", "reviewed-media", "--preview-fingerprint", "reviewed-preview")
			o, err := Parse(args, func(string) string { return "" })
			if err != nil {
				t.Fatal(err)
			}
			if _, err = executePrinting(context.Background(), api, o); err != nil || requests != 2 {
				t.Fatalf("request count=%d error=%v", requests, err)
			}
		})
	}
}
