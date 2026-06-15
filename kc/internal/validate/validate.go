// Package validate checks that input is a usable kubeconfig before upload.
package validate

import (
	"fmt"

	"sigs.k8s.io/yaml"
)

// Error is a validation failure naming which check failed.
type Error struct {
	Check int
	Msg   string
}

func (e *Error) Error() string { return fmt.Sprintf("kc: %s (check %d)", e.Msg, e.Check) }

type named struct {
	Name string `json:"name"`
}

type ctxEntry struct {
	Name    string `json:"name"`
	Context struct {
		Cluster string `json:"cluster"`
		User    string `json:"user"`
	} `json:"context"`
}

type kubeconfig struct {
	Kind           string     `json:"kind"`
	APIVersion     string     `json:"apiVersion"`
	CurrentContext string     `json:"current-context"`
	Clusters       []named    `json:"clusters"`
	Contexts       []ctxEntry `json:"contexts"`
	Users          []named    `json:"users"`
}

func isSeq(raw map[string]interface{}, key string) bool {
	v, ok := raw[key]
	if !ok {
		return false
	}
	_, ok = v.([]interface{})
	return ok
}

// Validate runs the five kubeconfig checks. It returns *RepairableError when
// the only problem is an empty contexts array with a dangling current-context.
func Validate(data []byte) error {
	// Check 1: parses as YAML, and is a mapping.
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return &Error{Check: 1, Msg: "input is not valid YAML: " + err.Error()}
	}
	if raw == nil {
		return &Error{Check: 2, Msg: "input is empty / not a kubeconfig mapping"}
	}

	var kc kubeconfig
	if err := yaml.Unmarshal(data, &kc); err != nil {
		return &Error{Check: 2, Msg: "input does not decode as a kubeconfig: " + err.Error()}
	}

	// Check 2: kubeconfig shape.
	if kc.Kind != "Config" {
		if kc.APIVersion != "v1" || !isSeq(raw, "clusters") || !isSeq(raw, "contexts") || !isSeq(raw, "users") {
			return &Error{Check: 2, Msg: fmt.Sprintf(
				"input does not look like a kubeconfig (kind=%q apiVersion=%q; expected kind=Config or apiVersion=v1 with clusters/contexts/users arrays)",
				kc.Kind, kc.APIVersion)}
		}
	}

	// Check 3: contexts non-empty.
	if len(kc.Contexts) == 0 {
		if kc.CurrentContext != "" {
			return &RepairableError{
				CurrentContext: kc.CurrentContext,
				Clusters:       names(kc.Clusters),
				Users:          names(kc.Users),
			}
		}
		return &Error{Check: 3, Msg: "contexts: [] (empty) — no usable contexts"}
	}

	// Check 4: current-context (if set) resolves.
	if kc.CurrentContext != "" && !hasContext(kc.Contexts, kc.CurrentContext) {
		return &Error{Check: 4, Msg: fmt.Sprintf("current-context %q does not match any entry in contexts", kc.CurrentContext)}
	}

	// Check 5: every context's cluster/user references resolve.
	clusters := nameSet(kc.Clusters)
	users := nameSet(kc.Users)
	for _, c := range kc.Contexts {
		if !clusters[c.Context.Cluster] {
			return &Error{Check: 5, Msg: fmt.Sprintf("context %q references unknown cluster %q", c.Name, c.Context.Cluster)}
		}
		if !users[c.Context.User] {
			return &Error{Check: 5, Msg: fmt.Sprintf("context %q references unknown user %q", c.Name, c.Context.User)}
		}
	}
	return nil
}

// RepairableError signals the one auto-fixable failure: an empty contexts
// array alongside a set current-context. Callers may offer to synthesise the
// missing context (see RepairEmptyContexts).
type RepairableError struct {
	CurrentContext string
	Clusters       []string
	Users          []string
}

func (e *RepairableError) Error() string {
	return fmt.Sprintf("kc: contexts: [] but current-context %q is set (auto-repair may be possible)", e.CurrentContext)
}

func names(ns []named) []string {
	out := make([]string, len(ns))
	for i, n := range ns {
		out[i] = n.Name
	}
	return out
}

func nameSet(ns []named) map[string]bool {
	m := make(map[string]bool, len(ns))
	for _, n := range ns {
		m[n.Name] = true
	}
	return m
}

func hasContext(ctxs []ctxEntry, name string) bool {
	for _, c := range ctxs {
		if c.Name == name {
			return true
		}
	}
	return false
}

// RepairEmptyContexts synthesises the missing context entry for a kubeconfig
// that has contexts: [] and a set current-context, when exactly one cluster
// and one user are defined. The rest of the document is preserved (re-
// serialised; content unchanged). Re-validate the result before use.
func RepairEmptyContexts(data []byte) ([]byte, error) {
	var kc kubeconfig
	if err := yaml.Unmarshal(data, &kc); err != nil {
		return nil, err
	}
	if kc.CurrentContext == "" {
		return nil, fmt.Errorf("cannot repair: current-context is empty")
	}
	if len(kc.Clusters) != 1 || len(kc.Users) != 1 {
		return nil, fmt.Errorf("cannot repair: expected exactly 1 cluster and 1 user, got %d and %d", len(kc.Clusters), len(kc.Users))
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	raw["contexts"] = []interface{}{
		map[string]interface{}{
			"name": kc.CurrentContext,
			"context": map[string]interface{}{
				"cluster": kc.Clusters[0].Name,
				"user":    kc.Users[0].Name,
			},
		},
	}
	return yaml.Marshal(raw)
}
