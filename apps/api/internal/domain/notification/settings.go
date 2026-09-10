package notification

import (
	"maps"
	"strings"
	"time"
)

type AssetTypeID string

type Settings struct {
	Defaults    ExpirationPreferences
	Timezone    string
	PushEnabled bool
	Overrides   map[AssetTypeID]ExpirationPreferences
}

func DefaultSettings(timezone string) Settings {
	return Settings{Defaults: DefaultExpirationPreferences(), Timezone: timezone, Overrides: map[AssetTypeID]ExpirationPreferences{}}
}
func (s Settings) Validate() error {
	if err := s.Defaults.Validate(); err != nil {
		return err
	}
	if s.Timezone == "" || s.Timezone == "Local" || strings.TrimSpace(s.Timezone) != s.Timezone {
		return ErrInvalidPreferences
	}
	if _, err := time.LoadLocation(s.Timezone); err != nil {
		return ErrInvalidPreferences
	}
	for id, value := range s.Overrides {
		if strings.TrimSpace(string(id)) == "" || value.Validate() != nil {
			return ErrInvalidPreferences
		}
	}
	return nil
}
func (s Settings) Clone() Settings {
	s.Overrides = maps.Clone(s.Overrides)
	if s.Overrides == nil {
		s.Overrides = map[AssetTypeID]ExpirationPreferences{}
	}
	return s
}
func (s Settings) Equal(other Settings) bool {
	return s.Defaults == other.Defaults && s.Timezone == other.Timezone && s.PushEnabled == other.PushEnabled && maps.Equal(s.Overrides, other.Overrides)
}
func (s Settings) ForType(id AssetTypeID) ExpirationPreferences {
	if value, found := s.Overrides[id]; found {
		return value
	}
	return s.Defaults
}
