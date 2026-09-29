# findbooks

Пошук творів у бібліотеках електронних книг — зокрема всередині збірок, альманахів і антологій.

## Встановлення

```bash
go build -o findbooks ./cmd/findbooks
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

У TUI: `enter` — відкрити, `ctrl+o` — показати у Finder, `ctrl+y` — копіювати шлях, `tab` — шукати за автором, `↑`/`↓` — вибір, `esc` — вихід.

FB2 (і `.fb2.zip`) індексуються разом зі змістом збірок; PDF, DJVU, DOC, RTF, TXT, EPUB — за ім'ям файлу і текою.
