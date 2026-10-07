package app

import (
	"flag"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
	"strconv"
	"strings"
)

// partitionOptions shares the actual registered arity with execution and help.
func partitionOptions(args []string, flags *flag.FlagSet) (flagArgs, positional []string, help bool, err error) {
	for i := 0; i < len(args); i++ {
		word := args[i]
		if word == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(word, "-") {
			positional = append(positional, word)
			continue
		}
		name, value, hasValue := strings.Cut(strings.TrimLeft(word, "-"), "=")
		definition := flags.Lookup(name)
		if definition == nil {
			if err == nil {
				err = ports.Failure("usage", "The command has an unsupported option. Run the command with --help for its options.")
			}
			continue
		}
		if name == "help" {
			help = true
			if hasValue {
				if b, e := strconv.ParseBool(value); e == nil {
					help = b
				}
			}
		}
		flagArgs = append(flagArgs, word)
		boolean, ok := definition.Value.(interface{ IsBoolFlag() bool })
		if !hasValue && (!ok || !boolean.IsBoolFlag()) {
			i++
			if i == len(args) {
				if err == nil {
					err = ports.Failure("usage", "Supply a value for --"+name+". Run the command with --help for its options.")
				}
				break
			}
			flagArgs = append(flagArgs, args[i])
		}
	}
	return
}

// ParseHelp is local: it never reads runtime configuration or command input.
func ParseHelp(args []string) (Options, bool, error) {
	var o Options
	flags := optionFlags(&o)
	flagArgs, command, requested, err := partitionOptions(args, flags)
	o.Command = command
	if !requested {
		return o, false, nil
	}
	if err != nil {
		return o, true, err
	}
	if err = parseOptionValues(flags, flagArgs); err != nil {
		return o, true, err
	}
	o.Command = command
	return o, true, nil
}

type HelpOption struct{ Name, Value, Description string }

// HelpOptions derives option names, descriptions and boolean arity from execution.
func HelpOptions() map[string]HelpOption {
	flags := optionFlags(&Options{})
	result := map[string]HelpOption{}
	flags.VisitAll(func(f *flag.Flag) {
		value := "VALUE"
		if b, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && b.IsBoolFlag() {
			value = ""
		}
		result[f.Name] = HelpOption{Name: f.Name, Value: value, Description: f.Usage}
	})
	return result
}
