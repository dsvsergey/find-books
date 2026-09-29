# findbooks Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** CLI/TUI на Go, що індексує бібліотеки FB2 (і не-FB2 за ім'ям файлу) на зовнішніх дисках і миттєво знаходить твір усередині збірок — навіть коли диск не підключено.

**Architecture:** Один глобальний SQLite-індекс (FTS5, `modernc.org/sqlite`, без cgo). Сканер обходить теку бібліотеки, потоково парсить FB2, евристикою виділяє твори з дерева секцій і пише їх в індекс інкрементально (розмір+mtime). Бібліотека прив'язана до тому (volume UUID) і відносного шляху; усі ОС-залежні виклики — лише в `internal/platform`. Cobra-команди для індексації/пошуку, Bubble Tea v2 TUI для інтерактивного пошуку.

**Tech Stack:** Go 1.27, `github.com/spf13/cobra` v1.10.2, `modernc.org/sqlite` v1.60.0, `charm.land/bubbletea/v2` v2.0.10, `charm.land/bubbles/v2` v2.2.1, `charm.land/lipgloss/v2` v2.0.6, `github.com/charmbracelet/x/ansi`, `github.com/sahilm/fuzzy` v0.1.3, `github.com/atotto/clipboard` v0.1.4, `golang.org/x/text` v0.42.0, `golang.org/x/sys`, `golang.org/x/term`.

**Spec:** `docs/superpowers/specs/2026-09-29-findbooks-design.md`

## Global Constraints

- Модуль `findbooks` (`go mod init findbooks`), Go 1.27; збірка має проходити з `CGO_ENABLED=0`.
- ОС-специфічні виклики (`diskutil`, `open`, `unix.*`) і build-теги — **лише** в `internal/platform`. Windows/Linux у v1 — заглушки, що повертають `platform.ErrUnsupported`; `GOOS=windows go build ./...` і `GOOS=linux go build ./...` мають компілюватися.
- Відносні шляхи в БД (`root_rel`, `rel_path`) — завжди з `/` (`filepath.ToSlash`); до ФС — через `filepath.FromSlash`.
- Одна функція нормалізації `textnorm.Normalize` для індексації і для запиту: NFC, нижній регістр, `ё→е`, прибрати комбіновані діакритики, все, що не літера/цифра → один пробіл.
- FTS5: `tokenize = 'unicode61 remove_diacritics 0'`, кожне слово запиту — префікс у лапках (`"слово"*`), слова через AND.
- Індекс за замовчуванням: `platform.DataDir()/index.db` (`~/.local/share/findbooks/index.db` на macOS); глобальний прапорець `--db` перевизначає шлях.
- Тексти для користувача — українською.
- Критерій успіху: `findbooks search чужие дети` на реальній бібліотеці повертає «Румбы фантастики. 1988 год. Том II → Игорь Пидоренко › ЧУЖИЕ ДЕТИ»; сам запит до індексу < 100 мс.
- Кожен коміт закінчується рядком `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Уточнення специфікації (прийняті в цьому плані)

1. **Гарячі клавіші TUI.** У специфікації `f`/`c`/`a` — але це звичайні літери, які користувач друкує в полі пошуку. Замість них: `ctrl+o` — показати у Finder, `ctrl+y` — копіювати шлях, `tab` — шукати за автором. `enter`, `esc`, `↑/↓` — як у специфікації.
2. **Нормалізація** винесена в окремий пакет `internal/textnorm`, бо її використовують `index`, `extract` і `tui` (специфікація клала її в `index`).
3. **`works_fts`** — звичайна FTS5-таблиця з уже нормалізованим текстом, `rowid = works.id`; рядки прибирає тригер `AFTER DELETE ON works` (каскад з `books`/`libraries`). Це простіше за external content і дає той самий результат.
4. **Колонка `book_title`** у FTS містить назву книги **і шлях до теки** файлу — так PDF/DJVU у теках на кшталт `Авторы/Беляев Александр/` знаходяться за автором.
5. **Евристика творів:** секція, чий заголовок дорівнює назві книги, прозора (як шумова) — інакше обгортка «Румбы фантастики» ховала б усю збірку. Шаблон «Глава/Часть/Книга/Том N» вимагає номер або порядкове слово, щоб «Книга мертвых» не вважалась шумом.

## Review Focus

1. **Запит із синтаксисом FTS5** (`"`, `*`, `NEAR(`, `title:x`, `-дети`, `AND`, `')--`) — має трактуватися як звичайний текст, без помилки SQL. Тест: Task 5 `TestSearchHostileInput`.
2. **NFD-імена файлів і наголоси** (`и+◌̆` замість `й` у файлах, скопійованих з інших систем; `за́мок`) — мають знаходитися звичайним набором. Тести: Task 2 (таблиця), Task 7 `TestRunIndexesLibrary` (файл з NFD-іменем).
3. **Диск бібліотеки відключений** під час `update`, `search`, `enter` у TUI — пропуск з попередженням, позначка `○`, підказка «підключіть диск», жодної помилки. Тести: Task 8 `TestEnterOfflineShowsHint`, Task 9 `TestOfflineLibrary`.
4. **Биті/величезні FB2** (обрізаний XML, сміття, мегабайти base64-обкладинок) — без падіння, файл індексується за ім'ям і потрапляє у звіт. Тести: Task 3 `TestParseTruncated`/`TestParseGarbage`/`TestParseSkipsLargeBinary`, Task 7 (broken.fb2).
5. **Бібліотека на системному диску** (`~/Books`, `/private/var/...` — на macOS це firmlink на том Data, змонтований у `/System/Volumes/Data`) — `add` має працювати, шлях має відновлюватися. Тест: Task 6 `TestLocateAndRootRoundTrip` (TempDir лежить саме там).

---

## Структура файлів

```
cmd/findbooks/main.go              точка входу: os.Exit(cli.Execute())
internal/platform/platform.go      ErrUnsupported, DataDir (усі ОС)
internal/platform/platform_darwin.go  VolumeID, MountPoint, Open, Reveal (macOS)
internal/platform/platform_windows.go заглушки
internal/platform/platform_linux.go   заглушки
internal/textnorm/textnorm.go      Normalize
internal/fb2/fb2.go                типи Book/Author/Section, Parse
internal/fb2/file.go               ParseFile (.fb2 і .fb2.zip)
internal/fb2/testdata/rumby.fb2    фікстура-збірка (використовують fb2, scan, cli тести)
internal/fb2/testdata/novel.fb2    фікстура-роман
internal/extract/extract.go        Works — евристика творів
internal/index/schema.go           SQL-схема
internal/index/store.go            Store, Open, бібліотеки
internal/index/books.go            Files, WriteBatch, MarkScanned
internal/index/search.go           MatchExpr, Search, BookWorks, TotalWorks
internal/library/library.go        Locate, Root, FilePath
internal/scan/scan.go              Run — інкрементальне сканування
internal/tui/model.go              Model, Update
internal/tui/rank.go               fuzzy-ранжування кандидатів
internal/tui/view.go               View
internal/tui/run.go                Run, DefaultActions
internal/cli/root.go               root-команда, Execute, --db
internal/cli/progress.go           прогрес-бар і звіт сканування
internal/cli/add.go update.go list.go remove.go search.go
```

---

### Task 1: Каркас модуля і пакет platform

**Files:**
- Create: `go.mod`, `.gitignore`, `internal/platform/platform.go`, `internal/platform/platform_darwin.go`, `internal/platform/platform_windows.go`, `internal/platform/platform_linux.go`
- Test: `internal/platform/platform_test.go`, `internal/platform/platform_darwin_test.go`

**Interfaces:**
- Consumes: —
- Produces:
  - `var platform.ErrUnsupported error`
  - `func platform.VolumeID(path string) (id, name string, err error)`
  - `func platform.MountPoint(volumeID string) (string, bool)`
  - `func platform.Open(path string) error`
  - `func platform.Reveal(path string) error`
  - `func platform.DataDir() (string, error)` — створює теку

- [ ] **Step 1: Ініціалізувати модуль**

```bash
cd /Users/dsv/Projects/find-books
go mod init findbooks
printf '/findbooks\n*.db\n*.db-*\n' > .gitignore
```

- [ ] **Step 2: Написати тести, що падають**

`internal/platform/platform_test.go`:
```go
package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDataDirHonorsXDG(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG_DATA_HOME is not used on Windows")
	}
	base := t.TempDir()
	t.Setenv("XDG_DATA_HOME", base)
	dir, err := DataDir()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(base, "findbooks"); dir != want {
		t.Fatalf("DataDir() = %q, want %q", dir, want)
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		t.Fatalf("DataDir did not create %s: %v", dir, err)
	}
}
```

`internal/platform/platform_darwin_test.go`:
```go
package platform

import (
	"path/filepath"
	"testing"
)

func TestVolumeIDOfRoot(t *testing.T) {
	id, name, err := VolumeID("/")
	if err != nil {
		t.Fatal(err)
	}
	if id == "" || name == "" {
		t.Fatalf("empty id=%q name=%q", id, name)
	}
}

func TestMountPointFindsVolumeOfTempDir(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	id, _, err := VolumeID(dir)
	if err != nil {
		t.Fatal(err)
	}
	mp, ok := MountPoint(id)
	if !ok {
		t.Fatalf("volume %s not found among mounts", id)
	}
	want, err := mountOf(dir)
	if err != nil {
		t.Fatal(err)
	}
	if mp != want {
		t.Fatalf("MountPoint = %q, statfs says %q", mp, want)
	}
}

func TestMountPointUnknownVolume(t *testing.T) {
	if mp, ok := MountPoint("00000000-0000-0000-0000-000000000000"); ok {
		t.Fatalf("unexpected mount point %q", mp)
	}
}

func TestPlistStringsTopLevelOnly(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>VolumeName</key><string>dsvDev</string>
<key>Nested</key><dict><key>VolumeUUID</key><string>WRONG</string></dict>
<key>Writable</key><true/>
<key>VolumeUUID</key><string>BB26BFDB-2128</string>
</dict></plist>`)
	got, err := plistStrings(data)
	if err != nil {
		t.Fatal(err)
	}
	if got["VolumeName"] != "dsvDev" || got["VolumeUUID"] != "BB26BFDB-2128" {
		t.Fatalf("got %v", got)
	}
}
```

- [ ] **Step 3: Переконатися, що тести падають**

Run: `go test ./internal/platform/`
Expected: FAIL — `undefined: DataDir`, `undefined: VolumeID` тощо.

- [ ] **Step 4: Реалізація**

`internal/platform/platform.go`:
```go
// Package platform isolates every OS-specific call used by findbooks.
// Nothing outside this package may shell out to OS tools or use build tags.
package platform

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// ErrUnsupported is returned by functions not yet implemented for this OS.
var ErrUnsupported = errors.New("platform: not supported on this OS yet")

// DataDir returns (and creates) the directory that holds index.db:
// %LOCALAPPDATA%\findbooks on Windows, $XDG_DATA_HOME/findbooks or
// ~/.local/share/findbooks elsewhere.
func DataDir() (string, error) {
	var base string
	if runtime.GOOS == "windows" {
		base = os.Getenv("LOCALAPPDATA")
		if base == "" {
			return "", errors.New("LOCALAPPDATA is not set")
		}
	} else {
		base = os.Getenv("XDG_DATA_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			base = filepath.Join(home, ".local", "share")
		}
	}
	dir := filepath.Join(base, "findbooks")
	return dir, os.MkdirAll(dir, 0o755)
}
```

`internal/platform/platform_darwin.go`:
```go
package platform

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// VolumeID returns the volume UUID and display name of the volume that
// contains path. Volumes without a UUID fall back to "mnt:<mount point>".
func VolumeID(path string) (id, name string, err error) {
	mnt, err := mountOf(path)
	if err != nil {
		return "", "", err
	}
	info, err := diskutilInfo(mnt)
	if err != nil {
		return "", "", err
	}
	id, name = info["VolumeUUID"], info["VolumeName"]
	if id == "" {
		id = "mnt:" + mnt
	}
	if name == "" {
		name = mnt
	}
	return id, name, nil
}

var (
	mountMu  sync.Mutex
	mounts   map[string]string // volume UUID -> mount point
	mountsAt time.Time
)

// mountRescanEvery bounds how often a miss triggers a full diskutil scan,
// so offline volumes in a result list do not cost a scan per lookup.
const mountRescanEvery = 5 * time.Second

// MountPoint reports where the volume is mounted right now.
func MountPoint(volumeID string) (string, bool) {
	if p, ok := strings.CutPrefix(volumeID, "mnt:"); ok {
		m, err := mountOf(p)
		return p, err == nil && m == p
	}
	mountMu.Lock()
	defer mountMu.Unlock()
	if mp, ok := mounts[volumeID]; ok {
		if m, err := mountOf(mp); err == nil && m == mp {
			return mp, true
		}
	}
	if mounts == nil || time.Since(mountsAt) > mountRescanEvery {
		mounts, mountsAt = scanMounts(), time.Now()
	}
	mp, ok := mounts[volumeID]
	return mp, ok
}

// Open opens path in its default application.
func Open(path string) error { return exec.Command("open", path).Run() }

// Reveal selects path in Finder.
func Reveal(path string) error { return exec.Command("open", "-R", path).Run() }

func mountOf(path string) (string, error) {
	var st unix.Statfs_t
	if err := unix.Statfs(path, &st); err != nil {
		return "", fmt.Errorf("statfs %s: %w", path, err)
	}
	return unix.ByteSliceToString(st.Mntonname[:]), nil
}

func scanMounts() map[string]string {
	res := map[string]string{}
	n, err := unix.Getfsstat(nil, unix.MNT_NOWAIT)
	if err != nil {
		return res
	}
	buf := make([]unix.Statfs_t, n)
	n, err = unix.Getfsstat(buf, unix.MNT_NOWAIT)
	if err != nil {
		return res
	}
	for _, st := range buf[:n] {
		if !strings.HasPrefix(unix.ByteSliceToString(st.Mntfromname[:]), "/dev/") {
			continue
		}
		mnt := unix.ByteSliceToString(st.Mntonname[:])
		info, err := diskutilInfo(mnt)
		if err != nil || info["VolumeUUID"] == "" {
			continue
		}
		res[info["VolumeUUID"]] = mnt
	}
	return res
}

func diskutilInfo(mnt string) (map[string]string, error) {
	out, err := exec.Command("diskutil", "info", "-plist", mnt).Output()
	if err != nil {
		return nil, fmt.Errorf("diskutil info %s: %w", mnt, err)
	}
	return plistStrings(out)
}

// plistStrings extracts the <key>/<string> pairs of a plist's top-level dict.
func plistStrings(data []byte) (map[string]string, error) {
	d := xml.NewDecoder(bytes.NewReader(data))
	res := map[string]string{}
	depth := 0
	key := ""
	for {
		tok, err := d.Token()
		if err == io.EOF {
			return res, nil
		}
		if err != nil {
			return nil, fmt.Errorf("plist: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if depth < 2 { // <plist> and the top-level <dict>
				depth++
				continue
			}
			switch t.Name.Local {
			case "key", "string":
				var s string
				if err := d.DecodeElement(&s, &t); err != nil {
					return nil, fmt.Errorf("plist: %w", err)
				}
				if t.Name.Local == "key" {
					key = s
				} else if key != "" {
					res[key], key = s, ""
				}
			default:
				key = ""
				if err := d.Skip(); err != nil {
					return nil, fmt.Errorf("plist: %w", err)
				}
			}
		case xml.EndElement:
			depth--
		}
	}
}
```

`internal/platform/platform_windows.go`:
```go
package platform

// Planned: volume serial via GetVolumeInformationW, drive-letter scan,
// `cmd /c start ""`, `explorer /select,`.

func VolumeID(path string) (id, name string, err error) { return "", "", ErrUnsupported }
func MountPoint(volumeID string) (string, bool)         { return "", false }
func Open(path string) error                            { return ErrUnsupported }
func Reveal(path string) error                          { return ErrUnsupported }
```

`internal/platform/platform_linux.go`:
```go
package platform

// Planned: UUID from /dev/disk/by-uuid, /proc/self/mounts, xdg-open.

func VolumeID(path string) (id, name string, err error) { return "", "", ErrUnsupported }
func MountPoint(volumeID string) (string, bool)         { return "", false }
func Open(path string) error                            { return ErrUnsupported }
func Reveal(path string) error                          { return ErrUnsupported }
```

Потім:
```bash
go get golang.org/x/sys@latest
go mod tidy
```

- [ ] **Step 5: Тести і крос-компіляція**

Run: `go test ./internal/platform/ && GOOS=windows GOARCH=amd64 go vet ./... && GOOS=linux GOARCH=amd64 go vet ./...`
Expected: `ok  findbooks/internal/platform`, vet без виводу.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum .gitignore internal/platform
git commit -m "feat(platform): volume lookup, open/reveal and data dir for macOS

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 2: Нормалізація тексту

**Files:**
- Create: `internal/textnorm/textnorm.go`
- Test: `internal/textnorm/textnorm_test.go`

**Interfaces:**
- Consumes: —
- Produces: `func textnorm.Normalize(s string) string`

- [ ] **Step 1: Тест**

`internal/textnorm/textnorm_test.go`:
```go
package textnorm

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct{ in, want string }{
		{"ЧУЖИЕ ДЕТИ", "чужие дети"},
		{"  Ёлка  ", "елка"},
		{"«Румбы фантастики»", "румбы фантастики"},
		{"Операция «ПРОГРЕССОР»", "операция прогрессор"},
		{"…ТОЖЕ РЕЗУЛЬТАТ", "тоже результат"},
		{"13-Й ПОДВИГ ГЕРАКЛА", "13 й подвиг геракла"},
		{"за\u0301мок", "замок"},             // stress accent
		{"Бои\u0306цов", "бойцов"},            // NFD й from macOS file names
		{"Її ґанок", "її ґанок"},              // Ukrainian letters survive
		{"* * *", ""},
		{"", ""},
		{"Tom's  Diner", "tom s diner"},
	}
	for _, tt := range tests {
		if got := Normalize(tt.in); got != tt.want {
			t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
```

- [ ] **Step 2: Перевірити падіння**

Run: `go test ./internal/textnorm/`
Expected: FAIL — `undefined: Normalize`.

- [ ] **Step 3: Реалізація**

`internal/textnorm/textnorm.go`:
```go
// Package textnorm folds titles and queries into one comparable form.
package textnorm

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Normalize composes s to NFC, lower-cases it, folds ё into е, drops
// combining marks (stress accents), and turns every run of characters that
// are not letters or digits into a single space, trimming both ends.
func Normalize(s string) string {
	s = norm.NFC.String(s)
	var b strings.Builder
	b.Grow(len(s))
	pendingSpace := false
	for _, r := range s {
		switch {
		case unicode.Is(unicode.Mn, r):
			continue
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if pendingSpace && b.Len() > 0 {
				b.WriteByte(' ')
			}
			pendingSpace = false
			r = unicode.ToLower(r)
			if r == 'ё' {
				r = 'е'
			}
			b.WriteRune(r)
		default:
			pendingSpace = true
		}
	}
	return b.String()
}
```

```bash
go get golang.org/x/text@v0.42.0
```

- [ ] **Step 4: Тести проходять**

Run: `go test ./internal/textnorm/`
Expected: `ok`.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum internal/textnorm
git commit -m "feat(textnorm): normalize titles and queries

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 3: Парсер FB2

**Files:**
- Create: `internal/fb2/fb2.go`, `internal/fb2/file.go`, `internal/fb2/testdata/rumby.fb2`, `internal/fb2/testdata/novel.fb2`
- Test: `internal/fb2/fb2_test.go`

**Interfaces:**
- Consumes: —
- Produces:
  - `type fb2.Author struct{ First, Middle, Last, Nick string }`, `func (a Author) Name() string` — «Ім'я Прізвище», інакше нікнейм
  - `type fb2.Section struct{ Title string; Children []*Section }`
  - `type fb2.Book struct{ Title string; Authors []Author; Year string; Sections []*Section }`, `func (b *Book) AuthorNames() string` — імена через `", "`
  - `func fb2.Parse(r io.Reader) (*Book, error)`
  - `func fb2.ParseFile(path string) (*Book, error)` — `.fb2` і `.fb2.zip`
  - Фікстури `internal/fb2/testdata/rumby.fb2`, `internal/fb2/testdata/novel.fb2` (Task 7 і 9 читають їх як `../fb2/testdata/...`)

- [ ] **Step 1: Фікстури**

`internal/fb2/testdata/rumby.fb2`:
```xml
<?xml version="1.0" encoding="utf-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0" xmlns:l="http://www.w3.org/1999/xlink">
 <description>
  <title-info>
   <genre>sf</genre>
   <author><first-name> Евгений</first-name><middle-name>Валентинович</middle-name><last-name> Носов</last-name></author>
   <author><first-name> Игорь</first-name><last-name> Пидоренко</last-name></author>
   <author><first-name>Олег</first-name><middle-name>Игоревич</middle-name><last-name>Чарушников</last-name></author>
   <book-title>Румбы фантастики. 1988 год. Том II</book-title>
   <date value="1988-01-01">1988</date>
   <lang>ru</lang>
  </title-info>
  <document-info>
   <author><nickname>Tanja45</nickname></author>
   <id>5F832003</id>
  </document-info>
 </description>
 <body>
  <title><p>Румбы фантастики. 1988 год. Том II</p></title>
  <section>
   <title><p>Евгений Носов</p></title>
   <section>
    <title><p>ЗЕМЛЕЙ РОЖДЕННЫЕ</p></title>
    <p>Текст.</p>
   </section>
  </section>
  <section>
   <title><p>Игорь Пидоренко</p></title>
   <section>
    <title><p>ЧУЖИЕ ДЕТИ</p></title>
    <section><title><p>1</p></title><p>Текст.</p></section>
    <section><title><p>2</p></title><p>Текст.</p></section>
   </section>
  </section>
  <section>
   <title><p>Олег Чарушников</p></title>
   <section>
    <title><p>НА «ОЛИМПЕ» ВСЕ СПОКОЙНО</p></title>
    <section><title><p>История первая</p><p>ТРУД СИЗИФА</p></title><p>Текст.</p></section>
   </section>
  </section>
  <section>
   <title><p>ОБ АВТОРАХ ЭТОГО СБОРНИКА</p></title>
   <p>Текст.</p>
  </section>
 </body>
 <body name="notes">
  <title><p>Примечания</p></title>
  <section id="n_1"><title><p>1</p></title><p>Сноска.</p></section>
 </body>
 <binary id="cover.jpg" content-type="image/jpeg">/9j/4AAQSkZJRgABAQ==</binary>
</FictionBook>
```

`internal/fb2/testdata/novel.fb2`:
```xml
<?xml version="1.0" encoding="utf-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0">
 <description>
  <title-info>
   <author><first-name>Александр</first-name><last-name>Беляев</last-name></author>
   <book-title>Человек-амфибия</book-title>
  </title-info>
  <publish-info><year>1928</year></publish-info>
 </description>
 <body>
  <section>
   <title><p>Часть первая</p></title>
   <section><title><p>Глава 1</p><p>«Морской дьявол»</p></title><p>Текст&nbsp;главы.</p></section>
   <section><title><p>Глава 2</p><p>Доктор Сальватор</p></title><p>Текст.</p></section>
  </section>
  <section>
   <title><p>Часть вторая</p></title>
   <section><title><p>Глава 1</p><p>Ихтиандр</p></title><p>Текст.</p></section>
  </section>
 </body>
</FictionBook>
```

- [ ] **Step 2: Тести**

`internal/fb2/fb2_test.go`:
```go
package fb2

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/text/encoding/charmap"
)

func titles(secs []*Section) []string {
	var out []string
	for _, s := range secs {
		out = append(out, s.Title)
	}
	return out
}

func TestParseCollection(t *testing.T) {
	b, err := ParseFile("testdata/rumby.fb2")
	if err != nil {
		t.Fatal(err)
	}
	if b.Title != "Румбы фантастики. 1988 год. Том II" {
		t.Errorf("Title = %q", b.Title)
	}
	if b.Year != "1988" {
		t.Errorf("Year = %q", b.Year)
	}
	if got, want := b.AuthorNames(), "Евгений Носов, Игорь Пидоренко, Олег Чарушников"; got != want {
		t.Errorf("AuthorNames = %q, want %q", got, want)
	}
	wantTop := []string{"Евгений Носов", "Игорь Пидоренко", "Олег Чарушников", "ОБ АВТОРАХ ЭТОГО СБОРНИКА"}
	if got := titles(b.Sections); !reflect.DeepEqual(got, wantTop) {
		t.Fatalf("top sections = %q, want %q (notes body must be skipped)", got, wantTop)
	}
	chuzhie := b.Sections[1].Children[0]
	if chuzhie.Title != "ЧУЖИЕ ДЕТИ" || !reflect.DeepEqual(titles(chuzhie.Children), []string{"1", "2"}) {
		t.Errorf("ЧУЖИЕ ДЕТИ subtree = %q %q", chuzhie.Title, titles(chuzhie.Children))
	}
	if got := b.Sections[2].Children[0].Children[0].Title; got != "История первая ТРУД СИЗИФА" {
		t.Errorf("multi-paragraph title = %q", got)
	}
}

func TestParseNovelYearFromPublishInfo(t *testing.T) {
	b, err := ParseFile("testdata/novel.fb2")
	if err != nil {
		t.Fatal(err)
	}
	if b.Title != "Человек-амфибия" || b.Year != "1928" {
		t.Errorf("Title=%q Year=%q", b.Title, b.Year)
	}
	if got := b.Sections[0].Children[0].Title; got != "Глава 1 «Морской дьявол»" {
		t.Errorf("chapter title = %q", got)
	}
}

func TestParseLegacyEncodings(t *testing.T) {
	src, err := os.ReadFile("testdata/novel.fb2")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		label string
		enc   *charmap.Charmap
	}{{"windows-1251", charmap.Windows1251}, {"koi8-r", charmap.KOI8R}} {
		doc := strings.Replace(string(src), `encoding="utf-8"`, `encoding="`+tc.label+`"`, 1)
		doc = strings.NewReplacer("«", `"`, "»", `"`).Replace(doc) // KOI8-R has no guillemets
		encoded, err := tc.enc.NewEncoder().String(doc)
		if err != nil {
			t.Fatal(err)
		}
		b, err := Parse(strings.NewReader(encoded))
		if err != nil {
			t.Fatalf("%s: %v", tc.label, err)
		}
		if b.Title != "Человек-амфибия" {
			t.Errorf("%s: Title = %q", tc.label, b.Title)
		}
	}
}

func TestParseUnknownEncoding(t *testing.T) {
	_, err := Parse(strings.NewReader(`<?xml version="1.0" encoding="x-klingon"?><FictionBook/>`))
	if err == nil {
		t.Fatal("expected error for unknown encoding")
	}
}

func TestParseGarbage(t *testing.T) {
	if _, err := Parse(strings.NewReader("not xml at all")); err == nil {
		t.Fatal("expected error for non-XML input")
	}
}

func TestParseTruncated(t *testing.T) {
	src, err := os.ReadFile("testdata/rumby.fb2")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(bytes.NewReader(src[:len(src)/2])); err == nil {
		t.Fatal("expected error for truncated XML")
	}
}

func TestParseSkipsLargeBinary(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="utf-8"?><FictionBook><description><title-info>` +
		`<book-title>Big</book-title></title-info></description>` +
		`<body><section><title><p>Рассказ</p></title></section></body>` +
		`<binary id="c" content-type="image/jpeg">`)
	buf.WriteString(strings.Repeat("QUFB", 2<<20)) // 8 MiB of base64
	buf.WriteString(`</binary></FictionBook>`)
	b, err := Parse(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if b.Title != "Big" || len(b.Sections) != 1 || b.Sections[0].Title != "Рассказ" {
		t.Fatalf("got %+v", b)
	}
}

func TestParseFileZip(t *testing.T) {
	src, err := os.ReadFile("testdata/rumby.fb2")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	withBook := filepath.Join(dir, "rumby.fb2.zip")
	writeZip(t, withBook, "Румбы.fb2", src)
	b, err := ParseFile(withBook)
	if err != nil {
		t.Fatal(err)
	}
	if b.Title != "Румбы фантастики. 1988 год. Том II" {
		t.Errorf("Title = %q", b.Title)
	}
	empty := filepath.Join(dir, "empty.fb2.zip")
	writeZip(t, empty, "readme.txt", []byte("hi"))
	if _, err := ParseFile(empty); err == nil {
		t.Fatal("expected error for zip without .fb2")
	}
}

func writeZip(t *testing.T, path, name string, data []byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestAuthorNameFallsBackToNick(t *testing.T) {
	if got := (Author{Nick: "Tanja45"}).Name(); got != "Tanja45" {
		t.Errorf("Name() = %q", got)
	}
	if got := (Author{First: "Кир", Middle: "Булычёв", Last: "Булычев"}).Name(); got != "Кир Булычев" {
		t.Errorf("Name() = %q", got)
	}
}
```

- [ ] **Step 3: Перевірити падіння**

Run: `go test ./internal/fb2/`
Expected: FAIL — `undefined: ParseFile`, `undefined: Section` тощо.

- [ ] **Step 4: Реалізація**

`internal/fb2/fb2.go`:
```go
// Package fb2 streams FictionBook 2 files and returns their metadata and the
// tree of section titles of the main body. Section text is never kept.
package fb2

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/htmlindex"
)

type Author struct{ First, Middle, Last, Nick string }

// Name returns "First Last", falling back to the nickname.
func (a Author) Name() string {
	parts := make([]string, 0, 2)
	for _, p := range []string{a.First, a.Last} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) == 0 {
		return a.Nick
	}
	return strings.Join(parts, " ")
}

// Section is a <section> of the main body. Title is the text of its <title>
// with paragraphs joined by spaces and whitespace collapsed ("" if none).
type Section struct {
	Title    string
	Children []*Section
}

type Book struct {
	Title    string
	Authors  []Author
	Year     string
	Sections []*Section
}

// AuthorNames joins the authors' names with ", ".
func (b *Book) AuthorNames() string {
	names := make([]string, 0, len(b.Authors))
	for _, a := range b.Authors {
		if n := a.Name(); n != "" {
			names = append(names, n)
		}
	}
	return strings.Join(names, ", ")
}

var yearRe = regexp.MustCompile(`\b(1[5-9]\d\d|20\d\d)\b`)

// Parse reads one FB2 document. Bodies with a name attribute (notes,
// comments) and <binary> payloads are skipped.
func Parse(r io.Reader) (*Book, error) {
	d := xml.NewDecoder(r)
	d.Strict = false
	d.AutoClose = xml.HTMLAutoClose
	d.Entity = xml.HTMLEntity
	d.CharsetReader = charsetReader
	p := &parser{d: d}
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("fb2: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if err := p.start(t); err != nil {
				return nil, fmt.Errorf("fb2: %w", err)
			}
		case xml.EndElement:
			p.end(t.Name.Local)
		case xml.CharData:
			p.text(string(t))
		}
	}
	if !p.seenRoot {
		return nil, errors.New("fb2: no <FictionBook> root element")
	}
	p.book.Title = collapse(p.bookTitle.String())
	p.book.Year = yearRe.FindString(p.date)
	if p.book.Year == "" {
		p.book.Year = yearRe.FindString(p.pubYear)
	}
	return &p.book, nil
}

func charsetReader(label string, input io.Reader) (io.Reader, error) {
	enc, err := htmlindex.Get(label)
	if err != nil {
		return nil, fmt.Errorf("unsupported encoding %q", label)
	}
	return enc.NewDecoder().Reader(input), nil
}

type parser struct {
	d         *xml.Decoder
	book      Book
	path      []string // local names of open elements, innermost last
	author    *Author  // title-info author being read
	open      []*Section
	title     *strings.Builder // text of the section title being read
	titleFor  *Section
	bookTitle strings.Builder
	date      string
	pubYear   string
	inMain    bool
	seenMain  bool
	seenRoot  bool
}

func (p *parser) parent() string {
	if len(p.path) == 0 {
		return ""
	}
	return p.path[len(p.path)-1]
}

// at reports whether the open elements end with names (innermost last).
func (p *parser) at(names ...string) bool {
	if len(p.path) < len(names) {
		return false
	}
	tail := p.path[len(p.path)-len(names):]
	for i := range names {
		if tail[i] != names[i] {
			return false
		}
	}
	return true
}

func (p *parser) start(t xml.StartElement) error {
	name := t.Name.Local
	switch {
	case name == "FictionBook":
		p.seenRoot = true
	case name == "binary":
		return p.d.Skip()
	case name == "body":
		if !p.seenMain && attr(t, "name") == "" {
			p.inMain, p.seenMain = true, true
		}
	case name == "section" && p.inMain:
		s := &Section{}
		if n := len(p.open); n > 0 {
			p.open[n-1].Children = append(p.open[n-1].Children, s)
		} else {
			p.book.Sections = append(p.book.Sections, s)
		}
		p.open = append(p.open, s)
	case name == "title" && p.inMain && len(p.open) > 0 && p.parent() == "section":
		p.title = &strings.Builder{}
		p.titleFor = p.open[len(p.open)-1]
	case name == "author" && p.parent() == "title-info":
		p.author = &Author{}
	case name == "date" && p.parent() == "title-info":
		p.date += attr(t, "value") + " "
	}
	if p.title != nil && (name == "p" || name == "empty-line") {
		p.title.WriteByte(' ')
	}
	p.path = append(p.path, name)
	return nil
}

func (p *parser) end(name string) {
	if len(p.path) > 0 {
		p.path = p.path[:len(p.path)-1]
	}
	switch {
	case name == "title" && p.title != nil:
		p.titleFor.Title = collapse(p.title.String())
		p.title, p.titleFor = nil, nil
	case name == "section" && p.inMain && len(p.open) > 0:
		p.open = p.open[:len(p.open)-1]
	case name == "body" && p.inMain:
		p.inMain = false
	case name == "author" && p.author != nil:
		a := *p.author
		p.book.Authors = append(p.book.Authors, Author{
			First: collapse(a.First), Middle: collapse(a.Middle),
			Last: collapse(a.Last), Nick: collapse(a.Nick),
		})
		p.author = nil
	}
}

func (p *parser) text(s string) {
	switch {
	case p.title != nil:
		p.title.WriteString(s)
	case p.at("title-info", "book-title"):
		p.bookTitle.WriteString(s)
	case p.author != nil && p.at("author", "first-name"):
		p.author.First += s
	case p.author != nil && p.at("author", "middle-name"):
		p.author.Middle += s
	case p.author != nil && p.at("author", "last-name"):
		p.author.Last += s
	case p.author != nil && p.at("author", "nickname"):
		p.author.Nick += s
	case p.at("title-info", "date"):
		p.date += s + " "
	case p.at("publish-info", "year"):
		p.pubYear += s
	}
}

func attr(t xml.StartElement, name string) string {
	for _, a := range t.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func collapse(s string) string { return strings.Join(strings.Fields(s), " ") }
```

`internal/fb2/file.go`:
```go
package fb2

import (
	"archive/zip"
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ParseFile parses a .fb2 file or the first .fb2 entry of a .fb2.zip archive.
func ParseFile(path string) (*Book, error) {
	if strings.EqualFold(filepathExt(path), ".zip") {
		return parseZip(path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(bufio.NewReaderSize(f, 64<<10))
}

func parseZip(path string) (*Book, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("fb2: %w", err)
	}
	defer zr.Close()
	for _, f := range zr.File {
		if !strings.EqualFold(filepathExt(f.Name), ".fb2") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("fb2: %w", err)
		}
		defer rc.Close()
		return Parse(bufio.NewReaderSize(rc, 64<<10))
	}
	return nil, fmt.Errorf("fb2: no .fb2 entry in %s", path)
}

func filepathExt(name string) string {
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		return name[i:]
	}
	return ""
}
```

- [ ] **Step 5: Тести проходять**

Run: `go test ./internal/fb2/`
Expected: `ok`.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum internal/fb2
git commit -m "feat(fb2): streaming parser for metadata and section tree

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 4: Евристика вибору творів

**Files:**
- Create: `internal/extract/extract.go`
- Test: `internal/extract/extract_test.go`

**Interfaces:**
- Consumes: `fb2.Book`, `fb2.Section`, `fb2.Author`, `(*fb2.Book).AuthorNames()`, `textnorm.Normalize`
- Produces:
  - `type extract.Work struct{ Title, Author, TreePath string }`
  - `func extract.Works(b *fb2.Book) (works []Work, isCollection bool)` — для не-збірки рівно один Work: `{b.Title, b.AuthorNames(), b.Title}`

- [ ] **Step 1: Тести**

`internal/extract/extract_test.go`:
```go
package extract

import (
	"reflect"
	"testing"

	"findbooks/internal/fb2"
)

func sec(title string, children ...*fb2.Section) *fb2.Section {
	return &fb2.Section{Title: title, Children: children}
}

func TestWorks(t *testing.T) {
	nosov := fb2.Author{First: "Евгений", Middle: "Валентинович", Last: "Носов"}
	pidorenko := fb2.Author{First: "Игорь", Last: "Пидоренко"}
	charushnikov := fb2.Author{First: "Олег", Last: "Чарушников"}
	sheckley := fb2.Author{First: "Роберт", Last: "Шекли"}

	tests := []struct {
		name     string
		book     fb2.Book
		want     []Work
		wantColl bool
	}{
		{
			name: "anthology with author sections",
			book: fb2.Book{
				Title:   "Румбы фантастики. 1988 год. Том II",
				Authors: []fb2.Author{nosov, pidorenko, charushnikov},
				Sections: []*fb2.Section{
					sec("Евгений Носов", sec("ЗЕМЛЕЙ РОЖДЕННЫЕ")),
					sec("Игорь Пидоренко", sec("ЧУЖИЕ ДЕТИ", sec("1"), sec("2"))),
					sec("Олег Чарушников", sec("НА «ОЛИМПЕ» ВСЕ СПОКОЙНО", sec("История первая ТРУД СИЗИФА"))),
					sec("Александр Каширин БИБЛИОГРАФИЯ ФАНТАСТИКИ [8]"),
					sec("ОБ АВТОРАХ ЭТОГО СБОРНИКА"),
				},
			},
			want: []Work{
				{"ЗЕМЛЕЙ РОЖДЕННЫЕ", "Евгений Носов", "Евгений Носов › ЗЕМЛЕЙ РОЖДЕННЫЕ"},
				{"ЧУЖИЕ ДЕТИ", "Игорь Пидоренко", "Игорь Пидоренко › ЧУЖИЕ ДЕТИ"},
				{"НА «ОЛИМПЕ» ВСЕ СПОКОЙНО", "Олег Чарушников", "Олег Чарушников › НА «ОЛИМПЕ» ВСЕ СПОКОЙНО"},
			},
			wantColl: true,
		},
		{
			name: "novel with parts and chapters",
			book: fb2.Book{
				Title:   "Человек-амфибия",
				Authors: []fb2.Author{{First: "Александр", Last: "Беляев"}},
				Sections: []*fb2.Section{
					sec("Часть первая", sec("Глава 1 «Морской дьявол»"), sec("Глава 2 Доктор Сальватор")),
					sec("Часть вторая", sec("Глава 1 Ихтиандр")),
				},
			},
			want:     []Work{{"Человек-амфибия", "Александр Беляев", "Человек-амфибия"}},
			wantColl: false,
		},
		{
			name: "single-author story collection",
			book: fb2.Book{
				Title:    "Рассказы",
				Authors:  []fb2.Author{sheckley},
				Sections: []*fb2.Section{sec("Запах мысли"), sec("Страж-птица"), sec("Примечания")},
			},
			want: []Work{
				{"Запах мысли", "Роберт Шекли", "Запах мысли"},
				{"Страж-птица", "Роберт Шекли", "Страж-птица"},
			},
			wantColl: true,
		},
		{
			name: "wrapper section equal to book title is transparent",
			book: fb2.Book{
				Title:    "Сборник X",
				Authors:  []fb2.Author{sheckley},
				Sections: []*fb2.Section{sec("Сборник «X»", sec("Рассказ А"), sec("Рассказ Б"))},
			},
			want: []Work{
				{"Рассказ А", "Роберт Шекли", "Рассказ А"},
				{"Рассказ Б", "Роберт Шекли", "Рассказ Б"},
			},
			wantColl: true,
		},
		{
			name: "multi-author anthology without author sections",
			book: fb2.Book{
				Title:    "Антология",
				Authors:  []fb2.Author{nosov, pidorenko},
				Sections: []*fb2.Section{sec("Рассказ А"), sec("Рассказ Б")},
			},
			want:     []Work{{"Рассказ А", "", "Рассказ А"}, {"Рассказ Б", "", "Рассказ Б"}},
			wantColl: true,
		},
		{
			name: "untitled and asterisk sections are transparent",
			book: fb2.Book{
				Title:    "Сборник",
				Authors:  []fb2.Author{sheckley},
				Sections: []*fb2.Section{sec("", sec("Рассказ А")), sec("* * *"), sec("Рассказ Б")},
			},
			want: []Work{
				{"Рассказ А", "Роберт Шекли", "Рассказ А"},
				{"Рассказ Б", "Роберт Шекли", "Рассказ Б"},
			},
			wantColl: true,
		},
		{
			name: "author section with initials",
			book: fb2.Book{
				Title:    "Альманах",
				Authors:  []fb2.Author{nosov, pidorenko},
				Sections: []*fb2.Section{sec("Е. Носов", sec("Повесть")), sec("И. Пидоренко", sec("Рассказ"))},
			},
			want: []Work{
				{"Повесть", "Евгений Носов", "Евгений Носов › Повесть"},
				{"Рассказ", "Игорь Пидоренко", "Игорь Пидоренко › Рассказ"},
			},
			wantColl: true,
		},
		{
			name: "one work plus back matter is not a collection",
			book: fb2.Book{
				Title:    "Повесть",
				Authors:  []fb2.Author{sheckley},
				Sections: []*fb2.Section{sec("Предисловие"), sec("Повесть о странном"), sec("Примечания")},
			},
			want:     []Work{{"Повесть", "Роберт Шекли", "Повесть"}},
			wantColl: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, coll := Works(&tt.book)
			if coll != tt.wantColl {
				t.Errorf("isCollection = %v, want %v", coll, tt.wantColl)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("works =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

func TestIsNoise(t *testing.T) {
	tests := map[string]bool{
		"1": true, "iv": true, "глава 5": true, "глава первая встреча": true,
		"часть вторая": true, "пролог": true, "примечания": true, "об авторе": true,
		"предисловие": true, "книга пятая": true,
		"книга мертвых": false, "чужие дети": false, "civil": false, "часть тела": false,
	}
	for in, want := range tests {
		if got := isNoise(in); got != want {
			t.Errorf("isNoise(%q) = %v, want %v", in, got, want)
		}
	}
}
```

- [ ] **Step 2: Перевірити падіння**

Run: `go test ./internal/extract/`
Expected: FAIL — `undefined: Works`.

- [ ] **Step 3: Реалізація**

`internal/extract/extract.go`:
```go
// Package extract decides which sections of an FB2 book are separate works.
package extract

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"findbooks/internal/fb2"
	"findbooks/internal/textnorm"
)

type Work struct{ Title, Author, TreePath string }

const treeSep = " › "

// Works walks the section tree. Untitled, noise (chapters, notes, prefaces)
// and book-title sections are transparent; a section titled with one of the
// book's authors sets the author for its children; any other titled section
// is a work and its subsections are not visited. With fewer than two works
// the book is not a collection and a single work for the whole book is
// returned.
func Works(b *fb2.Book) ([]Work, bool) {
	bookTitle := textnorm.Normalize(b.Title)
	var found []Work
	var walk func(secs []*fb2.Section, ctxAuthor string)
	walk = func(secs []*fb2.Section, ctxAuthor string) {
		for _, s := range secs {
			norm := textnorm.Normalize(s.Title)
			if norm == "" || norm == bookTitle || isNoise(norm) {
				walk(s.Children, ctxAuthor)
				continue
			}
			if name := matchAuthor(norm, b.Authors); name != "" {
				walk(s.Children, name)
				continue
			}
			found = append(found, newWork(s.Title, ctxAuthor, b.Authors))
		}
	}
	walk(b.Sections, "")
	if len(found) < 2 {
		return []Work{{Title: b.Title, Author: b.AuthorNames(), TreePath: b.Title}}, false
	}
	return found, true
}

func newWork(title, ctxAuthor string, authors []fb2.Author) Work {
	author := ctxAuthor
	if author == "" && len(authors) == 1 {
		author = authors[0].Name()
	}
	tree := title
	if ctxAuthor != "" {
		tree = ctxAuthor + treeSep + title
	}
	return Work{Title: title, Author: author, TreePath: tree}
}

var (
	noiseExact = map[string]bool{
		"пролог": true, "эпилог": true, "примечания": true, "примечание": true,
		"комментарии": true, "содержание": true, "оглавление": true,
		"от составителя": true, "от автора": true, "от редакции": true,
		"предисловие": true, "послесловие": true, "вместо предисловия": true,
		"вместо послесловия": true, "аннотация": true, "notes": true, "contents": true,
	}
	noiseContains = []string{"об авторах", "об авторе", "библиограф"}
	numberRe      = regexp.MustCompile(`^[0-9]+$`)
	romanRe       = regexp.MustCompile(`^m{0,3}(cm|cd|d?c{0,3})(xc|xl|l?x{0,3})(ix|iv|v?i{0,3})$`)
	chapterRe     = regexp.MustCompile(`^(глава|часть|книга|том|раздел|chapter|part) (([0-9]+|[ivxlcdm]+)( |$)|перв|втор|трет|четв|пят|шест|седьм|восьм|девят|десят|one|two|three)`)
)

// isNoise reports whether a normalized, non-empty title is structural or
// back matter rather than a work.
func isNoise(norm string) bool {
	if noiseExact[norm] || numberRe.MatchString(norm) || romanRe.MatchString(norm) || chapterRe.MatchString(norm) {
		return true
	}
	for _, s := range noiseContains {
		if strings.Contains(norm, s) {
			return true
		}
	}
	return false
}

// matchAuthor returns the author's display name when the normalized title
// consists only of that author's name parts (initials allowed) and contains
// the full last name.
func matchAuthor(norm string, authors []fb2.Author) string {
	words := strings.Fields(norm)
	if len(words) == 0 || len(words) > 5 {
		return ""
	}
	for _, a := range authors {
		last := strings.Fields(textnorm.Normalize(a.Last))
		if len(last) == 0 {
			continue
		}
		given := strings.Fields(textnorm.Normalize(a.First + " " + a.Middle))
		if containsAll(words, last) && allKnown(words, last, given) {
			return a.Name()
		}
	}
	return ""
}

func containsAll(words, need []string) bool {
	for _, n := range need {
		if !slices.Contains(words, n) {
			return false
		}
	}
	return true
}

func allKnown(words, last, given []string) bool {
	for _, w := range words {
		if slices.Contains(last, w) || slices.Contains(given, w) {
			continue
		}
		if utf8.RuneCountInString(w) == 1 && slices.ContainsFunc(given, func(g string) bool { return strings.HasPrefix(g, w) }) {
			continue
		}
		return false
	}
	return true
}
```

- [ ] **Step 4: Тести проходять**

Run: `go test ./internal/extract/`
Expected: `ok`.

- [ ] **Step 5: Commit**

```bash
git add internal/extract
git commit -m "feat(extract): pick works out of the FB2 section tree

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 5: SQLite-індекс

**Files:**
- Create: `internal/index/schema.go`, `internal/index/store.go`, `internal/index/books.go`, `internal/index/search.go`
- Test: `internal/index/store_test.go`, `internal/index/search_test.go`

**Interfaces:**
- Consumes: `textnorm.Normalize`
- Produces:
  - `var index.ErrNotFound, index.ErrExists error`
  - `type index.Library struct{ ID int64; Name, VolumeID, VolumeName, RootRel string; LastScan time.Time; Books, Works int }`
  - `type index.FileStamp struct{ BookID, Size, MTime int64 }`
  - `type index.WorkRecord struct{ Title, Author, TreePath string }`
  - `type index.BookRecord struct{ RelPath, Format string; Size, MTime int64; Title, Authors, Year string; IsCollection bool; Works []WorkRecord }`
  - `type index.Query struct{ Text, Author string; Limit int }` (Limit ≤ 0 → 50)
  - `type index.Hit struct{ WorkID int64; Title, Author, TreePath string; BookID int64; BookTitle, BookYear string; IsCollection bool; RelPath, Format, LibraryName, VolumeID, VolumeName, RootRel string }`
  - `func index.Open(path string) (*Store, error)`, `(*Store).Close() error`
  - `(*Store).AddLibrary(name, volumeID, volumeName, rootRel string) (Library, error)`
  - `(*Store).Library(name string) (Library, error)`, `(*Store).Libraries() ([]Library, error)`, `(*Store).RemoveLibrary(name string) error`
  - `(*Store).MarkScanned(libraryID int64, t time.Time) error`
  - `(*Store).Files(libraryID int64) (map[string]FileStamp, error)` — ключ `rel_path`
  - `(*Store).WriteBatch(libraryID int64, put []BookRecord, del []int64) error` — одна транзакція; `put` замінює книгу з тим самим `RelPath`; `del` — `book_id`
  - `(*Store).Search(q Query) ([]Hit, error)` — порожній запит → `nil, nil`
  - `(*Store).BookWorks(bookID int64) ([]string, error)` — назви творів у порядку книги
  - `(*Store).TotalWorks() (int, error)`
  - `func index.MatchExpr(text string) string`

- [ ] **Step 1: Тести**

`internal/index/store_test.go`:
```go
package index

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	st, err := Open(filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func addLib(t *testing.T, st *Store, name string) Library {
	t.Helper()
	lib, err := st.AddLibrary(name, "VOL-"+name, "Disk "+name, "Бібліотека/"+name)
	if err != nil {
		t.Fatal(err)
	}
	return lib
}

var (
	rumby = BookRecord{
		RelPath: "Сборники/Румбы/(1988) Румбы фантастики. Том II.fb2", Format: "fb2", Size: 100, MTime: 1000,
		Title: "Румбы фантастики. 1988 год. Том II", Authors: "Евгений Носов, Игорь Пидоренко", Year: "1988",
		IsCollection: true,
		Works: []WorkRecord{
			{"ЗЕМЛЕЙ РОЖДЕННЫЕ", "Евгений Носов", "Евгений Носов › ЗЕМЛЕЙ РОЖДЕННЫЕ"},
			{"ЧУЖИЕ ДЕТИ", "Игорь Пидоренко", "Игорь Пидоренко › ЧУЖИЕ ДЕТИ"},
		},
	}
	belyaev = BookRecord{
		RelPath: "Авторы/Беляев Александр/Звезда КЭЦ.pdf", Format: "pdf", Size: 5, MTime: 1,
		Title: "Звезда КЭЦ", Works: []WorkRecord{{Title: "Звезда КЭЦ", TreePath: "Звезда КЭЦ"}},
	}
	yolka = BookRecord{
		RelPath: "Ёлка.fb2", Format: "fb2", Size: 7, MTime: 7, Title: "Ёлка", Authors: "Иван Петров",
		Works: []WorkRecord{{"Ёлка", "Иван Петров", "Ёлка"}},
	}
)

func TestAddLibraryDuplicate(t *testing.T) {
	st := newStore(t)
	addLib(t, st, "A")
	if _, err := st.AddLibrary("A", "other", "x", "y"); !errors.Is(err, ErrExists) {
		t.Fatalf("same name: err = %v, want ErrExists", err)
	}
	if _, err := st.AddLibrary("B", "VOL-A", "x", "Бібліотека/A"); !errors.Is(err, ErrExists) {
		t.Fatalf("same volume+root: err = %v, want ErrExists", err)
	}
}

func TestLibrariesCountsAndScanTime(t *testing.T) {
	st := newStore(t)
	lib := addLib(t, st, "A")
	if err := st.WriteBatch(lib.ID, []BookRecord{rumby, belyaev}, nil); err != nil {
		t.Fatal(err)
	}
	when := time.Unix(1_800_000_000, 0)
	if err := st.MarkScanned(lib.ID, when); err != nil {
		t.Fatal(err)
	}
	libs, err := st.Libraries()
	if err != nil {
		t.Fatal(err)
	}
	if len(libs) != 1 || libs[0].Books != 2 || libs[0].Works != 3 || !libs[0].LastScan.Equal(when) {
		t.Fatalf("libraries = %+v", libs)
	}
	got, err := st.Library("A")
	if err != nil || got.RootRel != "Бібліотека/A" || got.VolumeID != "VOL-A" {
		t.Fatalf("Library(A) = %+v, %v", got, err)
	}
	if _, err := st.Library("nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	if n, err := st.TotalWorks(); err != nil || n != 3 {
		t.Fatalf("TotalWorks = %d, %v", n, err)
	}
}

func TestFilesReflectsWrites(t *testing.T) {
	st := newStore(t)
	lib := addLib(t, st, "A")
	if err := st.WriteBatch(lib.ID, []BookRecord{rumby}, nil); err != nil {
		t.Fatal(err)
	}
	files, err := st.Files(lib.ID)
	if err != nil {
		t.Fatal(err)
	}
	f, ok := files[rumby.RelPath]
	if !ok || f.Size != 100 || f.MTime != 1000 || f.BookID == 0 {
		t.Fatalf("files = %+v", files)
	}
}

func TestWriteBatchReplacesBook(t *testing.T) {
	st := newStore(t)
	lib := addLib(t, st, "A")
	if err := st.WriteBatch(lib.ID, []BookRecord{rumby}, nil); err != nil {
		t.Fatal(err)
	}
	changed := rumby
	changed.Size = 200
	changed.Works = []WorkRecord{{"НОВАЯ ПОВЕСТЬ", "Игорь Пидоренко", "Игорь Пидоренко › НОВАЯ ПОВЕСТЬ"}}
	if err := st.WriteBatch(lib.ID, []BookRecord{changed}, nil); err != nil {
		t.Fatal(err)
	}
	mustHits(t, st, Query{Text: "чужие"}, 0)
	hits := mustHits(t, st, Query{Text: "новая"}, 1)
	titles, err := st.BookWorks(hits[0].BookID)
	if err != nil || len(titles) != 1 || titles[0] != "НОВАЯ ПОВЕСТЬ" {
		t.Fatalf("BookWorks = %q, %v", titles, err)
	}
	if n, _ := st.TotalWorks(); n != 1 {
		t.Fatalf("TotalWorks = %d, want 1 (old works must be gone)", n)
	}
}

func TestDeleteAndRemoveLibrary(t *testing.T) {
	st := newStore(t)
	lib := addLib(t, st, "A")
	if err := st.WriteBatch(lib.ID, []BookRecord{rumby, yolka}, nil); err != nil {
		t.Fatal(err)
	}
	files, _ := st.Files(lib.ID)
	if err := st.WriteBatch(lib.ID, nil, []int64{files[rumby.RelPath].BookID}); err != nil {
		t.Fatal(err)
	}
	mustHits(t, st, Query{Text: "чужие"}, 0)
	mustHits(t, st, Query{Text: "елка"}, 1)
	if err := st.RemoveLibrary("A"); err != nil {
		t.Fatal(err)
	}
	mustHits(t, st, Query{Text: "елка"}, 0)
	if libs, _ := st.Libraries(); len(libs) != 0 {
		t.Fatalf("libraries left: %+v", libs)
	}
	if err := st.RemoveLibrary("A"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
```

`internal/index/search_test.go`:
```go
package index

import "testing"

func mustHits(t *testing.T, st *Store, q Query, want int) []Hit {
	t.Helper()
	hits, err := st.Search(q)
	if err != nil {
		t.Fatalf("Search(%+v): %v", q, err)
	}
	if len(hits) != want {
		t.Fatalf("Search(%+v) = %d hits %+v, want %d", q, len(hits), hits, want)
	}
	return hits
}

func seeded(t *testing.T) *Store {
	st := newStore(t)
	lib := addLib(t, st, "Фантастика")
	if err := st.WriteBatch(lib.ID, []BookRecord{rumby, belyaev, yolka}, nil); err != nil {
		t.Fatal(err)
	}
	return st
}

func TestSearchFindsWorkInsideCollection(t *testing.T) {
	st := seeded(t)
	h := mustHits(t, st, Query{Text: "чужие дети"}, 1)[0]
	if h.Title != "ЧУЖИЕ ДЕТИ" || h.Author != "Игорь Пидоренко" || h.TreePath != "Игорь Пидоренко › ЧУЖИЕ ДЕТИ" ||
		h.BookTitle != rumby.Title || h.BookYear != "1988" || !h.IsCollection || h.RelPath != rumby.RelPath ||
		h.Format != "fb2" || h.LibraryName != "Фантастика" || h.VolumeID != "VOL-Фантастика" ||
		h.VolumeName != "Disk Фантастика" || h.RootRel != "Бібліотека/Фантастика" {
		t.Fatalf("hit = %+v", h)
	}
}

func TestSearchIgnoresCaseYoAndPunctuation(t *testing.T) {
	st := seeded(t)
	for _, q := range []string{"ЧУЖИЕ", "чуж дет", "  Чужие,  дети! "} {
		mustHits(t, st, Query{Text: q}, 1)
	}
	mustHits(t, st, Query{Text: "елка"}, 1)
	mustHits(t, st, Query{Text: "Ёлка"}, 1)
}

func TestSearchByBookTitleAndFolder(t *testing.T) {
	st := seeded(t)
	mustHits(t, st, Query{Text: "румбы"}, 2)
	mustHits(t, st, Query{Text: "беляев"}, 1)
}

func TestSearchAuthorFilter(t *testing.T) {
	st := seeded(t)
	h := mustHits(t, st, Query{Author: "пидоренко"}, 1)[0]
	if h.Title != "ЧУЖИЕ ДЕТИ" {
		t.Fatalf("hit = %+v", h)
	}
	mustHits(t, st, Query{Text: "земле", Author: "пидоренко"}, 0)
	mustHits(t, st, Query{Text: "земле", Author: "носов"}, 1)
}

func TestSearchHostileInput(t *testing.T) {
	st := seeded(t)
	for _, q := range []string{`"`, `*`, `чуж*"`, `NEAR(a b)`, `title:x`, `-дети`, `AND`, `OR NOT`, `')--`, `{title}: x`} {
		if _, err := st.Search(Query{Text: q}); err != nil {
			t.Errorf("Search(%q): %v", q, err)
		}
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	st := seeded(t)
	hits, err := st.Search(Query{Text: " !! "})
	if err != nil || hits != nil {
		t.Fatalf("got %v, %v", hits, err)
	}
}

func TestMatchExpr(t *testing.T) {
	tests := map[string]string{
		"Чужие дети":  `"чужие"* "дети"*`,
		`"; DROP`:     `"drop"*`,
		"":            "",
		"«Ёлка»-2":    `"елка"* "2"*`,
	}
	for in, want := range tests {
		if got := MatchExpr(in); got != want {
			t.Errorf("MatchExpr(%q) = %q, want %q", in, got, want)
		}
	}
}
```

- [ ] **Step 2: Перевірити падіння**

Run: `go test ./internal/index/`
Expected: FAIL — `undefined: Open`, `undefined: BookRecord`…

- [ ] **Step 3: Реалізація**

`internal/index/schema.go`:
```go
package index

// works_fts holds already-normalized text; rowid = works.id. Rows are removed
// by the trigger, which also fires for cascaded deletes from books/libraries.
const schema = `
CREATE TABLE IF NOT EXISTS libraries (
	id           INTEGER PRIMARY KEY,
	name         TEXT NOT NULL UNIQUE,
	volume_id    TEXT NOT NULL,
	volume_name  TEXT NOT NULL,
	root_rel     TEXT NOT NULL,
	last_scan_at INTEGER NOT NULL DEFAULT 0,
	UNIQUE (volume_id, root_rel)
);
CREATE TABLE IF NOT EXISTS books (
	id            INTEGER PRIMARY KEY,
	library_id    INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
	rel_path      TEXT NOT NULL,
	format        TEXT NOT NULL,
	size          INTEGER NOT NULL,
	mtime         INTEGER NOT NULL,
	title         TEXT NOT NULL,
	authors       TEXT NOT NULL,
	year          TEXT NOT NULL,
	is_collection INTEGER NOT NULL,
	UNIQUE (library_id, rel_path)
);
CREATE TABLE IF NOT EXISTS works (
	id        INTEGER PRIMARY KEY,
	book_id   INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
	ord       INTEGER NOT NULL,
	title     TEXT NOT NULL,
	author    TEXT NOT NULL,
	tree_path TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS works_book ON works(book_id);
CREATE VIRTUAL TABLE IF NOT EXISTS works_fts USING fts5(
	title, author, book_title,
	tokenize = 'unicode61 remove_diacritics 0',
	prefix = '2 3'
);
CREATE TRIGGER IF NOT EXISTS works_fts_delete AFTER DELETE ON works BEGIN
	DELETE FROM works_fts WHERE rowid = old.id;
END;
`
```

`internal/index/store.go`:
```go
// Package index stores libraries, books and works in SQLite with an FTS5
// index over normalized work titles, authors and book titles.
package index

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var (
	ErrNotFound = errors.New("бібліотеку не знайдено")
	ErrExists   = errors.New("бібліотека з такою назвою або шляхом уже є")
)

type Store struct{ db *sql.DB }

type Library struct {
	ID                                 int64
	Name, VolumeID, VolumeName, RootRel string
	LastScan                           time.Time
	Books, Works                       int
}

// Open opens (creating if needed) the index at path.
func Open(path string) (*Store, error) {
	dsn := path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("index %s: %w", path, err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) AddLibrary(name, volumeID, volumeName, rootRel string) (Library, error) {
	res, err := s.db.Exec(`INSERT INTO libraries(name, volume_id, volume_name, root_rel) VALUES (?, ?, ?, ?)`,
		name, volumeID, volumeName, rootRel)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return Library{}, fmt.Errorf("%w: %s", ErrExists, name)
		}
		return Library{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Library{}, err
	}
	return Library{ID: id, Name: name, VolumeID: volumeID, VolumeName: volumeName, RootRel: rootRel}, nil
}

const libSelect = `SELECT l.id, l.name, l.volume_id, l.volume_name, l.root_rel, l.last_scan_at,
	(SELECT count(*) FROM books b WHERE b.library_id = l.id),
	(SELECT count(*) FROM works w JOIN books b ON b.id = w.book_id WHERE b.library_id = l.id)
FROM libraries l`

func scanLibrary(row interface{ Scan(...any) error }) (Library, error) {
	var l Library
	var last int64
	if err := row.Scan(&l.ID, &l.Name, &l.VolumeID, &l.VolumeName, &l.RootRel, &last, &l.Books, &l.Works); err != nil {
		return Library{}, err
	}
	if last > 0 {
		l.LastScan = time.Unix(last, 0)
	}
	return l, nil
}

func (s *Store) Library(name string) (Library, error) {
	l, err := scanLibrary(s.db.QueryRow(libSelect+` WHERE l.name = ?`, name))
	if errors.Is(err, sql.ErrNoRows) {
		return Library{}, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return l, err
}

func (s *Store) Libraries() ([]Library, error) {
	rows, err := s.db.Query(libSelect + ` ORDER BY l.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Library
	for rows.Next() {
		l, err := scanLibrary(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *Store) RemoveLibrary(name string) error {
	res, err := s.db.Exec(`DELETE FROM libraries WHERE name = ?`, name)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return nil
}

func (s *Store) MarkScanned(libraryID int64, t time.Time) error {
	_, err := s.db.Exec(`UPDATE libraries SET last_scan_at = ? WHERE id = ?`, t.Unix(), libraryID)
	return err
}
```

`internal/index/books.go`:
```go
package index

import (
	"database/sql"
	"fmt"
	"path"

	"findbooks/internal/textnorm"
)

type FileStamp struct{ BookID, Size, MTime int64 }

type WorkRecord struct{ Title, Author, TreePath string }

type BookRecord struct {
	RelPath, Format        string
	Size, MTime            int64
	Title, Authors, Year   string
	IsCollection           bool
	Works                  []WorkRecord
}

// Files returns the stored size/mtime of every book in the library, keyed by rel_path.
func (s *Store) Files(libraryID int64) (map[string]FileStamp, error) {
	rows, err := s.db.Query(`SELECT id, rel_path, size, mtime FROM books WHERE library_id = ?`, libraryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]FileStamp{}
	for rows.Next() {
		var rel string
		var f FileStamp
		if err := rows.Scan(&f.BookID, &rel, &f.Size, &f.MTime); err != nil {
			return nil, err
		}
		out[rel] = f
	}
	return out, rows.Err()
}

// WriteBatch deletes the books in del and inserts or replaces the books in
// put, all in one transaction.
func (s *Store) WriteBatch(libraryID int64, put []BookRecord, del []int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, id := range del {
		if _, err := tx.Exec(`DELETE FROM books WHERE id = ?`, id); err != nil {
			return err
		}
	}
	for _, b := range put {
		if err := putBook(tx, libraryID, b); err != nil {
			return fmt.Errorf("%s: %w", b.RelPath, err)
		}
	}
	return tx.Commit()
}

func putBook(tx *sql.Tx, libraryID int64, b BookRecord) error {
	if _, err := tx.Exec(`DELETE FROM books WHERE library_id = ? AND rel_path = ?`, libraryID, b.RelPath); err != nil {
		return err
	}
	res, err := tx.Exec(`INSERT INTO books(library_id, rel_path, format, size, mtime, title, authors, year, is_collection)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		libraryID, b.RelPath, b.Format, b.Size, b.MTime, b.Title, b.Authors, b.Year, b.IsCollection)
	if err != nil {
		return err
	}
	bookID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	bookText := textnorm.Normalize(b.Title + " " + folderOf(b.RelPath))
	for i, w := range b.Works {
		res, err := tx.Exec(`INSERT INTO works(book_id, ord, title, author, tree_path) VALUES (?, ?, ?, ?, ?)`,
			bookID, i, w.Title, w.Author, w.TreePath)
		if err != nil {
			return err
		}
		workID, err := res.LastInsertId()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO works_fts(rowid, title, author, book_title) VALUES (?, ?, ?, ?)`,
			workID, textnorm.Normalize(w.Title), textnorm.Normalize(w.Author), bookText); err != nil {
			return err
		}
	}
	return nil
}

func folderOf(rel string) string {
	if d := path.Dir(rel); d != "." {
		return d
	}
	return ""
}
```

`internal/index/search.go`:
```go
package index

import (
	"strings"

	"findbooks/internal/textnorm"
)

type Query struct {
	Text, Author string
	Limit        int
}

type Hit struct {
	WorkID                      int64
	Title, Author, TreePath     string
	BookID                      int64
	BookTitle, BookYear         string
	IsCollection                bool
	RelPath, Format             string
	LibraryName                 string
	VolumeID, VolumeName, RootRel string
}

// MatchExpr turns free text into an FTS5 expression: every normalized word
// becomes a quoted prefix term; terms are ANDed. Normalization strips quotes
// and operators, so user input can never inject FTS5 syntax.
func MatchExpr(text string) string {
	words := strings.Fields(textnorm.Normalize(text))
	for i, w := range words {
		words[i] = `"` + w + `"*`
	}
	return strings.Join(words, " ")
}

const searchSQL = `
SELECT w.id, w.title, w.author, w.tree_path,
       b.id, b.title, b.year, b.is_collection, b.rel_path, b.format,
       l.name, l.volume_id, l.volume_name, l.root_rel
FROM works_fts
JOIN works w ON w.id = works_fts.rowid
JOIN books b ON b.id = w.book_id
JOIN libraries l ON l.id = b.library_id
WHERE works_fts MATCH ?
ORDER BY bm25(works_fts, 10.0, 3.0, 1.0)
LIMIT ?`

func (s *Store) Search(q Query) ([]Hit, error) {
	expr := MatchExpr(q.Text)
	if a := MatchExpr(q.Author); a != "" {
		col := "author : (" + a + ")"
		if expr == "" {
			expr = col
		} else {
			expr = "(" + expr + ") AND " + col
		}
	}
	if expr == "" {
		return nil, nil
	}
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(searchSQL, expr, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hits []Hit
	for rows.Next() {
		var h Hit
		if err := rows.Scan(&h.WorkID, &h.Title, &h.Author, &h.TreePath,
			&h.BookID, &h.BookTitle, &h.BookYear, &h.IsCollection, &h.RelPath, &h.Format,
			&h.LibraryName, &h.VolumeID, &h.VolumeName, &h.RootRel); err != nil {
			return nil, err
		}
		hits = append(hits, h)
	}
	return hits, rows.Err()
}

// BookWorks returns the titles of a book's works in book order.
func (s *Store) BookWorks(bookID int64) ([]string, error) {
	rows, err := s.db.Query(`SELECT title FROM works WHERE book_id = ? ORDER BY ord`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) TotalWorks() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT count(*) FROM works`).Scan(&n)
	return n, err
}
```

```bash
go get modernc.org/sqlite@v1.60.0
go mod tidy
gofmt -w internal/index
```

- [ ] **Step 4: Тести проходять**

Run: `go test ./internal/index/`
Expected: `ok`.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum internal/index
git commit -m "feat(index): SQLite FTS5 store for libraries, books and works

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 6: Прив'язка бібліотеки до тому

**Files:**
- Create: `internal/library/library.go`
- Test: `internal/library/library_test.go`

**Interfaces:**
- Consumes: `platform.VolumeID`, `platform.MountPoint`
- Produces:
  - `type library.Location struct{ VolumeID, VolumeName, RootRel string }` — `RootRel` з `/`
  - `func library.Locate(dir string) (Location, error)` — тека має існувати і бути на підключеному томі
  - `func library.Root(volumeID, rootRel string) (string, bool)` — повний шлях, `false` якщо том не підключено
  - `func library.FilePath(root, relPath string) string`

- [ ] **Step 1: Тести**

`internal/library/library_test.go`:
```go
package library

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func skipUnlessDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("platform is implemented only for macOS in v1")
	}
}

// TempDir lives under /private/var/folders on the Data volume, which is
// mounted at /System/Volumes/Data — this exercises the firmlink fallback.
func TestLocateAndRootRoundTrip(t *testing.T) {
	skipUnlessDarwin(t)
	dir := t.TempDir()
	loc, err := Locate(dir)
	if err != nil {
		t.Fatal(err)
	}
	if loc.VolumeID == "" || loc.VolumeName == "" || strings.Contains(loc.RootRel, `\`) || strings.HasPrefix(loc.RootRel, "..") {
		t.Fatalf("location = %+v", loc)
	}
	root, ok := Root(loc.VolumeID, loc.RootRel)
	if !ok {
		t.Fatalf("Root(%+v) not mounted", loc)
	}
	a, err1 := os.Stat(root)
	b, err2 := os.Stat(dir)
	if err1 != nil || err2 != nil || !os.SameFile(a, b) {
		t.Fatalf("Root = %q is not %q (%v, %v)", root, dir, err1, err2)
	}
}

func TestLocateRejectsFileAndMissing(t *testing.T) {
	skipUnlessDarwin(t)
	file := filepath.Join(t.TempDir(), "x.fb2")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Locate(file); err == nil {
		t.Error("expected error for a file")
	}
	if _, err := Locate(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("expected error for a missing dir")
	}
}

func TestRootOfflineVolume(t *testing.T) {
	if root, ok := Root("00000000-0000-0000-0000-000000000000", "books"); ok {
		t.Fatalf("unexpected root %q", root)
	}
}

func TestFilePathJoinsSlashPath(t *testing.T) {
	if got, want := FilePath("/a", "b/c.fb2"), filepath.Join("/a", "b", "c.fb2"); got != want {
		t.Fatalf("FilePath = %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Перевірити падіння**

Run: `go test ./internal/library/`
Expected: FAIL — `undefined: Locate`.

- [ ] **Step 3: Реалізація**

`internal/library/library.go`:
```go
// Package library maps library folders to (volume, relative path) pairs so
// the index survives remounts and changing mount points or drive letters.
package library

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"findbooks/internal/platform"
)

type Location struct{ VolumeID, VolumeName, RootRel string }

// Locate resolves dir to the volume it lives on and its path on that volume.
func Locate(dir string) (Location, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Location{}, err
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return Location{}, fmt.Errorf("тека %s: %w", dir, err)
	}
	fi, err := os.Stat(real)
	if err != nil {
		return Location{}, err
	}
	if !fi.IsDir() {
		return Location{}, fmt.Errorf("%s — не тека", dir)
	}
	id, name, err := platform.VolumeID(real)
	if err != nil {
		return Location{}, fmt.Errorf("не вдалося визначити диск для %s: %w", dir, err)
	}
	mp, ok := platform.MountPoint(id)
	if !ok {
		return Location{}, fmt.Errorf("диск «%s» не знайдено серед підключених", name)
	}
	rel, err := relToMount(mp, real)
	if err != nil {
		return Location{}, err
	}
	return Location{VolumeID: id, VolumeName: name, RootRel: filepath.ToSlash(rel)}, nil
}

// relToMount returns real relative to the mount point mp. A path that is not
// lexically under mp (macOS firmlinks: /Users and /private/var live on the
// Data volume mounted at /System/Volumes/Data) is accepted when mp+real is
// the same directory.
func relToMount(mp, real string) (string, error) {
	rel, err := filepath.Rel(mp, real)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return rel, nil
	}
	alt := filepath.Join(mp, real)
	if a, err1 := os.Stat(alt); err1 == nil {
		if r, err2 := os.Stat(real); err2 == nil && os.SameFile(a, r) {
			return filepath.Rel(mp, alt)
		}
	}
	return "", fmt.Errorf("%s не лежить на томі %s", real, mp)
}

// Root returns the library's full path, or false if its volume is not mounted.
func Root(volumeID, rootRel string) (string, bool) {
	mp, ok := platform.MountPoint(volumeID)
	if !ok {
		return "", false
	}
	return filepath.Join(mp, filepath.FromSlash(rootRel)), true
}

// FilePath joins a library root with a slash-separated rel_path.
func FilePath(root, relPath string) string {
	return filepath.Join(root, filepath.FromSlash(relPath))
}
```

- [ ] **Step 4: Тести проходять**

Run: `go test ./internal/library/`
Expected: `ok`.

- [ ] **Step 5: Commit**

```bash
git add internal/library
git commit -m "feat(library): tie library roots to volume id and relative path

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 7: Інкрементальне сканування

**Files:**
- Create: `internal/scan/scan.go`
- Test: `internal/scan/scan_test.go`

**Interfaces:**
- Consumes: `index.Store` (`Files`, `WriteBatch`, `MarkScanned`), `index.BookRecord`, `index.WorkRecord`, `index.FileStamp`, `fb2.ParseFile`, `(*fb2.Book).AuthorNames`, `extract.Works`
- Produces:
  - `type scan.Progress struct{ Done, Total int }`
  - `type scan.FileError struct{ RelPath string; Err error }`
  - `type scan.Report struct{ Added, Updated, Removed, Unchanged int; Errors []FileError }`
  - `func scan.Run(ctx context.Context, st *index.Store, libraryID int64, root string, onProgress func(Progress)) (Report, error)` — `onProgress` викликається з горутини, що викликала `Run`; може бути `nil`

- [ ] **Step 1: Тести**

`internal/scan/scan_test.go`:
```go
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
	write(t, filepath.Join(root, "Бои\u0306цов - Рассказы.txt"), []byte("text")) // NFD name
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
```

- [ ] **Step 2: Перевірити падіння**

Run: `go test ./internal/scan/`
Expected: FAIL — `undefined: Run`.

- [ ] **Step 3: Реалізація**

`internal/scan/scan.go`:
```go
// Package scan walks a library folder and keeps the index in sync with it.
package scan

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"findbooks/internal/extract"
	"findbooks/internal/fb2"
	"findbooks/internal/index"
)

type Progress struct{ Done, Total int }

type FileError struct {
	RelPath string
	Err     error
}

type Report struct {
	Added, Updated, Removed, Unchanged int
	Errors                             []FileError
}

const batchSize = 200

var formats = map[string]string{
	".fb2": "fb2", ".pdf": "pdf", ".djvu": "djvu", ".djv": "djvu", ".doc": "doc",
	".docx": "docx", ".rtf": "rtf", ".txt": "txt", ".epub": "epub", ".mobi": "mobi",
}

const zipFB2 = ".fb2.zip"

func hasZipFB2Suffix(name string) bool {
	return len(name) >= len(zipFB2) && strings.EqualFold(name[len(name)-len(zipFB2):], zipFB2)
}

// formatOf returns the book format of a file name, or false for non-books.
func formatOf(name string) (string, bool) {
	if hasZipFB2Suffix(name) {
		return "fb2", true
	}
	f, ok := formats[strings.ToLower(filepath.Ext(name))]
	return f, ok
}

// stem strips the book extension (".fb2.zip" counts as one).
func stem(name string) string {
	if hasZipFB2Suffix(name) {
		return name[:len(name)-len(zipFB2)]
	}
	return strings.TrimSuffix(name, filepath.Ext(name))
}

type job struct {
	rel, abs, format string
	size, mtime      int64
	existed          bool
}

type result struct {
	rec     index.BookRecord
	existed bool
	err     error
}

// Run brings the index of one library in line with the folder at root:
// new and changed files (by size and mtime) are parsed in parallel and
// written in batches, vanished files are deleted. Unreadable files are
// indexed by file name and listed in Report.Errors.
func Run(ctx context.Context, st *index.Store, libraryID int64, root string, onProgress func(Progress)) (Report, error) {
	var rep Report
	stored, err := st.Files(libraryID)
	if err != nil {
		return rep, err
	}
	jobs, seen, err := collect(ctx, root, stored, &rep)
	if err != nil {
		return rep, err
	}
	var del []int64
	for rel, f := range stored {
		if !seen[rel] {
			del = append(del, f.BookID)
		}
	}
	rep.Removed = len(del)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	results := parseAll(ctx, jobs)

	batch := make([]index.BookRecord, 0, batchSize)
	flush := func() error {
		if len(batch) == 0 && len(del) == 0 {
			return nil
		}
		err := st.WriteBatch(libraryID, batch, del)
		batch, del = batch[:0], nil
		return err
	}
	done := 0
	for r := range results {
		done++
		if r.err != nil {
			rep.Errors = append(rep.Errors, FileError{RelPath: r.rec.RelPath, Err: r.err})
		}
		if r.existed {
			rep.Updated++
		} else {
			rep.Added++
		}
		batch = append(batch, r.rec)
		if len(batch) >= batchSize {
			if err := flush(); err != nil {
				cancel()
				for range results {
				}
				return rep, err
			}
		}
		if onProgress != nil {
			onProgress(Progress{Done: done, Total: len(jobs)})
		}
	}
	if err := ctx.Err(); err != nil {
		return rep, err
	}
	if err := flush(); err != nil {
		return rep, err
	}
	return rep, st.MarkScanned(libraryID, time.Now())
}

func collect(ctx context.Context, root string, stored map[string]index.FileStamp, rep *Report) ([]job, map[string]bool, error) {
	var jobs []job
	seen := map[string]bool{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if err != nil {
			if p == root {
				return err
			}
			rep.Errors = append(rep.Errors, FileError{RelPath: relPath(root, p), Err: err})
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if p != root && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		format, ok := formatOf(d.Name())
		if !ok {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			rep.Errors = append(rep.Errors, FileError{RelPath: relPath(root, p), Err: err})
			return nil
		}
		rel := relPath(root, p)
		seen[rel] = true
		old, existed := stored[rel]
		size, mtime := info.Size(), info.ModTime().Unix()
		if existed && old.Size == size && old.MTime == mtime {
			rep.Unchanged++
			return nil
		}
		jobs = append(jobs, job{rel: rel, abs: p, format: format, size: size, mtime: mtime, existed: existed})
		return nil
	})
	return jobs, seen, err
}

func relPath(root, p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return filepath.ToSlash(p)
	}
	return filepath.ToSlash(r)
}

func parseAll(ctx context.Context, jobs []job) <-chan result {
	jobCh := make(chan job)
	results := make(chan result)
	go func() {
		defer close(jobCh)
		for _, j := range jobs {
			select {
			case jobCh <- j:
			case <-ctx.Done():
				return
			}
		}
	}()
	var wg sync.WaitGroup
	for range runtime.NumCPU() {
		wg.Go(func() {
			for j := range jobCh {
				select {
				case results <- parse(j):
				case <-ctx.Done():
					return
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	return results
}

func parse(j job) (r result) {
	title := stem(path.Base(j.rel))
	fallback := index.BookRecord{
		RelPath: j.rel, Format: j.format, Size: j.size, MTime: j.mtime, Title: title,
		Works: []index.WorkRecord{{Title: title, TreePath: title}},
	}
	r = result{rec: fallback, existed: j.existed}
	if j.format != "fb2" {
		return r
	}
	defer func() {
		if p := recover(); p != nil {
			r = result{rec: fallback, existed: j.existed, err: fmt.Errorf("parser panic: %v", p)}
		}
	}()
	b, err := fb2.ParseFile(j.abs)
	if err != nil {
		r.err = err
		return r
	}
	if b.Title == "" {
		b.Title = title
	}
	works, isCollection := extract.Works(b)
	rec := index.BookRecord{
		RelPath: j.rel, Format: j.format, Size: j.size, MTime: j.mtime,
		Title: b.Title, Authors: b.AuthorNames(), Year: b.Year, IsCollection: isCollection,
	}
	for _, w := range works {
		rec.Works = append(rec.Works, index.WorkRecord{Title: w.Title, Author: w.Author, TreePath: w.TreePath})
	}
	r.rec = rec
	return r
}
```

- [ ] **Step 4: Тести проходять (з race detector)**

Run: `go test -race ./internal/scan/`
Expected: `ok`.

- [ ] **Step 5: Commit**

```bash
git add internal/scan
git commit -m "feat(scan): incremental parallel library scanning

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 8: Інтерактивний TUI

**Files:**
- Create: `internal/tui/model.go`, `internal/tui/rank.go`, `internal/tui/view.go`, `internal/tui/run.go`
- Test: `internal/tui/model_test.go`

**Interfaces:**
- Consumes: `index.Hit`, `index.Query`, `textnorm.Normalize`, `library.Root`, `library.FilePath`, `platform.Open`, `platform.Reveal`
- Produces:
  - `type tui.Searcher interface{ Search(index.Query) ([]index.Hit, error); BookWorks(int64) ([]string, error) }` — `*index.Store` її задовольняє
  - `type tui.Actions struct{ Root func(volumeID, rootRel string) (string, bool); Open, Reveal, Copy func(string) error }`
  - `func tui.DefaultActions() Actions`
  - `func tui.New(src Searcher, act Actions, total int) Model`
  - `func tui.Run(src Searcher, total int) error`

- [ ] **Step 1: Тести**

`internal/tui/model_test.go`:
```go
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
```

- [ ] **Step 2: Перевірити падіння**

Run: `go test ./internal/tui/`
Expected: FAIL — `undefined: New` тощо (спершу `go get` залежностей з кроку 3).

- [ ] **Step 3: Реалізація**

```bash
go get charm.land/bubbletea/v2@v2.0.10 charm.land/bubbles/v2@v2.2.1 charm.land/lipgloss/v2@v2.0.6 \
  github.com/charmbracelet/x/ansi github.com/sahilm/fuzzy@v0.1.3 github.com/atotto/clipboard@v0.1.4
```

`internal/tui/model.go`:
```go
// Package tui is the interactive search screen.
package tui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"findbooks/internal/index"
	"findbooks/internal/library"
)

type Searcher interface {
	Search(q index.Query) ([]index.Hit, error)
	BookWorks(bookID int64) ([]string, error)
}

// Actions are the side effects the screen triggers; tests replace them.
type Actions struct {
	Root   func(volumeID, rootRel string) (string, bool)
	Open   func(path string) error
	Reveal func(path string) error
	Copy   func(text string) error
}

const (
	debounce       = 50 * time.Millisecond
	candidateLimit = 500
)

type searchMsg struct{ seq int }

type resultsMsg struct {
	seq  int
	hits []index.Hit
	err  error
}

type Model struct {
	src      Searcher
	act      Actions
	total    int
	input    textinput.Model
	hits     []index.Hit
	cursor   int
	seq      int
	searched bool
	online   map[string]bool    // volume id -> mounted
	toc      map[int64][]string // book id -> work titles (cache)
	status   string
	err      error
	width    int
	height   int
}

func New(src Searcher, act Actions, total int) Model {
	in := textinput.New()
	in.Prompt = "🔎 "
	in.Placeholder = "назва твору, автор або збірка…"
	in.Focus()
	return Model{src: src, act: act, total: total, input: in, online: map[string]bool{}, toc: map[int64][]string{}}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.SetWidth(max(10, msg.Width-20))
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
		m.refreshOnline()
		m.loadTOC()
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, tea.Quit
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
	before := m.input.Value()
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != before {
		return m, tea.Batch(cmd, m.queueSearch())
	}
	return m, cmd
}

// queueSearch starts a new debounce period; only the latest one searches.
func (m *Model) queueSearch() tea.Cmd {
	m.seq++
	m.status = ""
	seq := m.seq
	return tea.Tick(debounce, func(time.Time) tea.Msg { return searchMsg{seq: seq} })
}

func (m Model) searchCmd(seq int, q string) tea.Cmd {
	src := m.src
	return func() tea.Msg {
		hits, err := src.Search(index.Query{Text: q, Limit: candidateLimit})
		return resultsMsg{seq: seq, hits: hits, err: err}
	}
}

func (m *Model) move(d int) {
	if len(m.hits) == 0 {
		return
	}
	m.cursor = min(max(m.cursor+d, 0), len(m.hits)-1)
	m.loadTOC()
}

func (m Model) selected() *index.Hit {
	if m.cursor < 0 || m.cursor >= len(m.hits) {
		return nil
	}
	return &m.hits[m.cursor]
}

func (m *Model) withFile(fn func(string) error, okStatus string) {
	h := m.selected()
	if h == nil {
		return
	}
	root, ok := m.act.Root(h.VolumeID, h.RootRel)
	if !ok {
		m.status = fmt.Sprintf("Диск «%s» не підключено — підключіть його, щоб відкрити файл", h.VolumeName)
		return
	}
	if err := fn(library.FilePath(root, h.RelPath)); err != nil {
		m.status = "Помилка: " + err.Error()
		return
	}
	m.status = okStatus
}

func (m *Model) refreshOnline() {
	m.online = map[string]bool{}
	for _, h := range m.hits {
		if _, seen := m.online[h.VolumeID]; !seen {
			_, ok := m.act.Root(h.VolumeID, h.RootRel)
			m.online[h.VolumeID] = ok
		}
	}
}

func (m *Model) loadTOC() {
	h := m.selected()
	if h == nil || !h.IsCollection {
		return
	}
	if _, ok := m.toc[h.BookID]; ok {
		return
	}
	if titles, err := m.src.BookWorks(h.BookID); err == nil {
		m.toc[h.BookID] = titles
	}
}
```

`internal/tui/rank.go`:
```go
package tui

import (
	"github.com/sahilm/fuzzy"

	"findbooks/internal/index"
	"findbooks/internal/textnorm"
)

type rankSource []index.Hit

func (s rankSource) String(i int) string { return textnorm.Normalize(s[i].Title + " " + s[i].Author) }
func (s rankSource) Len() int            { return len(s) }

// rank orders FTS candidates by fuzzy closeness to the query; candidates the
// fuzzy matcher rejects keep their bm25 order after the matched ones.
func rank(query string, hits []index.Hit) []index.Hit {
	q := textnorm.Normalize(query)
	if q == "" || len(hits) < 2 {
		return hits
	}
	matches := fuzzy.FindFrom(q, rankSource(hits))
	out := make([]index.Hit, 0, len(hits))
	used := make([]bool, len(hits))
	for _, mt := range matches {
		out = append(out, hits[mt.Index])
		used[mt.Index] = true
	}
	for i, h := range hits {
		if !used[i] {
			out = append(out, h)
		}
	}
	return out
}
```

`internal/tui/view.go`:
```go
package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"findbooks/internal/index"
)

var (
	styleSelected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	styleTitle    = lipgloss.NewStyle().Bold(true)
	styleDim      = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	styleOnline   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	styleOffline  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleLabel    = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Width(8)
	styleStatus   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	styleError    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

const helpLine = "enter відкрити · ctrl+o показати у Finder · ctrl+y копіювати шлях · tab за автором · ↑/↓ вибір · esc вихід"

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m Model) render() string {
	w := m.width
	if w <= 0 {
		w = 100
	}
	var b strings.Builder
	b.WriteString(m.input.View() + "  " + styleDim.Render(fmt.Sprintf("%d з %d", len(m.hits), m.total)) + "\n\n")
	if m.err != nil {
		b.WriteString(styleError.Render("Помилка пошуку: "+m.err.Error()) + "\n")
	}
	if m.searched && len(m.hits) == 0 && strings.TrimSpace(m.input.Value()) != "" {
		b.WriteString(styleDim.Render("Нічого не знайдено") + "\n")
	}
	first, last := m.window()
	for i := first; i < last; i++ {
		b.WriteString(m.renderHit(i, w))
	}
	if h := m.selected(); h != nil {
		b.WriteString(styleDim.Render(strings.Repeat("─", w)) + "\n")
		b.WriteString(m.renderDetails(*h, w))
	}
	if m.status != "" {
		b.WriteString("\n" + styleStatus.Render(m.status) + "\n")
	}
	b.WriteString("\n" + styleDim.Render(ansi.Truncate(helpLine, w, "…")))
	return b.String()
}

// window returns the slice of hits that fits the screen around the cursor.
func (m Model) window() (int, int) {
	visible := 10
	if m.height > 0 {
		visible = max(1, (m.height-12)/2)
	}
	first := 0
	if m.cursor >= visible {
		first = m.cursor - visible + 1
	}
	return first, min(len(m.hits), first+visible)
}

func (m Model) mark(volumeID string) string {
	if m.online[volumeID] {
		return styleOnline.Render("●")
	}
	return styleOffline.Render("○")
}

func (m Model) renderHit(i, w int) string {
	h := m.hits[i]
	line := h.Title
	if h.Author != "" {
		line += " — " + h.Author
	}
	marker, style := "  ", styleTitle
	if i == m.cursor {
		marker, style = "▸ ", styleSelected
	}
	where := "    файл: " + h.RelPath
	if h.IsCollection {
		where = "    в: " + h.BookTitle
	}
	volume := "[" + h.VolumeName + " "
	room := max(10, w-lipgloss.Width(volume)-4)
	return marker + style.Render(ansi.Truncate(line, w-2, "…")) + "\n" +
		styleDim.Render(ansi.Truncate(where, room, "…")+"  "+volume) + m.mark(h.VolumeID) + styleDim.Render("]") + "\n"
}

func (m Model) renderDetails(h index.Hit, w int) string {
	book := h.BookTitle
	if h.BookYear != "" {
		book += " (" + h.BookYear + ")"
	}
	rows := [][2]string{{"Книга:", book}, {"Шлях:", h.TreePath}, {"Файл:", h.LibraryName + ": " + h.RelPath}}
	if toc := m.toc[h.BookID]; len(toc) > 1 {
		rows = append(rows, [2]string{"Зміст:", strings.Join(toc, " · ")})
	}
	var b strings.Builder
	for _, r := range rows {
		b.WriteString(styleLabel.Render(r[0]) + ansi.Truncate(r[1], max(10, w-9), "…") + "\n")
	}
	return b.String()
}
```

`internal/tui/run.go`:
```go
package tui

import (
	"github.com/atotto/clipboard"

	tea "charm.land/bubbletea/v2"

	"findbooks/internal/library"
	"findbooks/internal/platform"
)

// DefaultActions wires the screen to the real OS.
func DefaultActions() Actions {
	return Actions{Root: library.Root, Open: platform.Open, Reveal: platform.Reveal, Copy: clipboard.WriteAll}
}

// Run shows the search screen until the user quits.
func Run(src Searcher, total int) error {
	_, err := tea.NewProgram(New(src, DefaultActions(), total)).Run()
	return err
}
```

```bash
go mod tidy
gofmt -w internal/tui
```

- [ ] **Step 4: Тести проходять**

Run: `go test ./internal/tui/`
Expected: `ok`. Якщо `TestRankPutsClosestFirst` падає через порядок `sahilm/fuzzy` — перевірити, що рядок для fuzzy нормалізований (`textnorm.Normalize`) і запит теж; не змінювати тест.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum internal/tui
git commit -m "feat(tui): interactive search with debounce, fuzzy ranking and details

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 9: CLI-команди

**Files:**
- Create: `cmd/findbooks/main.go`, `internal/cli/root.go`, `internal/cli/progress.go`, `internal/cli/add.go`, `internal/cli/update.go`, `internal/cli/list.go`, `internal/cli/remove.go`, `internal/cli/search.go`
- Test: `internal/cli/cli_test.go`

**Interfaces:**
- Consumes: `index.*`, `library.Locate/Root/FilePath`, `scan.Run/Progress/Report`, `platform.DataDir`, `tui.Run`
- Produces:
  - `func cli.NewRootCmd() *cobra.Command` — глобальний прапорець `--db`
  - `func cli.Execute() int` — код виходу
  - JSON-вивід `search --json`: масив `jsonHit` (поля `title, author, tree_path, book, year, collection, library, volume, rel_path, path, online`)

- [ ] **Step 1: Тести**

`internal/cli/cli_test.go`:
```go
package cli

import (
	"bytes"
	"encoding/json"
	"errors"
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
```

- [ ] **Step 2: Перевірити падіння**

Run: `go test ./internal/cli/`
Expected: FAIL — `undefined: NewRootCmd`.

- [ ] **Step 3: Реалізація**

```bash
go get github.com/spf13/cobra@v1.10.2 golang.org/x/term@latest
```

`internal/cli/root.go`:
```go
// Package cli defines the findbooks commands.
package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/spf13/cobra"

	"findbooks/internal/index"
	"findbooks/internal/platform"
	"findbooks/internal/tui"
)

var errNoLibraries = errors.New("індекс порожній — додайте бібліотеку: findbooks add <шлях>")

// runTUI is replaced in tests.
var runTUI = func(st *index.Store, total int) error { return tui.Run(st, total) }

type app struct{ dbPath string }

func (a *app) openStore() (*index.Store, error) {
	p := a.dbPath
	if p == "" {
		dir, err := platform.DataDir()
		if err != nil {
			return nil, err
		}
		p = filepath.Join(dir, "index.db")
	}
	return index.Open(p)
}

func NewRootCmd() *cobra.Command {
	a := &app{}
	root := &cobra.Command{
		Use:           "findbooks",
		Short:         "Пошук творів у бібліотеках електронних книг, зокрема всередині збірок",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer st.Close()
			total, err := st.TotalWorks()
			if err != nil {
				return err
			}
			if total == 0 {
				return errNoLibraries
			}
			return runTUI(st, total)
		},
	}
	root.PersistentFlags().StringVar(&a.dbPath, "db", "", "файл індексу (за замовчуванням — у теці даних користувача)")
	root.AddCommand(newAddCmd(a), newUpdateCmd(a), newListCmd(a), newRemoveCmd(a), newSearchCmd(a))
	return root
}

// Execute runs the CLI and returns the process exit code.
func Execute() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := NewRootCmd().ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "помилка:", err)
		return 1
	}
	return 0
}
```

`internal/cli/progress.go`:
```go
package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"charm.land/bubbles/v2/progress"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"findbooks/internal/index"
	"findbooks/internal/scan"
)

const maxReportedErrors = 20

type progressBar struct {
	w     io.Writer
	bar   progress.Model
	on    bool
	drawn bool
	last  time.Time
}

// newProgressBar draws only when w is a terminal, so pipes and tests stay clean.
func newProgressBar(w io.Writer) *progressBar {
	f, ok := w.(*os.File)
	return &progressBar{
		w:   w,
		on:  ok && term.IsTerminal(int(f.Fd())),
		bar: progress.New(progress.WithDefaultBlend(), progress.WithWidth(40)),
	}
}

func (p *progressBar) Update(pr scan.Progress) {
	if !p.on || pr.Total == 0 {
		return
	}
	if pr.Done < pr.Total && time.Since(p.last) < 100*time.Millisecond {
		return
	}
	p.last = time.Now()
	fmt.Fprintf(p.w, "\r%s %d/%d", p.bar.ViewAs(float64(pr.Done)/float64(pr.Total)), pr.Done, pr.Total)
	p.drawn = true
}

func (p *progressBar) Done() {
	if p.drawn {
		fmt.Fprintln(p.w)
	}
}

// runScan indexes one library with a progress bar and prints the report.
func runScan(cmd *cobra.Command, st *index.Store, lib index.Library, root string) error {
	errOut := cmd.ErrOrStderr()
	fmt.Fprintf(errOut, "Індексую «%s» (%s)…\n", lib.Name, root)
	bar := newProgressBar(errOut)
	rep, err := scan.Run(cmd.Context(), st, lib.ID, root, bar.Update)
	bar.Done()
	fmt.Fprintf(errOut, "Додано: %d, оновлено: %d, видалено: %d, без змін: %d\n",
		rep.Added, rep.Updated, rep.Removed, rep.Unchanged)
	if n := len(rep.Errors); n > 0 {
		fmt.Fprintf(errOut, "Проблемні файли (%d) — проіндексовано лише за ім'ям файлу:\n", n)
		for i, e := range rep.Errors {
			if i == maxReportedErrors {
				fmt.Fprintf(errOut, "  …та ще %d\n", n-i)
				break
			}
			fmt.Fprintf(errOut, "  %s: %v\n", e.RelPath, e.Err)
		}
	}
	return err
}
```

`internal/cli/add.go`:
```go
package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"findbooks/internal/library"
)

func newAddCmd(a *app) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "add <шлях>",
		Short: "Зареєструвати бібліотеку і проіндексувати її",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			loc, err := library.Locate(args[0])
			if err != nil {
				return err
			}
			root, ok := library.Root(loc.VolumeID, loc.RootRel)
			if !ok {
				return fmt.Errorf("диск «%s» не знайдено серед підключених", loc.VolumeName)
			}
			if name == "" {
				name = filepath.Base(root)
			}
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer st.Close()
			lib, err := st.AddLibrary(name, loc.VolumeID, loc.VolumeName, loc.RootRel)
			if err != nil {
				return err
			}
			return runScan(cmd, st, lib, root)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "назва бібліотеки (за замовчуванням — ім'я теки)")
	return cmd
}
```

`internal/cli/update.go`:
```go
package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"findbooks/internal/index"
	"findbooks/internal/library"
)

func newUpdateCmd(a *app) *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "update [назва]",
		Short: "Переіндексувати змінені файли (лише на підключених дисках)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if all == (len(args) == 1) {
				return errors.New("вкажіть назву бібліотеки або --all")
			}
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer st.Close()
			var libs []index.Library
			if all {
				libs, err = st.Libraries()
			} else {
				var lib index.Library
				lib, err = st.Library(args[0])
				libs = []index.Library{lib}
			}
			if err != nil {
				return err
			}
			if len(libs) == 0 {
				return errNoLibraries
			}
			for _, lib := range libs {
				root, ok := library.Root(lib.VolumeID, lib.RootRel)
				if !ok {
					fmt.Fprintf(cmd.ErrOrStderr(), "⚠ «%s»: диск «%s» не підключено — пропускаю\n", lib.Name, lib.VolumeName)
					continue
				}
				if err := runScan(cmd, st, lib, root); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "оновити всі підключені бібліотеки")
	return cmd
}
```

`internal/cli/list.go`:
```go
package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"findbooks/internal/library"
)

func onlineMark(volumeID, rootRel string) string {
	if _, ok := library.Root(volumeID, rootRel); ok {
		return "●"
	}
	return "○"
}

func newListCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Показати бібліотеки в індексі",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer st.Close()
			libs, err := st.Libraries()
			if err != nil {
				return err
			}
			if len(libs) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), errNoLibraries)
				return nil
			}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "НАЗВА\tДИСК\tКНИГ\tТВОРІВ\tОНОВЛЕНО")
			for _, l := range libs {
				updated := "—"
				if !l.LastScan.IsZero() {
					updated = l.LastScan.Format("2006-01-02 15:04")
				}
				fmt.Fprintf(tw, "%s\t%s %s\t%d\t%d\t%s\n", l.Name, l.VolumeName, onlineMark(l.VolumeID, l.RootRel), l.Books, l.Works, updated)
			}
			return tw.Flush()
		},
	}
}
```

`internal/cli/remove.go`:
```go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRemoveCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <назва>",
		Short: "Прибрати бібліотеку з індексу (файли не чіпаються)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer st.Close()
			if err := st.RemoveLibrary(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Бібліотеку «%s» прибрано з індексу\n", args[0])
			return nil
		},
	}
}
```

`internal/cli/search.go`:
```go
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"findbooks/internal/index"
	"findbooks/internal/library"
)

type jsonHit struct {
	Title      string `json:"title"`
	Author     string `json:"author"`
	TreePath   string `json:"tree_path"`
	Book       string `json:"book"`
	Year       string `json:"year"`
	Collection bool   `json:"collection"`
	Library    string `json:"library"`
	Volume     string `json:"volume"`
	RelPath    string `json:"rel_path"`
	Path       string `json:"path,omitempty"`
	Online     bool   `json:"online"`
}

func newSearchCmd(a *app) *cobra.Command {
	var (
		author string
		limit  int
		asJSON bool
	)
	cmd := &cobra.Command{
		Use:   "search <запит>",
		Short: "Знайти твір за назвою, автором або назвою книги",
		RunE: func(cmd *cobra.Command, args []string) error {
			text := strings.Join(args, " ")
			if strings.TrimSpace(text) == "" && strings.TrimSpace(author) == "" {
				return errors.New("вкажіть запит або --author")
			}
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer st.Close()
			total, err := st.TotalWorks()
			if err != nil {
				return err
			}
			if total == 0 {
				return errNoLibraries
			}
			hits, err := st.Search(index.Query{Text: text, Author: author, Limit: limit})
			if err != nil {
				return err
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), hits)
			}
			if len(hits) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), "Нічого не знайдено")
				return nil
			}
			return writeTable(cmd.OutOrStdout(), hits)
		},
	}
	cmd.Flags().StringVar(&author, "author", "", "обмежити пошук автором")
	cmd.Flags().IntVar(&limit, "limit", 20, "максимальна кількість результатів")
	cmd.Flags().BoolVar(&asJSON, "json", false, "вивести результат як JSON")
	return cmd
}

func writeJSON(w io.Writer, hits []index.Hit) error {
	out := make([]jsonHit, 0, len(hits))
	for _, h := range hits {
		j := jsonHit{
			Title: h.Title, Author: h.Author, TreePath: h.TreePath, Book: h.BookTitle, Year: h.BookYear,
			Collection: h.IsCollection, Library: h.LibraryName, Volume: h.VolumeName, RelPath: h.RelPath,
		}
		if root, ok := library.Root(h.VolumeID, h.RootRel); ok {
			j.Path, j.Online = library.FilePath(root, h.RelPath), true
		}
		out = append(out, j)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(out)
}

func writeTable(w io.Writer, hits []index.Hit) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ТВІР\tАВТОР\tКНИГА\tФАЙЛ")
	for _, h := range hits {
		book := h.BookTitle
		if h.BookYear != "" {
			book += " (" + h.BookYear + ")"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s: %s %s\n", h.Title, h.Author, book, h.LibraryName, h.RelPath, onlineMark(h.VolumeID, h.RootRel))
	}
	return tw.Flush()
}
```

`cmd/findbooks/main.go`:
```go
package main

import (
	"os"

	"findbooks/internal/cli"
)

func main() { os.Exit(cli.Execute()) }
```

```bash
go mod tidy
gofmt -w internal/cli cmd
```

- [ ] **Step 4: Усі тести проходять**

Run: `go test -race ./...`
Expected: усі пакети `ok`.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum cmd internal/cli
git commit -m "feat(cli): add, update, list, remove, search commands and TUI entry

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 10: Перевірка на реальній бібліотеці

**Files:**
- Create: `README.md`

**Interfaces:**
- Consumes: усе вище
- Produces: зібраний бінарник, підтверджений критерій успіху, README

- [ ] **Step 1: Збірка і крос-компіляція**

```bash
cd /Users/dsv/Projects/find-books
go vet ./...
CGO_ENABLED=0 go build -o findbooks ./cmd/findbooks
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o /dev/null ./cmd/findbooks
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /dev/null ./cmd/findbooks
```
Expected: без помилок.

- [ ] **Step 2: Проіндексувати реальну бібліотеку в тимчасовий індекс**

```bash
DB="$(mktemp -d)/index.db"
time ./findbooks --db "$DB" add '/Volumes/dsvDev/Бібліотека/Советская фантастика' --name 'Советская фантастика'
./findbooks --db "$DB" list
```
Expected: ~11 тис. книг, звіт про проблемні файли; записати час індексації і кількість творів.

- [ ] **Step 3: Критерій успіху**

```bash
time ./findbooks --db "$DB" search чужие дети
./findbooks --db "$DB" search --json 'земля рожденные' | head -30
time ./findbooks --db "$DB" update 'Советская фантастика'
```
Expected: у результатах «ЧУЖИЕ ДЕТИ — Игорь Пидоренко — Румбы фантастики. 1988 год. Том II (1988)»; `search` < 0.1 с (разом із запуском процесу); повторний `update` — лише «без змін» і завершується за секунди.

- [ ] **Step 4: Перевірити якість евристики**

```bash
./findbooks --db "$DB" search --limit 30 глава
./findbooks --db "$DB" search --limit 30 часть
./findbooks --db "$DB" search --limit 30 примечания
```
Expected: майже нема «творів» з назвами на кшталт «Глава 5». Якщо є систематичні хибні твори — записати приклади (файл + заголовок) і показати користувачу, **не** змінюючи евристику без погодження.

- [ ] **Step 5: Ручна перевірка TUI**

```bash
./findbooks --db "$DB"
```
Перевірити: набір «чужие дети» → результат за ~50 мс; ↑/↓; панель «Зміст:»; `enter` відкриває файл; `ctrl+o` показує у Finder; `ctrl+y` копіює шлях (перевірити `pbpaste`); `tab` — пошук за автором; `esc` — вихід.

- [ ] **Step 6: README**

`README.md`:
````markdown
# findbooks

Пошук творів у бібліотеках електронних книг — зокрема всередині збірок, альманахів і антологій.

## Встановлення

```bash
go install findbooks/cmd/findbooks@latest   # або: go build -o findbooks ./cmd/findbooks
```

## Використання

```bash
findbooks add '/Volumes/dsvDev/Бібліотека/Советская фантастика'   # зареєструвати й проіндексувати
findbooks                                                          # інтерактивний пошук
findbooks search чужие дети                                        # пошук з командного рядка
findbooks search --author пидоренко --json
findbooks update --all                                             # доіндексувати змінене
findbooks list
findbooks remove 'Советская фантастика'
```

Індекс лежить у `~/.local/share/findbooks/index.db` (інший — `--db <файл>`). Пошук працює і тоді, коли диск з бібліотекою не підключено: такі результати позначені `○`.

У TUI: `enter` — відкрити, `ctrl+o` — показати у Finder, `ctrl+y` — копіювати шлях, `tab` — шукати за автором, `esc` — вихід.

FB2 (і `.fb2.zip`) індексуються разом зі змістом збірок; PDF, DJVU, DOC, RTF, TXT, EPUB — за ім'ям файлу і текою.
````

- [ ] **Step 7: Оновити граф знань і закомітити**

```bash
command -v graphify >/dev/null && graphify update . || true
git add README.md
git commit -m "docs: README with usage

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

- [ ] **Step 8: Звіт користувачу**

Повідомити: час індексації, кількість книг/творів/збірок, час пошуку, результат для «чужие дети», список проблемних файлів (кількість + приклади), знайдені хибні спрацювання евристики.
