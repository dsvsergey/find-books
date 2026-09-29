# Graph Report - find-books  (2026-09-29)

## Corpus Check
- 84 files · ~73,792 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 753 nodes · 1541 edges · 50 communities (36 shown, 14 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 237 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `92ce8296`
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
- [[_COMMUNITY_task-5-brief|task-5-brief.md]]

## God Nodes (most connected - your core abstractions)
1. `libModel()` - 37 edges
2. `Errorf()` - 34 edges
3. `press()` - 33 edges
4. `drain()` - 32 edges
5. `T()` - 30 edges
6. `New()` - 27 edges
7. `Library` - 25 edges
8. `Run()` - 19 edges
9. `run()` - 17 edges
10. `Model` - 17 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Execute()`  [INFERRED]
  cmd/findbooks/main.go → internal/cli/root.go
- `ChooseFolder()` --calls--> `Errorf()`  [INFERRED]
  internal/platform/platform_darwin.go → internal/i18n/i18n.go
- `newAddCmd()` --calls--> `runScan()`  [INFERRED]
  internal/cli/add.go → internal/cli/progress.go
- `newAddCmd()` --calls--> `Register()`  [INFERRED]
  internal/cli/add.go → internal/libman/libman.go
- `TestConfigLang()` --calls--> `run()`  [INFERRED]
  internal/cli/lang_test.go → internal/cli/cli_test.go

## Import Cycles
- None detected.

## Communities (50 total, 14 thin omitted)

### Community 0 - "NewRootCmd"
Cohesion: 0.06
Nodes (52): app, jsonHit, scanIncomplete, main(), Command, newAddCmd(), scanIncompleteError(), copyFixture() (+44 more)

### Community 1 - "Parse"
Cohesion: 0.12
Nodes (27): Builder, Decoder, Author, Book, parser, Section, attr(), charsetReader() (+19 more)

### Community 2 - "Hit"
Cohesion: 0.11
Nodes (32): Current(), ansiStrip(), ctrlG(), Cmd, Model, T, TestCtrlGClearsStatus(), TestCtrlGIgnoredWhileChoosing() (+24 more)

### Community 3 - "model_test.go"
Cohesion: 0.24
Nodes (24): Cmd, KeyPressMsg, Model, T, key(), newModel(), searched(), settle() (+16 more)

### Community 4 - "Task 7: Incremental Parallel Library Scanning"
Cohesion: 0.07
Nodes (40): CancelFunc, Library, Context, Store, Register(), Scan(), Store, T (+32 more)

### Community 5 - "scan.go"
Cohesion: 0.11
Nodes (40): BookRecord, FileStamp, WorkRecord, folderOf(), Store, putBook(), collect(), deletions() (+32 more)

### Community 6 - "Library"
Cohesion: 0.14
Nodes (30): Hit, Query, Store, MatchExpr(), Store, T, mustHits(), seeded() (+22 more)

### Community 7 - "Task 4: Extract Works from FB2 Section Tree - Report"
Cohesion: 0.24
Nodes (35): drain(), Cmd, Model, T, isQuit(), libModel(), press(), TestAddCanceledBeforeRegistering() (+27 more)

### Community 8 - "Task 5 Implementation Report: SQLite FTS5 Store"
Cohesion: 0.10
Nodes (17): DB, Key, Lang, localizedError, Errorf(), Parse(), Set(), T (+9 more)

### Community 9 - "mustHits"
Cohesion: 0.11
Nodes (18): Catalog Data, Code Quality, Concerns, Correctness, Files Changed, GREEN Phase, Implementation Summary, RED Phase (+10 more)

### Community 10 - "Task 1 Report: Module Scaffold and Platform Package"
Cohesion: 0.12
Nodes (16): Commit, Concerns, Config Struct, Features, Files Changed, Files Created, Implementation Details, Key Design Decisions (+8 more)

### Community 11 - "Task 6 Report: Прив'язка бібліотеки до тому"
Cohesion: 0.14
Nodes (18): Work, allKnown(), containsAll(), hasGivenName(), isNoise(), matchAuthor(), newWork(), T (+10 more)

### Community 12 - "Task 2: textnorm Package Implementation Report"
Cohesion: 0.12
Nodes (27): langSource, Config, captureStdout(), T, TestExecuteBadConfigLang(), TestExecuteBadEnvLang(), TestExecuteBadLangFlag(), TestExecuteLangFlagWithConfigLangShow() (+19 more)

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
Cohesion: 0.13
Nodes (14): Binary probes, Commit list, Final verification, findbooks i18n — final fix wave report, Item 10 — Spec amendment, Item 1 — README gloss, Item 2 — `config` without a known subcommand, Item 3 — Serialize language saves (+6 more)

### Community 20 - "cli_test.go"
Cohesion: 0.26
Nodes (13): Writer, writeJSON(), FilePath(), Locate(), relToMount(), Root(), T, skipUnlessDarwin() (+5 more)

### Community 22 - "Model"
Cohesion: 0.15
Nodes (12): Concerns, errors.Is verification, `internal/index/store.go`, `internal/libman/libman.go`, `internal/library/library.go`, `internal/scan/scan.go`, Task 3 report: localized errors in index, library, libman, scan, TDD evidence (+4 more)

### Community 25 - "findbooks"
Cohesion: 0.17
Nodes (10): findbooks, Install, Language, Libraries in the TUI, findbooks, Бібліотеки в TUI, Використання, Встановлення (+2 more)

### Community 26 - "SDD ledger — plan: docs/superpowers/plans/2026-09-29-findbooks.md"
Cohesion: 0.17
Nodes (11): Concerns, Concerns, Covering tests and results, Deviations, Deviations from the brief, Fix round 1 (controller rulings from review), Task 5 report: TUI — тексти і `ctrl+g`, TDD evidence (+3 more)

### Community 27 - "task-10-brief.md"
Cohesion: 0.20
Nodes (9): Global Constraints, i18n (EN/UK) Implementation Plan, Review Focus, Task 1: Пакет `i18n` з повним каталогом, Task 2: Пакет `config`, Task 3: Локалізовані помилки в `index`, `library`, `libman`, `scan`, Task 4: CLI — вибір мови, `config lang`, тексти, Task 5: TUI — тексти і `ctrl+g` (+1 more)

### Community 28 - "task-1-brief.md"
Cohesion: 0.25
Nodes (7): Commit, Concerns, Final check — full output, Fix round 1, Task 6 report — README (EN) + README.uk.md, Verification against code, What was written

### Community 29 - "task-2-brief.md"
Cohesion: 0.29
Nodes (6): Concerns, Deviations from the brief, Task 4 Report — CLI language selection, `config lang`, translated texts, TDD evidence, Tests and results, What changed

### Community 32 - "SDD ledger — plan: docs/superpowers/plans/2026-09-29-i18n.md"
Cohesion: 0.50
Nodes (3): Pre-flight scan, Progress, SDD ledger — plan: docs/superpowers/plans/2026-09-29-i18n.md

### Community 42 - "runScan"
Cohesion: 0.25
Nodes (8): progressBar, Command, Model, Store, Time, Writer, newProgressBar(), runScan()

## Knowledge Gaps
- **147 isolated node(s):** `findbooks`, `jsonHit`, `quitTimeoutMsg`, `folderMsg`, `removedMsg` (+142 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **14 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Errorf()` connect `Task 5 Implementation Report: SQLite FTS5 Store` to `NewRootCmd`, `Parse`, `Task 7: Incremental Parallel Library Scanning`, `scan.go`, `Task 4: Extract Works from FB2 Section Tree - Report`, `platform_darwin.go`, `cli_test.go`?**
  _High betweenness centrality (0.130) - this node is a cross-community bridge._
- **Why does `New()` connect `Hit` to `NewRootCmd`, `Parse`, `model_test.go`, `Task 7: Incremental Parallel Library Scanning`, `Task 4: Extract Works from FB2 Section Tree - Report`, `Task 5 Implementation Report: SQLite FTS5 Store`, `runScan`, `platform_darwin.go`?**
  _High betweenness centrality (0.129) - this node is a cross-community bridge._
- **Why does `T()` connect `NewRootCmd` to `Hit`, `Task 7: Incremental Parallel Library Scanning`, `scan.go`, `Task 5 Implementation Report: SQLite FTS5 Store`, `runScan`, `cli_test.go`?**
  _High betweenness centrality (0.112) - this node is a cross-community bridge._
- **Are the 6 inferred relationships involving `libModel()` (e.g. with `TestCtrlGIgnoredWhileChoosing()` and `TestCtrlGIgnoredWhileJobRunning()`) actually correct?**
  _`libModel()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **Are the 30 inferred relationships involving `Errorf()` (e.g. with `.AddLibrary()` and `.ExtractVersion()`) actually correct?**
  _`Errorf()` has 30 INFERRED edges - model-reasoned connections that need verification._
- **Are the 3 inferred relationships involving `press()` (e.g. with `TestCtrlGIgnoredWhileChoosing()` and `TestCtrlGIgnoredWhileJobRunning()`) actually correct?**
  _`press()` has 3 INFERRED edges - model-reasoned connections that need verification._
- **Are the 4 inferred relationships involving `drain()` (e.g. with `TestCtrlGIgnoredWhileJobRunning()` and `TestCtrlGIgnoredWhileRemoving()`) actually correct?**
  _`drain()` has 4 INFERRED edges - model-reasoned connections that need verification._