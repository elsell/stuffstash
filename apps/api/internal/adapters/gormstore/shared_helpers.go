package gormstore

import "github.com/stuffstash/stuff-stash/internal/domain/audit"

func stringPtr(value string) *string {
	return &value
}

func stringFromPtr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func stringPtrFromAuditID(id audit.ID) *string {
	if id.String() == "" {
		return nil
	}
	return stringPtr(id.String())
}
