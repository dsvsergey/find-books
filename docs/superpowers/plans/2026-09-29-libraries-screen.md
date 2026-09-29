# Libraries Screen Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Керувати бібліотеками прямо з TUI: екран «Бібліотеки» з вибором теки через діалог Finder, індексацією з прогрес-баром, оновленням і видаленням; `findbooks` на порожньому індексі відкриває цей екран замість помилки.

**Architecture:** `platform.ChooseFolder` (osascript) дає шлях теки. Новий пакет `internal/libman` містить спільну логіку реєстрації (`Register`) і сканування з перевіркою підключення (`Scan`); CLI `add` переходить на `Register`. TUI отримує інтерфейс `Backend` (пошук + бібліотеки) і другий екран; довгі операції (діалог, сканування, завантаження списку) виконуються в `tea.Cmd`, прогрес сканування надходить через канал і команду-«слухача».

**Tech Stack:** Go 1.27, Bubble Tea v2 (`charm.land/bubbletea/v2` v2.0.10), Bubbles v2 (`progress`, `textinput`), Lip Gloss v2, Cobra, `modernc.org/sqlite`.

**Spec:** `docs/superpowers/specs/2026-09-29-findbooks-libraries-screen-design.md` (доповнює `docs/superpowers/specs/2026-09-29-findbooks-design.md`)

## Global Constraints

- ОС-специфічні виклики (`osascript`, `diskutil`, `open`) і build-теги — лише в `internal/platform`. `GOOS=windows` і `GOOS=linux` `CGO_ENABLED=0 go build ./...` мають компілюватися.
- Тексти для користувача — українською.
- Жодна операція екрана «Бібліотеки» не блокує UI-потік: діалог, сканування, завантаження списку і перевірка підключення — у `tea.Cmd`.
- Тести ніколи не пишуть поза тимчасовими теками (`XDG_DATA_HOME` через `TestMain` у пакетах, що доходять до `platform.DataDir()`), не чіпають `/Volumes` і `~/.local/share`, не відкривають справжній діалог (`osascript` замінюється в тестах).
- Поведінка і вивід CLI `add`/`update`/`list`/`remove`/`search` не змінюються; `findbooks` без аргументів на порожньому індексі відкриває TUI замість помилки.
- Кожен коміт закінчується рядком `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Уточнення специфікації (прийняті в цьому плані)

1. **API `libman`:** замість `Add`/`Update` — `Register(st, dir, name) (index.Library, string, error)` і `Scan(ctx, st, lib, onProgress) (scan.Report, error)`. Причина: CLI друкує «Індексую «X» (шлях)…» між реєстрацією і скануванням; дві функції дають це без колбеків. `Scan` = «Update» зі специфікації (повертає `ErrOffline`). CLI `update` лишається як є (його логіка — три рядки навколо `runScan`), CLI `add` переходить на `Register`.
2. **Стартовий екран:** TUI стартує на екрані «Бібліотеки», коли в індексі 0 творів (а не лише 0 бібліотек) — так само корисно після перерваного першого `add`. `esc` на екрані бібліотек при 0 творів виходить з програми.
3. **Українська розкладка:** клавіші екрана бібліотек приймають і латиницю, і відповідні клавіші української розкладки: `a`/`ф`, `u`/`г`, `d`/`в`, `y`/`н`.
4. **Повернення до пошуку:** після `esc` з бібліотек, якщо в полі пошуку є текст, пошук повторюється (індекс міг змінитись).

## Review Focus

1. **Діалог скасовано (Cancel/Esc у Finder)** — нічого не відбувається, без повідомлення про помилку. Тест: Task 3 `TestChooseFolderCanceled`; Task 1 `TestChooseFolderCanceled`.
2. **Українська розкладка** — `ф`/`г`/`в`/`н` працюють як `a`/`u`/`d`/`y`. Тест: Task 3 `TestUkrainianLayoutKeys`.
3. **`esc` або `ctrl+c` під час індексації** — сканування скасовується через context, бібліотека лишається зареєстрованою, статус підказує `u`; UI не блокується. Тест: Task 3 `TestEscCancelsIndexing`.
4. **Вибрано теку, яка вже є бібліотекою** — «Бібліотека «X» уже є», без падіння. Тест: Task 3 `TestAddExistingLibrary`; Task 2 `TestRegisterDuplicate`.
5. **Прибрано останню бібліотеку** — лічильник творів 0, `esc` виходить (не показує порожній пошук). Тест: Task 3 `TestDeleteLastLibraryThenEscQuits`.

---

## Структура файлів

```
internal/platform/platform.go          + ErrCanceled
internal/platform/platform_darwin.go   + ChooseFolder, runOsascript
internal/platform/platform_windows.go  + ChooseFolder (ErrUnsupported)
internal/platform/platform_linux.go    + ChooseFolder (ErrUnsupported)
internal/libman/libman.go              Register, Scan, ErrOffline
internal/libman/libman_test.go
internal/libman/main_test.go           TestMain: XDG_DATA_HOME у тимчасовій теці
internal/cli/add.go                    використовує libman.Register
internal/cli/root.go                   порожній індекс → TUI
internal/tui/model.go                  Backend, Actions.ChooseFolder, два екрани, маршрутизація
internal/tui/libraries.go              логіка екрана «Бібліотеки»: повідомлення, команди, клавіші
internal/tui/libraries_view.go         рендер екрана «Бібліотеки»
internal/tui/view.go                   render() → вибір екрана; підказка ctrl+l
internal/tui/run.go                    storeBackend, DefaultActions + ChooseFolder, Run
internal/tui/libraries_test.go
README.md                              розділ про екран «Бібліотеки»
```

---

### Task 1: `platform.ChooseFolder`

**Files:**
- Modify: `internal/platform/platform.go`, `internal/platform/platform_darwin.go`, `internal/platform/platform_windows.go`, `internal/platform/platform_linux.go`
- Test: `internal/platform/platform_darwin_test.go`

**Interfaces:**
- Consumes: —
- Produces:
  - `var platform.ErrCanceled error`
  - `func platform.ChooseFolder(prompt string) (string, error)` — абсолютний шлях без кінцевого `/` (крім `/`); скасування → `ErrCanceled`; windows/linux → `ErrUnsupported`

- [ ] **Step 1: Тести (додати в кінець `internal/platform/platform_darwin_test.go`)**

```go
func stubOsascript(t *testing.T, stdout, stderr string, err error) *[]string {
	t.Helper()
	var got []string
	orig := runOsascript
	runOsascript = func(args ...string) (string, string, error) {
		got = args
		return stdout, stderr, err
	}
	t.Cleanup(func() { runOsascript = orig })
	return &got
}

func TestChooseFolderReturnsPath(t *testing.T) {
	args := stubOsascript(t, "/Volumes/dsvDev/Книги/\n", "", nil)
	prompt := `Виберіть «теку» "x"`
	p, err := ChooseFolder(prompt)
	if err != nil {
		t.Fatal(err)
	}
	if p != "/Volumes/dsvDev/Книги" {
		t.Fatalf("path = %q", p)
	}
	if len(*args) != 3 || (*args)[0] != "-e" || (*args)[2] != prompt {
		t.Fatalf("osascript args = %q (prompt must be passed verbatim as argv)", *args)
	}
}

func TestChooseFolderRoot(t *testing.T) {
	stubOsascript(t, "/\n", "", nil)
	if p, err := ChooseFolder("x"); err != nil || p != "/" {
		t.Fatalf("got %q, %v", p, err)
	}
}

func TestChooseFolderCanceled(t *testing.T) {
	stubOsascript(t, "", "0:98: execution error: User cancelled. (-128)\n", errors.New("exit status 1"))
	if _, err := ChooseFolder("x"); !errors.Is(err, ErrCanceled) {
		t.Fatalf("err = %v, want ErrCanceled", err)
	}
}

func TestChooseFolderOtherError(t *testing.T) {
	stubOsascript(t, "", "boom happened\n", errors.New("exit status 1"))
	_, err := ChooseFolder("x")
	if err == nil || errors.Is(err, ErrCanceled) || !strings.Contains(err.Error(), "boom happened") {
		t.Fatalf("err = %v", err)
	}
}
```

Додати `"errors"` і `"strings"` до імпортів тестового файлу, якщо їх там ще немає.

- [ ] **Step 2: Перевірити падіння**

Run: `go test ./internal/platform/ -run ChooseFolder`
Expected: FAIL — `undefined: runOsascript`, `undefined: ChooseFolder`.

- [ ] **Step 3: Реалізація**

У `internal/platform/platform.go` поруч з `ErrUnsupported`:
```go
// ErrCanceled is returned when the user dismisses an OS dialog.
var ErrCanceled = errors.New("platform: canceled by user")
```

У `internal/platform/platform_darwin.go` (імпорти `bytes`, `fmt`, `os/exec`, `strings` уже є):
```go
// chooseFolderScript receives the prompt as argv so quotes in it cannot
// break the AppleScript source.
const chooseFolderScript = `on run argv
	return POSIX path of (choose folder with prompt (item 1 of argv))
end run`

// runOsascript runs osascript with args; replaced in tests.
var runOsascript = func(args ...string) (stdout, stderr string, err error) {
	cmd := exec.Command("osascript", args...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err = cmd.Run()
	return out.String(), errb.String(), err
}

// ChooseFolder shows the Finder "choose folder" dialog and returns the
// selected folder's absolute path, or ErrCanceled if the user dismissed it.
func ChooseFolder(prompt string) (string, error) {
	out, errOut, err := runOsascript("-e", chooseFolderScript, prompt)
	if err != nil {
		if strings.Contains(errOut, "(-128)") {
			return "", ErrCanceled
		}
		return "", fmt.Errorf("osascript: %w: %s", err, strings.TrimSpace(errOut))
	}
	p := strings.TrimSuffix(out, "\n")
	if len(p) > 1 {
		p = strings.TrimSuffix(p, "/")
	}
	return p, nil
}
```

У `internal/platform/platform_windows.go` і `internal/platform/platform_linux.go` додати:
```go
func ChooseFolder(prompt string) (string, error) { return "", ErrUnsupported }
```

- [ ] **Step 4: Тести і крос-компіляція**

Run: `go test ./internal/platform/ && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./... && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./... && gofmt -l internal/platform`
Expected: `ok`, збірки без помилок, gofmt нічого не друкує.

- [ ] **Step 5: Commit**

```bash
git add internal/platform
git commit -m "feat(platform): Finder folder chooser via osascript

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 2: Пакет `libman` і перехід CLI `add`

**Files:**
- Create: `internal/libman/libman.go`, `internal/libman/main_test.go`
- Modify: `internal/cli/add.go`
- Test: `internal/libman/libman_test.go`

**Interfaces:**
- Consumes: `library.Locate(dir) (library.Location, error)`, `library.Root(volumeID, rootRel) (string, bool)`, `(*index.Store).AddLibrary(name, volumeID, volumeName, rootRel) (index.Library, error)`, `index.ErrExists`, `scan.Run(ctx, st, libraryID, root, onProgress) (scan.Report, error)`, `scan.Progress`, `scan.Report`
- Produces:
  - `var libman.ErrOffline error`
  - `func libman.Register(st *index.Store, dir, name string) (index.Library, string, error)` — повертає бібліотеку і повний шлях її кореня; `name == ""` → `filepath.Base(root)`
  - `func libman.Scan(ctx context.Context, st *index.Store, lib index.Library, onProgress func(scan.Progress)) (scan.Report, error)` — диск не підключено → помилка, що `errors.Is(err, ErrOffline)` і містить назву диска

- [ ] **Step 1: Тести**

`internal/libman/main_test.go`:
```go
package libman

import (
	"os"
	"testing"
)

// TestMain keeps platform's volume cache out of the user's home directory.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "findbooks-test-*")
	if err != nil {
		panic(err)
	}
	os.Setenv("XDG_DATA_HOME", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
```

`internal/libman/libman_test.go`:
```go
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
```

- [ ] **Step 2: Перевірити падіння**

Run: `go test ./internal/libman/`
Expected: FAIL — `undefined: Register`, `undefined: Scan`, `undefined: ErrOffline`.

- [ ] **Step 3: Реалізація**

`internal/libman/libman.go`:
```go
// Package libman is the library management shared by the CLI and the TUI:
// registering a folder as a library and scanning a registered library.
package libman

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"findbooks/internal/index"
	"findbooks/internal/library"
	"findbooks/internal/scan"
)

// ErrOffline means the library's disk is not mounted.
var ErrOffline = errors.New("диск бібліотеки не підключено")

// Register locates dir on its volume and records it as a library named name
// (the folder name when name is empty). It returns the library and the
// full path of its root.
func Register(st *index.Store, dir, name string) (index.Library, string, error) {
	loc, err := library.Locate(dir)
	if err != nil {
		return index.Library{}, "", err
	}
	root, ok := library.Root(loc.VolumeID, loc.RootRel)
	if !ok {
		return index.Library{}, "", fmt.Errorf("диск «%s» не знайдено серед підключених", loc.VolumeName)
	}
	if name == "" {
		name = filepath.Base(root)
	}
	lib, err := st.AddLibrary(name, loc.VolumeID, loc.VolumeName, loc.RootRel)
	if err != nil {
		return index.Library{}, "", err
	}
	return lib, root, nil
}

// Scan brings the index of lib in line with its folder. It fails with
// ErrOffline when the library's disk is not mounted.
func Scan(ctx context.Context, st *index.Store, lib index.Library, onProgress func(scan.Progress)) (scan.Report, error) {
	root, ok := library.Root(lib.VolumeID, lib.RootRel)
	if !ok {
		return scan.Report{}, fmt.Errorf("%w: «%s»", ErrOffline, lib.VolumeName)
	}
	return scan.Run(ctx, st, lib.ID, root, onProgress)
}
```

`internal/cli/add.go` — у `RunE` замінити блок від `loc, err := library.Locate(args[0])` до `lib, err := st.AddLibrary(...)` на:
```go
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer closeStore(st, &err)
			lib, root, err := libman.Register(st, args[0], name)
			if err != nil {
				return err
			}
			if err := runScan(cmd, st, lib, root); err != nil {
				return scanIncompleteError(lib.Name, err)
			}
			return nil
```
Прибрати імпорти `path/filepath` і `findbooks/internal/library`, якщо вони більше не використовуються; додати `findbooks/internal/libman`. Решту `add.go` (`scanIncomplete`) не змінювати.

- [ ] **Step 4: Тести проходять**

Run: `go test ./internal/libman/ ./internal/cli/ && gofmt -l internal/libman internal/cli`
Expected: `ok` для обох пакетів; gofmt нічого не друкує. Наявні тести `cli` (зокрема `TestAddSearchListUpdateRemove`, `TestAddDuplicateName`, `TestAddScanFailureHintsUpdate`) проходять без змін.

- [ ] **Step 5: Commit**

```bash
git add internal/libman internal/cli/add.go
git commit -m "feat(libman): shared library registration and scanning; cli add uses it

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 3: Екран «Бібліотеки» в TUI

**Files:**
- Modify: `internal/tui/model.go`, `internal/tui/view.go`, `internal/tui/run.go`, `internal/tui/model_test.go`
- Create: `internal/tui/libraries.go`, `internal/tui/libraries_view.go`
- Test: `internal/tui/libraries_test.go`

**Interfaces:**
- Consumes: `index.Library{ID int64; Name, VolumeID, VolumeName, RootRel string; LastScan time.Time; Books, Works int}`, `index.ErrExists`, `scan.Progress{Done, Total int}`, `scan.Report{Added, Updated, Removed, Unchanged int; Errors []FileError}`, `libman.Register`, `libman.Scan`, `libman.ErrOffline`, `platform.ChooseFolder`, `platform.ErrCanceled`, `platform.ErrUnsupported`
- Produces:
  - `type tui.Backend interface` (див. код нижче)
  - `tui.Actions.ChooseFolder func(prompt string) (string, error)`
  - `func tui.New(b Backend, act Actions, total int) Model` — `total == 0` → стартовий екран «Бібліотеки»
  - `func tui.Run(st *index.Store, total int) error` — **сигнатура не змінюється** (Task 4 на неї спирається)

- [ ] **Step 1: Тести**

У `internal/tui/model_test.go` додати до `fakeSearcher` методи, щоб він задовольняв `Backend` (імпорти `context`, `errors`, `findbooks/internal/scan`):
```go
func (f *fakeSearcher) Libraries() ([]index.Library, error) { return nil, nil }
func (f *fakeSearcher) TotalWorks() (int, error)            { return 0, nil }
func (f *fakeSearcher) RemoveLibrary(string) error          { return nil }
func (f *fakeSearcher) AddFolder(context.Context, string, func(scan.Progress)) (index.Library, scan.Report, error) {
	return index.Library{}, scan.Report{}, errors.New("not used")
}
func (f *fakeSearcher) Rescan(context.Context, index.Library, func(scan.Progress)) (scan.Report, error) {
	return scan.Report{}, errors.New("not used")
}
```
і до `recorder.actions()` поле `ChooseFolder: func(string) (string, error) { return "", errors.New("not used") }`.

Додати тест у `model_test.go`:
```go
func TestSearchHelpMentionsLibraries(t *testing.T) {
	m, _, _ := newModel(true)
	if !strings.Contains(m.View().Content, "ctrl+l") {
		t.Fatal("search help must mention ctrl+l")
	}
}
```

`internal/tui/libraries_test.go`:
```go
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
```

- [ ] **Step 2: Перевірити падіння**

Run: `go test ./internal/tui/`
Expected: FAIL — `undefined: screenLibraries`, `unknown field ChooseFolder`, `New` не приймає `*fakeLibs` тощо.

- [ ] **Step 3: Реалізація — `internal/tui/model.go`**

Замінити верх файлу (пакетний коментар, імпорти, `Searcher`, `Actions`, константи, повідомлення, `Model`, `New`, `Init`, `Update`) на наведене нижче; функції `queueSearch`, `searchCmd`, `move`, `selected`, `withFile`, `loadTOC` лишаються, але в `searchCmd` і `loadTOC` поле `m.src` перейменовується на `m.b`.

```go
// Package tui is the interactive search screen and the libraries screen.
package tui

import (
	"context"
	"fmt"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"findbooks/internal/index"
	"findbooks/internal/library"
	"findbooks/internal/scan"
)

type Searcher interface {
	Search(q index.Query) ([]index.Hit, error)
	BookWorks(bookID int64) ([]string, error)
}

// Backend is everything the screens need from the index.
type Backend interface {
	Searcher
	Libraries() ([]index.Library, error)
	TotalWorks() (int, error)
	RemoveLibrary(name string) error
	// AddFolder registers dir as a library and runs its first scan. A
	// returned library with ID != 0 was registered even when err != nil.
	AddFolder(ctx context.Context, dir string, onProgress func(scan.Progress)) (index.Library, scan.Report, error)
	// Rescan updates a registered library.
	Rescan(ctx context.Context, lib index.Library, onProgress func(scan.Progress)) (scan.Report, error)
}

// Actions are the side effects the screens trigger; tests replace them.
type Actions struct {
	Root         func(volumeID, rootRel string) (string, bool)
	Open         func(path string) error
	Reveal       func(path string) error
	Copy         func(text string) error
	Exists       func(path string) bool // does the file exist on the mounted disk
	ChooseFolder func(prompt string) (string, error)
}

const (
	debounce       = 50 * time.Millisecond
	candidateLimit = 500
)

type screen int

const (
	screenSearch screen = iota
	screenLibraries
)

type searchMsg struct{ seq int }

type resultsMsg struct {
	seq    int
	hits   []index.Hit
	online map[string]bool
	err    error
}

type Model struct {
	b      Backend
	act    Actions
	total  int
	screen screen

	// search screen
	input    textinput.Model
	hits     []index.Hit
	cursor   int
	seq      int
	searched bool
	online   map[string]bool    // volume id -> mounted
	toc      map[int64][]string // book id -> work titles (cache)
	err      error

	// libraries screen
	libs          []index.Library
	libOnline     map[string]bool
	libCursor     int
	confirmDelete bool
	job           *job
	bar           progress.Model

	status string
	width  int
	height int
}

// New builds the UI; with an empty index (total == 0) it opens on the
// libraries screen.
func New(b Backend, act Actions, total int) Model {
	in := textinput.New()
	in.Prompt = "🔎 "
	in.Placeholder = "назва твору, автор або збірка…"
	in.Focus()
	m := Model{
		b: b, act: act, total: total, input: in,
		online: map[string]bool{}, toc: map[int64][]string{}, libOnline: map[string]bool{},
		bar: progress.New(progress.WithDefaultBlend(), progress.WithWidth(40)),
	}
	if total == 0 {
		m.screen = screenLibraries
	}
	return m
}

func (m Model) Init() tea.Cmd { return m.loadLibraries() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.SetWidth(max(10, msg.Width-20))
		m.bar.SetWidth(min(60, max(10, msg.Width-20)))
		return m, nil
	case searchMsg:
		if msg.seq != m.seq {
			return m, nil
		}
		return m, m.searchCmd(msg.seq, m.input.Value())
	case resultsMsg:
		if msg.seq != m.seq {
			return m, nil
		}
		m.err = msg.err
		m.hits = rank(m.input.Value(), msg.hits)
		m.cursor = 0
		m.searched = true
		m.online = msg.online
		m.loadTOC()
		return m, nil
	case librariesMsg, folderMsg, progressMsg, jobDoneMsg, removedMsg:
		return m.updateLibraries(msg)
	case tea.KeyPressMsg:
		if m.screen == screenLibraries {
			return m.libraryKey(msg)
		}
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
		case "ctrl+l":
			m.screen = screenLibraries
			m.status = ""
			return m, m.loadLibraries()
		case "up", "ctrl+p":
			m.move(-1)
			return m, nil
		case "down", "ctrl+n":
			m.move(1)
			return m, nil
		case "enter":
			m.withFile(m.act.Open, "")
			return m, nil
		case "ctrl+o":
			m.withFile(m.act.Reveal, "")
			return m, nil
		case "ctrl+y":
			m.withFile(m.act.Copy, "Шлях скопійовано")
			return m, nil
		case "tab":
			if h := m.selected(); h != nil && h.Author != "" {
				m.input.SetValue(h.Author)
				m.input.CursorEnd()
				return m, m.queueSearch()
			}
			return m, nil
		}
	}
	if m.screen != screenSearch {
		return m, nil
	}
	before := m.input.Value()
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != before {
		return m, tea.Batch(cmd, m.queueSearch())
	}
	return m, cmd
}
```

(Імпорт `fmt` і `library` лишаються — їх використовує `withFile`.)

- [ ] **Step 4: Реалізація — `internal/tui/libraries.go`**

```go
package tui

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"findbooks/internal/index"
	"findbooks/internal/libman"
	"findbooks/internal/platform"
	"findbooks/internal/scan"
)

const choosePrompt = "Виберіть теку бібліотеки"

type librariesMsg struct {
	libs   []index.Library
	online map[string]bool // volume id -> mounted
	total  int
	err    error
}

type folderMsg struct {
	path string
	err  error
}

type progressMsg scan.Progress

type jobDoneMsg struct {
	adding bool
	name   string        // display name (folder name for a new library)
	lib    index.Library // the library; ID 0 if an add failed before registering
	rep    scan.Report
	err    error
}

type removedMsg struct {
	name string
	err  error
}

// job is a running add or rescan. ch carries progressMsg values and one
// final jobDoneMsg, then is closed.
type job struct {
	name     string
	cancel   context.CancelFunc
	ch       chan tea.Msg
	progress scan.Progress
}

// keyAliases maps Ukrainian-layout keys to the Latin keys of the libraries screen.
var keyAliases = map[string]string{"ф": "a", "г": "u", "в": "d", "н": "y", "Ф": "a", "Г": "u", "В": "d", "Н": "y", "Y": "y"}

func (m Model) loadLibraries() tea.Cmd {
	b, root := m.b, m.act.Root
	return func() tea.Msg {
		libs, err := b.Libraries()
		if err != nil {
			return librariesMsg{err: err}
		}
		total, err := b.TotalWorks()
		online := map[string]bool{}
		for _, l := range libs {
			if _, seen := online[l.VolumeID]; !seen {
				_, ok := root(l.VolumeID, l.RootRel)
				online[l.VolumeID] = ok
			}
		}
		return librariesMsg{libs: libs, online: online, total: total, err: err}
	}
}

func (m Model) chooseFolder() tea.Cmd {
	choose := m.act.ChooseFolder
	return func() tea.Msg {
		p, err := choose(choosePrompt)
		return folderMsg{path: p, err: err}
	}
}

func (m Model) removeLibrary(name string) tea.Cmd {
	b := m.b
	return func() tea.Msg { return removedMsg{name: name, err: b.RemoveLibrary(name)} }
}

// startJob runs fn in the background. Progress is sent without blocking
// (the bar only needs the latest value); the final message always arrives.
func (m *Model) startJob(name string, fn func(ctx context.Context, onProgress func(scan.Progress)) jobDoneMsg) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan tea.Msg, 16)
	m.job = &job{name: name, cancel: cancel, ch: ch}
	m.status = ""
	go func() {
		done := fn(ctx, func(p scan.Progress) {
			select {
			case ch <- progressMsg(p):
			default:
			}
		})
		ch <- done
		close(ch)
	}()
	return listen(ch)
}

// listen waits for the next message of a running job.
func listen(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func (m Model) selectedLib() *index.Library {
	if m.libCursor < 0 || m.libCursor >= len(m.libs) {
		return nil
	}
	return &m.libs[m.libCursor]
}

func (m Model) updateLibraries(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case librariesMsg:
		if msg.err != nil {
			m.status = "Помилка: " + msg.err.Error()
			return m, nil
		}
		m.libs, m.libOnline, m.total = msg.libs, msg.online, msg.total
		m.libCursor = min(m.libCursor, max(0, len(m.libs)-1))
		return m, nil
	case folderMsg:
		switch {
		case errors.Is(msg.err, platform.ErrCanceled):
			return m, nil
		case errors.Is(msg.err, platform.ErrUnsupported):
			m.status = "Діалог вибору теки недоступний — використайте: findbooks add <шлях>"
			return m, nil
		case msg.err != nil:
			m.status = "Помилка: " + msg.err.Error()
			return m, nil
		}
		dir, name, b := msg.path, filepath.Base(msg.path), m.b
		return m, m.startJob(name, func(ctx context.Context, on func(scan.Progress)) jobDoneMsg {
			lib, rep, err := b.AddFolder(ctx, dir, on)
			return jobDoneMsg{adding: true, name: name, lib: lib, rep: rep, err: err}
		})
	case progressMsg:
		if m.job == nil {
			return m, nil
		}
		m.job.progress = scan.Progress(msg)
		return m, listen(m.job.ch)
	case jobDoneMsg:
		if m.job != nil {
			m.job.cancel()
		}
		m.job = nil
		m.status = jobStatus(msg)
		return m, m.loadLibraries()
	case removedMsg:
		if msg.err != nil {
			m.status = "Помилка: " + msg.err.Error()
		} else {
			m.status = fmt.Sprintf("Бібліотеку «%s» прибрано з індексу", msg.name)
		}
		return m, m.loadLibraries()
	}
	return m, nil
}

func (m Model) libraryKey(k tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := k.String()
	if alias, ok := keyAliases[key]; ok {
		key = alias
	}
	if key == "ctrl+c" {
		if m.job != nil {
			m.job.cancel()
		}
		return m, tea.Quit
	}
	if m.job != nil {
		if key == "esc" {
			m.job.cancel()
			m.status = "Скасовую…"
		}
		return m, nil
	}
	if m.confirmDelete {
		m.confirmDelete = false
		if l := m.selectedLib(); key == "y" && l != nil {
			return m, m.removeLibrary(l.Name)
		}
		m.status = "Скасовано"
		return m, nil
	}
	switch key {
	case "esc":
		if m.total == 0 {
			return m, tea.Quit
		}
		m.screen = screenSearch
		m.status = ""
		if m.input.Value() != "" {
			return m, m.queueSearch()
		}
		return m, nil
	case "up":
		m.libCursor = max(0, m.libCursor-1)
	case "down":
		m.libCursor = max(0, min(len(m.libs)-1, m.libCursor+1))
	case "a":
		m.status = ""
		return m, m.chooseFolder()
	case "u":
		if l := m.selectedLib(); l != nil {
			lib, b := *l, m.b
			return m, m.startJob(lib.Name, func(ctx context.Context, on func(scan.Progress)) jobDoneMsg {
				rep, err := b.Rescan(ctx, lib, on)
				return jobDoneMsg{name: lib.Name, lib: lib, rep: rep, err: err}
			})
		}
	case "d":
		if m.selectedLib() != nil {
			m.confirmDelete = true
			m.status = ""
		}
	}
	return m, nil
}

func jobStatus(d jobDoneMsg) string {
	switch {
	case d.err == nil:
		name := d.lib.Name
		if name == "" {
			name = d.name
		}
		return fmt.Sprintf("«%s»: %s", name, reportLine(d.rep))
	case d.adding && d.lib.ID != 0:
		return fmt.Sprintf("Бібліотеку «%s» зареєстровано, індексацію перервано — натисніть u, щоб продовжити", d.lib.Name)
	case errors.Is(d.err, index.ErrExists):
		return fmt.Sprintf("Бібліотека «%s» уже є", d.name)
	case errors.Is(d.err, libman.ErrOffline):
		return fmt.Sprintf("Диск «%s» не підключено", d.lib.VolumeName)
	case errors.Is(d.err, context.Canceled):
		return fmt.Sprintf("Оновлення «%s» перервано", d.name)
	default:
		return "Помилка: " + d.err.Error()
	}
}

func reportLine(r scan.Report) string {
	return fmt.Sprintf("додано %d, оновлено %d, видалено %d, без змін %d; проблемних файлів: %d",
		r.Added, r.Updated, r.Removed, r.Unchanged, len(r.Errors))
}
```

Примітка: у `TestUpdateOffline` фейк повертає `ErrOffline`, а `d.lib.VolumeName` = `"dsvDev"` з `fantastika` — звідси «Диск «dsvDev» не підключено».

- [ ] **Step 5: Реалізація — `internal/tui/libraries_view.go`**

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

const (
	libHelp      = "a додати · u оновити · d прибрати · ↑/↓ вибір · esc до пошуку"
	libHelpEmpty = "a додати бібліотеку · esc вихід"
)

func (m Model) libMark(volumeID string) string {
	if m.libOnline[volumeID] {
		return styleOnline.Render("●")
	}
	return styleOffline.Render("○")
}

func (m Model) renderLibraries() string {
	w := m.width
	if w <= 0 {
		w = 100
	}
	var b strings.Builder
	b.WriteString(styleTitle.Render("Бібліотеки") + "  " + styleDim.Render(fmt.Sprintf("творів в індексі: %d", m.total)) + "\n\n")
	if len(m.libs) == 0 {
		b.WriteString(styleDim.Render("Бібліотек ще немає. Натисніть a, щоб вибрати теку з книгами.") + "\n")
	}
	for i, l := range m.libs {
		marker, style := "  ", styleTitle
		if i == m.libCursor {
			marker, style = "▸ ", styleSelected
		}
		updated := "—"
		if !l.LastScan.IsZero() {
			updated = l.LastScan.Format("2006-01-02 15:04")
		}
		b.WriteString(marker + style.Render(ansi.Truncate(l.Name, w-2, "…")) + "\n")
		b.WriteString(styleDim.Render("    "+l.VolumeName+" ") + m.libMark(l.VolumeID) +
			styleDim.Render(ansi.Truncate(fmt.Sprintf("  книг: %d · творів: %d · оновлено: %s", l.Books, l.Works, updated), max(10, w-8), "…")) + "\n")
	}
	if m.job != nil {
		p := m.job.progress
		b.WriteString("\n" + fmt.Sprintf("Індексую «%s»…", m.job.name) + "\n")
		if p.Total == 0 {
			b.WriteString(styleDim.Render("шукаю файли…") + "\n")
		} else {
			b.WriteString(m.bar.ViewAs(float64(p.Done)/float64(p.Total)) + fmt.Sprintf(" %d/%d", p.Done, p.Total) + "\n")
		}
		b.WriteString(styleDim.Render("esc — скасувати") + "\n")
	}
	if l := m.selectedLib(); m.confirmDelete && l != nil {
		b.WriteString("\n" + styleStatus.Render(fmt.Sprintf("Прибрати «%s» з індексу? Файли не чіпаються. (y/n)", l.Name)) + "\n")
	}
	if m.status != "" {
		b.WriteString("\n" + styleStatus.Render(m.status) + "\n")
	}
	help := libHelp
	if m.total == 0 {
		help = libHelpEmpty
	}
	b.WriteString("\n" + styleDim.Render(ansi.Truncate(help, w, "…")))
	return b.String()
}
```

- [ ] **Step 6: Реалізація — `internal/tui/view.go`**

- `helpLine` замінити на:
```go
const helpLine = "enter відкрити · ctrl+l бібліотеки · ctrl+o показати у Finder · ctrl+y копіювати шлях · tab за автором · ↑/↓ вибір · esc вихід"
```
- На початку `render()` додати:
```go
	if m.screen == screenLibraries {
		return m.renderLibraries()
	}
```

- [ ] **Step 7: Реалізація — `internal/tui/run.go`**

```go
package tui

import (
	"context"
	"os"

	"github.com/atotto/clipboard"

	tea "charm.land/bubbletea/v2"

	"findbooks/internal/index"
	"findbooks/internal/libman"
	"findbooks/internal/library"
	"findbooks/internal/platform"
	"findbooks/internal/scan"
)

// DefaultActions wires the screens to the real OS.
func DefaultActions() Actions {
	return Actions{
		Root: library.Root, Open: platform.Open, Reveal: platform.Reveal, Copy: clipboard.WriteAll,
		Exists:       func(p string) bool { _, err := os.Stat(p); return err == nil },
		ChooseFolder: platform.ChooseFolder,
	}
}

// storeBackend adds library management (via libman) to the index store.
type storeBackend struct{ *index.Store }

func (s storeBackend) AddFolder(ctx context.Context, dir string, onProgress func(scan.Progress)) (index.Library, scan.Report, error) {
	lib, _, err := libman.Register(s.Store, dir, "")
	if err != nil {
		return index.Library{}, scan.Report{}, err
	}
	rep, err := libman.Scan(ctx, s.Store, lib, onProgress)
	return lib, rep, err
}

func (s storeBackend) Rescan(ctx context.Context, lib index.Library, onProgress func(scan.Progress)) (scan.Report, error) {
	return libman.Scan(ctx, s.Store, lib, onProgress)
}

// Run shows the UI until the user quits; an empty index opens on the
// libraries screen.
func Run(st *index.Store, total int) error {
	_, err := tea.NewProgram(New(storeBackend{st}, DefaultActions(), total)).Run()
	return err
}
```

Примітка: `tui.Run` тепер приймає `*index.Store` замість `Searcher`. У `internal/cli/root.go` виклик `tui.Run(st, total)` уже передає `*index.Store` — компілюється без змін.

- [ ] **Step 8: Тести проходять**

Run: `go test -race ./internal/tui/ ./internal/cli/ && go vet ./... && gofmt -l internal`
Expected: `ok`; vet і gofmt без виводу. Якщо файл `model.go` після змін перевищує ~260 рядків — це очікувано (маршрутизація двох екранів); логіку бібліотек не переносити назад у `model.go`.

- [ ] **Step 9: Commit**

```bash
git add internal/tui
git commit -m "feat(tui): libraries screen with Finder chooser, progress, update and remove

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 4: Порожній індекс відкриває TUI; README

**Files:**
- Modify: `internal/cli/root.go`, `internal/cli/cli_test.go`, `README.md`

**Interfaces:**
- Consumes: `runTUI(st *index.Store, total int) error`, `tui.Run(st *index.Store, total int) error`
- Produces: `findbooks` без аргументів завжди відкриває TUI (з `total == 0` — на екрані «Бібліотеки»); `errNoLibraries` лишається для `search`, `update`, `list`

- [ ] **Step 1: Тест — замінити `TestRootCommandStartsTUI` у `internal/cli/cli_test.go`**

```go
func TestRootCommandStartsTUI(t *testing.T) {
	db := filepath.Join(t.TempDir(), "index.db")
	var calls []int
	orig := runTUI
	runTUI = func(_ *index.Store, total int) error { calls = append(calls, total); return nil }
	t.Cleanup(func() { runTUI = orig })

	if _, _, err := run(t, db); err != nil {
		t.Fatalf("empty index: err = %v, want TUI to start", err)
	}
	st, err := index.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	lib, _ := st.AddLibrary("A", "VOL", "Disk", ".")
	st.WriteBatch(lib.ID, []index.BookRecord{{RelPath: "a.pdf", Format: "pdf", Title: "A", Works: []index.WorkRecord{{Title: "A", TreePath: "A"}}}}, nil)
	st.Close()
	if _, _, err := run(t, db); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(calls, []int{0, 1}) {
		t.Fatalf("runTUI totals = %v, want [0 1]", calls)
	}
}
```
(Додати `"slices"` до імпортів `cli_test.go`, якщо його немає.)

- [ ] **Step 2: Перевірити падіння**

Run: `go test ./internal/cli/ -run TestRootCommandStartsTUI`
Expected: FAIL — `empty index: err = індекс порожній …`.

- [ ] **Step 3: Реалізація — `internal/cli/root.go`**

У `RunE` root-команди прибрати перевірку
```go
			if total == 0 {
				return errNoLibraries
			}
```
щоб лишилось `return runTUI(st, total)`. `errNoLibraries` та інші місця його використання не чіпати.

- [ ] **Step 4: README**

У `README.md` у розділі «Використання» замінити рядок-коментар до `findbooks` на `# інтерактивний пошук і керування бібліотеками`, а після абзацу про клавіші TUI додати:

```markdown
### Бібліотеки в TUI

Якщо індекс порожній, `findbooks` відкриває екран «Бібліотеки»; з пошуку до нього веде `ctrl+l`.

- `a` — вибрати теку в діалозі Finder і проіндексувати її (прогрес видно на екрані, `esc` — скасувати);
- `u` — оновити вибрану бібліотеку;
- `d` — прибрати бібліотеку з індексу (файли не чіпаються; підтвердження `y`);
- `esc` — назад до пошуку.

Клавіші працюють і в українській розкладці (`ф`, `г`, `в`, `н`).
```

- [ ] **Step 5: Повна перевірка**

Run:
```bash
go test -race ./... && go vet ./... && gofmt -l internal cmd \
  && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./... \
  && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./... \
  && CGO_ENABLED=0 go build -o findbooks ./cmd/findbooks
```
Expected: усі пакети `ok`, решта без виводу/помилок.

- [ ] **Step 6: Commit**

```bash
git add internal/cli/root.go internal/cli/cli_test.go README.md
git commit -m "feat(cli): open the TUI libraries screen on an empty index; document it

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

- [ ] **Step 7: Ручна перевірка (користувач)**

Попросити користувача: `./findbooks` (на порожньому індексі) → `a` → діалог Finder → вибрати `/Volumes/dsvDev/Бібліотека/Советская фантастика` → прогрес → бібліотека у списку → `esc` → пошук «чужие дети»; також перевірити `ctrl+l`, `u`, `d` + `n`.
