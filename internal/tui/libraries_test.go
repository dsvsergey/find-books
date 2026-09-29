package tui

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"findbooks/internal/index"
	"findbooks/internal/libman"
	"findbooks/internal/platform"
	"findbooks/internal/scan"
)

type fakeLibs struct {
	fakeSearcher
	libs     []index.Library
	total    int
	removed  []string
	added    []string
	addFn    func(ctx context.Context, dir string, on func(scan.Progress)) (index.Library, scan.Report, error)
	rescanFn func(ctx context.Context, lib index.Library, on func(scan.Progress)) (scan.Report, error)
}

func (f *fakeLibs) Libraries() ([]index.Library, error) { return slices.Clone(f.libs), nil }
func (f *fakeLibs) TotalWorks() (int, error)            { return f.total, nil }
func (f *fakeLibs) RemoveLibrary(name string) error {
	f.removed = append(f.removed, name)
	f.libs = slices.DeleteFunc(f.libs, func(l index.Library) bool { return l.Name == name })
	if len(f.libs) == 0 {
		f.total = 0
	}
	return nil
}
func (f *fakeLibs) AddFolder(ctx context.Context, dir string, on func(scan.Progress)) (index.Library, scan.Report, error) {
	f.added = append(f.added, dir)
	return f.addFn(ctx, dir, on)
}
func (f *fakeLibs) Rescan(ctx context.Context, lib index.Library, on func(scan.Progress)) (scan.Report, error) {
	return f.rescanFn(ctx, lib, on)
}

var fantastika = index.Library{
	ID: 1, Name: "Фантастика", VolumeID: "dsvDev", VolumeName: "dsvDev", RootRel: "Бібліотека",
	LastScan: time.Date(2026, 9, 29, 12, 0, 0, 0, time.Local), Books: 5547, Works: 23251,
}

// libModel builds a model on fb whose folder chooser returns chosen/chooseErr.
func libModel(fb *fakeLibs, total int, chosen string, chooseErr error) tea.Model {
	rec := &recorder{online: true}
	act := rec.actions()
	act.ChooseFolder = func(string) (string, error) { return chosen, chooseErr }
	return New(fb, act, total)
}

// drain runs commands and feeds their messages back until nothing is left.
func drain(t *testing.T, m tea.Model, cmd tea.Cmd) tea.Model {
	t.Helper()
	for i := 0; cmd != nil; i++ {
		if i > 100 {
			t.Fatal("command loop did not settle")
		}
		msg := cmd()
		if msg == nil {
			break
		}
		m, cmd = m.Update(msg)
	}
	return m
}

func press(m tea.Model, s string) (tea.Model, tea.Cmd) {
	switch s {
	case "esc":
		return m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	case "up":
		return m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	case "down":
		return m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	case "ctrl+l":
		return m.Update(tea.KeyPressMsg{Code: 'l', Mod: tea.ModCtrl})
	case "ctrl+c":
		return m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	}
	r := []rune(s)[0]
	return m.Update(tea.KeyPressMsg{Code: r, Text: s})
}

func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestEmptyIndexStartsOnLibraries(t *testing.T) {
	m := libModel(&fakeLibs{}, 0, "", nil)
	if m.(Model).screen != screenLibraries {
		t.Fatal("empty index must start on the libraries screen")
	}
	m = drain(t, m, m.Init())
	if !strings.Contains(m.View().Content, "Бібліотек ще немає") {
		t.Fatalf("view:\n%s", m.View().Content)
	}
	if _, cmd := press(m, "esc"); !isQuit(cmd) {
		t.Fatal("esc on an empty index must quit")
	}
}

func TestCtrlLSwitchesScreens(t *testing.T) {
	fb := &fakeLibs{libs: []index.Library{fantastika}, total: 23251}
	m := libModel(fb, 23251, "", nil)
	if m.(Model).screen != screenSearch {
		t.Fatal("non-empty index must start on search")
	}
	m, cmd := press(m, "ctrl+l")
	m = drain(t, m, cmd)
	out := m.View().Content
	for _, want := range []string{"Фантастика", "●", "5547", "23251", "2026-09-29 12:00"} {
		if !strings.Contains(out, want) {
			t.Errorf("libraries view lacks %q:\n%s", want, out)
		}
	}
	m, _ = press(m, "esc")
	if m.(Model).screen != screenSearch {
		t.Fatal("esc must return to search")
	}
}

func TestAddLibraryFlow(t *testing.T) {
	fb := &fakeLibs{}
	fb.addFn = func(_ context.Context, dir string, on func(scan.Progress)) (index.Library, scan.Report, error) {
		on(scan.Progress{Done: 1, Total: 2})
		on(scan.Progress{Done: 2, Total: 2})
		lib := index.Library{ID: 2, Name: "Книги", VolumeID: "dsvDev", VolumeName: "dsvDev", Books: 2, Works: 5}
		fb.libs = append(fb.libs, lib)
		fb.total = 5
		return lib, scan.Report{Added: 2}, nil
	}
	m := libModel(fb, 0, "/Volumes/dsvDev/Книги", nil)
	m, cmd := press(m, "a")
	m = drain(t, m, cmd)
	mm := m.(Model)
	if !slices.Equal(fb.added, []string{"/Volumes/dsvDev/Книги"}) {
		t.Fatalf("added = %q", fb.added)
	}
	if mm.job != nil || mm.total != 5 || len(mm.libs) != 1 {
		t.Fatalf("job = %v, total = %d, libs = %+v", mm.job, mm.total, mm.libs)
	}
	if !strings.Contains(mm.status, "додано 2") {
		t.Fatalf("status = %q", mm.status)
	}
}

func TestUkrainianLayoutKeys(t *testing.T) {
	fb := &fakeLibs{libs: []index.Library{fantastika}, total: 1}
	m := libModel(fb, 0, "", platform.ErrCanceled)
	m = drain(t, m, m.Init())
	if _, cmd := press(m, "ф"); cmd == nil {
		t.Error("ф must act like a (open the folder chooser)")
	}
	m, _ = press(m, "в")
	if !m.(Model).confirmDelete {
		t.Fatal("в must act like d")
	}
	m, cmd := press(m, "н")
	m = drain(t, m, cmd)
	if !slices.Equal(fb.removed, []string{"Фантастика"}) {
		t.Fatalf("н must confirm like y; removed = %q", fb.removed)
	}
}

func TestChooseFolderCanceled(t *testing.T) {
	fb := &fakeLibs{}
	m := libModel(fb, 0, "", platform.ErrCanceled)
	m, cmd := press(m, "a")
	m = drain(t, m, cmd)
	if m.(Model).status != "" || len(fb.added) != 0 {
		t.Fatalf("status = %q, added = %q", m.(Model).status, fb.added)
	}
}

func TestChooseFolderUnsupported(t *testing.T) {
	m := libModel(&fakeLibs{}, 0, "", platform.ErrUnsupported)
	m, cmd := press(m, "a")
	m = drain(t, m, cmd)
	if !strings.Contains(m.(Model).status, "findbooks add") {
		t.Fatalf("status = %q", m.(Model).status)
	}
}

func TestEscCancelsIndexing(t *testing.T) {
	fb := &fakeLibs{}
	fb.addFn = func(ctx context.Context, _ string, on func(scan.Progress)) (index.Library, scan.Report, error) {
		on(scan.Progress{Done: 1, Total: 10})
		<-ctx.Done()
		return index.Library{ID: 7, Name: "Книги"}, scan.Report{}, ctx.Err()
	}
	m := libModel(fb, 0, "/Volumes/dsvDev/Книги", nil)
	m, cmd := press(m, "a")
	m, cmd = m.Update(cmd()) // folderMsg → job started, cmd listens
	m, cmd = m.Update(cmd()) // progressMsg
	if !strings.Contains(m.View().Content, "1/10") {
		t.Fatalf("progress not shown:\n%s", m.View().Content)
	}
	if _, other := press(m, "a"); other != nil {
		t.Fatal("keys other than esc must be ignored while indexing")
	}
	m, _ = press(m, "esc")
	m, _ = m.Update(cmd()) // jobDoneMsg after cancel
	mm := m.(Model)
	if mm.job != nil || !strings.Contains(mm.status, "натисніть u") {
		t.Fatalf("job = %v, status = %q", mm.job, mm.status)
	}
}

func TestAddExistingLibrary(t *testing.T) {
	fb := &fakeLibs{}
	fb.addFn = func(context.Context, string, func(scan.Progress)) (index.Library, scan.Report, error) {
		return index.Library{}, scan.Report{}, fmt.Errorf("%w: Книги", index.ErrExists)
	}
	m := libModel(fb, 0, "/Volumes/dsvDev/Книги", nil)
	m, cmd := press(m, "a")
	m = drain(t, m, cmd)
	if !strings.Contains(m.(Model).status, "Бібліотека «Книги» уже є") {
		t.Fatalf("status = %q", m.(Model).status)
	}
}

func TestDeleteConfirm(t *testing.T) {
	fb := &fakeLibs{libs: []index.Library{fantastika}, total: 23251}
	m := libModel(fb, 23251, "", nil)
	m, cmd := press(m, "ctrl+l")
	m = drain(t, m, cmd)
	m, _ = press(m, "d")
	if !strings.Contains(m.View().Content, "(y/n)") {
		t.Fatal("delete must ask for confirmation")
	}
	m, _ = press(m, "n")
	if len(fb.removed) != 0 || m.(Model).status != "Скасовано" {
		t.Fatalf("removed = %q, status = %q", fb.removed, m.(Model).status)
	}
	m, _ = press(m, "d")
	m, cmd = press(m, "y")
	m = drain(t, m, cmd)
	if !slices.Equal(fb.removed, []string{"Фантастика"}) || !strings.Contains(m.(Model).status, "прибрано") {
		t.Fatalf("removed = %q, status = %q", fb.removed, m.(Model).status)
	}
}

func TestDeleteLastLibraryThenEscQuits(t *testing.T) {
	fb := &fakeLibs{libs: []index.Library{fantastika}, total: 23251}
	m := libModel(fb, 23251, "", nil)
	m, cmd := press(m, "ctrl+l")
	m = drain(t, m, cmd)
	m, _ = press(m, "d")
	m, cmd = press(m, "y")
	m = drain(t, m, cmd)
	if m.(Model).total != 0 {
		t.Fatalf("total = %d after removing the last library", m.(Model).total)
	}
	if _, cmd := press(m, "esc"); !isQuit(cmd) {
		t.Fatal("esc with an empty index must quit")
	}
}

func TestUpdateOffline(t *testing.T) {
	fb := &fakeLibs{libs: []index.Library{fantastika}, total: 23251}
	fb.rescanFn = func(context.Context, index.Library, func(scan.Progress)) (scan.Report, error) {
		return scan.Report{}, fmt.Errorf("%w: «dsvDev»", libman.ErrOffline)
	}
	m := libModel(fb, 23251, "", nil)
	m, cmd := press(m, "ctrl+l")
	m = drain(t, m, cmd)
	m, cmd = press(m, "u")
	m = drain(t, m, cmd)
	if !strings.Contains(m.(Model).status, "Диск «dsvDev» не підключено") {
		t.Fatalf("status = %q", m.(Model).status)
	}
}

func TestUpdateReportsCounts(t *testing.T) {
	fb := &fakeLibs{libs: []index.Library{fantastika}, total: 23251}
	fb.rescanFn = func(_ context.Context, _ index.Library, on func(scan.Progress)) (scan.Report, error) {
		on(scan.Progress{Done: 1, Total: 1})
		return scan.Report{Updated: 1, Unchanged: 5546, Errors: []scan.FileError{{RelPath: "x.fb2"}}}, nil
	}
	m := libModel(fb, 23251, "", nil)
	m, cmd := press(m, "ctrl+l")
	m = drain(t, m, cmd)
	m, cmd = press(m, "u")
	m = drain(t, m, cmd)
	st := m.(Model).status
	if !strings.Contains(st, "оновлено 1") || !strings.Contains(st, "без змін 5546") || !strings.Contains(st, "проблемних файлів: 1") {
		t.Fatalf("status = %q", st)
	}
}
