# findbooks

English | [Українська](README.uk.md)

Find works in e-book libraries — including inside collections, almanacs and anthologies.

## Install

```bash
go build -o findbooks ./cmd/findbooks
```

## Usage

```bash
findbooks add '/Volumes/dsvDev/Library/Soviet Sci-Fi'   # register and index
findbooks                                                # interactive search and library management
findbooks search alien children                          # search from the command line
findbooks search --author pidorenko --json
findbooks update --all                                   # re-index changed files
findbooks list
findbooks remove 'Soviet Sci-Fi'
```

The index lives at `~/.local/share/findbooks/index.db` (`volumes.json` and `config.json` live alongside it; use `--db <file>` for a different index). Search still works when the library's disk is not mounted: such results are marked `○`.

## Language

The interface is English by default. Ukrainian can be selected:

- permanently: `findbooks config lang uk` (back to English — `findbooks config lang en`);
- for one run: `findbooks --lang uk …` or the `FINDBOOKS_LANG=uk` environment variable;
- in the TUI: `ctrl+g` toggles EN ↔ UK and remembers the choice.

Precedence: `--lang` > `FINDBOOKS_LANG` > saved setting > English.

In the TUI: `enter` — open, `ctrl+l` — libraries, `ctrl+o` — show in Finder, `ctrl+y` — copy path, `tab` — search by author, `ctrl+g` — language, `↑`/`↓` — select, `esc` — quit.

### Libraries in the TUI

If the index is empty, `findbooks` opens the "Libraries" screen; `ctrl+l` leads to it from search.

- `a` — pick a folder in the Finder dialog and index it (progress is shown on screen, `esc` cancels);
- `u` — update the selected library;
- `d` — remove a library from the index (files are not touched; confirm with `y`);
- `esc` — back to search (quits if the index is empty).

The keys also work on a Ukrainian keyboard layout (`ф`, `г`, `в`, `н`).

FB2 (and `.fb2.zip`) is indexed along with the table of contents of collections; PDF, DJVU, DOC, DOCX, RTF, TXT, EPUB, MOBI are indexed by file name and folder.
