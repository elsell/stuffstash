package app

import (
	"flag"
	"io"

	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

// parseOptionValues observes registered Set failures, never library error text.
func parseOptionValues(flags *flag.FlagSet, args []string) error {
	var failure error
	flags.SetOutput(io.Discard)
	flags.VisitAll(func(f *flag.Flag) {
		f.Value = &diagnosedFlagValue{Value: f.Value, name: f.Name, expected: flagValueType(f.Value), failed: func(name, expected string) {
			failure = ports.Failure("usage", "The --"+name+" value is not correct. Supply "+expected+" and try again.")
		}}
	})
	if err := flags.Parse(args); err != nil {
		if failure != nil {
			return failure
		}
		return ports.Failure("usage", "The options could not be parsed. Run the command with --help and correct its options.")
	}
	return nil
}

type diagnosedFlagValue struct {
	flag.Value
	name, expected string
	failed         func(string, string)
}

func (v *diagnosedFlagValue) Set(value string) error {
	err := v.Value.Set(value)
	if err != nil {
		v.failed(v.name, v.expected)
	}
	return err
}
func (v *diagnosedFlagValue) IsBoolFlag() bool {
	b, ok := v.Value.(interface{ IsBoolFlag() bool })
	return ok && b.IsBoolFlag()
}
func (v *diagnosedFlagValue) Get() any {
	if getter, ok := v.Value.(flag.Getter); ok {
		return getter.Get()
	}
	return nil
}
func flagValueType(value flag.Value) string {
	getter, ok := value.(flag.Getter)
	if !ok {
		return "a correct value (see this command's --help)"
	}
	switch getter.Get().(type) {
	case bool:
		return "true or false"
	case int, int64:
		return "an integer within the supported range"
	case uint, uint64:
		return "a non-negative integer within the supported range"
	case float64:
		return "a number within the supported range"
	case string:
		return "a text value"
	default:
		return "a correct value (see this command's --help)"
	}
}
