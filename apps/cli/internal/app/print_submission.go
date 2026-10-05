package app

import (
	"encoding/json"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func isPrintSubmission(o Options) bool {
	if len(o.Command) < 2 {
		return false
	}
	switch o.Command[0] + " " + o.Command[1] {
	case "labels print", "printers test", "print-jobs reprint":
		return true
	}
	return false
}
func decodePrintSubmission(o Options) (ports.LabelPrintSelection, error) {
	var input struct {
		PrinterID                string  `json:"printerId"`
		ExpectedMediaFingerprint string  `json:"expectedMediaFingerprint"`
		TemplateID               *string `json:"templateId"`
		TemplateVersion          *uint32 `json:"templateVersion"`
		TemplateOptions          *struct {
			ShowReference *bool `json:"showReference"`
		} `json:"templateOptions"`
		Copies             *int64  `json:"copies"`
		PreviewFingerprint *string `json:"previewFingerprint"`
		Schema             *string `json:"$schema"`
	}
	if json.Unmarshal(o.RequestBody, &input) != nil || input.PrinterID == "" || input.ExpectedMediaFingerprint == "" || input.TemplateID == nil || input.TemplateVersion == nil || *input.TemplateVersion == 0 || input.TemplateOptions == nil || input.TemplateOptions.ShowReference == nil || input.Copies == nil || *input.Copies <= 0 {
		return ports.LabelPrintSelection{}, ports.Failure("usage", "The print selection is invalid. Supply printerId, expectedMediaFingerprint, templateId, templateVersion (1 to 4294967295), templateOptions.showReference (true or false), and positive integer copies.")
	}
	if o.Command[0] == "printers" && input.PrinterID != o.Command[2] {
		return ports.LabelPrintSelection{}, ports.Failure("usage", "The input printerId does not match PRINTER_ID. Use the same printer in the command and input.")
	}
	return ports.LabelPrintSelection{PrinterID: input.PrinterID, RequestBody: o.RequestBody}, nil
}
