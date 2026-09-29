package preview

import (
	"path/filepath"

	"findbooks/internal/extract"
	"findbooks/internal/fb2"
	"findbooks/internal/index"
	"findbooks/internal/scan"
)

// parseFB2 is fb2.ParseTextFile; a variable so tests can make it panic.
var parseFB2 = fb2.ParseTextFile

func loadFB2(absPath string, h index.Hit) (Doc, error) {
	b, err := parseFB2(absPath)
	if err != nil {
		return Doc{}, err
	}
	doc := Doc{Title: h.Title, Author: h.Author}
	if !h.IsCollection {
		doc.Blocks = blocksOf(b.Sections)
		return doc, nil
	}
	// Same fallback as scan, so the heuristic sees the same book.
	if b.Title == "" {
		b.Title = scan.Stem(filepath.Base(absPath))
	}
	works, _ := extract.Works(b)
	w, exact := pick(works, h)
	if w.Section == nil {
		doc.Blocks, doc.Stale = blocksOf(b.Sections), true
		return doc, nil
	}
	doc.Blocks, doc.Stale = blocksOf([]*fb2.Section{w.Section}), !exact
	return doc, nil
}

// pick finds the indexed work: at its position when the title still
// matches (exact), else the first work with that title, else the zero Work.
func pick(works []extract.Work, h index.Hit) (extract.Work, bool) {
	if h.WorkOrd >= 0 && h.WorkOrd < len(works) && works[h.WorkOrd].Title == h.Title {
		return works[h.WorkOrd], true
	}
	for _, w := range works {
		if w.Title == h.Title {
			return w, false
		}
	}
	return extract.Work{}, false
}

// blocksOf returns the blocks of secs and all their subsections in
// document order.
func blocksOf(secs []*fb2.Section) []fb2.Block {
	var out []fb2.Block
	for _, s := range secs {
		out = append(out, s.Blocks...)
		out = append(out, blocksOf(s.Children)...)
	}
	return out
}
