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
	seq   int
	path  string
	hit   index.Hit
	doc   preview.Doc
	text  string // blocks already rendered off the UI goroutine, for width
	width int    // the text width text was rendered for
	err   error
}

// previewChrome counts the screen lines around the text: two header lines,
// two separators and the help line.
const previewChrome = 5

// previewTextWidth computes the wrapped text column width for a window of
// width w (0 meaning not yet known), floored so narrow windows still fit
// readable text. It is the one formula shared by the load command, which
// renders off the UI goroutine at ctrl+r time, and layoutPreview, so a
// render started before a resize stays comparable to the current width.
func previewTextWidth(w int) int {
	if w <= 0 {
		w = 100
	}
	return max(20, min(w-4, previewMaxWidth))
}

// openPreview starts loading the selected work; it sets a status and
// returns nil when the preview cannot be shown. The load command renders
// the text off the UI goroutine (renderBlocks on a large book is too slow
// to run inline in Update) at the width the window has right now.
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
	width := previewTextWidth(m.width)
	return func() tea.Msg {
		doc, err := load(p, hit)
		text := renderBlocks(doc.Blocks, width)
		return previewMsg{seq: seq, path: p, hit: hit, doc: doc, text: text, width: width, err: err}
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
	if previewTextWidth(m.width) == msg.width {
		// the window did not resize while loading: use the text rendered
		// off the UI goroutine as-is, no need to re-render it here.
		m.pvWidth = msg.width
		m.pvView.SetContent(msg.text)
	} else {
		m.pvWidth = -1 // force layoutPreview below to re-wrap for the current width
	}
	m.layoutPreview()
	return m
}

// layoutPreview sizes the viewport to the window and, only when the
// computed text width actually changed, re-wraps the text (a height-only
// resize just resizes the viewport, which is cheap). It keeps the scroll
// position either way.
func (m *Model) layoutPreview() {
	if m.screen != screenPreview {
		return
	}
	h := m.height
	if h <= 0 {
		h = 30
	}
	chrome := previewChrome
	if m.pv.Stale {
		chrome++
	}
	textW := previewTextWidth(m.width)
	off := m.pvView.YOffset()
	m.pvView.SetWidth(textW)
	m.pvView.SetHeight(max(3, h-chrome))
	if textW != m.pvWidth {
		m.pvWidth = textW
		m.pvView.SetContent(renderBlocks(m.pv.Blocks, textW))
	}
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
		// The illustration label is part of the content, so force a
		// re-wrap below even though the width did not change.
		m.pvWidth = -1
		m.layoutPreview()
		return m, cmd
	}
	return m, nil
}
