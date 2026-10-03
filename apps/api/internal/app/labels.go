package app

import "github.com/stuffstash/stuff-stash/internal/app/printing"

func (a App) Labels() *printing.LabelService { return a.labels }
