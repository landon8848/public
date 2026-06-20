package whenin

import (
	"math"
	"sort"
	"strings"
	"unicode/utf8"
)

const Threshold = 0.60

var kindRank = map[string]int{"country": 1, "city": 0}

func rlen(s string) int { return utf8.RuneCountInString(s) }

func scoreText(q, text string) float64 {
	if text == "" || q == "" {
		return 0.0
	}
	if q == text {
		return 1.0
	}
	if strings.HasPrefix(text, q) {
		return 0.85 + 0.10*(float64(rlen(q))/float64(rlen(text)))
	}
	for _, tok := range strings.Fields(text) {
		if strings.HasPrefix(tok, q) {
			return 0.75 + 0.10*(float64(rlen(q))/float64(rlen(text)))
		}
	}
	if idx := strings.Index(text, q); idx >= 0 {
		pos := utf8.RuneCountInString(text[:idx])
		return 0.60 - math.Min(0.15, 0.01*float64(pos))
	}
	// Typo fallback — deliberately strict so it corrects misspellings
	// ("berln"->berlin) without matching different shorter words
	// ("chico"/"chiclayo"->chicago). Three gates: long enough query,
	// close lengths, high similarity.
	lq, lt := rlen(q), rlen(text)
	if lq < 5 {
		return 0.0
	}
	if float64(min(lq, lt))/float64(max(lq, lt)) < 0.80 {
		return 0.0
	}
	r := ratio(q, text)
	if r >= 0.82 {
		return r
	}
	return 0.0
}

// score assumes q is already normalized via Norm.
func score(q string, c Candidate) float64 {
	best := scoreText(q, c.SearchName)
	for _, alias := range c.Aliases {
		if best == 1.0 {
			break
		}
		if s := scoreText(q, alias); s > best {
			best = s
		}
	}
	switch {
	case c.Population > 5_000_000:
		best *= 1.08
	case c.Population > 1_000_000:
		best *= 1.05
	}
	return math.Min(best, 1.0)
}

// TopN ranks candidates for a query, returning at most n hits at or above
// the threshold. Mirrors Python matcher.top_n.
func TopN(query string, candidates []Candidate, n int) []Candidate {
	q := Norm(query)
	if q == "" {
		return nil
	}
	var hits []Candidate
	for _, c := range candidates {
		s := score(q, c)
		if s >= Threshold {
			c.Score = s
			hits = append(hits, c)
		}
	}
	// Scores within one 0.02 bucket are a tie, then ordered by kind
	// (country > city), population, exact score, and finally name.
	sort.SliceStable(hits, func(i, j int) bool {
		a, b := hits[i], hits[j]
		if ba, bb := bucket(a.Score), bucket(b.Score); ba != bb {
			return ba > bb
		}
		if ka, kb := kindRank[a.Kind], kindRank[b.Kind]; ka != kb {
			return ka > kb
		}
		if a.Population != b.Population {
			return a.Population > b.Population
		}
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		return a.Name < b.Name
	})
	if len(hits) > n {
		hits = hits[:n]
	}
	return hits
}

// bucket = Python round(score/0.02) using round-half-to-even.
func bucket(score float64) int {
	return int(math.RoundToEven(score / 0.02))
}
