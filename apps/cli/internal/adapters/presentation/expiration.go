package presentation

import (
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"strings"
)

func (o Output) expiration(w ports.ExpirationWorkspace) error {
	if _, err := fmt.Fprintf(o.Stdout, "Expired: %d  Expiring soon: %d  All: %d  Timezone: %s\n", w.Counts.Expired, w.Counts.Soon, w.Counts.All, strconv.Quote(w.Timezone)); err != nil {
		return err
	}
	for _, item := range w.Items {
		date, state := "", ""
		if item.Expiration != nil {
			date = item.Expiration.Date
		}
		if item.ExpirationContext != nil {
			state = item.ExpirationContext.State
		}
		path := make([]string, 0, len(item.AncestorPath))
		for _, a := range item.AncestorPath {
			path = append(path, a.Title)
		}
		if _, err := fmt.Fprintf(o.Stdout, "%s  %s  %s  %s  %s\n", strconv.Quote(item.ID), strconv.Quote(item.Title), strconv.Quote(state), strconv.Quote(date), strconv.Quote(strings.Join(path, " / "))); err != nil {
			return err
		}
	}
	return nil
}
