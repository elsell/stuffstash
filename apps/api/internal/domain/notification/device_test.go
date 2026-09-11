package notification

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestDeviceTokenValidationAndRedaction(t *testing.T) {
	token, err := ParseDeviceToken("secret-native-token")
	if err != nil || token.Secret() != "secret-native-token" {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(token)
	if err != nil || strings.Contains(string(encoded), "secret-native-token") || strings.Contains(fmt.Sprintf("%+v", token), "secret-native-token") {
		t.Fatal("token exposed")
	}
	for _, input := range []string{"", " token", "token\n", strings.Repeat("x", 4097)} {
		if _, err := ParseDeviceToken(input); err == nil {
			t.Fatal("invalid token accepted")
		}
	}
}
