package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"findbooks/internal/fb2"
	"findbooks/internal/i18n"
)

// previewMaxWidth caps the text column so long lines stay readable.
const previewMaxWidth = 100

const blockIndent = "    "

var styleItalic = lipgloss.NewStyle().Italic(true)

// wrap splits text into lines: the first at most first cells wide, the
// others at most rest; a word longer than its line is cut. curW tracks the
// current line's width incrementally (each word's width is measured once)
// instead of re-measuring cur on every word, which made this quadratic per
// line for long paragraphs.
func wrap(text string, first, rest int) []string {
	var lines []string
	limit, cur, curW := first, "", 0
	for _, w := range strings.Fields(text) {
		ww := ansi.StringWidth(w)
		if cur != "" && curW+1+ww > limit {
			lines = append(lines, cur)
			cur, curW, limit = "", 0, rest
		}
		for cur == "" && ww > limit {
			head := ansi.Truncate(w, limit, "")
			if head == "" { // a wide rune does not fit: take it anyway
				_, size := utf8.DecodeRuneInString(w)
				head = w[:size]
			}
			lines = append(lines, head)
			w, limit = w[len(head):], rest
			ww = ansi.StringWidth(w)
		}
		if w == "" {
			continue
		}
		if cur != "" {
			cur += " "
			curW++
		}
		cur += w
		curW += ww
	}
	if cur != "" || len(lines) == 0 {
		lines = append(lines, cur)
	}
	return lines
}

// renderBlocks lays the blocks out as lines at most width cells wide.
func renderBlocks(blocks []fb2.Block, width int) string {
	width = max(20, width)
	inner := width - len(blockIndent)
	var out []string
	prevTitle := false
	for _, b := range blocks {
		isTitle := b.Kind == fb2.BlockTitle
		if isTitle && !prevTitle && len(out) > 0 {
			out = append(out, "")
		}
		prevTitle = isTitle
		switch b.Kind {
		case fb2.BlockEmpty:
			out = append(out, "")
		case fb2.BlockImage:
			out = append(out, blockIndent+styleDim.Render(i18n.T(i18n.KeyIllustration)))
		case fb2.BlockTitle, fb2.BlockSubtitle:
			for _, l := range wrap(b.Text, width, width) {
				out = append(out, styleTitle.Render(l))
			}
		case fb2.BlockEpigraph, fb2.BlockCite:
			for _, l := range wrap(b.Text, inner, inner) {
				out = append(out, blockIndent+styleItalic.Render(l))
			}
		case fb2.BlockPoemLine:
			for _, l := range wrap(b.Text, inner, inner) {
				out = append(out, blockIndent+l)
			}
		default:
			ls := wrap(b.Text, inner, width)
			ls[0] = blockIndent + ls[0]
			out = append(out, ls...)
		}
	}
	return strings.Join(out, "\n")
}

func (m Model) renderPreview() string {
	w := m.width
	if w <= 0 {
		w = 100
	}
	h := m.pvHit
	head := m.pv.Title
	if m.pv.Author != "" {
		head += " — " + m.pv.Author
	}
	where := h.BookTitle
	if h.BookYear != "" {
		where += " (" + h.BookYear + ")"
	}
	if h.IsCollection {
		where += " › " + h.TreePath
	}
	volume := "  [" + h.VolumeName + " "
	sep := styleDim.Render(strings.Repeat("─", w)) + "\n"
	var b strings.Builder
	b.WriteString(styleTitle.Render(ansi.Truncate(head, w, "…")) + "\n")
	b.WriteString(styleDim.Render(ansi.Truncate(where, max(10, w-lipgloss.Width(volume)-2), "…")+volume) +
		m.mark(h.VolumeID) + styleDim.Render("]") + "\n")
	b.WriteString(sep)
	if m.pv.Stale {
		b.WriteString(styleStatus.Render(ansi.Truncate(i18n.T(i18n.KeyPreviewStale), w, "…")) + "\n")
	}
	b.WriteString(m.pvView.View() + "\n")
	b.WriteString(sep)
	if m.status != "" {
		b.WriteString(styleStatus.Render(ansi.Truncate(m.status, w, "…")))
	} else {
		help := fmt.Sprintf("%d%% · ", int(m.pvView.ScrollPercent()*100)) + i18n.T(i18n.KeyLangToggle) + " · " + i18n.T(i18n.KeyPreviewHelp)
		b.WriteString(styleDim.Render(ansi.Truncate(help, w, "…")))
	}
	return b.String()
}
