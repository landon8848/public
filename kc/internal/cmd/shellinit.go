package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/landon8848/public/kc/internal/session"
)

func runShellInit(d *Deps, shellName string) error {
	sh, err := session.ParseShell(shellName)
	if err != nil {
		return err
	}
	code, err := session.ShellInit(sh)
	if err != nil {
		return err
	}
	fmt.Fprint(d.Out, code)
	return nil
}

func newShellInitCmd(d *Deps) *cobra.Command {
	return &cobra.Command{
		Use:       "shell-init <zsh|bash|fish>",
		Short:     "Emit the shell hook to eval in your rc file",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"zsh", "bash", "fish"},
		RunE: func(_ *cobra.Command, args []string) error {
			return runShellInit(d, args[0])
		},
	}
}
