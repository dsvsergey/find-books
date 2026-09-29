# Graph Report - find-books  (2026-09-29)

## Corpus Check
- 88 files · ~78,750 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 769 nodes · 1645 edges · 55 communities (38 shown, 17 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 257 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `bfa9e665`
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
- [[_COMMUNITY_Task 2 & 3 Implementation Report|Task 2 & 3 Implementation Report]]
- [[_COMMUNITY_NewRootCmd|NewRootCmd]]
- [[_COMMUNITY_newAddCmd|newAddCmd]]
- [[_COMMUNITY_newSearchCmd|newSearchCmd]]
- [[_COMMUNITY_newListCmd|newListCmd]]
- [[_COMMUNITY_.renderLibraries|.renderLibraries]]

## God Nodes (most connected - your core abstractions)
1. `libModel()` - 37 edges
2. `Errorf()` - 35 edges
3. `press()` - 33 edges
4. `drain()` - 32 edges
5. `T()` - 30 edges
6. `New()` - 27 edges
7. `Library` - 25 edges
8. `Hit` - 20 edges
9. `Run()` - 19 edges
10. `run()` - 17 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Execute()`  [INFERRED]
  cmd/findbooks/main.go → internal/cli/root.go
- `ChooseFolder()` --calls--> `Errorf()`  [INFERRED]
  internal/platform/platform_darwin.go → internal/i18n/i18n.go
- `newAddCmd()` --calls--> `runScan()`  [INFERRED]
  internal/cli/add.go → internal/cli/progress.go
- `newAddCmd()` --calls--> `closeStore()`  [INFERRED]
  internal/cli/add.go → internal/cli/root.go
- `newAddCmd()` --calls--> `exactArgs()`  [INFERRED]
  internal/cli/add.go → internal/cli/root.go

## Import Cycles
- None detected.

## Communities (55 total, 17 thin omitted)

### Community 0 - "NewRootCmd"
Cohesion: 0.15
Nodes (10): app, Command, newRemoveCmd(), closeStore(), Store, Command, newUpdateCmd(), DataDir() (+2 more)

### Community 1 - "Parse"
Cohesion: 0.06
Nodes (55): Builder, Decoder, tw, Work, Author, Book, Section, allKnown() (+47 more)

### Community 2 - "Hit"
Cohesion: 0.19
Nodes (10): Cmd, job, Msg, Model, Backend, langSavedMsg, resultsMsg, screen (+2 more)

### Community 3 - "model_test.go"
Cohesion: 0.17
Nodes (30): New(), Cmd, KeyPressMsg, Model, T, key(), newModel(), searched() (+22 more)

### Community 4 - "Task 7: Incremental Parallel Library Scanning"
Cohesion: 0.07
Nodes (40): CancelFunc, Library, Context, Store, Register(), Scan(), Store, T (+32 more)

### Community 5 - "scan.go"
Cohesion: 0.11
Nodes (40): BookRecord, FileStamp, WorkRecord, folderOf(), Store, putBook(), collect(), deletions() (+32 more)

### Community 6 - "Library"
Cohesion: 0.21
Nodes (26): Store, T, mustHits(), seeded(), TestMatchExpr(), TestSearchAuthorFilter(), TestSearchByBookTitleAndFolder(), TestSearchEmptyQuery() (+18 more)

### Community 7 - "Task 4: Extract Works from FB2 Section Tree - Report"
Cohesion: 0.15
Nodes (52): Current(), ansiStrip(), ctrlG(), Cmd, Model, T, TestCtrlGClearsStatus(), TestCtrlGIgnoredWhileChoosing() (+44 more)

### Community 8 - "Task 5 Implementation Report: SQLite FTS5 Store"
Cohesion: 0.10
Nodes (17): DB, Key, Lang, localizedError, Errorf(), Parse(), Set(), T (+9 more)

### Community 9 - "mustHits"
Cohesion: 0.44
Nodes (3): T(), Model, View

### Community 10 - "Task 1 Report: Module Scaffold and Platform Package"
Cohesion: 0.11
Nodes (32): Block, BlockKind, Hit, Query, blockKind(), parser, ParseTextFile(), Store (+24 more)

### Community 11 - "Task 6 Report: Прив'язка бібліотеки до тому"
Cohesion: 0.19
Nodes (23): langSource, copyFixture(), T, newLibrary(), run(), skipUnlessDarwin(), TestAddArgCountError(), TestAddDuplicateName() (+15 more)

### Community 12 - "Task 2: textnorm Package Implementation Report"
Cohesion: 0.22
Nodes (15): Config, captureStdout(), T, TestExecuteBadConfigLang(), TestExecuteBadEnvLang(), TestExecuteBadLangFlag(), TestExecuteLangFlagWithConfigLangShow(), Load() (+7 more)

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
Cohesion: 0.14
Nodes (13): extract, fb2, findbooks — перегляд тексту твору в TUI, i18n, index, preview (новий пакет), tui, Компоненти (+5 more)

### Community 20 - "cli_test.go"
Cohesion: 0.31
Nodes (11): FilePath(), Locate(), relToMount(), Root(), T, skipUnlessDarwin(), TestFilePathJoinsSlashPath(), TestLocateAndRootRoundTrip() (+3 more)

### Community 22 - "Model"
Cohesion: 0.15
Nodes (12): Global Constraints, Review Focus, Task 1: fb2 — розбір з текстом секцій, Task 2: extract — секція кожного твору, Task 3: index — позиція твору і відбиток файлу в Hit, Task 4: preview — завантаження тексту твору, Task 5: tui — рендер тексту і тексти i18n, Task 6: tui — екран перегляду (+4 more)

### Community 25 - "findbooks"
Cohesion: 0.17
Nodes (10): findbooks, Install, Language, Libraries in the TUI, findbooks, Бібліотеки в TUI, Використання, Встановлення (+2 more)

### Community 26 - "SDD ledger — plan: docs/superpowers/plans/2026-09-29-findbooks.md"
Cohesion: 0.17
Nodes (11): Commit, Concerns, Files changed, Graphify, GREEN, RED, Self-review findings, Task 1 report: fb2 — розбір з текстом секцій (+3 more)

### Community 27 - "task-10-brief.md"
Cohesion: 0.20
Nodes (9): Global Constraints, i18n (EN/UK) Implementation Plan, Review Focus, Task 1: Пакет `i18n` з повним каталогом, Task 2: Пакет `config`, Task 3: Локалізовані помилки в `index`, `library`, `libman`, `scan`, Task 4: CLI — вибір мови, `config lang`, тексти, Task 5: TUI — тексти і `ctrl+g` (+1 more)

### Community 42 - "runScan"
Cohesion: 0.25
Nodes (8): progressBar, Command, Model, Store, Time, Writer, newProgressBar(), runScan()

### Community 47 - "Task 2 & 3 Implementation Report"
Cohesion: 0.11
Nodes (17): Concerns, GREEN (Passing Test), GREEN (Passing Test), Implementation Details, Implementation Details, RED (Failing Test), RED (Failing Test), Self-Review Findings (+9 more)

### Community 48 - "NewRootCmd"
Cohesion: 0.17
Nodes (13): main(), Command, newConfigCmd(), exactArgs(), Execute(), exitCode(), Command, Writer (+5 more)

### Community 51 - "newAddCmd"
Cohesion: 0.25
Nodes (6): scanIncomplete, Command, newAddCmd(), scanIncompleteError(), TestExitCode(), TestScanIncompleteErrorMessage()

### Community 52 - "newSearchCmd"
Cohesion: 0.43
Nodes (6): jsonHit, Command, Writer, newSearchCmd(), writeJSON(), writeTable()

### Community 53 - "newListCmd"
Cohesion: 0.67
Nodes (3): Command, newListCmd(), onlineMark()

## Knowledge Gaps
- **121 isolated node(s):** `findbooks`, `jsonHit`, `quitTimeoutMsg`, `folderMsg`, `removedMsg` (+116 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **17 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Errorf()` connect `Task 5 Implementation Report: SQLite FTS5 Store` to `NewRootCmd`, `Parse`, `Task 7: Incremental Parallel Library Scanning`, `scan.go`, `Task 4: Extract Works from FB2 Section Tree - Report`, `mustHits`, `Task 1 Report: Module Scaffold and Platform Package`, `NewRootCmd`, `platform_darwin.go`, `newAddCmd`, `newSearchCmd`, `cli_test.go`?**
  _High betweenness centrality (0.165) - this node is a cross-community bridge._
- **Why does `New()` connect `model_test.go` to `NewRootCmd`, `Parse`, `Hit`, `Task 7: Incremental Parallel Library Scanning`, `Task 4: Extract Works from FB2 Section Tree - Report`, `Task 5 Implementation Report: SQLite FTS5 Store`, `mustHits`, `runScan`, `platform_darwin.go`, `newAddCmd`?**
  _High betweenness centrality (0.136) - this node is a cross-community bridge._
- **Why does `T()` connect `mustHits` to `NewRootCmd`, `Hit`, `model_test.go`, `Task 7: Incremental Parallel Library Scanning`, `scan.go`, `Task 4: Extract Works from FB2 Section Tree - Report`, `Task 5 Implementation Report: SQLite FTS5 Store`, `runScan`, `NewRootCmd`, `newAddCmd`, `newSearchCmd`, `newListCmd`, `cli_test.go`, `.renderLibraries`?**
  _High betweenness centrality (0.112) - this node is a cross-community bridge._
- **Are the 6 inferred relationships involving `libModel()` (e.g. with `TestCtrlGIgnoredWhileChoosing()` and `TestCtrlGIgnoredWhileJobRunning()`) actually correct?**
  _`libModel()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **Are the 31 inferred relationships involving `Errorf()` (e.g. with `.AddLibrary()` and `.ExtractVersion()`) actually correct?**
  _`Errorf()` has 31 INFERRED edges - model-reasoned connections that need verification._
- **Are the 3 inferred relationships involving `press()` (e.g. with `TestCtrlGIgnoredWhileChoosing()` and `TestCtrlGIgnoredWhileJobRunning()`) actually correct?**
  _`press()` has 3 INFERRED edges - model-reasoned connections that need verification._
- **Are the 4 inferred relationships involving `drain()` (e.g. with `TestCtrlGIgnoredWhileJobRunning()` and `TestCtrlGIgnoredWhileRemoving()`) actually correct?**
  _`drain()` has 4 INFERRED edges - model-reasoned connections that need verification._