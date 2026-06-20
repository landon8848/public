package whenin

// Candidate is one searchable place (or, in tz mode, a zone/code).
// JSON tags match the build/build_index.py output and workflow/data/index.json.
type Candidate struct {
	Name        string   `json:"name"`
	SearchName  string   `json:"search_name"`
	Kind        string   `json:"kind"` // "city" | "country" | "code" | "iana"
	IANA        string   `json:"iana"`
	Country     string   `json:"country"`
	CountryCode string   `json:"country_code"` // JSON null -> ""
	Aliases     []string `json:"aliases"`
	Population  int      `json:"population"`

	Score float64 `json:"-"` // set by the matcher, never serialized
}
