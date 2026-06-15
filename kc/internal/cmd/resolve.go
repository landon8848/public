package cmd

import (
	"fmt"
	"os"
)

// resolveName implements the active-config defaulting rule for read/use verbs:
// -n if given, else $KC_ACTIVE, else a hard error naming the verb.
func resolveName(verb, name string) (string, error) {
	if name != "" {
		return name, nil
	}
	if a := os.Getenv("KC_ACTIVE"); a != "" {
		return a, nil
	}
	return "", fmt.Errorf("kc %s: no -n given and no active config\n        run `kc set <name>` first, or `kc %s -n <name> ...`", verb, verb)
}
