package app

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

const paginationCursorVersion = 1

type paginationCursorPayload struct {
	Version    int    `json:"v"`
	Collection string `json:"collection"`
	Scope      string `json:"scope"`
	LastID     string `json:"lastId"`
}

func pageLimit(defaultLimit int, maxLimit int, requested int) int {
	if requested <= 0 {
		return defaultLimit
	}
	if requested > maxLimit {
		return maxLimit
	}
	return requested
}

func encodePageCursor(collection string, scope string, lastID string) *string {
	if strings.TrimSpace(lastID) == "" {
		return nil
	}
	payload, err := json.Marshal(paginationCursorPayload{
		Version:    paginationCursorVersion,
		Collection: collection,
		Scope:      scope,
		LastID:     lastID,
	})
	if err != nil {
		return nil
	}
	cursor := base64.RawURLEncoding.EncodeToString(payload)
	return &cursor
}

func decodePageCursor(collection string, scope string, cursor string) (string, error) {
	cursor = strings.TrimSpace(cursor)
	if cursor == "" {
		return "", nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return "", err
	}
	var payload paginationCursorPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return "", err
	}
	if payload.Version != paginationCursorVersion || payload.Collection != collection || payload.Scope != scope || strings.TrimSpace(payload.LastID) == "" {
		return "", errors.New("invalid pagination cursor")
	}
	return payload.LastID, nil
}
