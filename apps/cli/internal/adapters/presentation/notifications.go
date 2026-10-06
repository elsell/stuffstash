package presentation

import (
	"fmt"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"strings"
	"time"
)

func (o Output) notificationDetails(n ports.Notification) error {
	state := "Unread"
	if n.ReadAt != nil {
		state = "Read at " + n.ReadAt.Format(time.RFC3339Nano)
	}
	trail := make([]string, 0, len(n.ParentTrail))
	for _, a := range n.ParentTrail {
		trail = append(trail, a.Title)
	}
	fields := [][2]string{{"ID", n.ID}, {"Title", n.Title}, {"State", state}, {"Asset ID", n.AssetID}, {"Expiration", n.ExpirationDate}, {"Precision", n.ExpirationPrecision}, {"Milestone", n.Milestone}, {"Location", strings.Join(trail, " / ")}, {"Parent ID", n.ParentAssetID}, {"Type ID", n.CustomAssetTypeID}, {"Created", n.CreatedAt.Format(time.RFC3339Nano)}}
	if n.ParentTrailIncomplete {
		fields = append(fields, [2]string{"Location path", "Incomplete"})
	}
	return o.details(fields)
}
func (o Output) notificationList(r ports.Result[[]ports.Notification]) error {
	for _, n := range r.Data {
		state := "unread"
		if n.ReadAt != nil {
			state = "read"
		}
		trail := make([]string, 0, len(n.ParentTrail))
		for _, a := range n.ParentTrail {
			trail = append(trail, a.Title)
		}
		location := strings.Join(trail, " / ")
		if n.ParentTrailIncomplete {
			location += " (incomplete)"
		}
		if _, err := fmt.Fprintf(o.Stdout, "%s  %s  %s  %s  %s  asset %s  %s\n", strconv.Quote(n.ID), state, strconv.Quote(n.Title), strconv.Quote(n.Milestone), strconv.Quote(n.ExpirationDate), strconv.Quote(n.AssetID), strconv.Quote(location)); err != nil {
			return err
		}
	}
	return o.pagination(r.Pagination)
}
