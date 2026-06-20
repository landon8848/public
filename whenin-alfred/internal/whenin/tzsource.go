package whenin

import (
	_ "embed"
	"strings"
	"sync"
)

//go:embed zones.txt
var zonesData string

// tzCode is a curated abbreviation -> (full name, IANA zone).
// Fixed-offset codes use Etc/GMT zones. NOTE the POSIX sign inversion:
//
//	Etc/GMT+5 == UTC-5    Etc/GMT-9 == UTC+9
//
// Half-hour / DST-bearing codes map to a representative real zone.
type tzCode struct{ code, full, iana string }

var tzCodes = []tzCode{
	{"UTC", "Coordinated Universal Time", "UTC"},
	{"GMT", "Greenwich Mean Time", "Etc/GMT"},
	{"EST", "Eastern Standard Time (US)", "Etc/GMT+5"},
	{"EDT", "Eastern Daylight Time (US)", "Etc/GMT+4"},
	{"CST", "Central Standard Time (US)", "Etc/GMT+6"},
	{"CDT", "Central Daylight Time (US)", "Etc/GMT+5"},
	{"MST", "Mountain Standard Time (US)", "Etc/GMT+7"},
	{"MDT", "Mountain Daylight Time (US)", "Etc/GMT+6"},
	{"PST", "Pacific Standard Time (US)", "Etc/GMT+8"},
	{"PDT", "Pacific Daylight Time (US)", "Etc/GMT+7"},
	{"AKST", "Alaska Standard Time", "Etc/GMT+9"},
	{"HST", "Hawaii Standard Time", "Etc/GMT+10"},
	{"AST", "Atlantic Standard Time", "Etc/GMT+4"},
	{"NST", "Newfoundland Standard Time", "America/St_Johns"},
	{"WET", "Western European Time", "Etc/GMT"},
	{"WEST", "Western European Summer Time", "Etc/GMT-1"},
	{"BST", "British Summer Time", "Etc/GMT-1"},
	{"CET", "Central European Time", "Etc/GMT-1"},
	{"CEST", "Central European Summer Time", "Etc/GMT-2"},
	{"EET", "Eastern European Time", "Etc/GMT-2"},
	{"EEST", "Eastern European Summer Time", "Etc/GMT-3"},
	{"MSK", "Moscow Standard Time", "Europe/Moscow"},
	{"IST", "India Standard Time", "Asia/Kolkata"},
	{"PKT", "Pakistan Standard Time", "Asia/Karachi"},
	{"GST", "Gulf Standard Time", "Asia/Dubai"},
	{"WIB", "Western Indonesia Time", "Asia/Jakarta"},
	{"ICT", "Indochina Time", "Asia/Bangkok"},
	{"SGT", "Singapore Time", "Asia/Singapore"},
	{"HKT", "Hong Kong Time", "Asia/Hong_Kong"},
	{"PHT", "Philippine Time", "Asia/Manila"},
	{"JST", "Japan Standard Time", "Etc/GMT-9"},
	{"KST", "Korea Standard Time", "Etc/GMT-9"},
	{"AWST", "Australian Western Time", "Etc/GMT-8"},
	{"ACST", "Australian Central Time", "Australia/Adelaide"},
	{"AEST", "Australian Eastern Time", "Etc/GMT-10"},
	{"AEDT", "Australian Eastern Daylight", "Etc/GMT-11"},
	{"NZST", "New Zealand Standard Time", "Etc/GMT-12"},
	{"WAT", "West Africa Time", "Africa/Lagos"},
	{"CAT", "Central Africa Time", "Africa/Maputo"},
	{"EAT", "East Africa Time", "Africa/Nairobi"},
	{"SAST", "South Africa Standard Time", "Africa/Johannesburg"},
	{"ART", "Argentina Time", "America/Argentina/Buenos_Aires"},
	{"BRT", "Brasilia Time", "America/Sao_Paulo"},
}

// popularTZNames are the defaults for an empty `when tz ` query.
var popularTZNames = []string{
	"UTC", "GMT", "America/New_York", "America/Los_Angeles",
	"Europe/London", "Asia/Tokyo",
}

var (
	tzOnce  sync.Once
	tzCands []Candidate
)

// TZCandidates returns timezone candidates for `when tz …`: curated codes
// plus every IANA zone. Mirrors Python tzsource.tz_candidates.
func TZCandidates() []Candidate {
	tzOnce.Do(func() {
		for _, tc := range tzCodes {
			tzCands = append(tzCands, Candidate{
				Name:       tc.code,
				SearchName: Norm(tc.code),
				Kind:       "code",
				IANA:       tc.iana,
				Country:    tc.full,
			})
		}
		for _, zone := range strings.Split(strings.TrimSpace(zonesData), "\n") {
			zone = strings.TrimSpace(zone)
			if zone == "" {
				continue
			}
			leaf := zone
			if i := strings.LastIndex(zone, "/"); i >= 0 {
				leaf = zone[i+1:]
			}
			// Leaf alias is fine here (unlike the city index): in tz mode
			// no city competes, so `when tz tokyo` -> Asia/Tokyo.
			tzCands = append(tzCands, Candidate{
				Name:       zone,
				SearchName: Norm(zone),
				Kind:       "iana",
				IANA:       zone,
				Aliases:    []string{Norm(leaf)},
			})
		}
	})
	return tzCands
}

// PopularTZ returns the curated default tz rows for an empty tz query.
func PopularTZ() []Candidate {
	byName := map[string]Candidate{}
	for _, c := range TZCandidates() {
		byName[c.Name] = c
	}
	var out []Candidate
	for _, n := range popularTZNames {
		if c, ok := byName[n]; ok {
			out = append(out, c)
		}
	}
	return out
}
