// Package tui is the interactive search screen.
package tui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"findbooks/internal/index"
	"findbooks/internal/library"
)

type Searcher interface {
	Search(q index.Query) ([]index.Hit, error)
	BookWorks(bookID int64) ([]string, error)
}

// Actions are the side effects the screen triggers; tests replace them.
type Actions struct {
	Root   func(volumeID, rootRel string) (string, bool)
	Open   func(path string) error
	Reveal func(path string) error
	Copy   func(text string) error
	Exists func(path string) bool // does the file exist on the mounted disk
}

const (
	debounce       = 50 * time.Millisecond
	candidateLimit = 500
)

type searchMsg struct{ seq int }

type resultsMsg struct {
	seq    int
	hits   []index.Hit
	online map[string]bool
	err    error
}

type Model struct {
	src      Searcher
	act      Actions
	total    int
	input    textinput.Model
	hits     []index.Hit
	cursor   int
	seq      int
	searched bool
	online   map[string]bool    // volume id -> mounted
	toc      map[int64][]string // book id -> work titles (cache)
	status   string
	err      error
	width    int
	height   int
}

func New(src Searcher, act Actions, total int) Model {
	in := textinput.New()
	in.Prompt = "🔎 "
	in.Placeholder = "назва твору, автор або збірка…"
	in.Focus()
	return Model{src: src, act: act, total: total, input: in, online: map[string]bool{}, toc: map[int64][]string{}}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.SetWidth(max(10, msg.Width-20))
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
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
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
			m.withFile(m.act.Copy, "Шлях скопійовано")
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
	before := m.input.Value()
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != before {
		return m, tea.Batch(cmd, m.queueSearch())
	}
	return m, cmd
}

// queueSearch starts a new debounce period; only the latest one searches.
func (m *Model) queueSearch() tea.Cmd {
	m.seq++
	m.status = ""
	seq := m.seq
	return tea.Tick(debounce, func(time.Time) tea.Msg { return searchMsg{seq: seq} })
}

func (m Model) searchCmd(seq int, q string) tea.Cmd {
	src := m.src
	root := m.act.Root
	return func() tea.Msg {
		hits, err := src.Search(index.Query{Text: q, Limit: candidateLimit})
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
	root, ok := m.act.Root(h.VolumeID, h.RootRel)
	if !ok {
		m.status = fmt.Sprintf("Диск «%s» не підключено — підключіть його, щоб відкрити файл", h.VolumeName)
		return
	}
	p := library.FilePath(root, h.RelPath)
	if !m.act.Exists(p) {
		m.status = "Файл не знайдено: " + h.RelPath
		return
	}
	if err := fn(p); err != nil {
		m.status = "Помилка: " + err.Error()
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
	if titles, err := m.src.BookWorks(h.BookID); err == nil {
		m.toc[h.BookID] = titles
	}
}
