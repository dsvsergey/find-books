# Graph Report - find-books  (2026-09-29)

## Corpus Check
- 78 files · ~65,947 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 691 nodes · 1435 edges · 51 communities (38 shown, 13 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 213 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `ae162fc4`
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
- [[_COMMUNITY_DataDir|DataDir]]
- [[_COMMUNITY_SDD ledger — plan docssuperpowersplans2026-09-29-i18n|SDD ledger — plan: docs/superpowers/plans/2026-09-29-i18n.md]]
- [[_COMMUNITY_TestMain|TestMain]]
- [[_COMMUNITY_TestMain|TestMain]]
- [[_COMMUNITY_TestMain|TestMain]]
- [[_COMMUNITY_TestMain|TestMain]]
- [[_COMMUNITY_findbooks|findbooks]]
- [[_COMMUNITY_task-1-brief|task-1-brief.md]]
- [[_COMMUNITY_task-2-brief|task-2-brief.md]]
- [[_COMMUNITY_runScan|runScan]]
- [[_COMMUNITY_task-3-brief|task-3-brief.md]]
- [[_COMMUNITY_TestMain|TestMain]]
- [[_COMMUNITY_TestMain|TestMain]]
- [[_COMMUNITY_task-4-brief|task-4-brief.md]]
- [[_COMMUNITY_Model|Model]]
- [[_COMMUNITY_task-5-brief|task-5-brief.md]]

## God Nodes (most connected - your core abstractions)
1. `Errorf()` - 34 edges
2. `libModel()` - 34 edges
3. `press()` - 31 edges
4. `T()` - 30 edges
5. `drain()` - 29 edges
6. `Library` - 25 edges
7. `New()` - 25 edges
8. `Run()` - 19 edges
9. `Model` - 17 edges
10. `Parse()` - 16 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Execute()`  [INFERRED]
  cmd/findbooks/main.go → internal/cli/root.go
- `configLang()` --calls--> `Load()`  [INFERRED]
  internal/cli/lang.go → internal/config/config.go
- `ChooseFolder()` --calls--> `Errorf()`  [INFERRED]
  internal/platform/platform_darwin.go → internal/i18n/i18n.go
- `newAddCmd()` --calls--> `runScan()`  [INFERRED]
  internal/cli/add.go → internal/cli/progress.go
- `newAddCmd()` --calls--> `T()`  [INFERRED]
  internal/cli/add.go → internal/i18n/i18n.go

## Import Cycles
- None detected.

## Communities (51 total, 13 thin omitted)

### Community 0 - "NewRootCmd"
Cohesion: 0.15
Nodes (19): app, Command, newAddCmd(), Command, newListCmd(), onlineMark(), Command, newRemoveCmd() (+11 more)

### Community 1 - "Parse"
Cohesion: 0.08
Nodes (40): Builder, Decoder, Work, Author, Book, parser, Section, allKnown() (+32 more)

### Community 2 - "Hit"
Cohesion: 0.07
Nodes (25): jsonHit, Hit, Query, Command, Writer, newSearchCmd(), writeJSON(), writeTable() (+17 more)

### Community 3 - "model_test.go"
Cohesion: 0.17
Nodes (30): New(), Cmd, KeyPressMsg, Model, T, key(), newModel(), searched() (+22 more)

### Community 4 - "Task 7: Incremental Parallel Library Scanning"
Cohesion: 0.13
Nodes (23): Library, Context, Store, Register(), Scan(), Store, T, newLibraryDir() (+15 more)

### Community 5 - "scan.go"
Cohesion: 0.11
Nodes (40): BookRecord, FileStamp, WorkRecord, folderOf(), Store, putBook(), collect(), deletions() (+32 more)

### Community 6 - "Library"
Cohesion: 0.21
Nodes (25): Store, T, mustHits(), seeded(), TestMatchExpr(), TestSearchAuthorFilter(), TestSearchByBookTitleAndFolder(), TestSearchEmptyQuery() (+17 more)

### Community 7 - "Task 4: Extract Works from FB2 Section Tree - Report"
Cohesion: 0.24
Nodes (35): drain(), Cmd, Model, T, isQuit(), libModel(), press(), TestAddCanceledBeforeRegistering() (+27 more)

### Community 8 - "Task 5 Implementation Report: SQLite FTS5 Store"
Cohesion: 0.12
Nodes (21): Key, Lang, localizedError, Current(), Parse(), Set(), T, TestCatalogComplete() (+13 more)

### Community 9 - "mustHits"
Cohesion: 0.11
Nodes (18): Catalog Data, Code Quality, Concerns, Correctness, Files Changed, GREEN Phase, Implementation Summary, RED Phase (+10 more)

### Community 10 - "Task 1 Report: Module Scaffold and Platform Package"
Cohesion: 0.12
Nodes (16): Commit, Concerns, Config Struct, Features, Files Changed, Files Created, Implementation Details, Key Design Decisions (+8 more)

### Community 11 - "Task 6 Report: Прив'язка бібліотеки до тому"
Cohesion: 0.21
Nodes (7): DB, Errorf(), Store, Time, initSchema(), scanLibrary(), versionError()

### Community 12 - "Task 2: textnorm Package Implementation Report"
Cohesion: 0.31
Nodes (11): Config, Command, newConfigCmd(), Load(), Path(), Save(), T, reset() (+3 more)

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
Nodes (13): CLI, config, findbooks — вибір мови інтерфейсу (EN/UK), i18n, README, TUI, Вибір мови (cli), Глибші пакети (+5 more)

### Community 17 - "platform_darwin.go"
Cohesion: 0.11
Nodes (36): Int64, ChooseFolder(), diskutilInfo(), fingerprint(), fingerprintOf(), loadVolumesLocked(), mountOf(), MountPoint() (+28 more)

### Community 18 - "Task 8 Report: internal/tui"
Cohesion: 0.20
Nodes (9): Global Constraints, Libraries Screen Implementation Plan, Review Focus, Task 1: `platform.ChooseFolder`, Task 2: Пакет `libman` і перехід CLI `add`, Task 3: Екран «Бібліотеки» в TUI, Task 4: Порожній індекс відкриває TUI; README, Структура файлів (+1 more)

### Community 19 - "Task 9 Report: Cobra CLI (`internal/cli`, `cmd/findbooks/main.go`)"
Cohesion: 0.44
Nodes (13): copyFixture(), T, newLibrary(), run(), skipUnlessDarwin(), TestAddArgCountError(), TestAddDuplicateName(), TestAddScanFailureHintsUpdate() (+5 more)

### Community 20 - "cli_test.go"
Cohesion: 0.31
Nodes (11): FilePath(), Locate(), relToMount(), Root(), T, skipUnlessDarwin(), TestFilePathJoinsSlashPath(), TestLocateAndRootRoundTrip() (+3 more)

### Community 22 - "Model"
Cohesion: 0.15
Nodes (12): Concerns, errors.Is verification, `internal/index/store.go`, `internal/libman/libman.go`, `internal/library/library.go`, `internal/scan/scan.go`, Task 3 report: localized errors in index, library, libman, scan, TDD evidence (+4 more)

### Community 25 - "findbooks"
Cohesion: 0.40
Nodes (4): findbooks, Бібліотеки в TUI, Використання, Встановлення

### Community 26 - "SDD ledger — plan: docs/superpowers/plans/2026-09-29-findbooks.md"
Cohesion: 0.18
Nodes (7): scanIncomplete, main(), scanIncompleteError(), TestExitCode(), Execute(), exitCode(), Writer

### Community 27 - "task-10-brief.md"
Cohesion: 0.20
Nodes (9): Global Constraints, i18n (EN/UK) Implementation Plan, Review Focus, Task 1: Пакет `i18n` з повним каталогом, Task 2: Пакет `config`, Task 3: Локалізовані помилки в `index`, `library`, `libman`, `scan`, Task 4: CLI — вибір мови, `config lang`, тексти, Task 5: TUI — тексти і `ctrl+g` (+1 more)

### Community 28 - "task-1-brief.md"
Cohesion: 0.31
Nodes (7): configLang(), resolveLang(), T, TestConfigLang(), TestEnglishByDefault(), TestLangFlagSwitchesOutput(), TestResolveLang()

### Community 29 - "task-2-brief.md"
Cohesion: 0.29
Nodes (6): Concerns, Deviations from the brief, Task 4 Report — CLI language selection, `config lang`, translated texts, TDD evidence, Tests and results, What changed

### Community 31 - "DataDir"
Cohesion: 0.33
Nodes (3): DataDir(), T, TestDataDirHonorsXDG()

### Community 32 - "SDD ledger — plan: docs/superpowers/plans/2026-09-29-i18n.md"
Cohesion: 0.50
Nodes (3): Pre-flight scan, Progress, SDD ledger — plan: docs/superpowers/plans/2026-09-29-i18n.md

### Community 42 - "runScan"
Cohesion: 0.25
Nodes (8): progressBar, Command, Model, Store, Time, Writer, newProgressBar(), runScan()

### Community 47 - "Model"
Cohesion: 0.10
Nodes (21): CancelFunc, T(), Cmd, Context, job, KeyPressMsg, Msg, Model (+13 more)

## Knowledge Gaps
- **115 isolated node(s):** `findbooks`, `jsonHit`, `quitTimeoutMsg`, `folderMsg`, `removedMsg` (+110 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **13 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Errorf()` connect `Task 6 Report: Прив'язка бібліотеки до тому` to `NewRootCmd`, `Parse`, `Hit`, `Task 7: Incremental Parallel Library Scanning`, `scan.go`, `Task 4: Extract Works from FB2 Section Tree - Report`, `Task 5 Implementation Report: SQLite FTS5 Store`, `Model`, `platform_darwin.go`, `cli_test.go`, `SDD ledger — plan: docs/superpowers/plans/2026-09-29-findbooks.md`?**
  _High betweenness centrality (0.150) - this node is a cross-community bridge._
- **Why does `New()` connect `model_test.go` to `Parse`, `Hit`, `Task 7: Incremental Parallel Library Scanning`, `Task 4: Extract Works from FB2 Section Tree - Report`, `Task 5 Implementation Report: SQLite FTS5 Store`, `runScan`, `Model`, `platform_darwin.go`, `Task 9 Report: Cobra CLI (`internal/cli`, `cmd/findbooks/main.go`)`, `SDD ledger — plan: docs/superpowers/plans/2026-09-29-findbooks.md`, `DataDir`?**
  _High betweenness centrality (0.143) - this node is a cross-community bridge._
- **Why does `T()` connect `Model` to `NewRootCmd`, `Hit`, `model_test.go`, `scan.go`, `Task 5 Implementation Report: SQLite FTS5 Store`, `runScan`, `Task 6 Report: Прив'язка бібліотеки до тому`, `Task 2: textnorm Package Implementation Report`, `cli_test.go`, `SDD ledger — plan: docs/superpowers/plans/2026-09-29-findbooks.md`?**
  _High betweenness centrality (0.126) - this node is a cross-community bridge._
- **Are the 30 inferred relationships involving `Errorf()` (e.g. with `.AddLibrary()` and `.ExtractVersion()`) actually correct?**
  _`Errorf()` has 30 INFERRED edges - model-reasoned connections that need verification._
- **Are the 3 inferred relationships involving `libModel()` (e.g. with `TestCtrlGIgnoredWhileChoosing()` and `TestCtrlGOnLibrariesScreen()`) actually correct?**
  _`libModel()` has 3 INFERRED edges - model-reasoned connections that need verification._
- **Are the 26 inferred relationships involving `T()` (e.g. with `.Error()` and `newAddCmd()`) actually correct?**
  _`T()` has 26 INFERRED edges - model-reasoned connections that need verification._
- **What connects `findbooks`, `jsonHit`, `quitTimeoutMsg` to the rest of the system?**
  _115 weakly-connected nodes found - possible documentation gaps or missing edges._