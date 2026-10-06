package presentation

import (
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func (o Output) notificationPreferences(p ports.NotificationPreferences) error {
	if err := o.details([][2]string{{"Revision", strconv.FormatInt(p.Revision, 10)}, {"Timezone", p.Timezone}, {"Push enabled", strconv.FormatBool(p.PushEnabled)}}); err != nil {
		return err
	}
	if err := o.notificationPolicy("Defaults", p.Defaults); err != nil {
		return err
	}
	for _, v := range p.Overrides {
		if err := o.notificationPolicy("Type "+v.CustomAssetTypeID, v.Settings); err != nil {
			return err
		}
	}
	return nil
}
func (o Output) notificationPolicy(name string, p ports.ExpirationPolicy) error {
	return o.details([][2]string{{"Policy", name}, {"Enabled", strconv.FormatBool(p.Enabled)}, {"Upcoming", strconv.FormatBool(p.Upcoming)}, {"Expired", strconv.FormatBool(p.Expired)}, {"Advance days", strconv.FormatInt(p.AdvanceDays, 10)}})
}
