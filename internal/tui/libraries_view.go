package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

const (
	libHelpBase  = "a додати · u оновити · d прибрати · ↑/↓ вибір"
	libHelpEmpty = "a додати бібліотеку · esc вихід"
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
	b.WriteString(styleTitle.Render("Бібліотеки") + "  " + styleDim.Render(fmt.Sprintf("творів в індексі: %d", m.total)) + "\n\n")
	if len(m.libs) == 0 {
		b.WriteString(styleDim.Render("Бібліотек ще немає. Натисніть a, щоб вибрати теку з книгами.") + "\n")
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
			styleDim.Render(ansi.Truncate(fmt.Sprintf("  книг: %d · творів: %d · оновлено: %s", l.Books, l.Works, updated), max(10, w-8), "…")) + "\n")
	}
	if m.job != nil {
		p := m.job.progress
		b.WriteString("\n" + fmt.Sprintf("Індексую «%s»…", m.job.name) + "\n")
		if p.Total == 0 {
			b.WriteString(styleDim.Render("шукаю файли…") + "\n")
		} else {
			b.WriteString(m.bar.ViewAs(float64(p.Done)/float64(p.Total)) + fmt.Sprintf(" %d/%d", p.Done, p.Total) + "\n")
		}
		b.WriteString(styleDim.Render("esc — скасувати") + "\n")
	}
	if l := m.selectedLib(); m.confirmDelete && l != nil {
		b.WriteString("\n" + styleStatus.Render(fmt.Sprintf("Прибрати «%s» з індексу? Файли не чіпаються. (y/n)", l.Name)) + "\n")
	}
	if m.status != "" {
		b.WriteString("\n" + styleStatus.Render(m.status) + "\n")
	}
	help := libHelpEmpty
	if len(m.libs) > 0 {
		esc := "esc до пошуку"
		if m.total == 0 {
			esc = "esc вихід"
		}
		help = libHelpBase + " · " + esc
	}
	b.WriteString("\n" + styleDim.Render(ansi.Truncate(help, w, "…")))
	return b.String()
}
