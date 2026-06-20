package whenin

// ratio replicates Python difflib.SequenceMatcher(None, a, b).ratio().
//
// Our strings are short city names, so autojunk (which only flags chars in
// sequences >= 200 long) never triggers and there is no junk set — the junk
// extension steps in CPython's find_longest_match are therefore inert and
// omitted here. ratio = 2*M / (len(a)+len(b)), M = total matched runes.
func ratio(a, b string) float64 {
	ra := []rune(a)
	rb := []rune(b)
	total := len(ra) + len(rb)
	if total == 0 {
		return 1.0
	}
	return 2.0 * float64(matchingRunes(ra, rb)) / float64(total)
}

func matchingRunes(a, b []rune) int {
	b2j := make(map[rune][]int, len(b))
	for j, c := range b {
		b2j[c] = append(b2j[c], j)
	}

	type block struct{ alo, ahi, blo, bhi int }
	queue := []block{{0, len(a), 0, len(b)}}
	matched := 0
	for len(queue) > 0 {
		q := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		i, j, k := findLongestMatch(a, b2j, q.alo, q.ahi, q.blo, q.bhi)
		if k == 0 {
			continue
		}
		matched += k
		if q.alo < i && q.blo < j {
			queue = append(queue, block{q.alo, i, q.blo, j})
		}
		if i+k < q.ahi && j+k < q.bhi {
			queue = append(queue, block{i + k, q.ahi, j + k, q.bhi})
		}
	}
	return matched
}

func findLongestMatch(a []rune, b2j map[rune][]int, alo, ahi, blo, bhi int) (besti, bestj, bestsize int) {
	besti, bestj, bestsize = alo, blo, 0
	j2len := map[int]int{}
	for i := alo; i < ahi; i++ {
		newj2len := map[int]int{}
		for _, j := range b2j[a[i]] { // indices ascending by construction
			if j < blo {
				continue
			}
			if j >= bhi {
				break
			}
			k := j2len[j-1] + 1
			newj2len[j] = k
			if k > bestsize {
				besti, bestj, bestsize = i-k+1, j-k+1, k
			}
		}
		j2len = newj2len
	}
	return besti, bestj, bestsize
}
