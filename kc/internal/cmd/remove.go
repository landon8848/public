package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/landon8848/public/kc/internal/opclient"
	"github.com/landon8848/public/kc/internal/registry"
	"github.com/landon8848/public/kc/internal/session"
)

func runRemove(d *Deps, name string, assumeYes bool) error {
	if name == "" {
		return fmt.Errorf("kc remove: -n <name> is required")
	}
	reg, err := registry.Load(d.RegistryPath)
	if err != nil {
		return err
	}
	c, ok := reg.Get(name)
	if !ok {
		return fmt.Errorf("unknown config: %q", name)
	}
	if !assumeYes && !promptYesNo(d, fmt.Sprintf("Remove %q (and its 1Password document)?", name), false) {
		return fmt.Errorf("aborted")
	}
	// 1P first.
	if err := d.OP.ItemDelete(context.Background(), opclient.RefToUUID(c.Ref)); err != nil {
		return err
	}
	// Registry second.
	reg.Unset(name)
	if err := reg.Save(d.RegistryPath); err != nil {
		return fmt.Errorf("1P delete succeeded but registry save failed: %w", err)
	}
	// If removing the active config, emit teardown shell code (when eval'd).
	if os.Getenv("KC_ACTIVE") == name {
		if shellName := os.Getenv("KC_SHELL"); shellName != "" {
			if sh, err := session.ParseShell(shellName); err == nil {
				if code, err := session.ClearEnv(sh); err == nil {
					if tf := os.Getenv("KC_TEMPFILE"); tf != "" {
						fmt.Fprintf(d.Out, "rm -f -- %q\n", tf)
					}
					fmt.Fprint(d.Out, code)
				}
			}
		}
		fmt.Fprintf(d.Err, "kc: cleared active session (was %q)\n", name)
	}
	fmt.Fprintf(d.Err, "kc: removed %q\n", name)
	return nil
}

func newRemoveCmd(d *Deps) *cobra.Command {
	var name string
	var assumeYes bool
	cmd := &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm"},
		Short:   "Remove a config from 1Password and the registry",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runRemove(d, name, assumeYes)
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "config name (required)")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt")
	return cmd
}
