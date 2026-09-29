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
