package index

// works_fts holds already-normalized text; rowid = works.id. Rows are removed
// by the trigger, which also fires for cascaded deletes from books/libraries.
const schema = `
CREATE TABLE IF NOT EXISTS libraries (
	id           INTEGER PRIMARY KEY,
	name         TEXT NOT NULL UNIQUE,
	volume_id    TEXT NOT NULL,
	volume_name  TEXT NOT NULL,
	root_rel     TEXT NOT NULL,
	last_scan_at INTEGER NOT NULL DEFAULT 0,
	UNIQUE (volume_id, root_rel)
);
CREATE TABLE IF NOT EXISTS books (
	id            INTEGER PRIMARY KEY,
	library_id    INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
	rel_path      TEXT NOT NULL,
	format        TEXT NOT NULL,
	size          INTEGER NOT NULL,
	mtime         INTEGER NOT NULL,
	title         TEXT NOT NULL,
	authors       TEXT NOT NULL,
	year          TEXT NOT NULL,
	is_collection INTEGER NOT NULL,
	UNIQUE (library_id, rel_path)
);
CREATE TABLE IF NOT EXISTS works (
	id        INTEGER PRIMARY KEY,
	book_id   INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	ord       INTEGER NOT NULL,
	title     TEXT NOT NULL,
	author    TEXT NOT NULL,
	tree_path TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS works_book ON works(book_id);
CREATE VIRTUAL TABLE IF NOT EXISTS works_fts USING fts5(
	title, author, book_title,
	tokenize = 'unicode61 remove_diacritics 0',
	prefix = '2 3'
);
CREATE TRIGGER IF NOT EXISTS works_fts_delete AFTER DELETE ON works BEGIN
	DELETE FROM works_fts WHERE rowid = old.id;
END;
`
