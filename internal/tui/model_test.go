package tui

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"findbooks/internal/index"
)

type fakeSearcher struct {
	hits    []index.Hit
	toc     map[int64][]string
	queries []string
}

func (f *fakeSearcher) Search(q index.Query) ([]index.Hit, error) {
	f.queries = append(f.queries, q.Text)
	return f.hits, nil
}

func (f *fakeSearcher) BookWorks(id int64) ([]string, error) { return f.toc[id], nil }

type recorder struct {
	online                   bool
	missing                  bool // files do not exist on the (fake) disk
	opened, revealed, copied []string
}

func (r *recorder) actions() Actions {
	return Actions{
		Root: func(vol, rootRel string) (string, bool) {
			if !r.online {
				return "", false
			}
			return filepath.Join("/Volumes", vol, rootRel), true
		},
		Open:   func(p string) error { r.opened = append(r.opened, p); return nil },
		Reveal: func(p string) error { r.revealed = append(r.revealed, p); return nil },
		Copy:   func(p string) error { r.copied = append(r.copied, p); return nil },
		Exists: func(string) bool { return !r.missing },
	}
}

var (
	chuzhie = index.Hit{
		WorkID: 1, Title: "ЧУЖИЕ ДЕТИ", Author: "Игорь Пидоренко", TreePath: "Игорь Пидоренко › ЧУЖИЕ ДЕТИ",
		BookID: 10, BookTitle: "Румбы фантастики. 1988 год. Том II", BookYear: "1988", IsCollection: true,
		RelPath: "Сборники/Румбы.fb2", LibraryName: "Фантастика", VolumeID: "dsvDev", VolumeName: "dsvDev", RootRel: "Бібліотека",
	}
	planet = index.Hit{
		WorkID: 2, Title: "Дети чужой планеты", Author: "Иван Петров", TreePath: "Дети чужой планеты",
		BookID: 11, BookTitle: "Дети чужой планеты", RelPath: "Петров/Дети.fb2",
		LibraryName: "Фантастика", VolumeID: "dsvDev", VolumeName: "dsvDev", RootRel: "Бібліотека",
	}
)

func newModel(online bool) (Model, *fakeSearcher, *recorder) {
	src := &fakeSearcher{
		hits: []index.Hit{planet, chuzhie}, // deliberately not in best order
		toc:  map[int64][]string{10: {"ЗЕМЛЕЙ РОЖДЕННЫЕ", "ЧУЖИЕ ДЕТИ"}},
	}
	rec := &recorder{online: online}
	return New(src, rec.actions(), 1234), src, rec
}

func typeText(m tea.Model, s string) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	for _, r := range s {
		m, cmd = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m, cmd
}

func key(m tea.Model, k tea.KeyPressMsg) (tea.Model, tea.Cmd) { return m.Update(k) }

// settle fires the debounced search for the current sequence synchronously.
func settle(t *testing.T, m tea.Model) tea.Model {
	t.Helper()
	m, cmd := m.Update(searchMsg{seq: m.(Model).seq})
	if cmd == nil {
		t.Fatal("no search command for current sequence")
	}
	m, _ = m.Update(cmd())
	return m
}

func searched(t *testing.T, online bool, q string) (tea.Model, *fakeSearcher, *recorder) {
	t.Helper()
	m, src, rec := newModel(online)
	tm, _ := typeText(m, q)
	return settle(t, tm), src, rec
}

func TestTypingDebouncesSearch(t *testing.T) {
	m, src, _ := newModel(true)
	tm, cmd := typeText(m, "чуж")
	if cmd == nil || tm.(Model).seq != 3 {
		t.Fatalf("seq = %d, cmd = %v", tm.(Model).seq, cmd)
	}
	if _, stale := tm.Update(searchMsg{seq: 1}); stale != nil {
		t.Fatal("stale debounce tick must not search")
	}
	if len(src.queries) != 0 {
		t.Fatalf("searched before debounce: %q", src.queries)
	}
	tm = settle(t, tm)
	if !slices.Equal(src.queries, []string{"чуж"}) || len(tm.(Model).hits) != 2 {
		t.Fatalf("queries = %q, hits = %d", src.queries, len(tm.(Model).hits))
	}
}

func TestStaleResultsIgnored(t *testing.T) {
	m, _, _ := newModel(true)
	tm, _ := typeText(m, "чуж")
	tm, _ = tm.Update(resultsMsg{seq: 2, hits: []index.Hit{chuzhie}})
	if len(tm.(Model).hits) != 0 {
		t.Fatal("results of an older query were applied")
	}
}

// TestResultsMsgUsesPrecomputedOnlineMap pins that Update never calls
// Actions.Root synchronously: the online map must arrive ready-made on
// resultsMsg (computed inside the async search command), because Root can
// trigger a slow diskutil rescan on a cache miss and must never block the
// Bubble Tea event loop.
func TestResultsMsgUsesPrecomputedOnlineMap(t *testing.T) {
	var rootCalls int
	act := Actions{
		Root: func(vol, rootRel string) (string, bool) {
			rootCalls++
			return filepath.Join("/Volumes", vol, rootRel), true
		},
		Open:   func(string) error { return nil },
		Reveal: func(string) error { return nil },
		Copy:   func(string) error { return nil },
	}
	src := &fakeSearcher{hits: []index.Hit{chuzhie}}
	m := New(src, act, 1234)
	tm, _ := typeText(m, "чуж")
	before := rootCalls
	tm, _ = tm.Update(resultsMsg{
		seq:    tm.(Model).seq,
		hits:   []index.Hit{chuzhie},
		online: map[string]bool{"dsvDev": true},
	})
	if rootCalls != before {
		t.Fatalf("Root called %d time(s) during Update; must be computed by the search command, not Update", rootCalls-before)
	}
	if !strings.Contains(tm.(Model).View().Content, "●") {
		t.Fatal("online marker missing from view")
	}
}

func TestRankPutsClosestFirst(t *testing.T) {
	tm, _, _ := searched(t, true, "чужие дети")
	if got := tm.(Model).hits[0].Title; got != "ЧУЖИЕ ДЕТИ" {
		t.Fatalf("first hit = %q", got)
	}
}

func TestNavigationClamps(t *testing.T) {
	tm, _, _ := searched(t, true, "дети")
	tm, _ = key(tm, tea.KeyPressMsg{Code: tea.KeyUp})
	if c := tm.(Model).cursor; c != 0 {
		t.Fatalf("cursor = %d after up at top", c)
	}
	for range 3 {
		tm, _ = key(tm, tea.KeyPressMsg{Code: tea.KeyDown})
	}
	if c := tm.(Model).cursor; c != 1 {
		t.Fatalf("cursor = %d, want 1 (last)", c)
	}
}

func TestEnterOpensFileWhenOnline(t *testing.T) {
	tm, _, rec := searched(t, true, "чужие дети")
	key(tm, tea.KeyPressMsg{Code: tea.KeyEnter})
	want := filepath.Join("/Volumes", "dsvDev", "Бібліотека", "Сборники", "Румбы.fb2")
	if !slices.Equal(rec.opened, []string{want}) {
		t.Fatalf("opened = %q, want %q", rec.opened, want)
	}
}

func TestEnterOfflineShowsHint(t *testing.T) {
	tm, _, rec := searched(t, false, "чужие дети")
	tm, _ = key(tm, tea.KeyPressMsg{Code: tea.KeyEnter})
	if len(rec.opened) != 0 {
		t.Fatalf("opened offline file: %q", rec.opened)
	}
	status := tm.(Model).status
	if !strings.Contains(status, "dsvDev") || !strings.Contains(status, "не підключено") {
		t.Fatalf("status = %q", status)
	}
	if !strings.Contains(tm.(Model).View().Content, "○") {
		t.Fatal("offline marker missing from view")
	}
}

func TestRevealAndCopy(t *testing.T) {
	tm, _, rec := searched(t, true, "чужие дети")
	tm, _ = key(tm, tea.KeyPressMsg{Code: 'o', Mod: tea.ModCtrl})
	tm, _ = key(tm, tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})
	if len(rec.revealed) != 1 || len(rec.copied) != 1 || rec.copied[0] != rec.revealed[0] {
		t.Fatalf("revealed = %q, copied = %q", rec.revealed, rec.copied)
	}
	if tm.(Model).status != "Шлях скопійовано" {
		t.Fatalf("status = %q", tm.(Model).status)
	}
}

func TestTabSearchesByAuthor(t *testing.T) {
	tm, _, _ := searched(t, true, "чужие дети")
	before := tm.(Model).seq
	tm, cmd := key(tm, tea.KeyPressMsg{Code: tea.KeyTab})
	if got := tm.(Model).input.Value(); got != "Игорь Пидоренко" {
		t.Fatalf("input = %q", got)
	}
	if cmd == nil || tm.(Model).seq != before+1 {
		t.Fatal("tab must queue a new search")
	}
}

func TestEscQuits(t *testing.T) {
	m, _, _ := newModel(true)
	_, cmd := key(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("no command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("esc must quit")
	}
}

func TestViewShowsDetails(t *testing.T) {
	tm, _, _ := searched(t, true, "чужие дети")
	out := tm.(Model).View().Content
	for _, want := range []string{"ЧУЖИЕ ДЕТИ", "Румбы фантастики. 1988 год. Том II (1988)",
		"Игорь Пидоренко › ЧУЖИЕ ДЕТИ", "Зміст:", "ЗЕМЛЕЙ РОЖДЕННЫЕ", "2 з 1234", "●"} {
		if !strings.Contains(out, want) {
			t.Errorf("view lacks %q", want)
		}
	}
	if !tm.(Model).View().AltScreen {
		t.Error("view must use the alt screen")
	}
}

func TestMissingFileShowsStatus(t *testing.T) {
	m, _, rec := newModel(true)
	rec.missing = true
	m = New(m.src, rec.actions(), 1234)
	tm, _ := typeText(m, "чужие дети")
	tm = settle(t, tm)
	for _, k := range []tea.KeyPressMsg{{Code: tea.KeyEnter}, {Code: 'o', Mod: tea.ModCtrl}, {Code: 'y', Mod: tea.ModCtrl}} {
		tm, _ = key(tm, k)
		if got, want := tm.(Model).status, "Файл не знайдено: Сборники/Румбы.fb2"; got != want {
			t.Fatalf("status = %q, want %q", got, want)
		}
	}
	if len(rec.opened)+len(rec.revealed)+len(rec.copied) != 0 {
		t.Fatalf("action called for a missing file: %q %q %q", rec.opened, rec.revealed, rec.copied)
	}
}

func TestDefaultActionsExists(t *testing.T) {
	dir := t.TempDir()
	exists := DefaultActions().Exists
	if !exists(dir) {
		t.Fatalf("Exists(%q) = false", dir)
	}
	if exists(filepath.Join(dir, "нема.fb2")) {
		t.Fatal("Exists of a missing file = true")
	}
}
