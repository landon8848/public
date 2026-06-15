// Package session emits shell code for env mutation and the shell-init hooks.
package session

import (
	"fmt"
	"strings"
)

// Shell is a supported interactive shell.
type Shell string

const (
	ShellZsh  Shell = "zsh"
	ShellBash Shell = "bash"
	ShellFish Shell = "fish"
)

// ParseShell validates and converts a shell name.
func ParseShell(s string) (Shell, error) {
	switch Shell(s) {
	case ShellZsh, ShellBash, ShellFish:
		return Shell(s), nil
	default:
		return "", fmt.Errorf("unsupported shell %q (supported: zsh, bash, fish)", s)
	}
}

// sqQuote single-quotes a value for POSIX shells, escaping embedded quotes.
func sqQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// ExportEnv returns shell code that exports the active-session variables.
func ExportEnv(sh Shell, tempfile, name string) (string, error) {
	switch sh {
	case ShellZsh, ShellBash:
		return strings.Join([]string{
			"export KUBECONFIG=" + sqQuote(tempfile),
			"export KC_ACTIVE=" + sqQuote(name),
			"export KC_TEMPFILE=" + sqQuote(tempfile),
		}, "\n") + "\n", nil
	case ShellFish:
		return strings.Join([]string{
			"set -gx KUBECONFIG " + sqQuote(tempfile),
			"set -gx KC_ACTIVE " + sqQuote(name),
			"set -gx KC_TEMPFILE " + sqQuote(tempfile),
		}, "\n") + "\n", nil
	default:
		return "", fmt.Errorf("unsupported shell %q", sh)
	}
}

// ClearEnv returns shell code that unsets the active-session variables.
func ClearEnv(sh Shell) (string, error) {
	switch sh {
	case ShellZsh, ShellBash:
		return "unset KUBECONFIG KC_ACTIVE KC_TEMPFILE\n", nil
	case ShellFish:
		return "set -e KUBECONFIG KC_ACTIVE KC_TEMPFILE\n", nil
	default:
		return "", fmt.Errorf("unsupported shell %q", sh)
	}
}
