package scan

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	write(t, filepath.Join(root, "Бойцов - Рассказы.txt"), []byte("text")) // NFD name
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
	if rep.Unchanged != 5 || rep.Added != 0 || rep.Updated != 0 || rep.Removed != 0 || len(rep.Errors) != 0 {
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
	if rep.Updated != 1 || rep.Removed != 1 || rep.Unchanged != 3 || rep.Added != 0 {
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
