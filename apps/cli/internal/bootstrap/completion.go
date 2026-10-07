package bootstrap

import (
	"io"
	"sort"
	"strings"

	"github.com/stuffstash/stuff-stash/cli/internal/app"
	"github.com/stuffstash/stuff-stash/cli/internal/ports"
)

func writeCompletion(w io.Writer, command []string) error {
	if len(command) != 2 {
		return ports.Failure("usage", "Use completion bash, completion zsh, or completion fish.")
	}
	script, ok := completionScript(command[1])
	if !ok {
		return ports.Failure("usage", "Select bash, zsh or fish for shell completion.")
	}
	_, err := io.WriteString(w, script)
	return err
}

// completionQuery only reads reviewed static metadata. It never parses options
// for execution, consults runtime state, or emits supplied argument values.
func completionQuery(w io.Writer, args []string) error {
	if len(args) < 4 || args[2] != "--" {
		return ports.Failure("usage", "The local completion request is not correct. Generate a new script with stuffstash completion bash, zsh, or fish.")
	}
	if _, ok := completionScript(args[1]); !ok {
		return ports.Failure("usage", "The local completion shell is not supported. Generate a new script with stuffstash completion bash, zsh, or fish.")
	}
	candidates := completionCandidates(args[1], args[3], args[4:])
	if len(candidates) == 0 {
		return nil
	}
	_, err := io.WriteString(w, strings.Join(candidates, "\n")+"\n")
	return err
}

func completionCandidates(shell, current string, previous []string) []string {
	catalog := helpCatalog()
	flags := app.HelpOptions()
	path := ""
	var leaf *commandHelp
	pending, afterValue := false, false
	for _, word := range previous {
		if pending {
			// Readline may split --option=value into three words.
			if shell == "bash" && word == "=" {
				continue
			}
			pending, afterValue = false, true
			continue
		}
		if shell == "bash" && afterValue && (word == "=" || word == ":") {
			pending = true
			continue
		}
		afterValue = false
		if word == "--" {
			return nil
		}
		if strings.HasPrefix(word, "-") {
			name, _, inline := strings.Cut(strings.TrimLeft(word, "-"), "=")
			option, known := flags[name]
			if !known {
				return nil
			}
			pending = option.Value != "" && !inline
			afterValue = !pending
			continue
		}
		if leaf != nil {
			continue
		}
		next := word
		if path != "" {
			next = path + " " + word
		}
		known := false
		for i := range catalog {
			if catalog[i].Path == next {
				leaf = &catalog[i]
				known = true
				break
			}
			if strings.HasPrefix(catalog[i].Path, next+" ") {
				known = true
			}
		}
		if !known {
			return nil
		}
		path = next
	}
	if pending || strings.Contains(current, "=") || shell == "bash" && afterValue && current == ":" {
		return nil
	}
	choices := map[string]bool{}
	wantFlags := strings.HasPrefix(current, "-") || leaf != nil && current == ""
	if wantFlags {
		if leaf != nil {
			for _, name := range commandHelpOptions(*leaf) {
				choices["--"+name] = true
			}
		} else {
			for _, entry := range catalog {
				if path != "" && !strings.HasPrefix(entry.Path, path+" ") {
					continue
				}
				// Before an action is chosen, offer only shared scope/display options.
				shared := commandHelp{Scope: entry.Scope, Path: entry.Path}
				for _, name := range commandHelpOptions(shared) {
					choices["--"+name] = true
				}
			}
		}
	} else if leaf == nil {
		for _, entry := range catalog {
			suffix := entry.Path
			if path != "" {
				var ok bool
				suffix, ok = strings.CutPrefix(suffix, path+" ")
				if !ok {
					continue
				}
			}
			child, _, _ := strings.Cut(suffix, " ")
			choices[child] = true
		}
	}
	var result []string
	for choice := range choices {
		if strings.HasPrefix(choice, current) {
			result = append(result, choice)
		}
	}
	sort.Strings(result)
	return result
}
