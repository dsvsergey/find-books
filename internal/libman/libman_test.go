package libman

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"findbooks/internal/index"
	"findbooks/internal/scan"
)

func skipUnlessDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("platform is implemented only for macOS")
	}
}

func newStore(t *testing.T) *index.Store {
	t.Helper()
	st, err := index.Open(filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func newLibraryDir(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "fb2", "testdata", "rumby.fb2"))
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(t.TempDir(), "Книги")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "Румбы.fb2"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestRegisterAndScan(t *testing.T) {
	skipUnlessDarwin(t)
	st := newStore(t)
	dir := newLibraryDir(t)
	lib, root, err := Register(st, dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if lib.ID == 0 || lib.Name != "Книги" || filepath.Base(root) != "Книги" {
		t.Fatalf("lib = %+v, root = %q", lib, root)
	}
	var last struct{ done, total int }
	rep, err := Scan(context.Background(), st, lib, func(p scan.Progress) { last.done, last.total = p.Done, p.Total })
	if err != nil {
		t.Fatal(err)
	}
	if rep.Added != 1 || last.done != 1 || last.total != 1 {
		t.Fatalf("report = %+v, last progress = %+v", rep, last)
	}
	hits, err := st.Search(index.Query{Text: "чужие дети"})
	if err != nil || len(hits) != 1 {
		t.Fatalf("hits = %+v, err = %v", hits, err)
	}
}

func TestRegisterCustomName(t *testing.T) {
	skipUnlessDarwin(t)
	st := newStore(t)
	lib, _, err := Register(st, newLibraryDir(t), "Фантастика")
	if err != nil || lib.Name != "Фантастика" {
		t.Fatalf("lib = %+v, err = %v", lib, err)
	}
}

func TestRegisterDuplicate(t *testing.T) {
	skipUnlessDarwin(t)
	st := newStore(t)
	dir := newLibraryDir(t)
	if _, _, err := Register(st, dir, ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Register(st, dir, ""); !errors.Is(err, index.ErrExists) {
		t.Fatalf("err = %v, want ErrExists", err)
	}
}

func TestRegisterRejectsFile(t *testing.T) {
	skipUnlessDarwin(t)
	st := newStore(t)
	file := filepath.Join(newLibraryDir(t), "Румбы.fb2")
	if _, _, err := Register(st, file, ""); err == nil {
		t.Fatal("expected error for a file path")
	}
}

func TestScanOffline(t *testing.T) {
	st := newStore(t)
	lib, err := st.AddLibrary("Старе", "00000000-0000-0000-0000-00000000DEAD", "Старий диск", "books")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Scan(context.Background(), st, lib, nil)
	if !errors.Is(err, ErrOffline) || !strings.Contains(err.Error(), "Старий диск") {
		t.Fatalf("err = %v, want ErrOffline naming the disk", err)
	}
}
