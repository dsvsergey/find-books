package tui

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"findbooks/internal/fb2"
	"findbooks/internal/i18n"
	"findbooks/internal/index"
	"findbooks/internal/preview"
)

var (
	ctrlR     = tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl}
	rumbyPath = filepath.Join("/Volumes", "dsvDev", "Бібліотека", "Сборники", "Румбы.fb2")
)

func storyDoc(paras int) preview.Doc {
	d := preview.Doc{Title: "ЧУЖИЕ ДЕТИ", Author: "Игорь Пидоренко"}
	for i := range paras {
		d.Blocks = append(d.Blocks, fb2.Block{Kind: fb2.BlockPara, Text: fmt.Sprintf("Абзац номер %d.", i)})
	}
	return d
}

// opened searches "чужие дети" with rec.doc set to doc and opens the preview.
func opened(t *testing.T, doc preview.Doc) (tea.Model, *recorder) {
	t.Helper()
	tm, _, rec := searched(t, true, "чужие дети")
	rec.doc = doc
	tm, cmd := key(tm, ctrlR)
	if cmd == nil {
		t.Fatalf("ctrl+r started no load; status = %q", tm.(Model).status)
	}
	if st := tm.(Model).status; st != i18n.T(i18n.KeyPreviewLoading) {
		t.Fatalf("status while loading = %q", st)
	}
	tm, _ = tm.Update(cmd())
	return tm, rec
}

func TestCtrlROpensPreview(t *testing.T) {
	tm, rec := opened(t, storyDoc(3))
	m := tm.(Model)
	if m.screen != screenPreview || !slices.Equal(rec.previewed, []string{rumbyPath}) {
		t.Fatalf("screen = %v, previewed = %q", m.screen, rec.previewed)
	}
	view := m.View().Content
	for _, want := range []string{"ЧУЖИЕ ДЕТИ — Игорь Пидоренко", "Абзац номер 0.", "Румбы фантастики"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view lacks %q:\n%s", want, view)
		}
	}
	if m.status != "" {
		t.Fatalf("status = %q after load", m.status)
	}
}

func TestPreviewEscBackKeepsState(t *testing.T) {
	tm, _ := opened(t, storyDoc(3))
	before := tm.(Model)
	tm, cmd := key(tm, tea.KeyPressMsg{Code: tea.KeyEscape})
	m := tm.(Model)
	if cmd != nil || m.screen != screenSearch || m.input.Value() != "чужие дети" || m.cursor != before.cursor || len(m.hits) != len(before.hits) {
		t.Fatalf("after esc: screen=%v query=%q cursor=%d hits=%d cmd=%v", m.screen, m.input.Value(), m.cursor, len(m.hits), cmd)
	}
}

func TestPreviewRefusals(t *testing.T) {
	pdf := chuzhie
	pdf.Format = "pdf"
	tests := []struct {
		name   string
		setup  func(*fakeSearcher, *recorder)
		online bool
		want   string
	}{
		{"offline", func(*fakeSearcher, *recorder) {}, false, i18n.T(i18n.KeyDiskOfflineOpen, "dsvDev")},
		{"missing", func(_ *fakeSearcher, r *recorder) { r.missing = true }, true, i18n.T(i18n.KeyFileNotFound, chuzhie.RelPath)},
		{"unsupported", func(s *fakeSearcher, _ *recorder) { s.hits = []index.Hit{pdf} }, true, i18n.T(i18n.KeyPreviewUnsupported, "PDF")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, src, rec := newModel(tt.online)
			tt.setup(src, rec)
			tm, _ := typeText(m, "чужие дети")
			tm = settle(t, tm)
			tm, cmd := key(tm, ctrlR)
			if cmd != nil || tm.(Model).screen != screenSearch || len(rec.previewed) != 0 {
				t.Fatalf("preview started: cmd=%v screen=%v previewed=%q", cmd, tm.(Model).screen, rec.previewed)
			}
			if st := tm.(Model).status; st != tt.want {
				t.Fatalf("status = %q, want %q", st, tt.want)
			}
		})
	}
}

func TestPreviewLoadError(t *testing.T) {
	tm, _, rec := searched(t, true, "чужие дети")
	rec.previewErr = errors.New("boom")
	tm, cmd := key(tm, ctrlR)
	tm, _ = tm.Update(cmd())
	if m := tm.(Model); m.screen != screenSearch || m.status != i18n.T(i18n.KeyPreviewFailed, "boom") {
		t.Fatalf("screen = %v, status = %q", m.screen, m.status)
	}
}

func TestStalePreviewIgnored(t *testing.T) {
	tm, _, rec := searched(t, true, "чужие дети")
	rec.doc = storyDoc(1)
	tm, cmd := key(tm, ctrlR)
	tm, _ = key(tm, tea.KeyPressMsg{Code: tea.KeyDown}) // user moved on
	tm, _ = tm.Update(cmd())
	if m := tm.(Model); m.screen != screenSearch || m.status != "" {
		t.Fatalf("late preview applied: screen = %v, status = %q", m.screen, m.status)
	}
}

func TestEscWhileLoadingCancels(t *testing.T) {
	tm, _, _ := searched(t, true, "чужие дети")
	tm, load := key(tm, ctrlR)
	tm, cmd := key(tm, tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd != nil {
		t.Fatal("esc while loading must cancel, not quit")
	}
	tm, _ = tm.Update(load())
	if m := tm.(Model); m.screen != screenSearch || m.status != "" {
		t.Fatalf("screen = %v, status = %q", m.screen, m.status)
	}
}

func TestPreviewScrollAndResize(t *testing.T) {
	tm, _, rec := searched(t, true, "чужие дети")
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	rec.doc = storyDoc(200)
	tm, cmd := key(tm, ctrlR)
	tm, _ = tm.Update(cmd())
	vp := func() int { m := tm.(Model); return m.pvView.YOffset() } // YOffset has a pointer receiver
	tm, _ = key(tm, tea.KeyPressMsg{Code: tea.KeyPgDown})
	if vp() == 0 {
		t.Fatal("pgdown did not scroll")
	}
	tm, _ = key(tm, tea.KeyPressMsg{Code: tea.KeyEnd})
	if m := tm.(Model); !m.pvView.AtBottom() {
		t.Fatal("end did not reach the bottom")
	}
	tm, _ = key(tm, tea.KeyPressMsg{Code: tea.KeyHome})
	if vp() != 0 {
		t.Fatal("home did not reach the top")
	}
	tm, _ = key(tm, tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if vp() == 0 {
		t.Fatal("space did not page down")
	}
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	if w := tm.(Model).pvView.Width(); w != 36 {
		t.Fatalf("text width after resize = %d, want 36", w)
	}
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 200, Height: 20})
	if w := tm.(Model).pvView.Width(); w != previewMaxWidth {
		t.Fatalf("text width on a wide screen = %d, want %d", w, previewMaxWidth)
	}
}

// TestPreviewRewrapsOnlyOnWidthChange guards the perf fix: a width change
// must re-wrap the text (more, narrower lines), while a height-only resize
// must leave the rendered content untouched.
func TestPreviewRewrapsOnlyOnWidthChange(t *testing.T) {
	tm, _, rec := searched(t, true, "чужие дети")
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 120, Height: 20})
	long := strings.Repeat("слово ", 400)
	rec.doc = preview.Doc{Title: "T", Blocks: []fb2.Block{{Kind: fb2.BlockPara, Text: long}}}
	tm, cmd := key(tm, ctrlR)
	tm, _ = tm.Update(cmd())
	wide := tm.(Model).pvView.TotalLineCount()
	wideContent := tm.(Model).pvView.GetContent()

	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	narrow := tm.(Model).pvView.TotalLineCount()
	if narrow <= wide {
		t.Fatalf("narrow lines = %d, want more than wide lines = %d", narrow, wide)
	}
	narrowContent := tm.(Model).pvView.GetContent()
	if narrowContent == wideContent {
		t.Fatal("content unchanged after a width change")
	}

	tm, _ = tm.Update(tea.WindowSizeMsg{Width: 40, Height: 30})
	if got := tm.(Model).pvView.TotalLineCount(); got != narrow {
		t.Fatalf("height-only resize changed line count: got %d, want %d", got, narrow)
	}
	if got := tm.(Model).pvView.GetContent(); got != narrowContent {
		t.Fatal("height-only resize changed the rendered content")
	}
}

func TestPreviewActions(t *testing.T) {
	tm, rec := opened(t, storyDoc(3))
	tm, _ = key(tm, tea.KeyPressMsg{Code: tea.KeyEnter})
	tm, _ = key(tm, tea.KeyPressMsg{Code: 'o', Mod: tea.ModCtrl})
	tm, _ = key(tm, tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})
	want := []string{rumbyPath}
	if !slices.Equal(rec.opened, want) || !slices.Equal(rec.revealed, want) || !slices.Equal(rec.copied, want) {
		t.Fatalf("opened=%q revealed=%q copied=%q", rec.opened, rec.revealed, rec.copied)
	}
	m := tm.(Model)
	if m.screen != screenPreview || !strings.Contains(m.View().Content, i18n.T(i18n.KeyPathCopied)) {
		t.Fatalf("screen = %v; view lacks copied status", m.screen)
	}
}

func TestPreviewStaleWarning(t *testing.T) {
	doc := storyDoc(1)
	doc.Stale = true
	tm, _ := opened(t, doc)
	if !strings.Contains(tm.(Model).View().Content, i18n.T(i18n.KeyPreviewStale)) {
		t.Fatal("stale warning missing")
	}
}

func TestPreviewIgnoresLetters(t *testing.T) {
	tm, _ := opened(t, storyDoc(3))
	tm, _ = typeText(tm, "jq")
	if m := tm.(Model); m.screen != screenPreview || m.input.Value() != "чужие дети" {
		t.Fatalf("letters leaked: screen = %v, query = %q", m.screen, m.input.Value())
	}
}
