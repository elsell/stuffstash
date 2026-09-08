package app

import "testing"

func TestRealtimeVoiceBillingFailureAcrossProviderStages(t *testing.T) {
	for _, stage := range []string{realtimeVoiceFailureSpeechToText, realtimeVoiceFailureLanguageInference, realtimeVoiceFailureTextToSpeech} {
		err := realtimeVoiceProviderStageError{code: stage, err: safeRealtimeVoiceDiagnosticFailure{safe: "provider_billing_disabled"}}
		if got := realtimeVoiceErrorCode(err); got != "provider_billing_disabled" {
			t.Errorf("%s: got %s", stage, got)
		}
		if got := safeRealtimeVoiceErrorDetail(err); got != "provider_billing_disabled" {
			t.Errorf("unsafe/unhelpful diagnostic: %s", got)
		}
	}
}
