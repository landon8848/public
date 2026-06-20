package whenin

import (
	"strings"
	"unicode"
)

// Mode is the lookup mode selected by the leading subcommand word.
type Mode string

const (
	ModeCity    Mode = "city"
	ModeTZ      Mode = "tz"
	ModeHoliday Mode = "holiday"
)

// Leading subcommand word -> mode. Keyword is "when", so the user types
// "when in rome" / "when tz pst" / "when is france". A bare query with no
// recognized subcommand ("when rome") falls back to city lookup.
var subcommands = map[string]Mode{"in": ModeCity, "tz": ModeTZ, "is": ModeHoliday}

// Query is the parsed user input. The when/whenTZ NLP seam (v2) lives here;
// v1 leaves them unset.
type Query struct {
	LocationText string
	Mode         Mode
}

// ParseQuery dispatches on the leading subcommand word, then treats the rest
// as the thing to look up. Mirrors Python query.parse_query.
func ParseQuery(text string) Query {
	text = strings.TrimSpace(text)
	if text == "" {
		return Query{LocationText: "", Mode: ModeCity}
	}
	first, rest := text, ""
	if i := strings.IndexFunc(text, unicode.IsSpace); i >= 0 {
		first = text[:i]
		rest = strings.TrimSpace(text[i:])
	}
	if mode, ok := subcommands[strings.ToLower(first)]; ok {
		return Query{LocationText: rest, Mode: mode}
	}
	return Query{LocationText: text, Mode: ModeCity}
}
