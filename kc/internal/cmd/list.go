package cmd

import (
	"context"
	"fmt"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"github.com/landon8848/public/kc/internal/registry"
)

// relDate renders an RFC3339 timestamp as a relative phrase ("2 hours ago").
// Unparseable or empty input is returned unchanged.
func relDate(updated string) string {
	t, err := time.Parse(time.RFC3339, updated)
	if err != nil {
		return updated
	}
	return humanize.Time(t)
}

func runList(d *Deps, quick, verbose bool) error {
	reg, err := registry.Load(d.RegistryPath)
	if err != nil {
		return err
	}
	if len(reg.Configs) == 0 {
		fmt.Fprintln(d.Out, "no configs registered")
		return nil
	}
	names := make([]string, 0, len(reg.Configs))
	for n := range reg.Configs {
		names = append(names, n)
	}
	sort.Strings(names)

	// Metadata (UPDATED column) needs one op call; --quick skips it.
	titles := map[string]string{}
	if !quick && d.OP != nil {
		if metas, err := d.OP.ItemList(context.Background()); err == nil {
			for _, m := range metas {
				titles[m.ID] = m.UpdatedAt
			}
		}
	}

	tw := tabwriter.NewWriter(d.Out, 0, 0, 2, ' ', 0)
	if verbose {
		fmt.Fprintln(tw, "NAME\tUPDATED\tVAULT\tREF")
	} else {
		fmt.Fprintln(tw, "NAME\tUPDATED\tVAULT")
	}
	for _, n := range names {
		c := reg.Configs[n]
		if verbose {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", n, c.Updated, c.Vault, c.Ref)
		} else {
			fmt.Fprintf(tw, "%s\t%s\t%s\n", n, relDate(c.Updated), c.Vault)
		}
	}
	return tw.Flush()
}

func newListCmd(d *Deps) *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List registered configs",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runList(d, false, verbose)
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show reference and absolute timestamps")
	return cmd
}
