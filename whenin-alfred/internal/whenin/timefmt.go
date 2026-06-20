package whenin

import (
	"fmt"
	"time"

	_ "time/tzdata" // embed the IANA tz database: zero runtime deps
)

// NowIn resolves the target wall-clock time in zone iana.
//
// v1 always passes when=nil -> current time in the target zone. The when
// param is the v2 NLP seam: a time.Time already carrying its own location
// (Go's equivalent of Python's when/when_tz) is converted to iana.
func NowIn(iana string, when *time.Time) time.Time {
	loc, err := time.LoadLocation(iana)
	if err != nil {
		loc = time.UTC
	}
	if when == nil {
		return time.Now().In(loc)
	}
	return when.In(loc)
}

// FmtClock renders "3:04 PM" (12h) or "15:04" (24h).
func FmtClock(t time.Time, format string) string {
	if format == "24h" {
		return t.Format("15:04")
	}
	return t.Format("3:04 PM")
}

// FmtDay renders the abbreviated weekday, e.g. "Fri".
func FmtDay(t time.Time) string {
	return t.Format("Mon")
}

// FmtOffset renders "UTC+2" or "UTC-5:30" (no minutes when zero), matching
// Python timefmt.fmt_offset.
func FmtOffset(t time.Time) string {
	_, offset := t.Zone()
	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}
	h := offset / 3600
	m := (offset % 3600) / 60
	if m == 0 {
		return fmt.Sprintf("UTC%s%d", sign, h)
	}
	return fmt.Sprintf("UTC%s%d:%02d", sign, h, m)
}

// FmtAbbrev returns the zone abbreviation like "CEST". For Etc/* fixed zones
// this is numeric ("+09"); callers that have a better label prefer their own.
func FmtAbbrev(t time.Time) string {
	name, _ := t.Zone()
	return name
}

// Flag maps a 2-letter country code to its regional-indicator emoji, or ""
// when the code is missing/invalid.
func Flag(cc string) string {
	if len(cc) != 2 {
		return ""
	}
	const base = 0x1F1E6
	out := make([]rune, 0, 2)
	for _, r := range cc {
		if r >= 'a' && r <= 'z' {
			r -= 'a' - 'A'
		}
		if r < 'A' || r > 'Z' {
			return ""
		}
		out = append(out, rune(base+int(r-'A')))
	}
	return string(out)
}
