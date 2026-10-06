package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func validateJSONObject(body []byte) error {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return ports.Failure("input", "The input is empty. Supply one JSON object in the file or stdin.")
	}
	// RawMessage validates the entire original input without converting numbers.
	// Use its original byte offset, and never expose a parser's value-bearing text.
	var value json.RawMessage
	if err := json.Unmarshal(body, &value); err != nil {
		var syntax *json.SyntaxError
		if errors.As(err, &syntax) {
			return ports.Failure("input", "The JSON is invalid at byte "+strconv.FormatInt(syntax.Offset, 10)+". Correct the JSON and try again.")
		}
		return ports.Failure("input", "The JSON is invalid. Supply one complete JSON object.")
	}
	if trimmed[0] != '{' {
		return ports.Failure("input", "The input must be a JSON object. Use braces around the request fields.")
	}
	return nil
}
