package push

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func fcmRetryNotBefore(header string, status int, now time.Time) time.Time {
	var deadline time.Time
	header = strings.TrimSpace(header)
	if len(header) <= 128 && header != "" {
		if seconds, err := strconv.ParseUint(header, 10, 64); err == nil && seconds <= uint64(math.MaxInt64/int64(time.Second)) {
			if seconds > 0 {
				deadline = now.Add(time.Duration(seconds) * time.Second)
			}
		} else if date, err := http.ParseTime(header); err == nil && date.After(now) {
			deadline = date
		}
	}
	if status == http.StatusTooManyRequests && deadline.Before(now.Add(time.Minute)) {
		deadline = now.Add(time.Minute)
	}
	return deadline
}
