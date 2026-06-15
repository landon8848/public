package session

import "fmt"

const zshBashInit = `# kc shell integration — eval "$(kc shell-init %[1]s)"
kc() {
  case "$1" in
    set|remove|rm)
      eval "$(KC_SHELL=%[1]s command kc "$@")"
      ;;
    *)
      command kc "$@"
      ;;
  esac
}
_kc_cleanup() {
  [ -n "${KC_TEMPFILE:-}" ] && rm -f -- "$KC_TEMPFILE"
}
`

const zshHook = `if (( $+functions[add-zsh-hook] )) || autoload -Uz add-zsh-hook 2>/dev/null; then
  add-zsh-hook zshexit _kc_cleanup
fi
`

const bashHook = `trap _kc_cleanup EXIT
`

const fishInit = `# kc shell integration — kc shell-init fish | source
function kc
  switch $argv[1]
    case set remove rm
      eval (env KC_SHELL=fish command kc $argv)
    case '*'
      command kc $argv
  end
end
function _kc_cleanup --on-event fish_exit
  test -n "$KC_TEMPFILE"; and rm -f -- "$KC_TEMPFILE"
end
`

// ShellInit returns the shell snippet to source/eval once in the user's rc
// file. It defines the kc wrapper (routing env-mutating verbs through eval)
// and installs a cleanup hook that removes the active tempfile on shell exit.
func ShellInit(sh Shell) (string, error) {
	switch sh {
	case ShellZsh:
		return fmt.Sprintf(zshBashInit, "zsh") + zshHook, nil
	case ShellBash:
		return fmt.Sprintf(zshBashInit, "bash") + bashHook, nil
	case ShellFish:
		return fishInit, nil
	default:
		return "", fmt.Errorf("unsupported shell %q", sh)
	}
}
