// Package tui holds the interactive carry-through forms and pickers.
package tui

import (
	"os"

	"github.com/charmbracelet/huh"
)

// AddInputs are the flags supplied on the command line.
type AddInputs struct {
	Name string
	File string // "" means no source given
}

// AddPlan says which fields the walkthrough must prompt for.
type AddPlan struct {
	PromptName   bool
	PromptSource bool
}

// PlanAdd computes the carry-through plan from supplied inputs.
func PlanAdd(in AddInputs) AddPlan {
	return AddPlan{
		PromptName:   in.Name == "",
		PromptSource: in.File == "",
	}
}

// RunAddForm prompts for whatever PlanAdd says is missing and returns the
// resolved name, source-kind ("file"/"paste"), file path, and pasted bytes.
// Rendering goes to the controlling terminal, never stdout.
func RunAddForm(in AddInputs) (name, sourceKind, file string, pasted []byte, err error) {
	plan := PlanAdd(in)
	name, file = in.Name, in.File
	sourceKind = "file"
	if in.File != "" {
		sourceKind = "file"
	}

	var fields []huh.Field
	if plan.PromptName {
		fields = append(fields, huh.NewInput().Title("Name for this config").Value(&name))
	}
	var pastedStr string
	if plan.PromptSource {
		fields = append(fields,
			huh.NewSelect[string]().Title("Source").Options(
				huh.NewOption("file", "file"),
				huh.NewOption("paste", "paste"),
			).Value(&sourceKind),
		)
	}
	if len(fields) > 0 {
		form := huh.NewForm(huh.NewGroup(fields...))
		// huh writes to stderr by default via the bubbletea program output.
		if err = form.Run(); err != nil {
			return "", "", "", nil, err
		}
	}
	if sourceKind == "file" && file == "" {
		f := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Path to kubeconfig file").Value(&file)))
		if err = f.Run(); err != nil {
			return "", "", "", nil, err
		}
	}
	if sourceKind == "paste" {
		f := huh.NewForm(huh.NewGroup(huh.NewText().Title("Paste kubeconfig (Ctrl-D to submit)").Value(&pastedStr)))
		if err = f.Run(); err != nil {
			return "", "", "", nil, err
		}
		pasted = []byte(pastedStr)
	}
	_ = os.Stdout // reminder: never render to stdout here
	return name, sourceKind, file, pasted, nil
}
