package app

import (
	"flag"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func printSubmissionFlags(flags *flag.FlagSet, o *Options) {
	flags.StringVar(&o.ExpectedMediaFingerprint, "expected-media-fingerprint", "", "reviewed printer media fingerprint")
	flags.StringVar(&o.PreviewFingerprint, "preview-fingerprint", "", "reviewed label preview fingerprint")
}
func validatePrintSubmissionFlags(flags *flag.FlagSet, o Options) error {
	var problem error
	flags.Visit(func(f *flag.Flag) {
		if problem != nil {
			return
		}
		if !isPrintSubmission(o) {
			if f.Name == "expected-media-fingerprint" || f.Name == "preview-fingerprint" {
				problem = ports.Failure("usage", "Use --"+f.Name+" with labels print, printers test, or print-jobs reprint.")
			}
			return
		}
		switch f.Name {
		case "server", "tenant", "inventory", "context", "json", "no-input", "request-id", "color", "help", "idempotency-key", "input":
		case "printer", "template", "template-version", "copies", "show-reference", "expected-media-fingerprint", "preview-fingerprint":
			if o.InputPath != "" {
				problem = ports.Failure("usage", "Use either --input or print selection flags. Do not combine them.")
				return
			}
			if f.Name == "template-version" && o.TemplateVersion == 0 {
				problem = ports.Failure("usage", "Use --template-version from 1 to 4294967295, or remove the option to use the default.")
				return
			}
			if (f.Name == "printer" && o.PrinterID == "") || (f.Name == "template" && o.TemplateID == "") || (f.Name == "expected-media-fingerprint" && o.ExpectedMediaFingerprint == "") {
				problem = ports.Failure("usage", "Supply a value that is not empty for --"+f.Name+" or remove the option.")
			}
		default:
			problem = ports.Failure("usage", "Print submission does not accept --"+f.Name+". Remove the option.")
		}
	})
	if problem != nil {
		return problem
	}
	if isPrintSubmission(o) && len(o.Command) == 3 && o.Command[0] == "printers" && o.PrinterID != "" && o.PrinterID != o.Command[2] {
		return ports.Failure("usage", "The --printer value does not match PRINTER_ID. Use the same printer or remove --printer.")
	}
	return nil
}
