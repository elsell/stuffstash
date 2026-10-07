package app

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"path/filepath"
	"strconv"
)

type importSourceInput struct {
	Schema              *string `json:"$schema,omitempty"`
	SourceType          string  `json:"sourceType"`
	BaseURL             *string `json:"baseUrl,omitempty"`
	Username            *string `json:"username,omitempty"`
	Password            *string `json:"password,omitempty"`
	IncludeImages       *bool   `json:"includeImages,omitempty"`
	AllowInsecureTLS    *bool   `json:"allowInsecureTLS,omitempty"`
	AllowPrivateNetwork *bool   `json:"allowPrivateNetwork,omitempty"`
	FileName            *string `json:"fileName,omitempty"`
	ContentBase64       *string `json:"contentBase64,omitempty"`
}

func isImportSource(o Options) bool {
	return isImportJobCommand(o) && len(o.Command) > 1 && (o.Command[1] == "preview" || o.Command[1] == "start")
}
func (r Runner) prepareImportSource(ctx context.Context, o Options) (Options, error) {
	if o.InputPath == "" {
		if o.NoInput || o.JSON || r.Picker == nil || r.TextInput == nil {
			return o, ports.Failure("usage", "Supply --input FILE|- with the import source JSON. Keep source credentials out of command arguments.")
		}
		v, err := r.guidedImportSource(ctx)
		if err != nil {
			return o, err
		}
		o.RequestBody, err = json.Marshal(v)
		if err != nil {
			return o, err
		}
	}
	var v importSourceInput
	if err := decodePortability(o.RequestBody, &v); err != nil {
		return o, err
	}
	if v.SourceType != "legacy_homebox" && v.SourceType != "legacy_homebox_csv" {
		return o, ports.Failure("usage", "Set sourceType to legacy_homebox or legacy_homebox_csv in the input JSON.")
	}
	return o, nil
}
func (r Runner) guidedImportSource(ctx context.Context) (importSourceInput, error) {
	var v importSourceInput
	kind, err := r.Picker.Pick(ctx, "Import source", []ports.Choice{{ID: "legacy_homebox", Label: "Live legacy Homebox", Detail: "The server contacts this source"}, {ID: "legacy_homebox_csv", Label: "Legacy Homebox CSV", Detail: "Send an exported CSV file"}})
	if err != nil {
		return v, err
	}
	if kind != "legacy_homebox" && kind != "legacy_homebox_csv" {
		return v, context.Canceled
	}
	v.SourceType = kind
	if kind == "legacy_homebox_csv" {
		if r.StreamFiles == nil {
			return v, ports.Failure("configuration", "File input is not available. Update the CLI.")
		}
		path, e := r.TextInput.ReadText(ctx, "CSV file path", 4096)
		if e != nil {
			return v, e
		}
		content, e := r.readImportFile(ctx, path, maximumImportCSVBytes)
		if e != nil {
			return v, e
		}
		name := filepath.Base(path)
		encoded := base64.StdEncoding.EncodeToString(content)
		v.FileName = &name
		v.ContentBase64 = &encoded
	} else {
		if r.SecretInput == nil {
			return v, ports.Failure("usage", "Masked credential input is not available. Use a protected JSON file with --input.")
		}
		url, e := r.TextInput.ReadText(ctx, "Source base URL (no credentials in URL)", 2048)
		if e != nil {
			return v, e
		}
		v.BaseURL = &url
		username, e := r.SecretInput.ReadSecret(ctx, "Source username", 1024)
		if e != nil {
			return v, e
		}
		v.Username = &username
		password, e := r.SecretInput.ReadSecret(ctx, "Source password", 4096)
		if e != nil {
			return v, e
		}
		v.Password = &password
	}
	images, err := r.portabilityChoice(ctx, "Import supported images?", "Include images", "Do not include images")
	if err != nil {
		return v, err
	}
	v.IncludeImages = &images
	if kind == "legacy_homebox" {
		private, e := r.portabilityChoice(ctx, "Allow the server to contact private/local source networks?", "Explicitly allow private networks", "Keep private networks blocked")
		if e != nil {
			return v, e
		}
		v.AllowPrivateNetwork = &private
		insecure, e := r.portabilityChoice(ctx, "Allow untrusted source TLS certificates?", "Explicitly allow untrusted TLS", "Keep TLS verification required")
		if e != nil {
			return v, e
		}
		v.AllowInsecureTLS = &insecure
	}
	return v, nil
}
func (r Runner) importSourceCommand(ctx context.Context, o Options, token string) error {
	if r.ImportSources == nil {
		return ports.Failure("configuration", "Import source commands are not available. Update the CLI.")
	}
	api, err := r.ImportSources(o.Server, token)
	if err != nil {
		return err
	}
	action := o.Command[1]
	detail := "Create an import preview job. The server reads the supplied source; inventory records are not imported yet."
	if action == "start" {
		detail = "Start the reviewed import into this inventory. Supply the same source and security options used for preview; the server checks the source fingerprint."
	}
	var source importSourceInput
	_ = json.Unmarshal(o.RequestBody, &source)
	if source.AllowPrivateNetwork != nil && *source.AllowPrivateNetwork {
		detail += " Private-network access is explicitly enabled."
	}
	if source.AllowInsecureTLS != nil && *source.AllowInsecureTLS {
		detail += " Untrusted TLS certificates are explicitly allowed."
	}
	if err = r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Inventory: " + strconv.Quote(o.Scope.Inventory) + ". " + detail); err != nil {
		return err
	}
	if err = r.confirmAction(ctx, o, "Import: "+action, "Continue", detail); err != nil {
		return err
	}
	var result ports.Result[ports.ImportJob]
	if action == "preview" {
		result, err = api.PreviewImportJob(ctx, o.Scope, o.RequestBody)
	} else {
		result, err = api.StartImportJob(ctx, o.Scope, o.Command[2], o.RequestBody)
	}
	if err != nil {
		return err
	}
	r.Observer.Event(ctx, "cli.import_job."+action+".completed")
	return r.Output.Result(result)
}
