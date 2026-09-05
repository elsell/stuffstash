package agentmodel

import (
	"fmt"
	"strings"
	"testing"
)

func TestCandidateTagEvidenceIsBounded(t *testing.T) {
	observation := CandidateObservation{EvidenceRound: 1, ReferenceKey: SemanticReferenceSubject, CandidateID: "clothes", Title: "3–6 months clothes", Kind: "item", TagNames: []string{"Baby", "Clothes"}}
	if err := observation.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, names := range [][]string{{""}, {"Baby", "baby"}, {strings.Repeat("x", 81)}, {string([]byte{255})}, make([]string, 33)} {
		observation.TagNames = names
		if err := observation.Validate(); err == nil {
			t.Fatalf("invalid tag evidence accepted: %q", names)
		}
	}
	names := []string{"Baby", " baby ", "Clothes"}
	for i := 0; i < 40; i++ {
		names = append(names, fmt.Sprintf("tag-%d", i))
	}
	bounded := BoundedObservationTagNames(names)
	if len(bounded) != 32 || bounded[0] != "Baby" || bounded[1] != "Clothes" {
		t.Fatalf("unstable bound: %v", bounded)
	}
	names[0] = "Changed"
	if bounded[0] != "Baby" {
		t.Fatal("evidence aliases input")
	}
}
