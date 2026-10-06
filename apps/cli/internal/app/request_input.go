package app

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func (r Runner) prepareInput(ctx context.Context, o Options) (Options, error) {
	if isArchiveBody(o) || isImportSource(o) {
		return r.preparePortability(ctx, o)
	}
	write := acceptsBody(o)
	if o.InputPath != "" && !write {
		return o, ports.Failure("usage", "This command does not accept --input. Remove the option.")
	}
	if !write {
		return o, nil
	}
	if o.IdempotencyKey != "" && !isPrintSubmission(o) && !isPrinterCreation(o) && !(isAssetWrite(o) && o.Command[1] == "create") {
		return o, ports.Failure("usage", "This API operation does not support --idempotency-key. Remove the option.")
	}
	if o.InputPath != "" && (o.Timezone != "" || o.ConnectorName != "" || o.TagColor != nil || o.TagKey != nil || o.Title != "" || o.Kind != "" || o.Parent != "" || o.PrintLabel || o.Details != nil) {
		return o, ports.Failure("usage", "Use either field options or --input. Do not combine them.")
	}
	if o.InputPath != "" {
		if r.InputFiles == nil {
			return o, ports.Failure("configuration", "Input files are not available. Use field options instead.")
		}
		body, err := r.InputFiles.Read(ctx, o.InputPath)
		if err != nil {
			return o, err
		}
		if err := validateJSONObject(body); err != nil {
			return o, err
		}
		o.RequestBody = body
	} else {
		if isProviderWrite(o) {
			return r.prepareProviderWrite(ctx, o)
		}
		if customizationBody(o) {
			return r.prepareCustomization(ctx, o)
		}
		if isPairingApproval(o) {
			return r.preparePairingApproval(ctx, o)
		}
		if isTelemetry(o) {
			return prepareTelemetry(o)
		}
		if isEvaluationWrite(o) {
			return prepareEvaluationWrite(o)
		}
		if isLabelRender(o) {
			return prepareLabelRender(o)
		}
		if isVoiceProviderUpdate(o) {
			return r.prepareVoiceProviderUpdate(o)
		}
		if isPrinterAdministration(o) {
			return r.preparePrinterAdministration(ctx, o)
		}
		if isWorkflowWrite(o) {
			return prepareWorkflowInput(o)
		}
		if isPrintSubmission(o) {
			return o, nil
		}
		if isEvaluationCancellation(o) {
			return r.prepareEvaluationCancellation(o)
		}
		if isPrintResolution(o) {
			return r.preparePrintResolution(ctx, o)
		}
		if isImportCancel(o) {
			return r.prepareImportCancel(ctx, o)
		}
		if isInvitationTokenCommand(o) {
			return r.prepareInvitationToken(ctx, o)
		}
		if isInvitationExpiration(o) {
			return r.prepareInvitationExpiration(ctx, o)
		}
		if isGrantCreate(o) {
			return r.prepareGrantInput(ctx, o)
		}
		if isDeviceRegistration(o) {
			return r.prepareDeviceRegistration(ctx, o)
		}
		if isPreferenceWrite(o) {
			return r.preparePreferenceInput(ctx, o)
		}
		if isCheckoutWrite(o) {
			return prepareCheckoutInput(o)
		}
		if isAssetWrite(o) {
			return r.prepareAssetInput(ctx, o)
		}
		if isTagWrite(o) {
			return r.prepareTagInput(ctx, o)
		}
		if o.ConnectorName == "" && r.TextInput != nil && !o.NoInput && !o.JSON {
			title := "Household name"
			if o.Command[0] == "inventories" {
				title = "Inventory name"
			}
			name, err := r.TextInput.ReadText(ctx, title, 120)
			if err != nil {
				return o, err
			}
			o.ConnectorName = name
		}
		if o.ConnectorName == "" {
			return o, ports.Failure("usage", "Supply --name NAME or --input FILE for this command.")
		}
		o.RequestBody, _ = json.Marshal(map[string]string{"name": o.ConnectorName})
	}
	if isProviderWrite(o) {
		return r.prepareProviderWrite(ctx, o)
	}
	if customizationBody(o) {
		return r.prepareCustomization(ctx, o)
	}
	if isPairingApproval(o) {
		return r.preparePairingApproval(ctx, o)
	}
	if isTelemetry(o) {
		return prepareTelemetry(o)
	}
	if isEvaluationWrite(o) {
		return prepareEvaluationWrite(o)
	}
	if isLabelRender(o) {
		return prepareLabelRender(o)
	}
	if isVoiceProviderUpdate(o) {
		return r.prepareVoiceProviderUpdate(o)
	}
	if isPrinterAdministration(o) {
		return r.preparePrinterAdministration(ctx, o)
	}
	if isWorkflowWrite(o) {
		return prepareWorkflowInput(o)
	}
	if isPrintSubmission(o) {
		_, err := decodePrintSubmission(o)
		return o, err
	}
	if isEvaluationCancellation(o) {
		return r.prepareEvaluationCancellation(o)
	}
	if isImportCancel(o) {
		_, err := importCancellationMode(o.RequestBody)
		return o, err
	}
	if isInvitationTokenCommand(o) {
		return o, validateInvitationToken(o.RequestBody)
	}
	if isInvitationExpiration(o) {
		_, err := invitationExpiration(o.RequestBody)
		return o, err
	}
	if isGrantCreate(o) {
		_, err := decodeGrant(o.RequestBody)
		return o, err
	}
	if isPrintResolution(o) {
		_, err := decodePrintResolution(o.RequestBody, true)
		return o, err
	}
	return o, validateAssetCreateKey(o)
}
