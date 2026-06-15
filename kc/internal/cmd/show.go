package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/landon8848/public/kc/internal/registry"
)

func runShow(d *Deps, verbose bool) error {
	active := os.Getenv("KC_ACTIVE")
	if active == "" {
		fmt.Fprintln(d.Out, "no active config")
		return errSilent
	}
	reg, err := registry.Load(d.RegistryPath)
	if err != nil {
		return err
	}
	c, ok := reg.Get(active)
	ref := "<not in registry>"
	if ok {
		ref = c.Ref
	}
	fmt.Fprintf(d.Out, "active: %s\nref:    %s\nfile:   %s\n", active, ref, os.Getenv("KUBECONFIG"))
	if verbose && ok {
		// Verbose mirrors `list -v`: absolute RFC3339, not the humanized form.
		fmt.Fprintf(d.Out, "vault:  %s\nupdated: %s\n", c.Vault, c.Updated)
	}
	return nil
}

func newShowCmd(d *Deps) *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Print the active config",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runShow(d, verbose)
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "also show vault and last-updated time")
	return cmd
}
