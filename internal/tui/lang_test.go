package tui

import (
	"errors"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"findbooks/internal/i18n"
)

func ctrlG(m tea.Model) (tea.Model, tea.Cmd) {
	return m.Update(tea.KeyPressMsg{Code: 'g', Mod: tea.ModCtrl})
}

func TestCtrlGTogglesLanguage(t *testing.T) {
	t.Cleanup(func() { i18n.Set(i18n.UK) })
	m, _, rec := newModel(true)
	if !strings.Contains(m.View().Content, "ctrl+g → EN") {
		t.Fatal("Ukrainian help must offer switching to EN")
	}
	tm, cmd := ctrlG(m)
	if i18n.Current() != i18n.EN {
		t.Fatalf("language = %q after ctrl+g", i18n.Current())
	}
	out := tm.View().Content
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
