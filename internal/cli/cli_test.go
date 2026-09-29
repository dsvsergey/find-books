package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"findbooks/internal/index"
)

func skipUnlessDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("platform is implemented only for macOS in v1")
	}
}

func run(t *testing.T, db string, args ...string) (string, string, error) {
	t.Helper()
	cmd := NewRootCmd()
	var out, errb bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errb)
	cmd.SetArgs(append([]string{"--db", db}, args...))
	err := cmd.Execute()
	return out.String(), errb.String(), err
}

func copyFixture(t *testing.T, name, dst string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "fb2", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func newLibrary(t *testing.T) (root, db string) {
	t.Helper()
	root = t.TempDir()
	copyFixture(t, "rumby.fb2", filepath.Join(root, "Сборники", "Румбы.fb2"))
	copyFixture(t, "novel.fb2", filepath.Join(root, "Беляев", "Человек-амфибия.fb2"))
	return root, filepath.Join(t.TempDir(), "index.db")
}

func TestAddSearchListUpdateRemove(t *testing.T) {
	skipUnlessDarwin(t)
	root, db := newLibrary(t)

	_, stderr, err := run(t, db, "add", root, "--name", "Тест")
	if err != nil || !strings.Contains(stderr, "Додано: 2") {
		t.Fatalf("add: %v\n%s", err, stderr)
	}

	out, _, err := run(t, db, "list")
	if err != nil || !strings.Contains(out, "Тест") || !strings.Contains(out, "●") {
		t.Fatalf("list: %v\n%s", err, out)
	}

	out, _, err = run(t, db, "search", "чужие", "дети")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"ЧУЖИЕ ДЕТИ", "Игорь Пидоренко", "Румбы фантастики. 1988 год. Том II (1988)", "Тест: Сборники/Румбы.fb2"} {
		if !strings.Contains(out, want) {
			t.Errorf("search output lacks %q:\n%s", want, out)
		}
	}

	out, _, err = run(t, db, "search", "--json", "чужие дети")
	if err != nil {
		t.Fatal(err)
	}
	var hits []jsonHit
	if err := json.Unmarshal([]byte(out), &hits); err != nil {
		t.Fatalf("json: %v\n%s", err, out)
	}
	if len(hits) != 1 || !hits[0].Online || hits[0].TreePath != "Игорь Пидоренко › ЧУЖИЕ ДЕТИ" {
		t.Fatalf("hits = %+v", hits)
	}
	if _, err := os.Stat(hits[0].Path); err != nil {
		t.Fatalf("path %q does not exist: %v", hits[0].Path, err)
	}

	if _, stderr, err = run(t, db, "update", "Тест"); err != nil || !strings.Contains(stderr, "без змін: 2") {
		t.Fatalf("update: %v\n%s", err, stderr)
	}
	if _, _, err = run(t, db, "update"); err == nil {
		t.Fatal("update without name or --all must fail")
	}

	if _, _, err = run(t, db, "remove", "Тест"); err != nil {
		t.Fatal(err)
	}
	if _, _, err = run(t, db, "search", "чужие"); !errors.Is(err, errNoLibraries) {
		t.Fatalf("err = %v, want errNoLibraries", err)
	}
}

func TestAddDuplicateName(t *testing.T) {
	skipUnlessDarwin(t)
	root, db := newLibrary(t)
	if _, _, err := run(t, db, "add", root); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, db, "add", root); !errors.Is(err, index.ErrExists) {
		t.Fatalf("err = %v, want ErrExists", err)
	}
}

func TestOfflineLibrary(t *testing.T) {
	skipUnlessDarwin(t)
	root, db := newLibrary(t)
	if _, _, err := run(t, db, "add", root, "--name", "Тест"); err != nil {
		t.Fatal(err)
	}
	st, err := index.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	lib, err := st.AddLibrary("Офлайн", "00000000-0000-0000-0000-00000000DEAD", "Старий диск", "books")
	if err == nil {
		err = st.WriteBatch(lib.ID, []index.BookRecord{{
			RelPath: "Лем/Солярис.fb2", Format: "fb2", Size: 1, MTime: 1, Title: "Солярис", Authors: "Станислав Лем",
			Works: []index.WorkRecord{{Title: "Солярис", Author: "Станислав Лем", TreePath: "Солярис"}},
		}}, nil)
	}
	st.Close()
	if err != nil {
		t.Fatal(err)
	}

	_, stderr, err := run(t, db, "update", "--all")
	if err != nil || !strings.Contains(stderr, "диск «Старий диск» не підключено") {
		t.Fatalf("update --all: %v\n%s", err, stderr)
	}
	out, _, err := run(t, db, "search", "солярис")
	if err != nil || !strings.Contains(out, "○") {
		t.Fatalf("search: %v\n%s", err, out)
	}
	out, _, err = run(t, db, "search", "--json", "солярис")
	if err != nil {
		t.Fatal(err)
	}
	var hits []jsonHit
	if err := json.Unmarshal([]byte(out), &hits); err != nil || len(hits) != 1 || hits[0].Online || hits[0].Path != "" {
		t.Fatalf("hits = %+v, err = %v", hits, err)
	}
}

func TestAddArgCountError(t *testing.T) {
	db := filepath.Join(t.TempDir(), "index.db")
	if _, _, err := run(t, db, "add"); err == nil || !strings.Contains(err.Error(), "очікується аргументів: 1, отримано: 0") {
		t.Fatalf("err = %v, want arg count error", err)
	}
}

func TestSearchRequiresQuery(t *testing.T) {
	db := filepath.Join(t.TempDir(), "index.db")
	if _, _, err := run(t, db, "search"); err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestRootCommandStartsTUI(t *testing.T) {
	db := filepath.Join(t.TempDir(), "index.db")
	if _, _, err := run(t, db); !errors.Is(err, errNoLibraries) {
		t.Fatalf("empty index: err = %v, want errNoLibraries", err)
	}
	st, err := index.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	lib, _ := st.AddLibrary("A", "VOL", "Disk", ".")
	st.WriteBatch(lib.ID, []index.BookRecord{{RelPath: "a.pdf", Format: "pdf", Title: "A", Works: []index.WorkRecord{{Title: "A", TreePath: "A"}}}}, nil)
	st.Close()

	var gotTotal int
	orig := runTUI
	runTUI = func(_ *index.Store, total int) error { gotTotal = total; return nil }
	t.Cleanup(func() { runTUI = orig })
	if _, _, err := run(t, db); err != nil || gotTotal != 1 {
		t.Fatalf("err = %v, total = %d", err, gotTotal)
	}
}

func TestAddScanFailureHintsUpdate(t *testing.T) {
	skipUnlessDarwin(t)
	root, db := newLibrary(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // the first scan fails as if interrupted with Ctrl+C
	cmd := NewRootCmd()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--db", db, "add", root, "--name", "Тест"})
	err := cmd.ExecuteContext(ctx)
	if err == nil {
		t.Fatal("add with a failed scan must fail")
	}
	if !strings.Contains(err.Error(), "findbooks update «Тест»") || !strings.Contains(err.Error(), "зареєстровано") {
		t.Fatalf("err = %q, want a hint to run findbooks update", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v must still wrap context.Canceled", err)
	}
	if out, _, err := run(t, db, "list"); err != nil || !strings.Contains(out, "Тест") {
		t.Fatalf("library not registered: %v\n%s", err, out)
	}
	if _, stderr, err := run(t, db, "update", "Тест"); err != nil || !strings.Contains(stderr, "Додано: 2") {
		t.Fatalf("update after failed add: %v\n%s", err, stderr)
	}
}

func TestScanIncompleteErrorMessage(t *testing.T) {
	err := scanIncompleteError("Б", errors.New("диск зник"))
	want := "бібліотеку «Б» зареєстровано, але індексування не завершено: диск зник — продовжіть: findbooks update «Б»"
	if err.Error() != want {
		t.Fatalf("got %q\nwant %q", err, want)
	}
}

func TestExitCode(t *testing.T) {
	for _, tc := range []struct {
		err        error
		code       int
		wantStderr string
	}{
		{nil, 0, ""},
		{errors.New("бум"), 1, "помилка: бум\n"},
		{context.Canceled, 130, "перервано\n"},
		{fmt.Errorf("scan: %w", context.Canceled), 130, "перервано\n"},
		{scanIncompleteError("Б", context.Canceled), 130,
			"перервано\nбібліотеку «Б» зареєстровано, але індексування не завершено — продовжіть: findbooks update «Б»\n"},
	} {
		var buf bytes.Buffer
		if code := exitCode(tc.err, &buf); code != tc.code || buf.String() != tc.wantStderr {
			t.Errorf("exitCode(%v) = %d, %q; want %d, %q", tc.err, code, buf.String(), tc.code, tc.wantStderr)
		}
	}
}
