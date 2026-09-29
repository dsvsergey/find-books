package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"findbooks/internal/i18n"
)

func (m Model) libMark(volumeID string) string {
	if m.libOnline[volumeID] {
		return styleOnline.Render("●")
	}
	return styleOffline.Render("○")
}

func (m Model) renderLibraries() string {
	w := m.width
	if w <= 0 {
		w = 100
	}
	var b strings.Builder
	b.WriteString(styleTitle.Render(i18n.T(i18n.KeyLibTitle)) + "  " + styleDim.Render(i18n.T(i18n.KeyWorksInIndex, m.total)) + "\n\n")
	if len(m.libs) == 0 {
		b.WriteString(styleDim.Render(i18n.T(i18n.KeyNoLibrariesYet)) + "\n")
	}
	for i, l := range m.libs {
		marker, style := "  ", styleTitle
		if i == m.libCursor {
			marker, style = "▸ ", styleSelected
		}
		updated := "—"
		if !l.LastScan.IsZero() {
			updated = l.LastScan.Format("2006-01-02 15:04")
		}
		b.WriteString(marker + style.Render(ansi.Truncate(l.Name, w-2, "…")) + "\n")
		b.WriteString(styleDim.Render("    "+l.VolumeName+" ") + m.libMark(l.VolumeID) +
			styleDim.Render(ansi.Truncate(i18n.T(i18n.KeyLibStats, l.Books, l.Works, updated), max(10, w-8), "…")) + "\n")
	}
	if m.job != nil {
		p := m.job.progress
		b.WriteString("\n" + i18n.T(i18n.KeyIndexingJob, m.job.name) + "\n")
		if p.Total == 0 {
			b.WriteString(styleDim.Render(i18n.T(i18n.KeyFindingFiles)) + "\n")
		} else {
			b.WriteString(m.bar.ViewAs(float64(p.Done)/float64(p.Total)) + fmt.Sprintf(" %d/%d", p.Done, p.Total) + "\n")
		}
		b.WriteString(styleDim.Render(i18n.T(i18n.KeyEscCancel)) + "\n")
	}
	if l := m.selectedLib(); m.confirmDelete && l != nil {
		b.WriteString("\n" + styleStatus.Render(i18n.T(i18n.KeyConfirmRemove, l.Name)) + "\n")
	}
	if m.status != "" {
		b.WriteString("\n" + styleStatus.Render(m.status) + "\n")
	}
	help := i18n.T(i18n.KeyLibHelpEmpty)
	if len(m.libs) > 0 {
		esc := i18n.T(i18n.KeyEscToSearch)
		if m.total == 0 {
			esc = i18n.T(i18n.KeyEscQuit)
		}
		help = i18n.T(i18n.KeyLibHelpBase) + " · " + esc
	}
	help = i18n.T(i18n.KeyLangToggle) + " · " + help
	b.WriteString("\n" + styleDim.Render(ansi.Truncate(help, w, "…")))
	return b.String()
}
