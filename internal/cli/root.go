// Package cli defines the findbooks commands.
package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/spf13/cobra"

	"findbooks/internal/index"
	"findbooks/internal/platform"
	"findbooks/internal/tui"
)

var errNoLibraries = errors.New("індекс порожній — додайте бібліотеку: findbooks add <шлях>")

// runTUI is replaced in tests.
var runTUI = func(st *index.Store, total int) error { return tui.Run(st, total) }

type app struct{ dbPath string }

func (a *app) openStore() (*index.Store, error) {
	p := a.dbPath
	if p == "" {
		dir, err := platform.DataDir()
		if err != nil {
			return nil, err
		}
		p = filepath.Join(dir, "index.db")
	}
	return index.Open(p)
}

func NewRootCmd() *cobra.Command {
	a := &app{}
	root := &cobra.Command{
		Use:           "findbooks",
		Short:         "Пошук творів у бібліотеках електронних книг, зокрема всередині збірок",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer st.Close()
			total, err := st.TotalWorks()
			if err != nil {
				return err
			}
			if total == 0 {
				return errNoLibraries
			}
			return runTUI(st, total)
		},
	}
	root.PersistentFlags().StringVar(&a.dbPath, "db", "", "файл індексу (за замовчуванням — у теці даних користувача)")
	root.AddCommand(newAddCmd(a), newUpdateCmd(a), newListCmd(a), newRemoveCmd(a), newSearchCmd(a))
	return root
}

// Execute runs the CLI and returns the process exit code.
func Execute() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := NewRootCmd().ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "помилка:", err)
		return 1
	}
	return 0
}
