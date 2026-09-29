# Graph Report - find-books  (2026-09-29)

## Corpus Check
- 67 files · ~63,913 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 631 nodes · 1019 edges · 46 communities (33 shown, 13 thin omitted)
- Extraction: 90% EXTRACTED · 10% INFERRED · 0% AMBIGUOUS · INFERRED: 101 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `faef5c8d`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_NewRootCmd|NewRootCmd]]
- [[_COMMUNITY_Parse|Parse]]
- [[_COMMUNITY_Hit|Hit]]
- [[_COMMUNITY_model_test.go|model_test.go]]
- [[_COMMUNITY_Task 7 Incremental Parallel Library Scanning|Task 7: Incremental Parallel Library Scanning]]
- [[_COMMUNITY_scan.go|scan.go]]
- [[_COMMUNITY_Library|Library]]
- [[_COMMUNITY_Task 4 Extract Works from FB2 Section Tree - Report|Task 4: Extract Works from FB2 Section Tree - Report]]
- [[_COMMUNITY_Task 5 Implementation Report SQLite FTS5 Store|Task 5 Implementation Report: SQLite FTS5 Store]]
- [[_COMMUNITY_mustHits|mustHits]]
- [[_COMMUNITY_Task 1 Report Module Scaffold and Platform Package|Task 1 Report: Module Scaffold and Platform Package]]
- [[_COMMUNITY_Task 6 Report Прив'язка бібліотеки до тому|Task 6 Report: Прив'язка бібліотеки до тому]]
- [[_COMMUNITY_Task 2 textnorm Package Implementation Report|Task 2: textnorm Package Implementation Report]]
- [[_COMMUNITY_Task 3 FB2 Parser — Report|Task 3: FB2 Parser — Report]]
- [[_COMMUNITY_Структура файлів|Структура файлів]]
- [[_COMMUNITY_findbooks — дизайн|findbooks — дизайн]]
- [[_COMMUNITY_Works|Works]]
- [[_COMMUNITY_platform_darwin.go|platform_darwin.go]]
- [[_COMMUNITY_Task 8 Report internaltui|Task 8 Report: internal/tui]]
- [[_COMMUNITY_Task 9 Report Cobra CLI (`internalcli`, `cmdfindbooksmain.go`)|Task 9 Report: Cobra CLI (`internal/cli`, `cmd/findbooks/main.go`)]]
- [[_COMMUNITY_cli_test.go|cli_test.go]]
- [[_COMMUNITY_putBook|putBook]]
- [[_COMMUNITY_Model|Model]]
- [[_COMMUNITY_findbooks|findbooks]]
- [[_COMMUNITY_SDD ledger — plan docssuperpowersplans2026-09-29-findbooks|SDD ledger — plan: docs/superpowers/plans/2026-09-29-findbooks.md]]
- [[_COMMUNITY_task-10-brief|task-10-brief.md]]
- [[_COMMUNITY_task-1-brief|task-1-brief.md]]
- [[_COMMUNITY_task-2-brief|task-2-brief.md]]
- [[_COMMUNITY_task-3-brief|task-3-brief.md]]
- [[_COMMUNITY_task-4-brief|task-4-brief.md]]
- [[_COMMUNITY_task-5-brief|task-5-brief.md]]
- [[_COMMUNITY_task-6-brief|task-6-brief.md]]
- [[_COMMUNITY_task-7-brief|task-7-brief.md]]
- [[_COMMUNITY_task-8-brief|task-8-brief.md]]
- [[_COMMUNITY_task-9-brief|task-9-brief.md]]
- [[_COMMUNITY_findbooks|findbooks]]
- [[_COMMUNITY_Task 10 report build verification, real-library check, README|Task 10 report: build verification, real-library check, README]]
- [[_COMMUNITY_Task 10 fix a ordinals above ten misclassify chapters as works|Task 10 fix a: ordinals above ten misclassify chapters as works]]
- [[_COMMUNITY_runScan|runScan]]
- [[_COMMUNITY_Task 10 fix B persistent volume fingerprint cache|Task 10 fix B: persistent volume fingerprint cache]]
- [[_COMMUNITY_TestMain|TestMain]]
- [[_COMMUNITY_TestMain|TestMain]]

## God Nodes (most connected - your core abstractions)
1. `Run()` - 17 edges
2. `New()` - 16 edges
3. `Parse()` - 15 edges
4. `setup()` - 15 edges
5. `searched()` - 15 edges
6. `Hit` - 13 edges
7. `Model` - 13 edges
8. `NewRootCmd()` - 11 edges
9. `parser` - 11 edges
10. `mustHits()` - 11 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Execute()`  [INFERRED]
  cmd/findbooks/main.go → internal/cli/root.go
- `newAddCmd()` --calls--> `runScan()`  [INFERRED]
  internal/cli/add.go → internal/cli/progress.go
- `newAddCmd()` --calls--> `Locate()`  [INFERRED]
  internal/cli/add.go → internal/library/library.go
- `newAddCmd()` --calls--> `Root()`  [INFERRED]
  internal/cli/add.go → internal/library/library.go
- `TestScanIncompleteErrorMessage()` --calls--> `New()`  [INFERRED]
  internal/cli/cli_test.go → internal/tui/model.go

## Import Cycles
- None detected.

## Communities (46 total, 13 thin omitted)

### Community 0 - "NewRootCmd"
Cohesion: 0.07
Nodes (41): app, scanIncomplete, main(), Command, newAddCmd(), scanIncompleteError(), copyFixture(), T (+33 more)

### Community 1 - "Parse"
Cohesion: 0.12
Nodes (27): Builder, Decoder, Author, Book, parser, Section, attr(), charsetReader() (+19 more)

### Community 2 - "Hit"
Cohesion: 0.21
Nodes (10): Cmd, Model, New(), DefaultActions(), Run(), Msg, Actions, resultsMsg (+2 more)

### Community 3 - "model_test.go"
Cohesion: 0.22
Nodes (24): Cmd, Model, T, key(), newModel(), searched(), settle(), TestDefaultActionsExists() (+16 more)

### Community 4 - "Task 7: Incremental Parallel Library Scanning"
Cohesion: 0.07
Nodes (28): Architecture Highlights, Code Quality, Commit, Completeness vs Brief, Concerns, Covering Test, Exact Command and Output, Files Changed (+20 more)

### Community 5 - "scan.go"
Cohesion: 0.14
Nodes (35): Context, collect(), deletions(), formatOf(), Store, hasZipFB2Suffix(), isSystemDir(), parse() (+27 more)

### Community 6 - "Library"
Cohesion: 0.15
Nodes (21): DB, Library, Store, Time, initSchema(), Open(), scanLibrary(), addLib() (+13 more)

### Community 7 - "Task 4: Extract Works from FB2 Section Tree - Report"
Cohesion: 0.08
Nodes (23): Build Verification, Changes Made, Code Quality, Commit, Commits, Completeness vs Brief, Concerns, Data Structure (+15 more)

### Community 8 - "Task 5 Implementation Report: SQLite FTS5 Store"
Cohesion: 0.09
Nodes (22): Adherence to Constraints, Build Verification (CGO_ENABLED=0), Code Quality, Commit, Completeness vs Brief, Concerns, Dependencies Added, Files Created (+14 more)

### Community 9 - "mustHits"
Cohesion: 0.11
Nodes (23): jsonHit, Hit, Query, Command, Writer, newSearchCmd(), writeJSON(), writeTable() (+15 more)

### Community 10 - "Task 1 Report: Module Scaffold and Platform Package"
Cohesion: 0.10
Nodes (20): 1. Module Initialization, 2. Core Platform Package (`internal/platform/`), 3. Tests, ✓ Build Compliance, CGO_ENABLED=0 Compatibility, ✓ Code Quality, Cross-Compilation Verification, Files Changed (+12 more)

### Community 11 - "Task 6 Report: Прив'язка бібліотеки до тому"
Cohesion: 0.11
Nodes (18): Build Verification, Code Quality, Completeness, Concerns, Constraints, Files Created, Files Modified, Git Commit (+10 more)

### Community 12 - "Task 2: textnorm Package Implementation Report"
Cohesion: 0.11
Nodes (17): Build Verification, Code Quality, Commit, Completeness, Concerns, Files Created/Modified, Implementation Details, No Extras (YAGNI) (+9 more)

### Community 13 - "Task 3: FB2 Parser — Report"
Cohesion: 0.11
Nodes (17): Build Verification, Completeness, Concerns, Exports Available to Later Tasks, Files Changed, GREEN: Tests Pass After Implementation, Implementation, No Extras (YAGNI) (+9 more)

### Community 14 - "Структура файлів"
Cohesion: 0.12
Nodes (15): findbooks Implementation Plan, Global Constraints, Review Focus, Task 10: Перевірка на реальній бібліотеці, Task 1: Каркас модуля і пакет platform, Task 2: Нормалізація тексту, Task 3: Парсер FB2, Task 4: Евристика вибору творів (+7 more)

### Community 15 - "findbooks — дизайн"
Cohesion: 0.13
Nodes (14): findbooks — дизайн, TUI, Архітектура, Евристика вибору творів (FB2), Команди, Мета, Модель даних, Нормалізація тексту (+6 more)

### Community 16 - "Works"
Cohesion: 0.14
Nodes (18): Work, allKnown(), containsAll(), hasGivenName(), isNoise(), matchAuthor(), newWork(), T (+10 more)

### Community 17 - "platform_darwin.go"
Cohesion: 0.13
Nodes (29): Int64, diskutilInfo(), fingerprint(), fingerprintOf(), loadVolumesLocked(), mountOf(), MountPoint(), plistStrings() (+21 more)

### Community 18 - "Task 8 Report: internal/tui"
Cohesion: 0.14
Nodes (13): Concerns, Concerns, Covering test, Deviations from the brief, Exact commands and output, Files changed, Files changed, Fix round 1 (review finding: synchronous `Root` call inside `Update`) (+5 more)

### Community 19 - "Task 9 Report: Cobra CLI (`internal/cli`, `cmd/findbooks/main.go`)"
Cohesion: 0.15
Nodes (12): 1. Argument-count errors must be in Ukrainian, 2. `Store.Close` errors must not be swallowed on write commands, Concerns, Deviations from the brief, Files changed, Files changed (fix round 1), Fix round 1 (controller-ruled review findings), Self-review (+4 more)

### Community 20 - "cli_test.go"
Cohesion: 0.31
Nodes (11): FilePath(), Locate(), relToMount(), Root(), T, skipUnlessDarwin(), TestFilePathJoinsSlashPath(), TestLocateAndRootRoundTrip() (+3 more)

### Community 21 - "putBook"
Cohesion: 0.31
Nodes (7): BookRecord, FileStamp, WorkRecord, folderOf(), Store, putBook(), Tx

### Community 22 - "Model"
Cohesion: 0.18
Nodes (10): 1. Pre-test state of ~/.local/share/findbooks/, 2. Run tests, 3. Post-test state of ~/.local/share/findbooks/, 4. Run full test suite, 5. Format check, Changes Made, Commit, Summary (+2 more)

### Community 25 - "findbooks"
Cohesion: 0.50
Nodes (3): findbooks, Використання, Встановлення

### Community 26 - "SDD ledger — plan: docs/superpowers/plans/2026-09-29-findbooks.md"
Cohesion: 0.50
Nodes (3): Pre-flight scan, Progress, SDD ledger — plan: docs/superpowers/plans/2026-09-29-findbooks.md

### Community 40 - "Task 10 report: build verification, real-library check, README"
Cohesion: 0.17
Nodes (11): 5 random books flagged as collections, with their work lists (`sqlite3`, read-only; cross-check with `search --json` is redundant here since this is a direct, complete listing rather than a ranked/limited search), Concerns for the controller/user, Files changed, Step 1 — Build and cross-compile, Step 2 — Index the real library, Step 3 — Success criterion, Step 4 — Heuristic quality, Step 5 — TUI (skipped, per ruling 5) (+3 more)

### Community 41 - "Task 10 fix a: ordinals above ten misclassify chapters as works"
Cohesion: 0.25
Nodes (7): Commit, Files changed, Full suite and formatting, GREEN, RED, Task 10 fix a: ordinals above ten misclassify chapters as works, What changed

### Community 42 - "runScan"
Cohesion: 0.29
Nodes (8): progressBar, Command, Model, Store, Time, Writer, newProgressBar(), runScan()

### Community 43 - "Task 10 fix B: persistent volume fingerprint cache"
Cohesion: 0.22
Nodes (8): Concerns, Cross builds, full test suite, gofmt, Fsid stability (design precondition), GREEN: `go test -race -v ./internal/platform/ ./internal/library/`, RED (new tests before implementation), Task 10 fix B: persistent volume fingerprint cache, Timing: `CGO_ENABLED=0 go build -o findbooks ./cmd/findbooks`, then `time ./findbooks --db "$DB" search чужие дети` x3 (XDG_DATA_HOME=scratchpad/fb-real/xdg), What changed (internal/platform only)

## Knowledge Gaps
- **203 isolated node(s):** `findbooks`, `jsonHit`, `searchMsg`, `Pre-flight scan`, `Progress` (+198 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **13 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `New()` connect `Hit` to `NewRootCmd`, `Parse`, `model_test.go`, `mustHits`, `runScan`, `platform_darwin.go`?**
  _High betweenness centrality (0.132) - this node is a cross-community bridge._
- **Why does `Hit` connect `mustHits` to `Works`, `Hit`, `model_test.go`, `scan.go`?**
  _High betweenness centrality (0.075) - this node is a cross-community bridge._
- **Why does `Parse()` connect `Parse` to `Hit`?**
  _High betweenness centrality (0.065) - this node is a cross-community bridge._
- **Are the 9 inferred relationships involving `Run()` (e.g. with `TestRunCanceled()` and `TestRunDetectsChangesAndRemovals()`) actually correct?**
  _`Run()` has 9 INFERRED edges - model-reasoned connections that need verification._
- **Are the 12 inferred relationships involving `New()` (e.g. with `TestExitCode()` and `TestScanIncompleteErrorMessage()`) actually correct?**
  _`New()` has 12 INFERRED edges - model-reasoned connections that need verification._
- **Are the 8 inferred relationships involving `Parse()` (e.g. with `New()` and `TestParseGarbage()`) actually correct?**
  _`Parse()` has 8 INFERRED edges - model-reasoned connections that need verification._
- **What connects `findbooks`, `jsonHit`, `searchMsg` to the rest of the system?**
  _203 weakly-connected nodes found - possible documentation gaps or missing edges._