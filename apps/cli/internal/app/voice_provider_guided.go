package app

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"strings"
)

type voiceSelectionSlot struct {
	capability, label string
	value             **string
}

func (r Runner) guideVoiceProvider(ctx context.Context, o Options, token string, api ports.VoiceProviderAPI) (Options, error) {
	if r.ProviderProfilesAPI == nil {
		return o, ports.Failure("configuration", "Provider profile selection is not available. Supply --input FILE with explicit choices.")
	}
	current, err := api.VoiceProviderConfiguration(ctx, o.Scope.Tenant)
	if err != nil {
		return o, err
	}
	profilesAPI, err := r.ProviderProfilesAPI(o.Server, token)
	if err != nil {
		return o, err
	}
	profiles, err := profilesAPI.ProviderProfiles(ctx, o.Scope.Tenant)
	if err != nil {
		return o, err
	}
	input := voiceProviderInput{}
	for _, slot := range []voiceSelectionSlot{{"speech_to_text", "Speech input", &input.SpeechToText}, {"language_inference", "Language inference", &input.LanguageInference}, {"text_to_speech", "Spoken output", &input.TextToSpeech}} {
		var existing *string
		if current.Data != nil {
			for _, v := range current.Data.Slots {
				if v.Capability == slot.capability && v.SelectionSource == "explicit" {
					existing = v.SelectedProfileID
					if existing == nil {
						switch slot.capability {
						case "speech_to_text":
							existing = current.Data.ProfileIDs.SpeechToText
						case "language_inference":
							existing = current.Data.ProfileIDs.LanguageInference
						case "text_to_speech":
							existing = current.Data.ProfileIDs.TextToSpeech
						}
					}
					if v.Readiness == "invalid_selection" && (existing == nil || strings.TrimSpace(*existing) == "") {
						return o, ports.Failure("usage", "The server did not return an unavailable explicit profile ID. Use --input FILE to specify all three intended choices; guided replacement cannot safely preserve the missing ID.")
					}
				}
			}
		}
		keep := "Keep automatic selection"
		if existing != nil {
			keep = "Keep current " + strconv.Quote(*existing) + " (unavailable profile)"
			for _, p := range profiles.Data {
				if p.ID == *existing && p.Capability == slot.capability && p.LifecycleState != "archived" {
					keep = "Keep current " + strconv.Quote(*existing)
					break
				}
			}
		}
		choices := []ports.Choice{{ID: "cancel", Label: "Cancel", Detail: "Keep all saved selections unchanged"}, {ID: "keep", Label: keep}, {ID: "automatic", Label: "Use automatic selection", Detail: "Remove the explicit choice; the server selects a profile"}}
		values := map[string]*string{"keep": existing, "automatic": nil}
		for i, p := range profiles.Data {
			if p.Capability != slot.capability || p.LifecycleState == "archived" {
				continue
			}
			key := "profile-" + strconv.Itoa(i)
			id := p.ID
			values[key] = &id
			choices = append(choices, ports.Choice{ID: key, Label: strconv.Quote(p.DisplayName) + " (" + strconv.Quote(p.ID) + ")", Detail: "State: " + strconv.Quote(p.LifecycleState) + ". Credentials: " + strconv.Quote(p.CredentialStatus)})
		}
		selected, err := r.Picker.Pick(ctx, slot.label, choices)
		if err != nil {
			return o, err
		}
		if selected == "cancel" {
			return o, context.Canceled
		}
		value, ok := values[selected]
		if !ok {
			return o, ports.Failure("input", "The selected profile is not available. Run voice-provider update again.")
		}
		*slot.value = value
	}
	o.RequestBody, err = json.Marshal(input)
	return o, err
}
