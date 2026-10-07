package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"strings"
)

type voiceProviderInput struct {
	Schema            *string `json:"$schema,omitempty"`
	SpeechToText      *string `json:"speechToTextProfileId"`
	LanguageInference *string `json:"languageInferenceProfileId"`
	TextToSpeech      *string `json:"textToSpeechProfileId"`
}

func isVoiceProviderUpdate(o Options) bool {
	return len(o.Command) == 2 && o.Command[0] == "voice-provider" && o.Command[1] == "update"
}
func decodeVoiceProviderInput(body []byte) (voiceProviderInput, error) {
	var v voiceProviderInput
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	if !json.Valid(body) || d.Decode(&v) != nil {
		return v, ports.Failure("usage", "Voice selection input accepts only speechToTextProfileId, languageInferenceProfileId and textToSpeechProfileId as strings or null, plus optional $schema.")
	}
	return v, nil
}
func (r Runner) prepareVoiceProviderUpdate(o Options) (Options, error) {
	if o.InputPath == "" {
		if r.Picker == nil || o.JSON || o.NoInput {
			return o, ports.Failure("usage", "Supply --input FILE or --input - with all intended voice selections and --yes for scripts. Omitted slots revert to automatic selection.")
		}
		return o, nil
	}
	_, err := decodeVoiceProviderInput(o.RequestBody)
	return o, err
}
func voiceSelectionDescription(id *string) string {
	if id == nil || strings.TrimSpace(*id) == "" {
		return "automatic (server selects a profile)"
	}
	return "explicit " + strconv.Quote(strings.TrimSpace(*id))
}
func (r Runner) updateVoiceProvider(ctx context.Context, o Options, token string, api ports.VoiceProviderAPI) error {
	if o.InputPath == "" {
		var err error
		o, err = r.guideVoiceProvider(ctx, o, token, api)
		if err != nil {
			return err
		}
	}
	input, err := decodeVoiceProviderInput(o.RequestBody)
	if err != nil {
		return err
	}
	notice := "Server: " + strconv.Quote(o.Server) + ". Household: " + strconv.Quote(o.Scope.Tenant) + ". Speech input: " + voiceSelectionDescription(input.SpeechToText) + ". Language inference: " + voiceSelectionDescription(input.LanguageInference) + ". Spoken output: " + voiceSelectionDescription(input.TextToSpeech) + ". Replace all three selections. Omitted, null or empty slots use automatic selection, not disabled voice. The API has no version check; this can replace another administrator's recent selections."
	if err := r.Output.Notice(notice); err != nil {
		return err
	}
	if err := r.confirmAction(ctx, o, "Replace voice provider selections", "Replace all selections", "Apply the three reviewed explicit or automatic selections for this household."); err != nil {
		return err
	}
	result, err := api.UpdateVoiceProviderConfiguration(ctx, o.Scope.Tenant, o.RequestBody)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var failure *ports.Error
		if errors.As(err, &failure) {
			switch failure.Category {
			case "conflict":
				return ports.Failure("conflict", "The server rejected the voice selections. Run voice-provider show and review all three choices before you try again.")
			case "network", "protocol", "unavailable", "api":
				return ports.Failure(failure.Category, "The selection update result is unknown. Run voice-provider show before you try again; do not assume the previous selections remain.")
			}
		}
		return err
	}
	r.Observer.Event(ctx, "cli.voice_provider.selection.updated")
	return r.Output.Result(result)
}
