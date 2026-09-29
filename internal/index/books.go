package index

import (
	"database/sql"
	"fmt"
	"path"

	"findbooks/internal/textnorm"
)

type FileStamp struct{ BookID, Size, MTime int64 }

type WorkRecord struct{ Title, Author, TreePath string }

type BookRecord struct {
	RelPath, Format      string
	Size, MTime          int64
	Title, Authors, Year string
	IsCollection         bool
	Works                []WorkRecord
}

// Files returns the stored size/mtime of every book in the library, keyed by rel_path.
func (s *Store) Files(libraryID int64) (map[string]FileStamp, error) {
	rows, err := s.db.Query(`SELECT id, rel_path, size, mtime FROM books WHERE library_id = ?`, libraryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]FileStamp{}
	for rows.Next() {
		var rel string
		var f FileStamp
		if err := rows.Scan(&f.BookID, &rel, &f.Size, &f.MTime); err != nil {
			return nil, err
		}
		out[rel] = f
	}
	return out, rows.Err()
}

// WriteBatch deletes the books in del and inserts or replaces the books in
// put, all in one transaction.
func (s *Store) WriteBatch(libraryID int64, put []BookRecord, del []int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, id := range del {
		if _, err := tx.Exec(`DELETE FROM books WHERE id = ?`, id); err != nil {
			return err
		}
	}
	for _, b := range put {
		if err := putBook(tx, libraryID, b); err != nil {
			return fmt.Errorf("%s: %w", b.RelPath, err)
		}
	}
	return tx.Commit()
}

func putBook(tx *sql.Tx, libraryID int64, b BookRecord) error {
	if _, err := tx.Exec(`DELETE FROM books WHERE library_id = ? AND rel_path = ?`, libraryID, b.RelPath); err != nil {
		return err
	}
	res, err := tx.Exec(`INSERT INTO books(library_id, rel_path, format, size, mtime, title, authors, year, is_collection)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		libraryID, b.RelPath, b.Format, b.Size, b.MTime, b.Title, b.Authors, b.Year, b.IsCollection)
	if err != nil {
		return err
	}
	bookID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	bookText := textnorm.Normalize(b.Title + " " + folderOf(b.RelPath))
	for i, w := range b.Works {
		res, err := tx.Exec(`INSERT INTO works(book_id, ord, title, author, tree_path) VALUES (?, ?, ?, ?, ?)`,
			bookID, i, w.Title, w.Author, w.TreePath)
		if err != nil {
			return err
		}
		workID, err := res.LastInsertId()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO works_fts(rowid, title, author, book_title) VALUES (?, ?, ?, ?)`,
			workID, textnorm.Normalize(w.Title), textnorm.Normalize(w.Author), bookText); err != nil {
			return err
		}
	}
	return nil
}

func folderOf(rel string) string {
	if d := path.Dir(rel); d != "." {
		return d
	}
	return ""
}
