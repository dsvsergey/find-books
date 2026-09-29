package tui

import (
	"github.com/sahilm/fuzzy"

	"findbooks/internal/index"
	"findbooks/internal/textnorm"
)

type rankSource []index.Hit

func (s rankSource) String(i int) string { return textnorm.Normalize(s[i].Title + " " + s[i].Author) }
func (s rankSource) Len() int            { return len(s) }

// rank orders FTS candidates by fuzzy closeness to the query; candidates the
// fuzzy matcher rejects keep their bm25 order after the matched ones.
func rank(query string, hits []index.Hit) []index.Hit {
	q := textnorm.Normalize(query)
	if q == "" || len(hits) < 2 {
		return hits
	}
	matches := fuzzy.FindFrom(q, rankSource(hits))
	out := make([]index.Hit, 0, len(hits))
	used := make([]bool, len(hits))
	for _, mt := range matches {
		out = append(out, hits[mt.Index])
		used[mt.Index] = true
	}
	for i, h := range hits {
		if !used[i] {
			out = append(out, h)
		}
	}
	return out
}
