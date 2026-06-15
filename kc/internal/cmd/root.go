// Package cmd wires the kc command-line surface.
package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/landon8848/public/kc/internal/opclient"
	"github.com/landon8848/public/kc/internal/paths"
	"github.com/landon8848/public/kc/internal/registry"
)

// version is the build version. It defaults to a dev value and is overridden
// at release time via -ldflags "-X .../internal/cmd.version=<tag>" (GoReleaser).
var version = "0.1.0"

// errSilent signals a non-zero exit without printing (message already shown).
var errSilent = errors.New("")

func defaultDeps() *Deps {
	return &Deps{
		RegistryPath: paths.RegistryPath(),
		OP:           &opclient.Client{Runner: opclient.ExecRunner{Bin: "op"}, Vault: "Private", Tag: "kc", Bin: "op"},
		In:           os.Stdin,
		Out:          os.Stdout,
		Err:          os.Stderr,
	}
}

func newRootCmd(d *Deps) *cobra.Command {
	root := &cobra.Command{
		Use:           "kc",
		Short:         "1Password-backed kubeconfig manager",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	// Pre-define --version with a capital -V so cobra doesn't claim lowercase
	// -v (which it does by default); -v is reserved for --verbose across verbs.
	root.Flags().BoolP("version", "V", false, "version for kc")
	root.AddCommand(
		newListCmd(d),
		newShowCmd(d),
		newSetCmd(d),
		newAddCmd(d),
		newEditCmd(d),
		newRemoveCmd(d),
		newRunCmd(d),
		newShellInitCmd(d),
	)

	// Dynamic completion of config names for -n across verbs.
	nameComp := func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		reg, err := registry.Load(d.RegistryPath)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		names := make([]string, 0, len(reg.Configs))
		for n := range reg.Configs {
			names = append(names, n)
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	}
	for _, c := range root.Commands() {
		if c.Flags().Lookup("name") != nil {
			_ = c.RegisterFlagCompletionFunc("name", nameComp)
		}
	}
	return root
}

// Execute runs the kc CLI and returns a process exit code.
func Execute() int {
	d := defaultDeps()
	if err := newRootCmd(d).Execute(); err != nil {
		if err != errSilent && err.Error() != "" {
			os.Stderr.WriteString("kc: " + err.Error() + "\n")
		}
		return 1
	}
	return 0
}
