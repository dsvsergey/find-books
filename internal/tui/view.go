package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"findbooks/internal/i18n"
	"findbooks/internal/index"
)

var (
	styleSelected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	styleTitle    = lipgloss.NewStyle().Bold(true)
	styleDim      = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	styleOnline   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	styleOffline  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleLabel    = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Width(10)
	styleStatus   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	styleError    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m Model) render() string {
	if m.screen == screenLibraries {
		return m.renderLibraries()
	}
	w := m.width
	if w <= 0 {
		w = 100
	}
	var b strings.Builder
	b.WriteString(m.input.View() + "  " + styleDim.Render(i18n.T(i18n.KeyHitsCount, len(m.hits), m.total)) + "\n\n")
	if m.err != nil {
		b.WriteString(styleError.Render(i18n.T(i18n.KeySearchError, m.err.Error())) + "\n")
	}
	if m.searched && len(m.hits) == 0 && strings.TrimSpace(m.input.Value()) != "" {
		b.WriteString(styleDim.Render(i18n.T(i18n.KeyNothingFound)) + "\n")
	}
	first, last := m.window()
	for i := first; i < last; i++ {
		b.WriteString(m.renderHit(i, w))
	}
	if h := m.selected(); h != nil {
		b.WriteString(styleDim.Render(strings.Repeat("─", w)) + "\n")
		b.WriteString(m.renderDetails(*h, w))
	}
	if m.status != "" {
		b.WriteString("\n" + styleStatus.Render(m.status) + "\n")
	}
	help := i18n.T(i18n.KeyLangToggle) + " · " + i18n.T(i18n.KeySearchHelp)
	b.WriteString("\n" + styleDim.Render(ansi.Truncate(help, w, "…")))
	return b.String()
}

// window returns the slice of hits that fits the screen around the cursor.
func (m Model) window() (int, int) {
	visible := 10
	if m.height > 0 {
		visible = max(1, (m.height-12)/2)
	}
	first := 0
	if m.cursor >= visible {
		first = m.cursor - visible + 1
	}
	return first, min(len(m.hits), first+visible)
}

func (m Model) mark(volumeID string) string {
	if m.online[volumeID] {
		return styleOnline.Render("●")
	}
	return styleOffline.Render("○")
}

func (m Model) renderHit(i, w int) string {
	h := m.hits[i]
	line := h.Title
	if h.Author != "" {
		line += " — " + h.Author
	}
	marker, style := "  ", styleTitle
	if i == m.cursor {
		marker, style = "▸ ", styleSelected
	}
	where := i18n.T(i18n.KeyWhereFile, h.RelPath)
	if h.IsCollection {
		where = i18n.T(i18n.KeyWhereIn, h.BookTitle)
	}
	volume := "[" + h.VolumeName + " "
	room := max(10, w-lipgloss.Width(volume)-4)
	return marker + style.Render(ansi.Truncate(line, w-2, "…")) + "\n" +
		styleDim.Render(ansi.Truncate(where, room, "…")+"  "+volume) + m.mark(h.VolumeID) + styleDim.Render("]") + "\n"
}

func (m Model) renderDetails(h index.Hit, w int) string {
	book := h.BookTitle
	if h.BookYear != "" {
		book += " (" + h.BookYear + ")"
	}
	rows := [][2]string{
		{i18n.T(i18n.KeyLabelBook), book},
		{i18n.T(i18n.KeyLabelPath), h.TreePath},
		{i18n.T(i18n.KeyLabelFile), h.LibraryName + ": " + h.RelPath},
	}
	if toc := m.toc[h.BookID]; len(toc) > 1 {
		rows = append(rows, [2]string{i18n.T(i18n.KeyLabelContents), strings.Join(toc, " · ")})
	}
	var b strings.Builder
	for _, r := range rows {
		b.WriteString(styleLabel.Render(r[0]) + ansi.Truncate(r[1], max(10, w-11), "…") + "\n")
	}
	return b.String()
}
