package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"findbooks/internal/i18n"
)

func newRemoveCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   i18n.T(i18n.KeyCmdRemoveUse),
		Short: i18n.T(i18n.KeyCmdRemoveShort),
		Args:  exactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer closeStore(st, &err)
			if err := st.RemoveLibrary(args[0]); err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), i18n.T(i18n.KeyRemoved, args[0])+"\n")
			return nil
		},
	}
}
