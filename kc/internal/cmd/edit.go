package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/landon8848/public/kc/internal/opclient"
	"github.com/landon8848/public/kc/internal/registry"
)

func runEditBytes(d *Deps, name string, raw []byte) error {
	name, err := resolveName("edit", name)
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
	data, err := validateWithRepair(d, raw, false)
	if err != nil {
		return err
	}
	if err := d.OP.DocumentEdit(context.Background(), opclient.RefToUUID(c.Ref), data); err != nil {
		return err
	}
	c.Updated = time.Now().UTC().Format(time.RFC3339)
	reg.Set(name, c)
	if err := reg.Save(d.RegistryPath); err != nil {
		return err
	}
	fmt.Fprintf(d.Err, "kc: updated %q\n", name)
	return nil
}

// runEditInteractive pulls the doc to a tempfile, opens $EDITOR, validates, pushes.
func runEditInteractive(d *Deps, name string) error {
	name, err := resolveName("edit", name)
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
	tmp, err := os.CreateTemp("", "kc-edit.*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	tmp.Write(data)
	tmp.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	ed := exec.Command(editor, tmp.Name())
	ed.Stdin, ed.Stdout, ed.Stderr = os.Stdin, os.Stderr, os.Stderr
	if err := ed.Run(); err != nil {
		return err
	}
	edited, err := os.ReadFile(tmp.Name())
	if err != nil {
		return err
	}
	return runEditBytes(d, name, edited)
}

func newEditCmd(d *Deps) *cobra.Command {
	var name, file string
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a config in $EDITOR, or replace it from a file/stdin",
		RunE: func(_ *cobra.Command, _ []string) error {
			if file != "" {
				raw, err := readSource(file, d.In)
				if err != nil {
					return err
				}
				return runEditBytes(d, name, raw)
			}
			return runEditInteractive(d, name)
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "", "config name (defaults to active)")
	cmd.Flags().StringVarP(&file, "file", "f", "", "replace from file (- for stdin) instead of $EDITOR")
	_ = io.Discard
	return cmd
}
