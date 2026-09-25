package integration

import (
	"fmt"
)

// GenerateShellHook outputs shell script to bind Ctrl+R to Eidetic.
func GenerateShellHook(shell string) (string, error) {
	switch shell {
	case "bash":
		return `# Eidetic Bash Integration
__eidetic_history() {
    local selected
    selected=$(eidetic --query "$READLINE_LINE")
    if [ -n "$selected" ]; then
        READLINE_LINE="$selected"
        READLINE_POINT=${#READLINE_LINE}
    fi
}
bind -x '"\C-r": __eidetic_history'
`, nil

	case "zsh":
		return `# Eidetic Zsh Integration
__eidetic_history() {
    local selected
    selected=$(eidetic --query "$BUFFER")
    if [ -n "$selected" ]; then
        BUFFER="$selected"
        CURSOR=${#BUFFER}
    fi
    zle reset-prompt
}
zle -N __eidetic_history
bindkey '^R' __eidetic_history
`, nil

	case "fish":
		return `# Eidetic Fish Integration
function __eidetic_history
    set -l query (commandline)
    set -l selected (eidetic --query "$query")
    if test -n "$selected"
        commandline -r "$selected"
    end
    commandline -f repaint
end
bind \cr __eidetic_history
`, nil

	default:
		return "", fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
	}
}
