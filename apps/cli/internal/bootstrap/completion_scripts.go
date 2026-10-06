package bootstrap

func completionScript(shell string) (string, bool) {
	switch shell {
	case "bash":
		return bashCompletion, true
	case "zsh":
		return zshCompletion, true
	case "fish":
		return fishCompletion, true
	default:
		return "", false
	}
}

const bashCompletion = `# Stuff Stash Bash completion. Save this file and source it from shell startup.
_stuffstash_complete() {
    COMPREPLY=()
    (( COMP_CWORD >= 1 )) || return 0
    local current="${COMP_WORDS[COMP_CWORD]}" candidate
    local -a previous=("${COMP_WORDS[@]:1:COMP_CWORD-1}")
    while IFS= read -r candidate; do
        COMPREPLY+=("$candidate")
    done < <(command stuffstash __complete bash -- "$current" "${previous[@]}" 2>/dev/null)
}
complete -F _stuffstash_complete stuffstash
`

const zshCompletion = `# Stuff Stash Zsh completion. Source this file after compinit in .zshrc.
_stuffstash_complete() {
    emulate -L zsh
    local -a previous candidates
    local i result
    for (( i=2; i<CURRENT; i++ )); do
        previous+=("${words[i]}")
    done
    result=$(command stuffstash __complete zsh -- "${words[CURRENT]}" "${previous[@]}" 2>/dev/null)
    [[ -n "$result" ]] || return 0
    candidates=("${(@f)result}")
    compadd -- "${candidates[@]}"
}
compdef _stuffstash_complete stuffstash
`

const fishCompletion = `# Stuff Stash Fish completion. Save as ~/.config/fish/completions/stuffstash.fish.
function __stuffstash_complete
    set -l words (commandline -opc)
    set -l current (commandline -ct)
    command stuffstash __complete fish -- "$current" $words[2..-1] 2>/dev/null
end
complete -c stuffstash -f -a '(__stuffstash_complete)'
`
