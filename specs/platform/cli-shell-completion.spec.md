# Local shell completion

## Interface

`stuffstash completion bash`, `completion zsh`, and `completion fish` write an
installable shell script to stdout. `completion --help` and shell-specific help
explain installation. Scripts are plain text even with `--json` or `--no-input`.
Generation does not read runtime environment configuration, contexts, credentials,
files, stdin, devices or network services. It does not install anything itself.

The scripts complete implemented command names, nested verbs and relevant option
names from the same reviewed help catalog and registered flag metadata used by
local help. They never complete resource IDs, files, input values or secrets.
After the option terminator `--`, completion stops. A pending option value or an
inline `--option=value` current token yields no suggestions. Existing flag arity
is used to skip completed values, including values that resemble command names.
Bash's separate `=` and `:` word-break tokens are handled conservatively.

## Safety and architecture

A hidden `__complete SHELL -- CURRENT [PREVIOUS_WORDS...]` query serves shell
wrappers. It is an internal protocol, not an operation or public command. Bootstrap
routes it before all option parsing and normal execution. It reads only static
metadata and produces static candidates; arbitrary words can never dispatch a
command, read runtime state or appear in its output. Centralizing token handling
avoids three shell parsers with divergent security behavior.

Wrappers invoke the literal `command stuffstash __complete` with quoted argv.
They do not use eval, execute entered words, expand values a second time, invoke
operations, query APIs, or enable fallback filename completion. The hidden query
uses a fixed argument boundary and validates the requested shell. Suggestions at
a group include its child commands and shared scope/display options; a leaf uses
that command's reviewed options. No separate command inventory is maintained.

## Shell installation and sources

Bash: save the emitted script and source it from shell startup.
Zsh: source the script after `autoload -Uz compinit; compinit` in shell startup.
Fish: save it as `~/.config/fish/completions/stuffstash.fish`.

Syntax is based on primary shell references:

- [Bash programmable completion](https://www.gnu.org/software/bash/manual/html_node/Programmable-Completion.html)
- [Bash completion builtins](https://www.gnu.org/software/bash/manual/html_node/Programmable-Completion-Builtins.html)
- [Zsh completion widgets](https://zsh.sourceforge.io/Doc/Release/Completion-Widgets.html)
- [Zsh completion system](https://zsh.sourceforge.io/Doc/Release/Completion-System.html)
- [Fish complete](https://fishshell.com/docs/current/cmds/complete.html)
- [Fish commandline](https://fishshell.com/docs/current/cmds/commandline.html)

Fish uses raw tokenization (`commandline -opc`), supported by Fish 3.6 and later,
to avoid expanding variables or command substitutions. The current manual marks
raw tokenization deprecated; it is intentional here to preserve literal input.

## Verification

Critical tests prove generation and hidden queries bypass environment/state
access, produce only catalog-derived choices, handle nested commands and option
values, and suppress suggestions after `--`. A native Bash test sources the emitted
script, uses a real local CLI helper process, checks completions and verifies that
command-substitution-like values remain inert. Shell syntax is checked with the
available native runtime. Temporary extracted Ubuntu packages pinned to Zsh
5.9-6ubuntu3 and Fish 4.0.1-1 verify generated wrappers without installing packages
or adding project dependencies. Fish checks use its actual `complete -C` engine;
Zsh checks load `compinit` and capture the generated wrapper's `compadd` arguments.
Both verify nested verbs, scoped flags, pending values, `--`, command-like values
and inert command-substitution-like values. A native Zsh PTY/ZLE Tab check also
expands `stuffstash workflows revisions sh` to `show`. This is scoped native
evidence, not certification of every terminal or platform.
