package presentation

import (
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func (o Output) voiceProvider(v *ports.VoiceProviderConfiguration) error {
	if v == nil {
		_, err := fmt.Fprintln(o.Stdout, "No voice provider configuration is available.")
		return err
	}
	fields := [][2]string{{"Household", v.TenantID}, {"Readiness", v.Readiness}}
	for _, entry := range []struct {
		label string
		value *string
	}{{"Language inference profile", v.ProfileIDs.LanguageInference}, {"Speech-to-text profile", v.ProfileIDs.SpeechToText}, {"Text-to-speech profile", v.ProfileIDs.TextToSpeech}} {
		value := "Not selected"
		if entry.value != nil {
			value = *entry.value
		}
		fields = append(fields, [2]string{entry.label, value})
	}
	if v.UpdatedAt != nil {
		fields = append(fields, [2]string{"Updated", *v.UpdatedAt})
	}
	if err := o.details(fields); err != nil {
		return err
	}
	if len(v.Slots) == 0 {
		_, err := fmt.Fprintln(o.Stdout, "No voice provider slots returned.")
		return err
	}
	for i, slot := range v.Slots {
		fields := [][2]string{{"Slot", strconv.Itoa(i + 1)}, {"Label", slot.Label}, {"Capability", slot.Capability}, {"Readiness", slot.Readiness}, {"Selection source", slot.SelectionSource}, {"Recommended action", slot.RecommendedAction}}
		if slot.SelectedProfileID != nil {
			fields = append(fields, [2]string{"Selected profile ID", *slot.SelectedProfileID})
		}
		if len(slot.Issues) == 0 {
			fields = append(fields, [2]string{"Issues", "None reported"})
		}
		for _, issue := range slot.Issues {
			fields = append(fields, [2]string{"Issue", issue})
		}
		if slot.SelectedProfile == nil {
			fields = append(fields, [2]string{"Selected profile", "Not selected"})
		} else {
			fields = append(fields, voiceProviderSummaryFields("Selected", *slot.SelectedProfile)...)
		}
		if len(slot.DuplicateProfiles) == 0 {
			fields = append(fields, [2]string{"Duplicate profiles", "None reported"})
		}
		for j, profile := range slot.DuplicateProfiles {
			fields = append(fields, voiceProviderSummaryFields("Duplicate "+strconv.Itoa(j+1), profile)...)
		}
		if err := o.details(fields); err != nil {
			return err
		}
	}
	return nil
}
func voiceProviderSummaryFields(prefix string, v ports.VoiceProviderSummary) [][2]string {
	fields := [][2]string{{prefix + " ID", v.ID}, {prefix + " name", v.DisplayName}, {prefix + " capability", v.Capability}, {prefix + " provider", v.ProviderKind}, {prefix + " model", v.ModelName}, {prefix + " lifecycle", v.LifecycleState}, {prefix + " credential status", v.CredentialStatus}}
	if v.CredentialPurpose != nil {
		fields = append(fields, [2]string{prefix + " credential purpose", *v.CredentialPurpose})
	}
	if v.LastTestedAt != nil {
		fields = append(fields, [2]string{prefix + " last tested", *v.LastTestedAt})
	}
	return fields
}
