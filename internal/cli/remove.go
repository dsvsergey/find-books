package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRemoveCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <назва>",
		Short: "Прибрати бібліотеку з індексу (файли не чіпаються)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer st.Close()
			if err := st.RemoveLibrary(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Бібліотеку «%s» прибрано з індексу\n", args[0])
			return nil
		},
	}
}
