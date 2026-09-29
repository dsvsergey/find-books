package tui

import (
	"context"
	"errors"
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

// --- fix round 1: guard against concurrent jobs ---

func TestChoosingBlocksOtherKeys(t *testing.T) {
	fb := &fakeLibs{libs: []index.Library{fantastika}, total: 23251}
	rescanCalled := false
	fb.rescanFn = func(context.Context, index.Library, func(scan.Progress)) (scan.Report, error) {
		rescanCalled = true
		return scan.Report{}, nil
	}
	m := libModel(fb, 23251, "/Volumes/dsvDev/Книги", nil)
	m, cmd := press(m, "ctrl+l")
	m = drain(t, m, cmd)

	m, chooseCmd := press(m, "a")
	if chooseCmd == nil {
		t.Fatal("a must start the folder chooser")
	}
	if !m.(Model).choosing {
		t.Fatal("choosing must be set while the dialog is pending")
	}

	m, uCmd := press(m, "u")
	if uCmd != nil || rescanCalled {
		t.Fatal("keys other than ctrl+c must be ignored while the folder dialog is pending")
	}

	m, escCmd := press(m, "esc")
	if escCmd != nil || m.(Model).screen != screenLibraries {
		t.Fatal("esc must be ignored (not leave the screen) while the folder dialog is pending")
	}

	if _, ccCmd := press(m, "ctrl+c"); !isQuit(ccCmd) {
		t.Fatal("ctrl+c must still quit while the folder dialog is pending")
	}
}

func TestFolderMsgIgnoredWhileJobRunning(t *testing.T) {
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

	m, _ = m.Update(folderMsg{path: "/Volumes/dsvDev/Інша"})
	mm := m.(Model)
	if mm.job != realJob {
		t.Fatal("a folderMsg must not replace a job that is already running")
	}
	if len(fb.added) != 0 {
		t.Fatalf("folderMsg must not start AddFolder while a job is running: added = %q", fb.added)
	}
}

func TestStaleJobDoneMsgIgnored(t *testing.T) {
	fb := &fakeLibs{libs: []index.Library{fantastika}, total: 23251}
	fb.rescanFn = func(ctx context.Context, _ index.Library, on func(scan.Progress)) (scan.Report, error) {
		<-ctx.Done()
		return scan.Report{}, ctx.Err()
	}
	m := libModel(fb, 23251, "", nil)
	m, cmd := press(m, "ctrl+l")
	m = drain(t, m, cmd)
	m, _ = press(m, "u")
	mm := m.(Model)
	realJob := mm.job
	if realJob == nil {
		t.Fatal("job must be running")
	}
	before := mm.status

	// A jobDoneMsg tagged with a different *job simulates a message that
	// crossed over from an earlier, already-superseded job.
	other := &job{name: "stale"}
	m, _ = m.Update(jobDoneMsg{j: other, name: "stale", err: errors.New("boom")})
	mm = m.(Model)
	if mm.status != before {
		t.Fatalf("status changed from a stale jobDoneMsg: %q", mm.status)
	}
	if mm.job != realJob {
		t.Fatal("the current job must be kept when a stale jobDoneMsg arrives")
	}
}

// --- fix round 1: quitting mid-job waits for the scan to stop ---

func TestCtrlCDuringJobWaitsForStop(t *testing.T) {
	fb := &fakeLibs{}
	fb.addFn = func(ctx context.Context, _ string, on func(scan.Progress)) (index.Library, scan.Report, error) {
		on(scan.Progress{Done: 1, Total: 10})
		<-ctx.Done()
		return index.Library{ID: 7, Name: "Книги"}, scan.Report{}, ctx.Err()
	}
	m := libModel(fb, 0, "/Volumes/dsvDev/Книги", nil)
	m, cmd := press(m, "a")
	m, cmd = m.Update(cmd()) // folderMsg → job started, cmd listens
	m, cmd = m.Update(cmd()) // progressMsg, cmd listens again

	m, ccCmd := press(m, "ctrl+c")
	if isQuit(ccCmd) {
		t.Fatal("ctrl+c during a job must not quit immediately")
	}
	if !m.(Model).quitting {
		t.Fatal("quitting must be set while the job stops")
	}
	if m.(Model).job == nil {
		t.Fatal("the job must stay tracked until it actually finishes")
	}

	_, doneCmd := m.Update(cmd()) // jobDoneMsg, delivered once the canceled job returns
	if !isQuit(doneCmd) {
		t.Fatal("the pending jobDoneMsg must quit once ctrl+c requested it")
	}
}

func TestCtrlCWithoutJobQuitsImmediately(t *testing.T) {
	fb := &fakeLibs{libs: []index.Library{fantastika}, total: 23251}
	m := libModel(fb, 23251, "", nil)
	m, cmd := press(m, "ctrl+l")
	m = drain(t, m, cmd)
	if _, cmd := press(m, "ctrl+c"); !isQuit(cmd) {
		t.Fatal("ctrl+c without a running job must quit immediately")
	}
}

// --- fix round 1: cheap fixes ---

func TestUkrainianGActsLikeU(t *testing.T) {
	fb := &fakeLibs{libs: []index.Library{fantastika}, total: 23251}
	called := false
	fb.rescanFn = func(context.Context, index.Library, func(scan.Progress)) (scan.Report, error) {
		called = true
		return scan.Report{}, nil
	}
	m := libModel(fb, 23251, "", nil)
	m, cmd := press(m, "ctrl+l")
	m = drain(t, m, cmd)
	m, cmd = press(m, "г")
	m = drain(t, m, cmd)
	if !called {
		t.Fatal("г must act like u (rescan)")
	}
}

func TestAddFailsAfterRegisterShowsResumeHint(t *testing.T) {
	fb := &fakeLibs{}
	fb.addFn = func(context.Context, string, func(scan.Progress)) (index.Library, scan.Report, error) {
		return index.Library{ID: 9, Name: "Книги"}, scan.Report{}, errors.New("диск зник")
	}
	m := libModel(fb, 0, "/Volumes/dsvDev/Книги", nil)
	m, cmd := press(m, "a")
	m = drain(t, m, cmd)
	if !strings.Contains(m.(Model).status, "натисніть u") {
		t.Fatalf("status = %q", m.(Model).status)
	}
}

func TestAddCanceledBeforeRegistering(t *testing.T) {
	fb := &fakeLibs{}
	fb.addFn = func(ctx context.Context, _ string, on func(scan.Progress)) (index.Library, scan.Report, error) {
		<-ctx.Done()
		return index.Library{}, scan.Report{}, ctx.Err()
	}
	m := libModel(fb, 0, "/Volumes/dsvDev/Книги", nil)
	m, cmd := press(m, "a")
	m, cmd = m.Update(cmd()) // folderMsg → job started, cmd listens
	m, _ = press(m, "esc")   // cancels before AddFolder returns a library
	m, _ = m.Update(cmd())   // jobDoneMsg after cancel
	if !strings.Contains(m.(Model).status, "Додавання «Книги» перервано") {
		t.Fatalf("status = %q", m.(Model).status)
	}
}
