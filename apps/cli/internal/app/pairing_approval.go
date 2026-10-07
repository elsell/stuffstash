package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func (r Runner) pairingApprovalCommand(ctx context.Context, o Options, token string) error {
	input, err := readPairingInput(o)
	if err != nil {
		return err
	}
	if rotationApproval(o) || o.InputPath == "" {
		input.Review.TenantID, input.Review.InventoryID = o.Scope.Tenant, o.Scope.Inventory
	} else if err := pairingScopeMatches(o, input.Review); err != nil {
		return err
	}
	if r.PairingApprovalAPI == nil {
		return ports.Failure("configuration", "Pairing approval is not available. Update the CLI.")
	}
	api, err := r.PairingApprovalAPI(o.Server, token)
	if err != nil {
		return err
	}
	reviewBody, err := json.Marshal(input.Review)
	if err != nil {
		return err
	}
	if o.Command[3] == "review" && o.InputPath != "" {
		reviewBody = o.RequestBody
	}
	review, err := api.ReviewPrintPairing(ctx, input.PairingID, reviewBody)
	if err != nil {
		return err
	}
	if review.Data.ID != input.PairingID {
		return ports.Failure("protocol", "The server returned a different pairing. Stop and verify the pairing ID.")
	}
	if o.Command[3] == "review" {
		r.Observer.Event(ctx, "cli.print_pairing.review.completed")
		return r.Output.Result(review)
	}
	if review.Data.Rotation != rotationApproval(o) {
		return ports.Failure("usage", "The pairing type does not match this command. Use pairings approve for registration or rotations approve for credential replacement.")
	}
	if err := r.describePairingReview(o, review.Data); err != nil {
		return err
	}
	if !rotationApproval(o) {
		if o.InputPath == "" {
			input.Bindings, err = r.choosePairingBindings(ctx, o, token, review.Data)
			if err != nil {
				return err
			}
		}
		if err := validateReviewedBindings(review.Data, input.Bindings); err != nil {
			return err
		}
		for _, b := range input.Bindings {
			if err := r.Output.Notice("Bind candidate " + strconv.Quote(b.CandidateID) + " to printer " + strconv.Quote(b.PrinterID)); err != nil {
				return err
			}
		}
	} else {
		if err := r.Output.Notice("Connector: " + strconv.Quote(o.Command[4]) + ". Generation: " + strconv.FormatUint(input.Generation, 10) + ". Approval replaces this connector's credential."); err != nil {
			return err
		}
	}
	if err := r.confirmAction(ctx, o, "Approve reviewed pairing", "Approve", "Approve the displayed pairing, scope and bindings or credential replacement."); err != nil {
		return err
	}
	body := o.RequestBody
	if o.InputPath == "" {
		if rotationApproval(o) {
			body, err = json.Marshal(pairingRotationInput{Schema: input.Review.Schema, Generation: input.Generation, PairingID: input.PairingID, UserCode: input.Review.UserCode})
		} else {
			body, err = json.Marshal(pairingApprovalInput{pairingReviewInput: input.Review, Bindings: input.Bindings})
		}
		if err != nil {
			return err
		}
	}
	var result any
	if rotationApproval(o) {
		result, err = api.ApproveConnectorRotation(ctx, o.Scope, o.Command[4], body)
	} else {
		result, err = api.ApprovePrintPairing(ctx, input.PairingID, body)
	}
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var failure *ports.Error
		if errors.As(err, &failure) {
			switch failure.Category {
			case "conflict":
				return ports.Failure("conflict", "Approval conflicts with current state. Examine the connector and review the pairing before you try again. Supply the expected generation.")
			case "network", "protocol", "unavailable", "api":
				return ports.Failure(failure.Category, "Approval outcome is unknown. Examine connectors print list or show before attempting approval again.")
			}
		}
		return err
	}
	r.Observer.Event(ctx, "cli.print_pairing.approval.completed")
	return r.Output.Result(result)
}
func (r Runner) describePairingReview(o Options, v ports.PairingReview) error {
	if err := r.Output.Notice("Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Inventory: " + strconv.Quote(o.Scope.Inventory) + ". Pairing: " + strconv.Quote(v.ID) + ". Name: " + strconv.Quote(v.Name) + ". Fingerprint: " + strconv.Quote(v.PublicKeyFingerprint) + ". Rotation: " + strconv.FormatBool(v.Rotation)); err != nil {
		return err
	}
	for _, c := range v.Candidates {
		if err := r.Output.Notice("Candidate: " + strconv.Quote(c.ID) + ". Name: " + strconv.Quote(c.Name) + ". Adapter: " + strconv.Quote(c.AdapterID)); err != nil {
			return err
		}
	}
	return nil
}
func validateReviewedBindings(review ports.PairingReview, bindings []ports.PairingBinding) error {
	if err := validatePairingBindings(bindings); err != nil {
		return err
	}
	known := map[string]bool{}
	for _, c := range review.Candidates {
		if c.ID == "" || known[c.ID] {
			return ports.Failure("protocol", "The pairing review has invalid candidate IDs. Review the pairing again.")
		}
		known[c.ID] = true
	}
	for _, b := range bindings {
		if !known[b.CandidateID] {
			return ports.Failure("usage", "A binding names a candidate absent from the pairing review. Review the pairing and correct the input.")
		}
	}
	return nil
}
