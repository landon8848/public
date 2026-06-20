package whenin

import "fmt"

const (
	globeEmoji = "\U0001F310" // 🌐 fallback when no country flag
	clockEmoji = "\U0001F552" // 🕒 used for tz-mode (code / iana) rows
)

// AlfredText is the optional text block (clipboard / large-type) of an item.
type AlfredText struct {
	Copy      string `json:"copy"`
	Largetype string `json:"largetype"`
}

// AlfredItem is one Script Filter result row.
type AlfredItem struct {
	UID          string      `json:"uid,omitempty"`
	Title        string      `json:"title"`
	Subtitle     string      `json:"subtitle"`
	Arg          string      `json:"arg,omitempty"`
	Autocomplete string      `json:"autocomplete,omitempty"`
	Valid        bool        `json:"valid"`
	Text         *AlfredText `json:"text,omitempty"`
}

// ToAlfred renders a candidate as a Script Filter item. Mirrors Python
// formatter.to_alfred. No uid: Alfred then preserves our matcher ranking
// instead of applying its own frequency reorder.
func ToAlfred(c Candidate, timeFormat string) AlfredItem {
	dt := NowIn(c.IANA, nil)
	clock := FmtClock(dt, timeFormat)
	day := FmtDay(dt)
	offset := FmtOffset(dt)

	// For `code` rows the entry name *is* the abbreviation (PST, JST…);
	// its Etc/* zone would otherwise render as "-08".
	abbrev := FmtAbbrev(dt)
	if c.Kind == "code" {
		abbrev = c.Name
	}

	var emoji, subtitle string
	switch c.Kind {
	case "city", "country":
		emoji = Flag(c.CountryCode)
		if emoji == "" {
			emoji = globeEmoji
		}
		subtitle = fmt.Sprintf("%s · %s (%s) · %s", c.Country, abbrev, offset, c.IANA)
	case "code":
		emoji = clockEmoji
		subtitle = fmt.Sprintf("%s · %s", c.Country, offset)
	default: // iana
		emoji = clockEmoji
		subtitle = fmt.Sprintf("%s (%s)", abbrev, offset)
	}

	title := fmt.Sprintf("%s %s — %s %s", emoji, c.Name, clock, day)
	copyStr := fmt.Sprintf("%s %s in %s (%s, %s)", clock, day, c.Name, abbrev, offset)

	return AlfredItem{
		Title:        title,
		Subtitle:     subtitle,
		Arg:          copyStr,
		Autocomplete: c.Name,
		Valid:        true,
		Text:         &AlfredText{Copy: copyStr, Largetype: copyStr},
	}
}

// NoMatch is the row shown when nothing scores above the threshold.
func NoMatch(queryText string) AlfredItem {
	return AlfredItem{
		UID:      "no-match",
		Title:    fmt.Sprintf("No match for ‘%s’", queryText),
		Subtitle: "Try a city or country name",
		Valid:    false,
	}
}

// HolidayStub is the placeholder for the not-yet-built `when is …` mode.
func HolidayStub() AlfredItem {
	return AlfredItem{
		UID:      "holiday-soon",
		Title:    "Holiday lookup — coming soon",
		Subtitle: "`when is <country>` will list upcoming public holidays",
		Valid:    false,
	}
}

// ErrorItem is the top-level guard row; arg carries details to copy.
func ErrorItem(message, details string) AlfredItem {
	return AlfredItem{
		UID:      "error",
		Title:    "⚠️  Something went wrong",
		Subtitle: fmt.Sprintf("%s · Press Enter to copy details", message),
		Arg:      details,
		Valid:    true,
		Text:     &AlfredText{Copy: details, Largetype: details},
	}
}
