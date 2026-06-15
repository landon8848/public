package tui

import "github.com/charmbracelet/huh"

// PickConfig shows a selector of config names and returns the chosen one.
// Rendering targets the terminal/stderr, never stdout.
func PickConfig(names []string, active string) (string, error) {
	if len(names) == 0 {
		return "", nil
	}
	opts := make([]huh.Option[string], 0, len(names))
	for _, n := range names {
		label := n
		if n == active {
			label = n + " *"
		}
		opts = append(opts, huh.NewOption(label, n))
	}
	var chosen string
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("Select config").Options(opts...).Value(&chosen),
	))
	if err := form.Run(); err != nil {
		return "", err
	}
	return chosen, nil
}
