package scan

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"findbooks/internal/extract"
	"findbooks/internal/index"
)

func write(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "fb2", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

const novelRel = "Авторы/Беляев Александр/Человек-амфибия.fb2"

// setup builds a library with 5 indexable files and several ignored ones.
func setup(t *testing.T) (*index.Store, index.Library, string) {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, "Сборники", "Румбы 1988.fb2"), fixture(t, "rumby.fb2"))
	write(t, filepath.Join(root, filepath.FromSlash(novelRel)), fixture(t, "novel.fb2"))
	write(t, filepath.Join(root, "Авторы", "Беляев Александр", "Звезда КЭЦ.pdf"), []byte("%PDF-1.4"))
	write(t, filepath.Join(root, "broken.fb2"), []byte("<FictionBook><body><section>"))
	write(t, filepath.Join(root, "Бойцов - Рассказы.txt"), []byte("text")) // NFD name
	// ignored:
	write(t, filepath.Join(root, ".DS_Store"), []byte("x"))
	write(t, filepath.Join(root, "Авторы", "Беляев Александр", "._Человек-амфибия.fb2"), []byte("x"))
	write(t, filepath.Join(root, ".Trashes", "old.fb2"), fixture(t, "novel.fb2"))
	write(t, filepath.Join(root, "cover.jpg"), []byte("x"))

	st, err := index.Open(filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	lib, err := st.AddLibrary("test", "VOL", "Test", ".")
	if err != nil {
		t.Fatal(err)
	}
	return st, lib, root
}

func hits(t *testing.T, st *index.Store, q string) []index.Hit {
	t.Helper()
	h, err := st.Search(index.Query{Text: q})
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestRunIndexesLibrary(t *testing.T) {
	st, lib, root := setup(t)
	// Guard: verify NFD fixture is decomposed
	entries, _ := os.ReadDir(root)
	found := false
	for _, e := range entries {
		if strings.Contains(e.Name(), "̆") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("NFD fixture was normalized to NFC; test cannot verify decomposed names")
	}
	var last Progress
	rep, err := Run(context.Background(), st, lib.ID, root, func(p Progress) { last = p })
	if err != nil {
		t.Fatal(err)
	}
	if rep.Added != 5 || rep.Updated != 0 || rep.Removed != 0 || rep.Unchanged != 0 {
		t.Fatalf("report = %+v", rep)
	}
	if len(rep.Errors) != 1 || rep.Errors[0].RelPath != "broken.fb2" {
		t.Fatalf("errors = %+v", rep.Errors)
	}
	if last != (Progress{Done: 5, Total: 5}) {
		t.Fatalf("last progress = %+v", last)
	}
	h := hits(t, st, "чужие дети")
	if len(h) != 1 || h[0].TreePath != "Игорь Пидоренко › ЧУЖИЕ ДЕТИ" || h[0].RelPath != "Сборники/Румбы 1988.fb2" {
		t.Fatalf("чужие дети: %+v", h)
	}
	if h := hits(t, st, "человек амфибия"); len(h) != 1 || h[0].IsCollection {
		t.Fatalf("novel: %+v", h)
	}
	if h := hits(t, st, "дьявол"); len(h) != 0 {
		t.Fatalf("chapters must not be indexed: %+v", h)
	}
	for _, q := range []string{"звезда кэц", "broken", "бойцов"} {
		if h := hits(t, st, q); len(h) != 1 {
			t.Errorf("%q: %+v", q, h)
		}
	}
	if h := hits(t, st, "trashes"); len(h) != 0 {
		t.Errorf("hidden dir indexed: %+v", h)
	}
	libs, _ := st.Libraries()
	if libs[0].LastScan.IsZero() {
		t.Error("last scan time not recorded")
	}
}

func TestRunIsIncremental(t *testing.T) {
	st, lib, root := setup(t)
	if _, err := Run(context.Background(), st, lib.ID, root, nil); err != nil {
		t.Fatal(err)
	}
	rep, err := Run(context.Background(), st, lib.ID, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	// broken.fb2 failed to parse, so it was stored with size -1 and is
	// retried (and fails again) on every run; the other 4 files are unchanged.
	if rep.Unchanged != 4 || rep.Added != 0 || rep.Updated != 1 || rep.Removed != 0 ||
		len(rep.Errors) != 1 || rep.Errors[0].RelPath != "broken.fb2" {
		t.Fatalf("second run report = %+v", rep)
	}
}

func TestRunDetectsChangesAndRemovals(t *testing.T) {
	st, lib, root := setup(t)
	if _, err := Run(context.Background(), st, lib.ID, root, nil); err != nil {
		t.Fatal(err)
	}
	novel := filepath.Join(root, filepath.FromSlash(novelRel))
	changed := strings.Replace(string(fixture(t, "novel.fb2")), "Человек-амфибия", "Продавец воздуха", 1)
	write(t, novel, []byte(changed))
	later := time.Now().Add(time.Hour)
	if err := os.Chtimes(novel, later, later); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "Авторы", "Беляев Александр", "Звезда КЭЦ.pdf")); err != nil {
		t.Fatal(err)
	}
	rep, err := Run(context.Background(), st, lib.ID, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Updated: the novel plus broken.fb2, which is retried on every run.
	if rep.Updated != 2 || rep.Removed != 1 || rep.Unchanged != 2 || rep.Added != 0 {
		t.Fatalf("report = %+v", rep)
	}
	if len(hits(t, st, "продавец воздуха")) != 1 || len(hits(t, st, "амфибия")) != 0 || len(hits(t, st, "кэц")) != 0 {
		t.Fatal("index does not reflect the changes")
	}
}

func TestRunCanceled(t *testing.T) {
	st, lib, root := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Run(ctx, st, lib.ID, root, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestFormatAndStem(t *testing.T) {
	for name, want := range map[string]string{"a.FB2": "fb2", "a.fb2.ZIP": "fb2", "a.djv": "djvu", "a.Pdf": "pdf"} {
		if got, ok := formatOf(name); !ok || got != want {
			t.Errorf("formatOf(%q) = %q, %v", name, got, ok)
		}
	}
	if _, ok := formatOf("a.zip"); ok {
		t.Error("plain .zip must not be a book")
	}
	if got := stem("Книга.fb2.zip"); got != "Книга" {
		t.Errorf("stem = %q", got)
	}
}

func TestRunRetriesFailedParse(t *testing.T) {
	st, lib, root := setup(t)
	// broken.fb2 would be re-parsed on every run too; drop it so the counts
	// below reflect only the file under test.
	if err := os.Remove(filepath.Join(root, "broken.fb2")); err != nil {
		t.Fatal(err)
	}
	rumby := filepath.Join(root, "Сборники", "Румбы 1988.fb2")
	if err := os.Chmod(rumby, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(rumby, 0o644) })
	rep, err := Run(context.Background(), st, lib.ID, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	var rumbyErr bool
	for _, e := range rep.Errors {
		if e.RelPath == "Сборники/Румбы 1988.fb2" {
			rumbyErr = true
		}
	}
	if !rumbyErr {
		t.Fatalf("unreadable file not reported: %+v", rep.Errors)
	}
	if h := hits(t, st, "румбы 1988"); len(h) != 1 {
		t.Fatalf("filename-only record missing: %+v", h)
	}
	if h := hits(t, st, "чужие дети"); len(h) != 0 {
		t.Fatalf("unreadable file must not have works yet: %+v", h)
	}

	if err := os.Chmod(rumby, 0o644); err != nil {
		t.Fatal(err)
	}
	rep, err = Run(context.Background(), st, lib.ID, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Updated != 1 || rep.Added != 0 || rep.Removed != 0 {
		t.Fatalf("second run report = %+v", rep)
	}
	if h := hits(t, st, "чужие дети"); len(h) != 1 {
		t.Fatalf("restored file was not re-parsed: %+v", h)
	}
}

func TestRunKeepsBooksInUnreadableDir(t *testing.T) {
	st, lib, root := setup(t)
	if _, err := Run(context.Background(), st, lib.ID, root, nil); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "Авторы", "Беляев Александр")
	if err := os.Chmod(dir, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })
	rep, err := Run(context.Background(), st, lib.ID, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Removed != 0 {
		t.Fatalf("books under an unreadable dir were removed: %+v", rep)
	}
	var dirErr bool
	for _, e := range rep.Errors {
		if e.RelPath == "Авторы/Беляев Александр" {
			dirErr = true
		}
	}
	if !dirErr {
		t.Fatalf("unreadable dir not reported: %+v", rep.Errors)
	}
	for _, q := range []string{"человек амфибия", "звезда кэц"} {
		if h := hits(t, st, q); len(h) != 1 {
			t.Errorf("%q: %+v", q, h)
		}
	}
}

func TestDeletions(t *testing.T) {
	stored := map[string]index.FileStamp{
		"a.fb2":       {BookID: 1},
		"dir/b.fb2":   {BookID: 2},
		"dir/x/c.fb2": {BookID: 3},
		"dirx/d.fb2":  {BookID: 4},
		"kept.fb2":    {BookID: 5},
	}
	seen := map[string]bool{"kept.fb2": true}
	got := deletions(stored, seen, []string{"dir"})
	slices.Sort(got)
	if want := []int64{1, 4}; !slices.Equal(got, want) {
		t.Fatalf("deletions = %v, want %v", got, want)
	}
}

func TestRunRootVanishedDeletesNothing(t *testing.T) {
	st, lib, root := setup(t)
	if _, err := Run(context.Background(), st, lib.ID, root, nil); err != nil {
		t.Fatal(err)
	}
	// Simulate the root disappearing between the walk and the delete step.
	statRoot = func(string) error { return os.ErrNotExist }
	t.Cleanup(func() { statRoot = defaultStatRoot })
	rep, err := Run(context.Background(), st, lib.ID, root, nil)
	if err == nil || !strings.Contains(err.Error(), "тека бібліотеки недоступна") {
		t.Fatalf("err = %v", err)
	}
	if rep.Removed != 0 {
		t.Fatalf("report = %+v", rep)
	}
	if h := hits(t, st, "человек амфибия"); len(h) != 1 {
		t.Fatalf("book deleted: %+v", h)
	}
}

func TestRunReparsesAllOnExtractVersionChange(t *testing.T) {
	st, lib, root := setup(t)
	// broken.fb2 is retried on every run; drop it so "unchanged" is exact.
	if err := os.Remove(filepath.Join(root, "broken.fb2")); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(context.Background(), st, lib.ID, root, nil); err != nil {
		t.Fatal(err)
	}
	if v, err := st.ExtractVersion(lib.ID); err != nil || v != extract.Version {
		t.Fatalf("ExtractVersion after run = %d, %v; want %d", v, err, extract.Version)
	}
	if err := st.SetExtractVersion(lib.ID, 0); err != nil {
		t.Fatal(err)
	}
	rep, err := Run(context.Background(), st, lib.ID, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Updated != 4 || rep.Unchanged != 0 || rep.Added != 0 || rep.Removed != 0 {
		t.Fatalf("after version change: report = %+v", rep)
	}
	rep, err = Run(context.Background(), st, lib.ID, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Unchanged != 4 || rep.Updated != 0 || rep.Added != 0 || rep.Removed != 0 {
		t.Fatalf("following run: report = %+v", rep)
	}
}
