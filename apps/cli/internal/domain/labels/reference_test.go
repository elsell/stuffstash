package labels

import "testing"

func TestReferenceRejectsAmbiguousInputAndKeepsOnlyIdentity(t *testing.T) {
	instance := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	label := "01ARZ3NDEKTSV4RRFFQ69G5FAW"
	for _, prefix := range []string{"https://old.example/prefix/l/v1/", "stuffstash://labels/v1/"} {
		got, err := Parse(prefix + instance + "/" + label)
		if err != nil || got.Instance != instance || got.Label != label {
			t.Fatalf("%+v %v", got, err)
		}
	}
	for _, value := range []string{"http://host/l/v1/", "https://user:password@host/l/v1/", "https://host/a/../l/v1/", "https://host//l/v1/", "https://host/%61/l/v1/", "https://host/l/v2/"} {
		if _, err := Parse(value + instance + "/" + label); err == nil {
			t.Fatal("accepted", value)
		}
	}
	valid := "https://host/l/v1/" + instance + "/" + label
	for _, value := range []string{valid + "?", valid + "#", valid + "/", " " + valid, valid + " ", valid[:len(valid)-1] + "I"} {
		if _, err := Parse(value); err == nil {
			t.Fatal("accepted", value)
		}
	}
}
