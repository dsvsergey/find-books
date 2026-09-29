# Graph Report - find-books  (2026-09-29)

## Corpus Check
- 96 files · ~86,659 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 867 nodes · 1821 edges · 60 communities (42 shown, 18 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 292 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `0124d014`
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
- [[_COMMUNITY_mustHits|mustHits]]
- [[_COMMUNITY_fb2_test.go|fb2_test.go]]
- [[_COMMUNITY_Task 7 документація, перевірка, граф — Report|Task 7: документація, перевірка, граф — Report]]
- [[_COMMUNITY_extract_test.go|extract_test.go]]
- [[_COMMUNITY_Block|Block]]
- [[_COMMUNITY_ParseText|ParseText]]
- [[_COMMUNITY_rank|rank]]
- [[_COMMUNITY_TestNormalize|TestNormalize]]

## God Nodes (most connected - your core abstractions)
1. `libModel()` - 37 edges
2. `T()` - 36 edges
3. `Errorf()` - 35 edges
4. `press()` - 33 edges
5. `drain()` - 32 edges
6. `New()` - 29 edges
7. `Library` - 25 edges
8. `Hit` - 23 edges
9. `searched()` - 22 edges
10. `Model` - 20 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Execute()`  [INFERRED]
  cmd/findbooks/main.go → internal/cli/root.go
- `ChooseFolder()` --calls--> `Errorf()`  [INFERRED]
  internal/platform/platform_darwin.go → internal/i18n/i18n.go
- `newAddCmd()` --calls--> `runScan()`  [INFERRED]
  internal/cli/add.go → internal/cli/progress.go
- `newAddCmd()` --calls--> `T()`  [INFERRED]
  internal/cli/add.go → internal/i18n/i18n.go
- `newAddCmd()` --calls--> `Register()`  [INFERRED]
  internal/cli/add.go → internal/libman/libman.go

## Import Cycles
- None detected.

## Communities (60 total, 18 thin omitted)

### Community 0 - "NewRootCmd"
Cohesion: 0.24
Nodes (13): Work, Section, allKnown(), containsAll(), hasGivenName(), isNoise(), matchAuthor(), newWork() (+5 more)

### Community 1 - "Parse"
Cohesion: 0.19
Nodes (11): Builder, Decoder, Author, Book, attr(), charsetReader(), collapse(), parser (+3 more)

### Community 2 - "Hit"
Cohesion: 0.06
Nodes (34): CancelFunc, T(), Cmd, Context, job, KeyPressMsg, Msg, Model (+26 more)

### Community 3 - "model_test.go"
Cohesion: 0.15
Nodes (40): Cmd, KeyPressMsg, Model, T, key(), newModel(), searched(), settle() (+32 more)

### Community 4 - "Task 7: Incremental Parallel Library Scanning"
Cohesion: 0.13
Nodes (23): Library, Context, Store, Register(), Scan(), Store, T, newLibraryDir() (+15 more)

### Community 5 - "scan.go"
Cohesion: 0.11
Nodes (40): BookRecord, FileStamp, WorkRecord, folderOf(), Store, putBook(), collect(), deletions() (+32 more)

### Community 6 - "Library"
Cohesion: 0.15
Nodes (21): DB, Errorf(), Store, Time, initSchema(), Open(), scanLibrary(), addLib() (+13 more)

### Community 7 - "Task 4: Extract Works from FB2 Section Tree - Report"
Cohesion: 0.15
Nodes (52): ansiStrip(), ctrlG(), Cmd, Model, T, TestCtrlGClearsStatus(), TestCtrlGIgnoredWhileChoosing(), TestCtrlGIgnoredWhileJobRunning() (+44 more)

### Community 8 - "Task 5 Implementation Report: SQLite FTS5 Store"
Cohesion: 0.07
Nodes (28): langSource, Key, Lang, localizedError, resolveLang(), Current(), Parse(), Set() (+20 more)

### Community 9 - "mustHits"
Cohesion: 0.11
Nodes (18): Concerns, Files changed, Files changed (this round), Finding (from task review, Important, plan-mandated), Fix applied (per controller ruling), Fix round 1 — panic-recovery path untested, GREEN, GREEN (+10 more)

### Community 10 - "Task 1 Report: Module Scaffold and Platform Package"
Cohesion: 0.24
Nodes (23): Hit, changed(), Load(), Supported(), chuzhieHit(), T, TestLoadChangedFileIsStale(), TestLoadEmptyBookTitleUsesFileStem() (+15 more)

### Community 11 - "Task 6 Report: Прив'язка бібліотеки до тому"
Cohesion: 0.16
Nodes (25): scanIncomplete, scanIncompleteError(), copyFixture(), T, newLibrary(), run(), skipUnlessDarwin(), TestAddArgCountError() (+17 more)

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
Cohesion: 0.18
Nodes (10): Concerns, Deviations from the brief, Files changed, GREEN, RED, Self-review findings, Task 6 report — tui: екран перегляду, TDD Evidence (+2 more)

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
Cohesion: 0.06
Nodes (45): app, jsonHit, main(), Command, newAddCmd(), Command, newConfigCmd(), Command (+37 more)

### Community 51 - "newAddCmd"
Cohesion: 0.22
Nodes (8): Concerns, Files changed, Full verification before commit, Self-review findings, Task 5 Report: tui — рендер тексту і тексти i18n, TDD Evidence, Tests and results, What I implemented

### Community 52 - "mustHits"
Cohesion: 0.23
Nodes (15): Query, Store, MatchExpr(), Store, T, mustHits(), seeded(), TestMatchExpr() (+7 more)

### Community 53 - "fb2_test.go"
Cohesion: 0.24
Nodes (15): T, TestAuthorNameFallsBackToNick(), TestParseCollection(), TestParseFileZip(), TestParseGarbage(), TestParseLegacyEncodings(), TestParseNovelYearFromPublishInfo(), TestParseSkipsLargeBinary() (+7 more)

### Community 54 - "Task 7: документація, перевірка, граф — Report"
Cohesion: 0.12
Nodes (16): Changes Made, Command Executed, Concerns, Conclusion, Files Changed, Files Modified, Git Commit, graphify update (+8 more)

### Community 55 - "extract_test.go"
Cohesion: 0.42
Nodes (8): tw, T, plain(), sec(), TestIsNoise(), TestMatchAuthor(), TestWorks(), TestWorksSections()

### Community 56 - "Block"
Cohesion: 0.39
Nodes (4): Block, BlockKind, blockKind(), parser

### Community 57 - "ParseText"
Cohesion: 0.36
Nodes (8): Reader, ParseText(), ParseTextFile(), T, TestParseKeepsNoText(), TestParseTextBlocks(), TestParseTextFileZipAndCP1251(), TestParseTextNested()

## Knowledge Gaps
- **162 isolated node(s):** `findbooks`, `jsonHit`, `quitTimeoutMsg`, `folderMsg`, `removedMsg` (+157 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **18 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Errorf()` connect `Library` to `Parse`, `Hit`, `Task 7: Incremental Parallel Library Scanning`, `scan.go`, `Task 4: Extract Works from FB2 Section Tree - Report`, `Task 5 Implementation Report: SQLite FTS5 Store`, `Task 1 Report: Module Scaffold and Platform Package`, `Task 6 Report: Прив'язка бібліотеки до тому`, `NewRootCmd`, `platform_darwin.go`, `fb2_test.go`?**
  _High betweenness centrality (0.131) - this node is a cross-community bridge._
- **Why does `New()` connect `Task 4: Extract Works from FB2 Section Tree - Report` to `Parse`, `Hit`, `model_test.go`, `Task 7: Incremental Parallel Library Scanning`, `Task 5 Implementation Report: SQLite FTS5 Store`, `runScan`, `Task 6 Report: Прив'язка бібліотеки до тому`, `NewRootCmd`, `platform_darwin.go`?**
  _High betweenness centrality (0.125) - this node is a cross-community bridge._
- **Why does `T()` connect `Hit` to `scan.go`, `Library`, `Task 4: Extract Works from FB2 Section Tree - Report`, `Task 5 Implementation Report: SQLite FTS5 Store`, `runScan`, `Task 6 Report: Прив'язка бібліотеки до тому`, `NewRootCmd`?**
  _High betweenness centrality (0.119) - this node is a cross-community bridge._
- **Are the 6 inferred relationships involving `libModel()` (e.g. with `TestCtrlGIgnoredWhileChoosing()` and `TestCtrlGIgnoredWhileJobRunning()`) actually correct?**
  _`libModel()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **Are the 32 inferred relationships involving `T()` (e.g. with `.Error()` and `newAddCmd()`) actually correct?**
  _`T()` has 32 INFERRED edges - model-reasoned connections that need verification._
- **Are the 31 inferred relationships involving `Errorf()` (e.g. with `.AddLibrary()` and `.ExtractVersion()`) actually correct?**
  _`Errorf()` has 31 INFERRED edges - model-reasoned connections that need verification._
- **Are the 3 inferred relationships involving `press()` (e.g. with `TestCtrlGIgnoredWhileChoosing()` and `TestCtrlGIgnoredWhileJobRunning()`) actually correct?**
  _`press()` has 3 INFERRED edges - model-reasoned connections that need verification._