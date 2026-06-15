// Package registry stores the name -> 1Password reference mapping as TOML.
package registry

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is one registered kubeconfig entry.
type Config struct {
	Ref     string `toml:"ref"`
	Vault   string `toml:"vault"`
	Updated string `toml:"updated"` // RFC3339
}

// Registry is the full on-disk state.
type Registry struct {
	Configs map[string]Config `toml:"configs"`
}

// Load reads the registry from path. A missing file is not an error — it
// yields an empty registry (no configs registered yet).
func Load(path string) (*Registry, error) {
	r := &Registry{Configs: map[string]Config{}}
	_, err := toml.DecodeFile(path, r)
	if errors.Is(err, fs.ErrNotExist) {
		return &Registry{Configs: map[string]Config{}}, nil
	}
	if err != nil {
		return nil, err
	}
	if r.Configs == nil {
		r.Configs = map[string]Config{}
	}
	return r, nil
}

// Get returns the config registered under name.
func (r *Registry) Get(name string) (Config, bool) {
	c, ok := r.Configs[name]
	return c, ok
}

// Set adds or replaces an entry in memory.
func (r *Registry) Set(name string, c Config) {
	if r.Configs == nil {
		r.Configs = map[string]Config{}
	}
	r.Configs[name] = c
}

// Unset removes an entry in memory. No-op if absent.
func (r *Registry) Unset(name string) {
	delete(r.Configs, name)
}

// Save writes the whole registry to path atomically (temp file + rename),
// creating the parent directory if needed.
func (r *Registry) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "registry-*.toml")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after a successful rename
	if err := toml.NewEncoder(tmp).Encode(r); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
