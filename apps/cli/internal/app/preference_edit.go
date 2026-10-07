package app

import (
	"context"
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"unicode/utf8"
)

func (r Runner) editPreferenceBody(ctx context.Context, o Options, current ports.NotificationPreferences) ([]byte, error) {
	if current.Revision < 1 {
		return nil, ports.Failure("usage", "Initialize the notification preferences before you edit them.")
	}
	policy := current.Defaults
	if o.Command[1] == "override" {
		for _, v := range current.Overrides {
			if v.CustomAssetTypeID == o.Command[2] {
				policy = v.Settings
				break
			}
		}
	}
	enabled, err := r.preferenceBool(ctx, "Expiration reminders", policy.Enabled)
	if err != nil {
		return nil, err
	}
	policy.Enabled = enabled
	policy.Upcoming, err = r.preferenceBool(ctx, "Upcoming reminders", policy.Upcoming)
	if err != nil {
		return nil, err
	}
	policy.Expired, err = r.preferenceBool(ctx, "Expired reminders", policy.Expired)
	if err != nil {
		return nil, err
	}
	text, err := r.preferenceText(ctx, "Advance days", strconv.FormatInt(policy.AdvanceDays, 10), 4)
	if err != nil {
		return nil, err
	}
	if text != "" {
		policy.AdvanceDays, err = strconv.ParseInt(text, 10, 64)
		if err != nil || policy.AdvanceDays < 0 || policy.AdvanceDays > 3650 {
			return nil, ports.Failure("usage", "Advance days must be an integer from 0 to 3650.")
		}
	}
	if o.Command[1] == "override" {
		return json.Marshal(struct {
			Revision int64                  `json:"revision"`
			Settings ports.ExpirationPolicy `json:"settings"`
		}{current.Revision, policy})
	}
	push, err := r.preferenceBool(ctx, "Push notifications", current.PushEnabled)
	if err != nil {
		return nil, err
	}
	zone, err := r.preferenceText(ctx, "Timezone", current.Timezone, 100)
	if err != nil {
		return nil, err
	}
	if zone == "" {
		zone = current.Timezone
	}
	if zone == "" || utf8.RuneCountInString(zone) > 100 {
		return nil, ports.Failure("usage", "Supply a timezone name with 1 to 100 characters.")
	}
	return json.Marshal(struct {
		Revision    int64                  `json:"revision"`
		Defaults    ports.ExpirationPolicy `json:"defaults"`
		Timezone    string                 `json:"timezone"`
		PushEnabled bool                   `json:"pushEnabled"`
	}{current.Revision, policy, zone, push})
}
func (r Runner) preferenceBool(ctx context.Context, title string, current bool) (bool, error) {
	choices := []ports.Choice{{ID: strconv.FormatBool(current), Label: map[bool]string{true: "On", false: "Off"}[current], Detail: "Current setting"}, {ID: strconv.FormatBool(!current), Label: map[bool]string{true: "On", false: "Off"}[!current]}}
	value, err := r.Picker.Pick(ctx, title, choices)
	if err != nil {
		return false, err
	}
	result, err := strconv.ParseBool(value)
	if err != nil {
		return false, ports.Failure("input", "The selected setting is not correct. Select On or Off.")
	}
	return result, nil
}

func (r Runner) preferenceText(ctx context.Context, title, current string, maximum int) (string, error) {
	selected, err := r.Picker.Pick(ctx, title, []ports.Choice{{ID: "keep", Label: "Keep current", Detail: current}, {ID: "change", Label: "Change"}})
	if err != nil {
		return "", err
	}
	if selected == "keep" {
		return current, nil
	}
	if selected != "change" {
		return "", context.Canceled
	}
	return r.TextInput.ReadText(ctx, title, maximum)
}
