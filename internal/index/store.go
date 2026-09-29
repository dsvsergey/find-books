// Package index stores libraries, books and works in SQLite with an FTS5
// index over normalized work titles, authors and book titles.
package index

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var (
	ErrNotFound = errors.New("бібліотеку не знайдено")
	ErrExists   = errors.New("бібліотека з такою назвою або шляхом уже є")
)

type Store struct{ db *sql.DB }

type Library struct {
	ID                                  int64
	Name, VolumeID, VolumeName, RootRel string
	LastScan                            time.Time
	Books, Works                        int
}

// Open opens (creating if needed) the index at path.
func Open(path string) (*Store, error) {
	dsn := path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("index %s: %w", path, err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) AddLibrary(name, volumeID, volumeName, rootRel string) (Library, error) {
	res, err := s.db.Exec(`INSERT INTO libraries(name, volume_id, volume_name, root_rel) VALUES (?, ?, ?, ?)`,
		name, volumeID, volumeName, rootRel)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return Library{}, fmt.Errorf("%w: %s", ErrExists, name)
		}
		return Library{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Library{}, err
	}
	return Library{ID: id, Name: name, VolumeID: volumeID, VolumeName: volumeName, RootRel: rootRel}, nil
}

const libSelect = `SELECT l.id, l.name, l.volume_id, l.volume_name, l.root_rel, l.last_scan_at,
	(SELECT count(*) FROM books b WHERE b.library_id = l.id),
	(SELECT count(*) FROM works w JOIN books b ON b.id = w.book_id WHERE b.library_id = l.id)
FROM libraries l`

func scanLibrary(row interface{ Scan(...any) error }) (Library, error) {
	var l Library
	var last int64
	if err := row.Scan(&l.ID, &l.Name, &l.VolumeID, &l.VolumeName, &l.RootRel, &last, &l.Books, &l.Works); err != nil {
		return Library{}, err
	}
	if last > 0 {
		l.LastScan = time.Unix(last, 0)
	}
	return l, nil
}

func (s *Store) Library(name string) (Library, error) {
	l, err := scanLibrary(s.db.QueryRow(libSelect+` WHERE l.name = ?`, name))
	if errors.Is(err, sql.ErrNoRows) {
		return Library{}, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return l, err
}

func (s *Store) Libraries() ([]Library, error) {
	rows, err := s.db.Query(libSelect + ` ORDER BY l.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Library
	for rows.Next() {
		l, err := scanLibrary(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Store) RemoveLibrary(name string) error {
	res, err := s.db.Exec(`DELETE FROM libraries WHERE name = ?`, name)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return nil
}

func (s *Store) MarkScanned(libraryID int64, t time.Time) error {
	_, err := s.db.Exec(`UPDATE libraries SET last_scan_at = ? WHERE id = ?`, t.Unix(), libraryID)
	return err
}
