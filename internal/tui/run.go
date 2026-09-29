package tui

import (
	"os"

	"github.com/atotto/clipboard"

	tea "charm.land/bubbletea/v2"

	"findbooks/internal/library"
	"findbooks/internal/platform"
)

// DefaultActions wires the screen to the real OS.
func DefaultActions() Actions {
	return Actions{
		Root: library.Root, Open: platform.Open, Reveal: platform.Reveal, Copy: clipboard.WriteAll,
		Exists: func(p string) bool { _, err := os.Stat(p); return err == nil },
	}
}

// Run shows the search screen until the user quits.
func Run(src Searcher, total int) error {
	_, err := tea.NewProgram(New(src, DefaultActions(), total)).Run()
	return err
}
