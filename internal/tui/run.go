package tui

import (
	"context"
	"os"

	"github.com/atotto/clipboard"

	tea "charm.land/bubbletea/v2"

	"findbooks/internal/config"
	"findbooks/internal/i18n"
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
		SaveLang: func(l i18n.Lang) error {
			c, _ := config.Load()
			c.Lang = string(l)
			return config.Save(c)
		},
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
