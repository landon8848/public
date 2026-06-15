package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/landon8848/public/kc/internal/opclient"
	"github.com/landon8848/public/kc/internal/registry"
	"github.com/landon8848/public/kc/internal/session"
	"github.com/landon8848/public/kc/internal/tui"
)

func runSet(d *Deps, name string, verbose bool) error {
	shellName := os.Getenv("KC_SHELL")
	if shellName == "" {
		fmt.Fprintln(d.Err, "kc set: requires the shell hook.")
		fmt.Fprintln(d.Err, `  add to your rc file:  eval "$(kc shell-init zsh)"  (or bash/fish)`)
		return errSilent
	}
	sh, err := session.ParseShell(shellName)
	if err != nil {
		return err
	}

	reg, err := registry.Load(d.RegistryPath)
	if err != nil {
		return err
	}
	c, ok := reg.Get(name)
	if !ok {
		return fmt.Errorf("unknown config: %q. Try: kc list", name)
	}

	data, err := d.OP.DocumentGet(context.Background(), opclient.RefToUUID(c.Ref))
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return fmt.Errorf("fetched document is empty for %q (%s)", name, c.Ref)
	}

	tmp, err := os.CreateTemp("", "kc.*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	tmp.Close()
	if err := os.Chmod(tmp.Name(), 0o400); err != nil {
		os.Remove(tmp.Name())
		return err
	}

	code, err := session.ExportEnv(sh, tmp.Name(), name)
	if err != nil {
		os.Remove(tmp.Name())
		return err
	}
	// The wrapper's eval runs this in the user's shell. Previous tempfile is
	// cleaned by the new value of $KC_TEMPFILE plus the shell-exit hook; emit
	// an rm of any prior tempfile first for immediate re-set hygiene.
	if prior := os.Getenv("KC_TEMPFILE"); prior != "" && prior != tmp.Name() {
		fmt.Fprintf(d.Out, "rm -f -- %q\n", prior)
	}
	fmt.Fprint(d.Out, code)
	if verbose {
		// Detail stays on stderr; stdout carries only the shell code to eval.
		fmt.Fprintf(d.Err, "kc: ref      %s\nkc: tempfile %s\n", c.Ref, tmp.Name())
	}
	fmt.Fprintf(d.Err, "kc: active config = %q\n", name)
	return nil
}

func newSetCmd(d *Deps) *cobra.Command {
	var name string
	var verbose bool
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Activate a config for this shell",
		RunE: func(_ *cobra.Command, _ []string) error {
			if name == "" {
				// Bare `kc set` → interactive picker (Task 14 wires the TUI).
				picked, err := pickConfig(d)
				if err != nil {
					return err
				}
				name = picked
			}
			return runSet(d, name, verbose)
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "config name")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "print ref + tempfile path to stderr")
	return cmd
}

func pickConfig(d *Deps) (string, error) {
	reg, err := registry.Load(d.RegistryPath)
	if err != nil {
		return "", err
	}
	names := make([]string, 0, len(reg.Configs))
	for n := range reg.Configs {
		names = append(names, n)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return "", fmt.Errorf("no configs registered; add one with `kc add`")
	}
	return tui.PickConfig(names, os.Getenv("KC_ACTIVE"))
}
