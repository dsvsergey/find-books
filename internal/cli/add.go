package cli

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"findbooks/internal/i18n"
	"findbooks/internal/libman"
)

func newAddCmd(a *app) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   i18n.T(i18n.KeyCmdAddUse),
		Short: i18n.T(i18n.KeyCmdAddShort),
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
	cmd.Flags().StringVar(&name, "name", "", i18n.T(i18n.KeyFlagName))
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
	if !errors.Is(e.err, context.Canceled) { // KeyInterrupted is printed by Execute
		cause = ": " + e.err.Error()
	}
	return i18n.T(i18n.KeyScanIncomplete, e.name, cause, e.name)
}

func (e *scanIncomplete) Unwrap() error { return e.err }
