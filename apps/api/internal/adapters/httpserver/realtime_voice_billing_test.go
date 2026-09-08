package httpserver

import (
	"net/http/httptest"
	"strings"
	"testing"
)

type billingDisabledVoiceFailure struct{}

func (billingDisabledVoiceFailure) Error() string { return "raw provider response secret-project" }
func (billingDisabledVoiceFailure) SafeRealtimeVoiceDiagnostic() string {
	return "provider_billing_disabled"
}

func TestRealtimeVoiceQueryReportsBillingFailureWithoutProviderDetails(t *testing.T) {
	t.Parallel()

	application := newSeededTestAppWithVoice(t, seededState{
		tenants:     []seedTenant{{id: "tenant-home", name: "Home", owner: "user-1"}},
		inventories: []seedInventory{{id: "inventory-home", tenantID: "tenant-home", name: "Home inventory", owner: "user-1"}},
		ids:         []string{"voice-session-id"},
	}, fakeSpeechToText{transcript: "Where is my water bottle?"}, failingLanguageModel{err: billingDisabledVoiceFailure{}}, fakeTextToSpeech{chunks: [][]byte{[]byte("spoken-audio")}})

	server := httptest.NewServer(NewServerWithOptions("127.0.0.1:0", application, Options{RateLimitDisabled: true}).Handler)
	t.Cleanup(server.Close)

	events := runRealtimeVoiceQuestionUntil(t, server.URL, "tenant-home", "inventory-home", "user-1", "session.failed")
	assertNoRealtimeEventType(t, events, "agent.diagnostic")
	failed := findRealtimeEvent(t, events, "session.failed")
	if failed["code"] != "provider_billing_disabled" {
		t.Fatalf("expected provider_billing_disabled for provider failure, got %+v", failed)
	}
	if strings.Contains(failed["message"].(string), "raw provider response") {
		t.Fatalf("provider details leaked in safe failure message: %+v", failed)
	}
}
