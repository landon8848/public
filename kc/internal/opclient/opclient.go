// Package opclient is a typed wrapper over the 1Password `op` CLI. All process
// execution goes through Runner so tests can substitute a fake.
package opclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Runner executes a command and returns its stdout, stderr, and error.
type Runner interface {
	Run(ctx context.Context, args []string, stdin io.Reader) (stdout, stderr []byte, err error)
}

// ExecRunner runs a real binary named by Client.Bin.
type ExecRunner struct{ Bin string }

func (r ExecRunner) Run(ctx context.Context, args []string, stdin io.Reader) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, r.Bin, args...)
	cmd.Stdin = stdin
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	return out.Bytes(), errb.Bytes(), err
}

// Client wraps op operations for a single vault.
type Client struct {
	Runner Runner
	Vault  string
	Tag    string
	Bin    string // for ExecRunner construction by callers
}

// RefToUUID extracts the item UUID from an op://VAULT/UUID/FILENAME reference.
// The op item/document subcommands want a bare UUID and reject op:// strings.
func RefToUUID(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) < 4 {
		return ""
	}
	return parts[3]
}

// DocumentGet fetches a document's bytes by item UUID.
func (c *Client) DocumentGet(ctx context.Context, uuid string) ([]byte, error) {
	out, errb, err := c.Runner.Run(ctx, []string{"document", "get", uuid}, nil)
	if err != nil {
		return nil, fmt.Errorf("op document get failed: %s", strings.TrimSpace(string(errb)))
	}
	return out, nil
}

// DocumentCreate uploads data as a new document and returns its op:// ref.
func (c *Client) DocumentCreate(ctx context.Context, name string, data []byte) (string, error) {
	args := []string{
		"document", "create",
		"--vault", c.Vault,
		"--title", name + "-kubeconfig",
		"--tags", c.Tag,
		"--file-name", name + ".yaml",
		"--format", "json",
	}
	out, errb, err := c.Runner.Run(ctx, args, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("op document create failed: %s", strings.TrimSpace(string(errb)))
	}
	var resp struct {
		UUID string `json:"uuid"`
		ID   string `json:"id"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return "", fmt.Errorf("op document create: cannot parse output: %w", err)
	}
	uuid := resp.UUID
	if uuid == "" {
		uuid = resp.ID
	}
	if uuid == "" {
		return "", fmt.Errorf("op document create: no uuid in output: %s", string(out))
	}
	return fmt.Sprintf("op://%s/%s/%s.yaml", c.Vault, uuid, name), nil
}

// DocumentEdit replaces the content of an existing document by UUID.
func (c *Client) DocumentEdit(ctx context.Context, uuid string, data []byte) error {
	_, errb, err := c.Runner.Run(ctx, []string{"document", "edit", uuid}, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("op document edit failed: %s", strings.TrimSpace(string(errb)))
	}
	return nil
}

// ItemDelete deletes an item by UUID.
func (c *Client) ItemDelete(ctx context.Context, uuid string) error {
	_, errb, err := c.Runner.Run(ctx, []string{"item", "delete", uuid}, nil)
	if err != nil {
		return fmt.Errorf("op item delete failed: %s", strings.TrimSpace(string(errb)))
	}
	return nil
}

// ItemMeta is the listing metadata for a kc-managed item.
type ItemMeta struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	UpdatedAt string `json:"updated_at"`
}

// ItemList returns metadata for all kc-tagged items in the vault.
func (c *Client) ItemList(ctx context.Context) ([]ItemMeta, error) {
	args := []string{"item", "list", "--vault", c.Vault, "--tags", c.Tag, "--format", "json"}
	out, errb, err := c.Runner.Run(ctx, args, nil)
	if err != nil {
		return nil, fmt.Errorf("op item list failed: %s", strings.TrimSpace(string(errb)))
	}
	var metas []ItemMeta
	if err := json.Unmarshal(out, &metas); err != nil {
		return nil, fmt.Errorf("op item list: cannot parse output: %w", err)
	}
	return metas, nil
}
