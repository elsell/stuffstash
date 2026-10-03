package app

import (
	"github.com/stuffstash/stuff-stash/internal/app/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func (a App) WithPrintJobs(repository ports.PrintJobRepository, config printing.JobConfig) App {
	a.printJobs = printing.NewJobService(a.labels, repository, a.printerRepository, config)
	return a
}
func (a App) PrintJobs() *printing.JobService { return a.printJobs }
