// Package paths resolves on-disk locations for kc state.
package paths

import (
	"os"
	"path/filepath"
)

// RegistryPath returns $XDG_DATA_HOME/kc/registry.toml, falling back to
// ~/.local/share/kc/registry.toml when XDG_DATA_HOME is unset.
func RegistryPath() string {
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		dir = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}
	return filepath.Join(dir, "kc", "registry.toml")
}
