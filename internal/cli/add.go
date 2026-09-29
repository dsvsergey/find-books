package cli

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"findbooks/internal/library"
)

func newAddCmd(a *app) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "add <шлях>",
		Short: "Зареєструвати бібліотеку і проіндексувати її",
		Args:  exactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			loc, err := library.Locate(args[0])
			if err != nil {
				return err
			}
			root, ok := library.Root(loc.VolumeID, loc.RootRel)
			if !ok {
				return fmt.Errorf("диск «%s» не знайдено серед підключених", loc.VolumeName)
			}
			if name == "" {
				name = filepath.Base(root)
			}
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer closeStore(st, &err)
			lib, err := st.AddLibrary(name, loc.VolumeID, loc.VolumeName, loc.RootRel)
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
