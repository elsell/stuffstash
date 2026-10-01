package config

import (
	"fmt"
	"os"
	"strconv"
)

type ExportConfiguration struct {
	records, bytes int
	captured       bool
	err            error
}

func (c ExportConfiguration) Limits() (int, int, error) {
	if c.err != nil {
		return 0, 0, c.err
	}
	if !c.captured {
		return 10000, 64 * 1024 * 1024, nil
	}
	return c.records, c.bytes, nil
}
func loadExportConfiguration() ExportConfiguration {
	c := ExportConfiguration{records: 10000, bytes: 64 * 1024 * 1024, captured: true}
	for _, entry := range []struct {
		name   string
		target *int
	}{{"STUFF_STASH_EXPORT_MAX_RECORDS", &c.records}, {"STUFF_STASH_EXPORT_MAX_BYTES", &c.bytes}} {
		raw := os.Getenv(entry.name)
		if raw == "" {
			continue
		}
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 || value == int(^uint(0)>>1) {
			c.err = fmt.Errorf("%s must be a positive bounded integer", entry.name)
			return c
		}
		*entry.target = value
	}
	return c
}
