package config

import (
	"errors"
	labelapp "github.com/stuffstash/stuff-stash/internal/app/printing"
	"strconv"
	"time"
)

type LabelConfig struct{ BaseURL, PublicWebBaseURL, RenderTTL, MaxRenderBytes, CleanupInterval, MaxPixels, MaxURLBytes, MaxTitleRunes, MaxReferenceRunes string }
type LabelSettings struct {
	BaseURL                                                  string
	RenderTTL                                                time.Duration
	MaxRenderBytes                                           int
	CleanupInterval                                          time.Duration
	MaxPixels, MaxURLBytes, MaxTitleRunes, MaxReferenceRunes int
}

func loadLabels() LabelConfig {
	return LabelConfig{
		BaseURL: envOrDefault("STUFF_STASH_LABEL_BASE_URL", ""), PublicWebBaseURL: envOrDefault("STUFF_STASH_PUBLIC_WEB_BASE_URL", ""),
		RenderTTL: envOrDefault("STUFF_STASH_LABEL_RENDER_TTL", "1h"), MaxRenderBytes: envOrDefault("STUFF_STASH_LABEL_RENDER_MAX_BYTES", "1048576"), CleanupInterval: envOrDefault("STUFF_STASH_LABEL_CLEANUP_INTERVAL", "1m"),
		MaxPixels: envOrDefault("STUFF_STASH_LABEL_MAX_PIXELS", "4000000"), MaxURLBytes: envOrDefault("STUFF_STASH_LABEL_MAX_URL_BYTES", "2048"), MaxTitleRunes: envOrDefault("STUFF_STASH_LABEL_MAX_TITLE_RUNES", "2048"), MaxReferenceRunes: envOrDefault("STUFF_STASH_LABEL_MAX_REFERENCE_RUNES", "128"),
	}
}
func (c LabelConfig) Settings() (LabelSettings, error) {
	result := LabelSettings{BaseURL: c.BaseURL, RenderTTL: time.Hour, MaxRenderBytes: 1048576, CleanupInterval: time.Minute, MaxPixels: 4000000, MaxURLBytes: 2048, MaxTitleRunes: 2048, MaxReferenceRunes: 128}
	if result.BaseURL == "" {
		result.BaseURL = c.PublicWebBaseURL
	}
	if result.BaseURL != "" && labelapp.ValidateLabelBaseURL(result.BaseURL) != nil {
		return result, errors.New("label base URL must be an unambiguous HTTPS base")
	}
	for _, item := range []struct {
		raw string
		dst *time.Duration
		max time.Duration
	}{{c.RenderTTL, &result.RenderTTL, 24 * time.Hour}, {c.CleanupInterval, &result.CleanupInterval, time.Hour}} {
		if item.raw != "" {
			value, err := time.ParseDuration(item.raw)
			if err != nil || value <= 0 || value > item.max {
				return result, errors.New("invalid label duration")
			}
			*item.dst = value
		}
	}
	for _, item := range []struct {
		raw string
		dst *int
		max int
	}{{c.MaxRenderBytes, &result.MaxRenderBytes, 16 * 1024 * 1024}, {c.MaxPixels, &result.MaxPixels, 16000000}, {c.MaxURLBytes, &result.MaxURLBytes, 4096}, {c.MaxTitleRunes, &result.MaxTitleRunes, 4096}, {c.MaxReferenceRunes, &result.MaxReferenceRunes, 512}} {
		if item.raw != "" {
			value, err := strconv.Atoi(item.raw)
			if err != nil || value <= 0 || value > item.max {
				return result, errors.New("invalid label size limit")
			}
			*item.dst = value
		}
	}
	return result, nil
}
