package whenin

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// stripDiacritics mirrors Python normalize.strip_diacritics: NFKD then
// drop combining marks (Unicode category Mn).
func stripDiacritics(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range norm.NFKD.String(s) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Norm is the canonical search normalization. It MUST stay byte-identical
// in behaviour to the Python normalize.norm used at index-build time, or
// matches silently fail.
//
// Treats / _ - as spaces so "Europe/Berlin", "Port-au-Prince" and
// "New_York" match their space-separated index forms, then collapses
// whitespace runs.
func Norm(s string) string {
	s = stripDiacritics(s)
	// casefold approximation: ToLower covers our (post-diacritic) charset;
	// handle ß explicitly since Python casefold maps it to "ss".
	s = strings.ReplaceAll(s, "ß", "ss")
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r == '/' || r == '_' || r == '-' {
			return ' '
		}
		return r
	}, s)
	return strings.Join(strings.Fields(s), " ")
}
