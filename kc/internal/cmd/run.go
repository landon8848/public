package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/landon8848/public/kc/internal/opclient"
	"github.com/landon8848/public/kc/internal/registry"
)

// prepareRun resolves the named config and fetches its bytes. With verbose it
// announces the resolved name + ref on stderr before any exec happens.
func prepareRun(d *Deps, name string, verbose bool) (string, []byte, error) {
	name, err := resolveName("run", name)
	if err != nil {
		return "", nil, err
	}
	reg, err := registry.Load(d.RegistryPath)
	if err != nil {
		return "", nil, err
	}
	c, ok := reg.Get(name)
	if !ok {
		return "", nil, fmt.Errorf("unknown config: %q. Try: kc list", name)
	}
	if verbose {
		fmt.Fprintf(d.Err, "kc: running against %q (%s)\n", name, c.Ref)
	}
	data, err := d.OP.DocumentGet(context.Background(), opclient.RefToUUID(c.Ref))
	if err != nil {
		return "", nil, err
	}
	return name, data, nil
}

// runRun fetches the document and runs kubectl with KUBECONFIG=/dev/fd/N — the
// config exists only as a pipe FD, never on disk.
func runRun(d *Deps, name string, verbose bool, kubectlArgs []string) error {
	_, data, err := prepareRun(d, name, verbose)
	if err != nil {
		return err
	}

	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}
	go func() {
		pw.Write(data)
		pw.Close()
	}()
	cmd := exec.Command("kubectl", kubectlArgs...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, d.Out, d.Err
	cmd.ExtraFiles = []*os.File{pr} // becomes fd 3 in the child
	cmd.Env = append(os.Environ(), "KUBECONFIG=/dev/fd/3")
	return cmd.Run()
}

func newRunCmd(d *Deps) *cobra.Command {
	var name string
	var verbose bool
	cmd := &cobra.Command{
		Use:                "run",
		Short:              "One-shot kubectl against a config (nothing on disk)",
		DisableFlagParsing: false,
		RunE: func(c *cobra.Command, args []string) error {
			// Everything after `--` is in args; cobra puts it there via ArgsLenAtDash.
			dash := c.ArgsLenAtDash()
			if dash < 0 || dash >= len(args) {
				return fmt.Errorf("usage: kc run [-n <name>] -- <kubectl args>")
			}
			return runRun(d, name, verbose, args[dash:])
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "config name (defaults to active)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "announce the resolved config + ref on stderr before running")
	return cmd
}
