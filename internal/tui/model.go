// Package tui is the interactive search screen and the libraries screen.
package tui

import (
	"context"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"findbooks/internal/i18n"
	"findbooks/internal/index"
	"findbooks/internal/library"
	"findbooks/internal/preview"
	"findbooks/internal/scan"
)

type Searcher interface {
	Search(q index.Query) ([]index.Hit, error)
	BookWorks(bookID int64) ([]string, error)
}

// Backend is everything the screens need from the index.
type Backend interface {
	Searcher
	Libraries() ([]index.Library, error)
	TotalWorks() (int, error)
	RemoveLibrary(name string) error
	// AddFolder registers dir as a library and runs its first scan. A
	// returned library with ID != 0 was registered even when err != nil.
	AddFolder(ctx context.Context, dir string, onProgress func(scan.Progress)) (index.Library, scan.Report, error)
	// Rescan updates a registered library.
	Rescan(ctx context.Context, lib index.Library, onProgress func(scan.Progress)) (scan.Report, error)
}

// Actions are the side effects the screens trigger; tests replace them.
type Actions struct {
	Root         func(volumeID, rootRel string) (string, bool)
	Open         func(path string) error
	Reveal       func(path string) error
	Copy         func(text string) error
	Exists       func(path string) bool // does the file exist on the mounted disk
	ChooseFolder func(prompt string) (string, error)
	SaveLang     func(i18n.Lang) error
	// Getenv looks up an environment variable; nil is treated like a
	// function that always returns "". Used to detect that FINDBOOKS_LANG
	// overrides the setting SaveLang just persisted.
	Getenv func(string) string
	// Preview loads the text of the work h from the file at path.
	Preview func(path string, h index.Hit) (preview.Doc, error)
}

const (
	debounce       = 50 * time.Millisecond
	candidateLimit = 500
	// defaultInputWidth is used for the search field before the first
	// tea.WindowSizeMsg arrives; without it textinput reports Width() == 0
	// and truncates the placeholder to a single character. It mirrors the
	// fallback render width (100) minus the same margin Update applies on
	// resize (max(10, width-20)).
	defaultInputWidth = 80
)

type screen int

const (
	screenSearch screen = iota
	screenLibraries
	screenPreview
)

type searchMsg struct{ seq int }

// langSavedMsg reports the result of the background call to Actions.SaveLang
// triggered by toggleLang.
type langSavedMsg struct{ err error }

type resultsMsg struct {
	seq    int
	hits   []index.Hit
	online map[string]bool
	err    error
}

type Model struct {
	b      Backend
	act    Actions
	total  int
	screen screen

	// search screen
	input    textinput.Model
	hits     []index.Hit
	cursor   int
	seq      int
	searched bool
	online   map[string]bool    // volume id -> mounted
	toc      map[int64][]string // book id -> work titles (cache)
	err      error

	// libraries screen
	libs          []index.Library
	libOnline     map[string]bool
	libCursor     int
	confirmDelete bool
	choosing      bool  // waiting for the Finder folder dialog to resolve
	removing      bool  // waiting for a removeLibrary to resolve
	quitting      bool  // ctrl+c pressed during a job; quit once it stops
	selectLibID   int64 // select this library the next time librariesMsg loads
	job           *job
	bar           progress.Model

	// preview screen
	pv        preview.Doc
	pvHit     index.Hit
	pvPath    string
	pvView    viewport.Model
	pvSeq     int  // only the previewMsg with this seq is applied
	pvLoading bool // a preview load for pvSeq is in flight

	status string
	width  int
	height int
}

// New builds the UI; with an empty index (total == 0) it opens on the
// libraries screen.
func New(b Backend, act Actions, total int) Model {
	in := textinput.New()
	in.Prompt = "🔎 "
	in.Placeholder = i18n.T(i18n.KeySearchPlaceholder)
	in.SetWidth(defaultInputWidth) // Update replaces this with the real width
	in.Focus()
	m := Model{
		b: b, act: act, total: total, input: in,
		online: map[string]bool{}, toc: map[int64][]string{}, libOnline: map[string]bool{},
		bar: progress.New(progress.WithDefaultBlend(), progress.WithWidth(40)),
	}
	if total == 0 {
		m.screen = screenLibraries
	}
	return m
}

func (m Model) Init() tea.Cmd { return m.loadLibraries() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.SetWidth(max(10, msg.Width-20))
		m.bar.SetWidth(min(60, max(10, msg.Width-20)))
		m.layoutPreview()
		return m, nil
	case searchMsg:
		if msg.seq != m.seq {
			return m, nil
		}
		return m, m.searchCmd(msg.seq, m.input.Value())
	case resultsMsg:
		if msg.seq != m.seq {
			return m, nil
		}
		m.err = msg.err
		m.hits = rank(m.input.Value(), msg.hits)
		m.cursor = 0
		m.searched = true
		m.online = msg.online
		m.loadTOC()
		return m, nil
	case librariesMsg, folderMsg, progressMsg, jobDoneMsg, removedMsg:
		return m.updateLibraries(msg)
	case previewMsg:
		return m.previewLoaded(msg), nil
	case quitTimeoutMsg:
		return m, tea.Quit
	case langSavedMsg:
		switch {
		case msg.err != nil:
			m.status = i18n.T(i18n.KeyLangSaveFailed, msg.err.Error())
		case m.act.Getenv != nil:
			if envVal := m.act.Getenv("FINDBOOKS_LANG"); envVal != "" {
				if envLang, perr := i18n.Parse(envVal); perr != nil || envLang != i18n.Current() {
					m.status = i18n.T(i18n.KeyEnvOverridesLang, envVal)
				}
			}
		}
		return m, nil
	case tea.KeyPressMsg:
		if m.screen == screenPreview {
			return m.previewKey(msg)
		}
		if m.screen == screenLibraries {
			return m.libraryKey(msg)
		}
		if m.pvLoading { // any key abandons a pending preview; esc only that
			m.pvLoading = false
			m.pvSeq++
			m.status = ""
			if msg.String() == "esc" {
				return m, nil
			}
		}
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "ctrl+l":
			m.screen = screenLibraries
			m.status = ""
			return m, m.loadLibraries()
		case "ctrl+g":
			return m, m.toggleLang()
		case "ctrl+r":
			return m, m.openPreview()
		case "up", "ctrl+p":
			m.move(-1)
			return m, nil
		case "down", "ctrl+n":
			m.move(1)
			return m, nil
		case "enter":
			m.withFile(m.act.Open, "")
			return m, nil
		case "ctrl+o":
			m.withFile(m.act.Reveal, "")
			return m, nil
		case "ctrl+y":
			m.withFile(m.act.Copy, i18n.T(i18n.KeyPathCopied))
			return m, nil
		case "tab":
			if h := m.selected(); h != nil && h.Author != "" {
				m.input.SetValue(h.Author)
				m.input.CursorEnd()
				return m, m.queueSearch()
			}
			return m, nil
		}
	}
	if m.screen != screenSearch {
		return m, nil
	}
	before := m.input.Value()
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != before {
		return m, tea.Batch(cmd, m.queueSearch())
	}
	return m, cmd
}

// toggleLang switches EN↔UK, refreshes texts cached in widgets and saves the
// choice in the background.
func (m *Model) toggleLang() tea.Cmd {
	next := i18n.UK
	if i18n.Current() == i18n.UK {
		next = i18n.EN
	}
	i18n.Set(next)
	m.input.Placeholder = i18n.T(i18n.KeySearchPlaceholder)
	m.status = ""
	save := m.act.SaveLang
	return func() tea.Msg { return langSavedMsg{err: save(next)} }
}

// queueSearch starts a new debounce period; only the latest one searches.
func (m *Model) queueSearch() tea.Cmd {
	m.seq++
	m.status = ""
	seq := m.seq
	return tea.Tick(debounce, func(time.Time) tea.Msg { return searchMsg{seq: seq} })
}

func (m Model) searchCmd(seq int, q string) tea.Cmd {
	b := m.b
	root := m.act.Root
	return func() tea.Msg {
		hits, err := b.Search(index.Query{Text: q, Limit: candidateLimit})
		online := map[string]bool{}
		for _, h := range hits {
			if _, seen := online[h.VolumeID]; !seen {
				_, ok := root(h.VolumeID, h.RootRel)
				online[h.VolumeID] = ok
			}
		}
		return resultsMsg{seq: seq, hits: hits, online: online, err: err}
	}
}

func (m *Model) move(d int) {
	if len(m.hits) == 0 {
		return
	}
	m.cursor = min(max(m.cursor+d, 0), len(m.hits)-1)
	m.loadTOC()
}

func (m Model) selected() *index.Hit {
	if m.cursor < 0 || m.cursor >= len(m.hits) {
		return nil
	}
	return &m.hits[m.cursor]
}

func (m *Model) withFile(fn func(string) error, okStatus string) {
	h := m.selected()
	if h == nil {
		return
	}
	if p, ok := m.filePath(*h); ok {
		m.runOn(fn, p, okStatus)
	}
}

// filePath resolves h's file on its mounted disk; when the disk is offline
// or the file is gone it sets the status and returns false.
func (m *Model) filePath(h index.Hit) (string, bool) {
	root, ok := m.act.Root(h.VolumeID, h.RootRel)
	if !ok {
		m.status = i18n.T(i18n.KeyDiskOfflineOpen, h.VolumeName)
		return "", false
	}
	p := library.FilePath(root, h.RelPath)
	if !m.act.Exists(p) {
		m.status = i18n.T(i18n.KeyFileNotFound, h.RelPath)
		return "", false
	}
	return p, true
}

func (m *Model) runOn(fn func(string) error, p, okStatus string) {
	if err := fn(p); err != nil {
		m.status = i18n.T(i18n.KeyErrorStatus, err.Error())
		return
	}
	m.status = okStatus
}

func (m *Model) loadTOC() {
	h := m.selected()
	if h == nil || !h.IsCollection {
		return
	}
	if _, ok := m.toc[h.BookID]; ok {
		return
	}
	if titles, err := m.b.BookWorks(h.BookID); err == nil {
		m.toc[h.BookID] = titles
	}
}
