package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"findbooks/internal/libman"
)

func newAddCmd(a *app) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "add <шлях>",
		Short: "Зареєструвати бібліотеку і проіндексувати її",
		Args:  exactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer closeStore(st, &err)
			lib, root, err := libman.Register(st, args[0], name)
			if err != nil {
				return err
			}
			if err := runScan(cmd, st, lib, root); err != nil {
				return scanIncompleteError(lib.Name, err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "назва бібліотеки (за замовчуванням — ім'я теки)")
	return cmd
}

// scanIncomplete is returned by add when the library was registered but its
// first scan failed or was interrupted; it tells the user how to resume.
type scanIncomplete struct {
	name string
	err  error
}

func scanIncompleteError(name string, err error) error { return &scanIncomplete{name: name, err: err} }

func (e *scanIncomplete) Error() string {
	cause := ""
	if !errors.Is(e.err, context.Canceled) { // "перервано" is printed by Execute
		cause = ": " + e.err.Error()
	}
	return fmt.Sprintf("бібліотеку «%s» зареєстровано, але індексування не завершено%s — продовжіть: findbooks update «%s»",
		e.name, cause, e.name)
}

func (e *scanIncomplete) Unwrap() error { return e.err }
