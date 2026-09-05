package audit

import "testing"

func TestConversationAuditTargetsRoundTrip(t *testing.T) {
	for _, value := range []string{"conversation_workflow", "conversation_evaluation_case", "conversation_evaluation_run"} {
		t.Run(value, func(t *testing.T) {
			target, ok := NewTargetType(value)
			if !ok || target.String() != value {
				t.Fatalf("conversation target rejected: %q", value)
			}
		})
	}
	if _, ok := NewTargetType("conversation_unknown"); ok {
		t.Fatal("unknown target accepted")
	}
}
