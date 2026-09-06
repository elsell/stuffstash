package agentmodel

import (
	"fmt"
	"strings"
	"testing"
)

func TestInventoryTagEvidenceIsBounded(t *testing.T) {
	for _, names := range [][]string{{""}, {strings.Repeat("x", 81)}, {string([]byte{255})}} {
		if got := BoundedObservationTagNames(names); len(got) != 0 {
			t.Fatalf("invalid inventory tag retained: %q", got)
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
