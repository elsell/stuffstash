package bootstrap

func localHelp() []commandHelp {
	return []commandHelp{
		{Path: "completion bash", Scope: helpLocal, Summary: "Write Bash shell completion to stdout.", Input: "Save the output and source the file from your shell startup. Completion suggests command and option names without resource or secret values.", Output: "Plain shell script on stdout, including with --json. Nothing is installed automatically.", Example: "completion bash > stuffstash.bash"},
		{Path: "completion zsh", Scope: helpLocal, Summary: "Write Zsh shell completion to stdout.", Input: "Save the output and source the file after autoload -Uz compinit; compinit in .zshrc.", Output: "Plain shell script on stdout, including with --json. Nothing is installed automatically.", Example: "completion zsh > stuffstash.zsh"},
		{Path: "completion fish", Scope: helpLocal, Summary: "Write Fish shell completion to stdout.", Input: "Save the output as ~/.config/fish/completions/stuffstash.fish.", Output: "Plain shell script on stdout, including with --json. Nothing is installed automatically.", Example: "completion fish > ~/.config/fish/completions/stuffstash.fish"},
		{Path: "login", Scope: helpServer, Summary: "Sign in through the server OIDC provider.", Options: "device-code", Input: "Use browser sign-in, or --device-code to approve from another device.", Example: "login --server https://stash.example --device-code", Output: "Stores the signed-in session and reports status. Headless hosts can set STUFF_STASH_CLI_CREDENTIAL_FILE explicitly; --json prints status as JSON."},
		{Path: "logout", Scope: helpServer, Summary: "Remove the local session and clear its remembered scope.", Example: "logout --server https://stash.example"},
		{Path: "version", Scope: helpLocal, Summary: "Show CLI version and platform capabilities."},
		{Path: "context list", Scope: helpLocal, Summary: "List saved contexts."},
		{Path: "context current", Scope: helpLocal, Summary: "Show the selected context."},
		{Path: "context use", Arguments: "NAME", Scope: helpLocal, Summary: "Use a saved context."},
		{Path: "context delete", Arguments: "NAME", Scope: helpLocal, Summary: "Delete a saved context."},
		{Path: "server show", Scope: helpServer, Summary: "Discover public server identity and protocol information."},
		{Path: "server auth-config", Scope: helpServer, Summary: "Discover the public sign-in configuration."},
		{Path: "printers discover", Scope: helpLocal, Summary: "Discover local USB printers without sign-in."},
		{Path: "printers catalog", Scope: helpLocal, Summary: "List the built-in printer catalog without sign-in."},
	}
}
