package app

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"strings"
)

type pairingReviewInput struct {
	Schema      *string `json:"$schema,omitempty"`
	UserCode    string  `json:"userCode"`
	TenantID    string  `json:"tenantId"`
	InventoryID string  `json:"inventoryId"`
}
type pairingApprovalInput struct {
	pairingReviewInput
	Bindings []ports.PairingBinding `json:"bindings"`
}
type pairingRotationInput struct {
	Schema     *string `json:"$schema,omitempty"`
	Generation uint64  `json:"generation"`
	PairingID  string  `json:"pairingId"`
	UserCode   string  `json:"userCode"`
}
type pairingInput struct {
	Review     pairingReviewInput
	Bindings   []ports.PairingBinding
	Generation uint64
	PairingID  string
}

func isPairingApproval(o Options) bool {
	return len(o.Command) >= 3 && o.Command[0] == "connectors" && o.Command[1] == "print" && (o.Command[2] == "pairings" || o.Command[2] == "rotations")
}
func rotationApproval(o Options) bool { return o.Command[2] == "rotations" }
func validatePairingCommand(o Options, scope bool) error {
	if len(o.Command) != 5 || o.Command[4] == "" || !(o.Command[3] == "approve" || o.Command[2] == "pairings" && o.Command[3] == "review") {
		return ports.Failure("usage", "Use connectors print pairings review|approve PAIRING_ID or connectors print rotations approve CONNECTOR_ID.")
	}
	if scope && (o.Scope.Tenant == "" || o.Scope.Inventory == "") {
		return ports.Failure("usage", "Choose the household and inventory for this pairing.")
	}
	return nil
}
func pairingApprovalFlags(flags *flag.FlagSet, o Options) error {
	if !isPairingApproval(o) {
		return nil
	}
	if err := validatePairingCommand(o, false); err != nil {
		return err
	}
	invalid := ""
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "server", "tenant", "inventory", "context", "credential-file", "allow-loopback-http", "json", "no-input", "request-id", "color", "help", "input":
		case "yes":
			if o.Command[3] != "approve" {
				invalid = f.Name
			}
		default:
			invalid = f.Name
		}
	})
	if invalid != "" {
		return ports.Failure("usage", "Pairing review and approval do not accept --"+invalid+". Use structured input for the user code and request fields.")
	}
	return nil
}
func decodePairing(body []byte, target any) error {
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	if !json.Valid(body) || d.Decode(target) != nil {
		return ports.Failure("usage", "Pairing input has unknown fields or incorrect field types. Use the request shape in --help.")
	}
	return nil
}
func pairingScopeMatches(o Options, v pairingReviewInput) error {
	if o.Scope.Tenant != "" && v.TenantID != o.Scope.Tenant || o.Scope.Inventory != "" && v.InventoryID != o.Scope.Inventory {
		return ports.Failure("usage", "Pairing input scope differs from the selected household or inventory. Review the scope and correct the input.")
	}
	return nil
}
func readPairingInput(o Options) (pairingInput, error) {
	var v pairingInput
	if rotationApproval(o) {
		var input pairingRotationInput
		if err := decodePairing(o.RequestBody, &input); err != nil {
			return v, err
		}
		v = pairingInput{Review: pairingReviewInput{Schema: input.Schema, UserCode: input.UserCode}, Generation: input.Generation, PairingID: input.PairingID}
		if input.Generation == 0 || strings.TrimSpace(input.PairingID) == "" {
			return v, ports.Failure("usage", "Rotation approval requires pairingId and a positive uint64 generation.")
		}
	} else if o.Command[3] == "approve" {
		var input pairingApprovalInput
		if err := decodePairing(o.RequestBody, &input); err != nil {
			return v, err
		}
		v.Review, v.Bindings, v.PairingID = input.pairingReviewInput, input.Bindings, o.Command[4]
		if o.InputPath != "" {
			if err := validatePairingBindings(v.Bindings); err != nil {
				return v, err
			}
		}
	} else {
		if err := decodePairing(o.RequestBody, &v.Review); err != nil {
			return v, err
		}
		v.PairingID = o.Command[4]
	}
	if strings.TrimSpace(v.Review.UserCode) == "" {
		return v, ports.Failure("usage", "Pairing input requires userCode. Use --input FILE|- or an interactive terminal.")
	}
	if !rotationApproval(o) && o.InputPath != "" {
		if strings.TrimSpace(v.Review.TenantID) == "" || strings.TrimSpace(v.Review.InventoryID) == "" {
			return v, ports.Failure("usage", "Pairing input requires tenantId and inventoryId.")
		}
		if err := pairingScopeMatches(o, v.Review); err != nil {
			return v, err
		}
	}
	return v, nil
}
func validatePairingBindings(bindings []ports.PairingBinding) error {
	if len(bindings) < 1 || len(bindings) > 16 {
		return ports.Failure("usage", "Supply one to sixteen printer bindings.")
	}
	candidates, printers := map[string]bool{}, map[string]bool{}
	for _, b := range bindings {
		if strings.TrimSpace(b.CandidateID) == "" || strings.TrimSpace(b.PrinterID) == "" || candidates[b.CandidateID] || printers[b.PrinterID] {
			return ports.Failure("usage", "Each binding needs a unique candidateId and a unique printerId.")
		}
		candidates[b.CandidateID], printers[b.PrinterID] = true, true
	}
	return nil
}
func (r Runner) preparePairingApproval(ctx context.Context, o Options) (Options, error) {
	if o.InputPath == "" {
		if o.NoInput || o.JSON || r.SecretInput == nil || r.Picker == nil {
			return o, ports.Failure("usage", "Supply pairing JSON with --input FILE|- or use an interactive terminal.")
		}
		code, err := r.SecretInput.ReadSecret(ctx, "Pairing user code", 4095)
		if err != nil {
			return o, err
		}
		if rotationApproval(o) {
			if r.TextInput == nil {
				return o, ports.Failure("configuration", "Interactive text input is unavailable. Use --input FILE|-.")
			}
			id, err := r.TextInput.ReadText(ctx, "Pairing ID", 4095)
			if err != nil {
				return o, err
			}
			raw, err := r.TextInput.ReadText(ctx, "Current connector generation", 20)
			if err != nil {
				return o, err
			}
			generation, err := strconv.ParseUint(raw, 10, 64)
			if err != nil || generation == 0 {
				return o, ports.Failure("usage", "Enter the exact positive connector generation from connectors print show.")
			}
			o.RequestBody, _ = json.Marshal(pairingRotationInput{UserCode: code, PairingID: id, Generation: generation})
		} else {
			o.RequestBody, _ = json.Marshal(pairingApprovalInput{pairingReviewInput: pairingReviewInput{UserCode: code}})
			if o.Command[3] == "review" {
				o.RequestBody, _ = json.Marshal(pairingReviewInput{UserCode: code})
			}
		}
	}
	if _, err := readPairingInput(o); err != nil {
		return o, err
	}
	if o.Command[3] == "approve" && !o.Yes && (o.NoInput || o.JSON || r.Picker == nil) {
		return o, ports.Failure("usage", "Approval needs confirmation. Review the pairing and add --yes.")
	}
	return o, nil
}
