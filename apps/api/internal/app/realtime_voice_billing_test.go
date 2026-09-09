package app

import (
	"context"
	agentmodelapp "github.com/stuffstash/stuff-stash/internal/app/agentmodel"
	"github.com/stuffstash/stuff-stash/internal/ports"
	"testing"
)

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

func TestRealtimeVoiceExplainsKnownFailures(t *testing.T) {
	for _, tc := range []struct {
		err  error
		code string
	}{
		{context.DeadlineExceeded, "request_timeout"},
		{agentmodelapp.ErrConversationBudgetExhausted, "conversation_budget_exhausted"},
		{agentmodelapp.ErrConversationContextExhausted, "conversation_context_exhausted"},
		{realtimeVoiceProviderStageError{code: realtimeVoiceFailureLanguageInference, err: ports.ErrInvalidProviderInput}, "invalid_provider_output"},
	} {
		if got := realtimeVoiceErrorCode(tc.err); got != tc.code {
			t.Errorf("got %s want %s", got, tc.code)
		}
	}
}
