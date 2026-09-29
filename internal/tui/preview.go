package tui

import (
	"errors"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"findbooks/internal/i18n"
	"findbooks/internal/index"
	"findbooks/internal/preview"
)

type previewMsg struct {
	seq  int
	path string
	hit  index.Hit
	doc  preview.Doc
	err  error
}

// previewChrome counts the screen lines around the text: two header lines,
// two separators and the help line.
const previewChrome = 5

// openPreview starts loading the selected work; it sets a status and
// returns nil when the preview cannot be shown.
func (m *Model) openPreview() tea.Cmd {
	h := m.selected()
	if h == nil {
		return nil
	}
	if !preview.Supported(h.Format) {
		m.status = i18n.T(i18n.KeyPreviewUnsupported, strings.ToUpper(h.Format))
		return nil
	}
	p, ok := m.filePath(*h)
	if !ok {
		return nil
	}
	m.pvSeq++
	m.pvLoading = true
	m.status = i18n.T(i18n.KeyPreviewLoading)
	seq, hit, load := m.pvSeq, *h, m.act.Preview
	return func() tea.Msg {
		doc, err := load(p, hit)
		return previewMsg{seq: seq, path: p, hit: hit, doc: doc, err: err}
	}
}

func (m Model) previewLoaded(msg previewMsg) Model {
	if msg.seq != m.pvSeq || !m.pvLoading || m.screen != screenSearch {
		return m
	}
	m.pvLoading = false
	switch {
	case errors.Is(msg.err, preview.ErrUnsupported):
		m.status = i18n.T(i18n.KeyPreviewUnsupported, strings.ToUpper(msg.hit.Format))
		return m
	case msg.err != nil:
		m.status = i18n.T(i18n.KeyPreviewFailed, msg.err.Error())
		return m
	}
	m.status = ""
	m.screen = screenPreview
	m.pv, m.pvHit, m.pvPath = msg.doc, msg.hit, msg.path
	m.pvView = viewport.New()
	m.layoutPreview()
	return m
}

// layoutPreview sizes the text to the window and re-wraps it, keeping the
// scroll position.
func (m *Model) layoutPreview() {
	if m.screen != screenPreview {
		return
	}
	w, h := m.width, m.height
	if w <= 0 {
		w = 100
	}
	if h <= 0 {
		h = 30
	}
	chrome := previewChrome
	if m.pv.Stale {
		chrome++
	}
	textW := min(w-4, previewMaxWidth)
	off := m.pvView.YOffset()
	m.pvView.SetWidth(textW)
	m.pvView.SetHeight(max(3, h-chrome))
	m.pvView.SetContent(renderBlocks(m.pv.Blocks, textW))
	m.pvView.SetYOffset(off)
}

func (m Model) previewKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.screen = screenSearch
		m.status = ""
	case "up":
		m.pvView.ScrollUp(1)
	case "down":
		m.pvView.ScrollDown(1)
	case "pgup":
		m.pvView.PageUp()
	case "pgdown", "space":
		m.pvView.PageDown()
	case "home":
		m.pvView.GotoTop()
	case "end":
		m.pvView.GotoBottom()
	case "enter":
		m.runOn(m.act.Open, m.pvPath, "")
	case "ctrl+o":
		m.runOn(m.act.Reveal, m.pvPath, "")
	case "ctrl+y":
		m.runOn(m.act.Copy, m.pvPath, i18n.T(i18n.KeyPathCopied))
	case "ctrl+g":
		cmd := m.toggleLang()
		m.layoutPreview() // the illustration label is part of the content
		return m, cmd
	}
	return m, nil
}
