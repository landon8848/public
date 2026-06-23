// Command whenin is the Alfred Script Filter entry point for "When In…".
//
// Usage: whenin "<query>"
// Emits Alfred Script Filter JSON on stdout. It never crashes to Alfred:
// any panic becomes a graceful error result row.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/landon8848/public/whenin-alfred/internal/whenin"
)

const resultLimit = 9

// popularPin pins empty-query suggestions by (search_name, iana) so ambiguous
// names resolve deterministically (London UK not London ON, etc.).
var popularPin = []struct{ searchName, iana string }{
	{"london", "Europe/London"},
	{"new york city", "America/New_York"},
	{"los angeles", "America/Los_Angeles"},
	{"rome", "Europe/Rome"},
	{"tokyo", "Asia/Tokyo"},
	{"sydney", "Australia/Sydney"},
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			details := fmt.Sprintf("%v\n\n%s", r, debug.Stack())
			logError(details)
			emit([]whenin.AlfredItem{
				whenin.ErrorItem(fmt.Sprintf("%v", r), details),
			})
		}
	}()
	run()
}

func run() {
	timeFormat := strings.ToLower(strings.TrimSpace(os.Getenv("TIME_FORMAT")))
	if timeFormat != "12h" && timeFormat != "24h" {
		timeFormat = "12h"
	}

	raw := ""
	if len(os.Args) > 1 {
		raw = os.Args[1]
	}
	query := whenin.ParseQuery(raw)

	switch query.Mode {
	case whenin.ModeHoliday:
		emit([]whenin.AlfredItem{whenin.HolidayStub()})
		return

	case whenin.ModeTZ:
		if query.LocationText == "" {
			emit(render(whenin.PopularTZ(), timeFormat))
			return
		}
		results := whenin.TopN(query.LocationText, whenin.TZCandidates(), resultLimit)
		emitResultsOrNoMatch(results, query.LocationText, timeFormat)
		return
	}

	// city mode (default)
	cands, err := whenin.AllCandidates()
	if err != nil {
		panic(err)
	}

	if query.LocationText == "" {
		emit(render(pinnedPopular(cands), timeFormat))
		return
	}

	results := whenin.TopN(query.LocationText, cands, resultLimit)
	emitResultsOrNoMatch(results, query.LocationText, timeFormat)
}

func pinnedPopular(cands []whenin.Candidate) []whenin.Candidate {
	rank := map[[2]string]int{}
	for i, p := range popularPin {
		rank[[2]string{p.searchName, p.iana}] = i
	}
	picks := make([]whenin.Candidate, len(popularPin))
	found := make([]bool, len(popularPin))
	for _, c := range cands {
		if i, ok := rank[[2]string{c.SearchName, c.IANA}]; ok && !found[i] {
			picks[i] = c
			found[i] = true
		}
	}
	out := picks[:0]
	for i, c := range picks {
		if found[i] {
			out = append(out, c)
		}
	}
	return out
}

func render(cands []whenin.Candidate, timeFormat string) []whenin.AlfredItem {
	items := make([]whenin.AlfredItem, 0, len(cands))
	for _, c := range cands {
		items = append(items, whenin.ToAlfred(c, timeFormat))
	}
	return items
}

func emitResultsOrNoMatch(results []whenin.Candidate, queryText, timeFormat string) {
	if len(results) == 0 {
		emit([]whenin.AlfredItem{whenin.NoMatch(queryText)})
		return
	}
	emit(render(results, timeFormat))
}

func emit(items []whenin.AlfredItem) {
	out, err := json.Marshal(map[string][]whenin.AlfredItem{"items": items})
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(out)
}

func logError(details string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	base := filepath.Join(home, "Library", "Caches",
		"com.runningwithcrayons.Alfred", "Workflow Data",
		"com.landonpowell.whenin")
	if err := os.MkdirAll(base, 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(base, "whenin.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(details + "\n")
}
