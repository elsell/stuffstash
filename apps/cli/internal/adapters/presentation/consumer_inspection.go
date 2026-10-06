package presentation

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
)

func (o Output) consumerPrinter(v ports.ConsumerPrinter) error {
	body, err := json.Marshal(v.Printer)
	if err != nil {
		return err
	}
	return o.details([][2]string{{"Printer", v.Printer.ID}, {"Name", v.Printer.Name}, {"Device", v.DeviceID}, {"Binding generation", strconv.FormatUint(v.BindingGeneration, 10)}, {"Details", string(body)}})
}
func (o Output) consumerAttempt(v *ports.ConsumerAttempt) error {
	if v == nil {
		return o.details([][2]string{{"Attempt", "Not returned"}})
	}
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return o.details([][2]string{{"Attempt", v.AttemptID}, {"Job", v.JobID}, {"Printer", v.PrinterID}, {"Status", v.Status}, {"Revision", strconv.FormatUint(v.Revision, 10)}, {"Details", string(body)}})
}
