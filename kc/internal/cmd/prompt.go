package cmd

import (
	"bufio"
	"fmt"
	"strings"
)

// promptYesNo reads a y/n answer from d.In. Default applies on blank input.
func promptYesNo(d *Deps, question string, defaultYes bool) bool {
	hint := "[y/N]"
	if defaultYes {
		hint = "[Y/n]"
	}
	fmt.Fprintf(d.Err, "%s %s ", question, hint)
	sc := bufio.NewScanner(d.In)
	if !sc.Scan() {
		return defaultYes
	}
	ans := strings.ToLower(strings.TrimSpace(sc.Text()))
	if ans == "" {
		return defaultYes
	}
	return ans == "y" || ans == "yes"
}
