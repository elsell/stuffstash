package config

import (
	"errors"
	"strings"
)

// ValidateS3Transport rejects invalid explicit public TLS settings before storage construction.
func (c Config) ValidateS3Transport() error {
	switch strings.ToLower(strings.TrimSpace(c.s3PublicSecureRaw)) {
	case "", "1", "true", "yes", "on", "0", "false", "no", "off":
		return nil
	default:
		return errors.New("invalid STUFF_STASH_S3_PUBLIC_SECURE setting")
	}
}
