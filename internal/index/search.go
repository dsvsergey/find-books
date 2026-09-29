package index

import (
	"strings"

	"findbooks/internal/textnorm"
)

type Query struct {
	Text, Author string
	Limit        int
}

type Hit struct {
	WorkID                        int64
	WorkOrd                       int // position of the work in its book
	Title, Author, TreePath       string
	BookID                        int64
	BookTitle, BookYear           string
	IsCollection                  bool
	RelPath, Format               string
	Size, MTime                   int64 // as indexed; Size -1 if the parse failed
	LibraryName                   string
	VolumeID, VolumeName, RootRel string
}

// MatchExpr turns free text into an FTS5 expression: every normalized word
// becomes a quoted prefix term; terms are ANDed. Normalization strips quotes
// and operators, so user input can never inject FTS5 syntax.
func MatchExpr(text string) string {
	words := strings.Fields(textnorm.Normalize(text))
	for i, w := range words {
		words[i] = `"` + w + `"*`
	}
	return strings.Join(words, " ")
}

const searchSQL = `
SELECT w.id, w.ord, w.title, w.author, w.tree_path,
       b.id, b.title, b.year, b.is_collection, b.rel_path, b.format, b.size, b.mtime,
       l.name, l.volume_id, l.volume_name, l.root_rel
FROM works_fts
JOIN works w ON w.id = works_fts.rowid
JOIN books b ON b.id = w.book_id
JOIN libraries l ON l.id = b.library_id
WHERE works_fts MATCH ?
ORDER BY bm25(works_fts, 10.0, 3.0, 1.0)
LIMIT ?`

func (s *Store) Search(q Query) ([]Hit, error) {
	expr := MatchExpr(q.Text)
	if a := MatchExpr(q.Author); a != "" {
		col := "author : (" + a + ")"
		if expr == "" {
			expr = col
		} else {
			expr = "(" + expr + ") AND " + col
		}
	}
	if expr == "" {
		return nil, nil
	}
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(searchSQL, expr, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hits []Hit
	for rows.Next() {
		var h Hit
		if err := rows.Scan(&h.WorkID, &h.WorkOrd, &h.Title, &h.Author, &h.TreePath,
			&h.BookID, &h.BookTitle, &h.BookYear, &h.IsCollection, &h.RelPath, &h.Format, &h.Size, &h.MTime,
			&h.LibraryName, &h.VolumeID, &h.VolumeName, &h.RootRel); err != nil {
			return nil, err
		}
		hits = append(hits, h)
	}
	return hits, rows.Err()
}

// BookWorks returns the titles of a book's works in book order.
func (s *Store) BookWorks(bookID int64) ([]string, error) {
	rows, err := s.db.Query(`SELECT title FROM works WHERE book_id = ? ORDER BY ord`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) TotalWorks() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT count(*) FROM works`).Scan(&n)
	return n, err
}
