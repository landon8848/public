package cmd

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
	"github.com/landon8848/public/kc/internal/registry"
	"github.com/landon8848/public/kc/internal/tui"
	"github.com/landon8848/public/kc/internal/validate"
)

// validateWithRepair runs validation; on the repairable empty-contexts case it
// prompts (assumeYes skips the prompt) and returns the repaired bytes.
func validateWithRepair(d *Deps, data []byte, assumeYes bool) ([]byte, error) {
	err := validate.Validate(data)
	if err == nil {
		return data, nil
	}
	re, ok := err.(*validate.RepairableError)
	if !ok {
		return nil, err
	}
	if len(re.Clusters) != 1 || len(re.Users) != 1 {
		return nil, fmt.Errorf("empty contexts detected, but auto-repair needs exactly 1 cluster and 1 user (got %d and %d)", len(re.Clusters), len(re.Users))
	}
	fmt.Fprintf(d.Err, "Detected empty-contexts kubeconfig with dangling current-context:\n  current-context: %s\n  clusters: %v\n  users: %v\n", re.CurrentContext, re.Clusters, re.Users)
	if !assumeYes && !promptYesNo(d, "Auto-repair (synthesise missing context) before importing?", true) {
		return nil, fmt.Errorf("import aborted (repair declined)")
	}
	repaired, err := validate.RepairEmptyContexts(data)
	if err != nil {
		return nil, err
	}
	if err := validate.Validate(repaired); err != nil {
		return nil, fmt.Errorf("repaired kubeconfig still invalid: %w", err)
	}
	return repaired, nil
}

func runAdd(d *Deps, name string, r io.Reader, assumeYes bool) error {
	raw, err := readSource("", r)
	if err != nil {
		return err
	}
	return runAddBytes(d, name, raw, assumeYes)
}

func runAddBytes(d *Deps, name string, raw []byte, assumeYes bool) error {
	reg, err := registry.Load(d.RegistryPath)
	if err != nil {
		return err
	}
	if _, exists := reg.Get(name); exists {
		return fmt.Errorf("%q is already registered; use `kc edit -n %s`", name, name)
	}
	data, err := validateWithRepair(d, raw, assumeYes)
	if err != nil {
		return err
	}
	// Two-phase commit: 1Password first.
	ref, err := d.OP.DocumentCreate(context.Background(), name, data)
	if err != nil {
		return err
	}
	reg.Set(name, registry.Config{Ref: ref, Vault: d.OP.Vault, Updated: time.Now().UTC().Format(time.RFC3339)})
	if err := reg.Save(d.RegistryPath); err != nil {
		return fmt.Errorf("1P upload succeeded but registry save failed: %w; manually add %q -> %s", err, name, ref)
	}
	fmt.Fprintf(d.Err, "kc: registered %q -> %s\n", name, ref)
	return nil
}

// effectiveFile resolves the kubeconfig source path from the -f/--file flag and
// any positional argument. A positional path is accepted as a convenience so
// `kc add ~/.kube/config` works without -f; supplying both is ambiguous.
func effectiveFile(file string, args []string) (string, error) {
	if len(args) > 0 {
		if file != "" {
			return "", fmt.Errorf("kc add: pass the kubeconfig either as -f/--file or as a positional path, not both")
		}
		return args[0], nil
	}
	return file, nil
}

func newAddCmd(d *Deps) *cobra.Command {
	var name, file string
	var assumeYes bool
	cmd := &cobra.Command{
		Use:   "add [path]",
		Short: "Import a new kubeconfig",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			src, err := effectiveFile(file, args)
			if err != nil {
				return err
			}
			if name != "" && src != "" {
				raw, err := readSource(src, d.In)
				if err != nil {
					return err
				}
				return runAddBytes(d, name, raw, assumeYes)
			}
			n, srcKind, file2, pasted, err := tui.RunAddForm(tui.AddInputs{Name: name, File: src})
			if err != nil {
				return err
			}
			var raw []byte
			if srcKind == "paste" {
				raw = pasted
			} else {
				raw, err = readSource(file2, d.In)
			}
			if err != nil {
				return err
			}
			return runAddBytes(d, n, raw, assumeYes)
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "config name")
	cmd.Flags().StringVarP(&file, "file", "f", "", "read kubeconfig from file (omit or - for stdin); or pass the path positionally")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "assume yes to repair prompt")
	return cmd
}
