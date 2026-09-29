# Graph Report - find-books  (2026-09-29)

## Corpus Check
- 60 files · ~57,870 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 604 nodes · 1231 edges · 38 communities (28 shown, 10 thin omitted)
- Extraction: 90% EXTRACTED · 10% INFERRED · 0% AMBIGUOUS · INFERRED: 120 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `0127821e`
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
- [[_COMMUNITY_findbooks|findbooks]]
- [[_COMMUNITY_runScan|runScan]]
- [[_COMMUNITY_TestMain|TestMain]]
- [[_COMMUNITY_TestMain|TestMain]]
- [[_COMMUNITY_Model|Model]]

## God Nodes (most connected - your core abstractions)
1. `libModel()` - 32 edges
2. `press()` - 30 edges
3. `drain()` - 28 edges
4. `Library` - 25 edges
5. `New()` - 24 edges
6. `Run()` - 17 edges
7. `Model` - 16 edges
8. `Parse()` - 15 edges
9. `setup()` - 15 edges
10. `searched()` - 15 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Execute()`  [INFERRED]
  cmd/findbooks/main.go → internal/cli/root.go
- `newAddCmd()` --calls--> `runScan()`  [INFERRED]
  internal/cli/add.go → internal/cli/progress.go
- `newAddCmd()` --calls--> `Register()`  [INFERRED]
  internal/cli/add.go → internal/libman/libman.go
- `TestScanIncompleteErrorMessage()` --calls--> `New()`  [INFERRED]
  internal/cli/cli_test.go → internal/tui/model.go
- `TestExitCode()` --calls--> `New()`  [INFERRED]
  internal/cli/cli_test.go → internal/tui/model.go

## Import Cycles
- None detected.

## Communities (38 total, 10 thin omitted)

### Community 0 - "NewRootCmd"
Cohesion: 0.08
Nodes (38): app, scanIncomplete, main(), Command, newAddCmd(), scanIncompleteError(), copyFixture(), T (+30 more)

### Community 1 - "Parse"
Cohesion: 0.12
Nodes (27): Builder, Decoder, Author, Book, parser, Section, attr(), charsetReader() (+19 more)

### Community 2 - "Hit"
Cohesion: 0.21
Nodes (9): Cmd, job, Msg, Model, Backend, resultsMsg, screen, Searcher (+1 more)

### Community 3 - "model_test.go"
Cohesion: 0.17
Nodes (30): New(), Cmd, KeyPressMsg, Model, T, key(), newModel(), searched() (+22 more)

### Community 4 - "Task 7: Incremental Parallel Library Scanning"
Cohesion: 0.08
Nodes (39): CancelFunc, Library, Context, Store, Register(), Scan(), Store, T (+31 more)

### Community 5 - "scan.go"
Cohesion: 0.11
Nodes (40): BookRecord, FileStamp, WorkRecord, folderOf(), Store, putBook(), collect(), deletions() (+32 more)

### Community 6 - "Library"
Cohesion: 0.14
Nodes (20): DB, Store, Time, initSchema(), Open(), scanLibrary(), addLib(), Store (+12 more)

### Community 7 - "Task 4: Extract Works from FB2 Section Tree - Report"
Cohesion: 0.24
Nodes (35): drain(), Cmd, Model, T, isQuit(), libModel(), press(), TestAddCanceledBeforeRegistering() (+27 more)

### Community 8 - "Task 5 Implementation Report: SQLite FTS5 Store"
Cohesion: 0.11
Nodes (18): 1. Guard against concurrent jobs (`internal/tui/libraries.go`, `internal/tui/model.go`), 2. Quitting mid-job waits for the scan to stop (`internal/tui/libraries.go`, `internal/tui/model.go`), 3. Cheap fixes (`internal/tui/libraries.go`), Concerns, Concerns, Covering Tests, Deviations, Deviations (+10 more)

### Community 9 - "mustHits"
Cohesion: 0.21
Nodes (15): Hit, Query, Store, MatchExpr(), Store, T, mustHits(), seeded() (+7 more)

### Community 10 - "Task 1 Report: Module Scaffold and Platform Package"
Cohesion: 0.12
Nodes (16): Commit Information, Concerns, Cross-Compilation Verification, Files Changed, GREEN (Passing Tests), RED (Failing Tests), Self-Review Findings, Strengths (+8 more)

### Community 11 - "Task 6 Report: Прив'язка бібліотеки до тому"
Cohesion: 0.12
Nodes (15): Comprehensive Test Suite, Concerns, Constraints Met, Full Verification Checklist (Step 5), Git Commit, Implementation, Key Implementation Details, Specific Test: TestRootCommandStartsTUI (+7 more)

### Community 12 - "Task 2: textnorm Package Implementation Report"
Cohesion: 0.13
Nodes (14): Architecture, Code Quality, Concerns, Created, Files Changed, Full test suite verification, Modified, Self-Review Findings (+6 more)

### Community 13 - "Task 3: FB2 Parser — Report"
Cohesion: 0.18
Nodes (10): findbooks — екран «Бібліотеки» в TUI, libman, platform, TUI, Компоненти, Мета, Обробка помилок, Поза межами (+2 more)

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
Cohesion: 0.11
Nodes (35): Int64, diskutilInfo(), fingerprint(), fingerprintOf(), loadVolumesLocked(), mountOf(), MountPoint(), plistStrings() (+27 more)

### Community 18 - "Task 8 Report: internal/tui"
Cohesion: 0.20
Nodes (9): Global Constraints, Libraries Screen Implementation Plan, Review Focus, Task 1: `platform.ChooseFolder`, Task 2: Пакет `libman` і перехід CLI `add`, Task 3: Екран «Бібліотеки» в TUI, Task 4: Порожній індекс відкриває TUI; README, Структура файлів (+1 more)

### Community 19 - "Task 9 Report: Cobra CLI (`internal/cli`, `cmd/findbooks/main.go`)"
Cohesion: 0.50
Nodes (3): Pre-flight scan, Progress, SDD ledger — plan: docs/superpowers/plans/2026-09-29-libraries-screen.md

### Community 20 - "cli_test.go"
Cohesion: 0.15
Nodes (20): jsonHit, Command, newListCmd(), onlineMark(), Command, Writer, newSearchCmd(), writeJSON() (+12 more)

### Community 25 - "findbooks"
Cohesion: 0.40
Nodes (4): findbooks, Бібліотеки в TUI, Використання, Встановлення

### Community 42 - "runScan"
Cohesion: 0.25
Nodes (8): progressBar, Command, Model, Store, Time, Writer, newProgressBar(), runScan()

## Knowledge Gaps
- **102 isolated node(s):** `findbooks`, `jsonHit`, `quitTimeoutMsg`, `folderMsg`, `removedMsg` (+97 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **10 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `New()` connect `model_test.go` to `NewRootCmd`, `Parse`, `Hit`, `Task 7: Incremental Parallel Library Scanning`, `Task 4: Extract Works from FB2 Section Tree - Report`, `runScan`, `platform_darwin.go`, `cli_test.go`?**
  _High betweenness centrality (0.242) - this node is a cross-community bridge._
- **Why does `Library` connect `Task 7: Incremental Parallel Library Scanning` to `runScan`, `Hit`, `scan.go`, `Library`?**
  _High betweenness centrality (0.130) - this node is a cross-community bridge._
- **Why does `Parse()` connect `Parse` to `model_test.go`?**
  _High betweenness centrality (0.084) - this node is a cross-community bridge._
- **What connects `findbooks`, `jsonHit`, `quitTimeoutMsg` to the rest of the system?**
  _102 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `NewRootCmd` be split into smaller, more focused modules?**
  _Cohesion score 0.08163265306122448 - nodes in this community are weakly interconnected._
- **Should `Parse` be split into smaller, more focused modules?**
  _Cohesion score 0.11711711711711711 - nodes in this community are weakly interconnected._
- **Should `Task 7: Incremental Parallel Library Scanning` be split into smaller, more focused modules?**
  _Cohesion score 0.0750925436277102 - nodes in this community are weakly interconnected._