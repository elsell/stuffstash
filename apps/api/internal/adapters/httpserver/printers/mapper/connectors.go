package mapper

import (
	"github.com/stuffstash/stuff-stash/internal/adapters/httpserver/printers/dto"
	"github.com/stuffstash/stuff-stash/internal/domain/printing"
	"github.com/stuffstash/stuff-stash/internal/ports"
)

func Connector(c printing.Connector) dto.Connector {
	return dto.Connector{Generation: c.Generation, PrinterIDs: []string{}, ID: string(c.ID), Name: c.Name, State: string(c.State), AuthorizationPending: c.Generation != c.SyncedGeneration, LastSeenAt: c.LastSeenAt}
}

func ConnectorRegistration(r ports.ConnectorRegistration) dto.Connector {
	result := Connector(r.Connector)
	for _, b := range r.Bindings {
		if !b.Revoked {
			result.PrinterIDs = append(result.PrinterIDs, string(b.PrinterID))
		}
	}
	return result
}
