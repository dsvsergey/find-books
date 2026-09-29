package tui

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"findbooks/internal/i18n"
	"findbooks/internal/index"
	"findbooks/internal/scan"
)

func ctrlG(m tea.Model) (tea.Model, tea.Cmd) {
	return m.Update(tea.KeyPressMsg{Code: 'g', Mod: tea.ModCtrl})
}

func TestCtrlGTogglesLanguage(t *testing.T) {
	t.Cleanup(func() { i18n.Set(i18n.UK) })
	m, _, rec := newModel(true)
	if !strings.Contains(ansiStrip(m.View().Content), "ctrl+g → EN") {
		t.Fatal("Ukrainian help must offer switching to EN")
	}
	tm, cmd := ctrlG(m)
	if i18n.Current() != i18n.EN {
		t.Fatalf("language = %q after ctrl+g", i18n.Current())
	}
	out := ansiStrip(tm.View().Content)
	for _, want := range []string{"enter open", "ctrl+g → УКР", "work title, author or collection"} {
		if !strings.Contains(out, want) {
			t.Errorf("English view lacks %q:\n%s", want, out)
		}
	}
	if cmd == nil {
		t.Fatal("ctrl+g must save the language")
	}
	tm, _ = tm.Update(cmd())
	if !slices.Equal(rec.saved, []i18n.Lang{i18n.EN}) {
		t.Fatalf("saved = %v", rec.saved)
	}
	ctrlG(tm)
	if i18n.Current() != i18n.UK {
		t.Fatal("second ctrl+g must switch back")
	}
}

// TestCtrlGIgnoredWhileJobRunning pins that the libraries-screen guard
// blocking most keys while a job runs also covers ctrl+g: the language must
// not change and no save command must be returned.
func TestCtrlGIgnoredWhileJobRunning(t *testing.T) {
	t.Cleanup(func() { i18n.Set(i18n.UK) })
	fb := &fakeLibs{libs: []index.Library{fantastika}, total: 23251}
	fb.rescanFn = func(ctx context.Context, _ index.Library, on func(scan.Progress)) (scan.Report, error) {
		<-ctx.Done()
		return scan.Report{}, ctx.Err()
	}
	m := libModel(fb, 23251, "", nil)
	m, cmd := press(m, "ctrl+l")
	m = drain(t, m, cmd)
	m, _ = press(m, "u")
	realJob := m.(Model).job
	if realJob == nil {
		t.Fatal("job must be running")
	}
	t.Cleanup(realJob.cancel) // unblock the goroutine parked on <-ctx.Done()

	m, gCmd := ctrlG(m)
	if gCmd != nil {
		t.Fatal("ctrl+g must be ignored (no save command) while a job runs")
	}
	if i18n.Current() != i18n.UK {
		t.Fatal("ctrl+g must not change the language while a job runs")
	}
}

// TestLibrariesHelpStartsWithLangToggle pins that the toggle hint is first
// on the libraries screen's help line too (a long help line is truncated to
// the screen width, so a trailing hint would be cut off).
func TestLibrariesHelpStartsWithLangToggle(t *testing.T) {
	t.Cleanup(func() { i18n.Set(i18n.UK) })
	m := libModel(&fakeLibs{}, 0, "", nil)
	m = drain(t, m, m.Init())
	out := ansiStrip(m.View().Content)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	help := lines[len(lines)-1]
	if !strings.HasPrefix(help, "ctrl+g → EN") {
		t.Fatalf("help line = %q, want it to start with the language toggle", help)
	}
}

// TestCtrlGClearsStatus pins that toggling the language does not leave a
// status message written in the language that's no longer current.
func TestCtrlGClearsStatus(t *testing.T) {
	t.Cleanup(func() { i18n.Set(i18n.UK) })
	m, _, _ := newModel(true)
	m.status = "стара мова"
	tm, _ := ctrlG(m)
	if st := tm.(Model).status; st != "" {
		t.Fatalf("status = %q, want cleared by ctrl+g", st)
	}
}

func TestCtrlGOnLibrariesScreen(t *testing.T) {
	t.Cleanup(func() { i18n.Set(i18n.UK) })
	m := libModel(&fakeLibs{}, 0, "", nil)
	m = drain(t, m, m.Init())
	m, _ = ctrlG(m)
	if !strings.Contains(m.View().Content, "Libraries") || !strings.Contains(m.View().Content, "No libraries yet") {
		t.Fatalf("libraries screen not English:\n%s", m.View().Content)
	}
}

func TestCtrlGIgnoredWhileChoosing(t *testing.T) {
	t.Cleanup(func() { i18n.Set(i18n.UK) })
	m := libModel(&fakeLibs{}, 0, "/x", nil)
	m, _ = press(m, "a") // choosing
	ctrlG(m)
	if i18n.Current() != i18n.UK {
		t.Fatal("ctrl+g must be ignored while the folder dialog is open")
	}
}

func TestLangSaveFailureShowsStatus(t *testing.T) {
	t.Cleanup(func() { i18n.Set(i18n.UK) })
	fb := &fakeSearcher{}
	rec := &recorder{online: true}
	act := rec.actions()
	act.SaveLang = func(i18n.Lang) error { return errors.New("disk full") }
	m := tea.Model(New(fb, act, 5))
	m, cmd := ctrlG(m)
	m, _ = m.Update(cmd())
	if st := m.(Model).status; !strings.Contains(st, "Could not save the language") || !strings.Contains(st, "disk full") {
		t.Fatalf("status = %q", st)
	}
}

func TestDetailsLabelsEnglish(t *testing.T) {
	i18n.Set(i18n.EN)
	t.Cleanup(func() { i18n.Set(i18n.UK) })
	tm, _, _ := searched(t, true, "чужие дети")
	out := tm.(Model).View().Content
	for _, label := range []string{"Book:", "Path:", "File:", "Contents:"} {
		found := false
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, label) && len(strings.TrimSpace(ansiStrip(line))) > len(label) {
				found = true
			}
		}
		if !found {
			t.Errorf("label %q is not followed by its value on the same line:\n%s", label, out)
		}
	}
}

func ansiStrip(s string) string { return ansi.Strip(s) }
