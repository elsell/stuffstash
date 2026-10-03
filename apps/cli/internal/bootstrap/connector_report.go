package bootstrap

import (
	"github.com/stuffstash/stuff-stash/cli/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/cli/internal/version"
	"runtime"
	"slices"
	"strings"
)

func connectorSoftwareReport(getenv func(string) string) *printing.ConnectorReport {
	tag, commit, _ := strings.Cut(version.Build, ":")
	report := &printing.ConnectorReport{Version: tag, Commit: commit, Platform: runtime.GOOS, Architecture: runtime.GOARCH, Adapters: []printing.Descriptor{}}
	for _, registered := range BuiltinPrinters(getenv) {
		descriptor := registered.Printer.Descriptor()
		if slices.Contains(descriptor.Platforms, runtime.GOOS) {
			report.Adapters = append(report.Adapters, descriptor)
		}
	}
	return report
}
