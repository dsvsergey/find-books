package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"findbooks/internal/index"
	"findbooks/internal/library"
)

func newUpdateCmd(a *app) *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "update [назва]",
		Short: "Переіндексувати змінені файли (лише на підключених дисках)",
		Args:  maxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if all == (len(args) == 1) {
				return errors.New("вкажіть назву бібліотеки або --all")
			}
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer closeStore(st, &err)
			var libs []index.Library
			if all {
				libs, err = st.Libraries()
			} else {
				var lib index.Library
				lib, err = st.Library(args[0])
				libs = []index.Library{lib}
			}
			if err != nil {
				return err
			}
			if len(libs) == 0 {
				return errNoLibraries
			}
			for _, lib := range libs {
				root, ok := library.Root(lib.VolumeID, lib.RootRel)
				if !ok {
					fmt.Fprintf(cmd.ErrOrStderr(), "⚠ «%s»: диск «%s» не підключено — пропускаю\n", lib.Name, lib.VolumeName)
					continue
				}
				if err := runScan(cmd, st, lib, root); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "оновити всі підключені бібліотеки")
	return cmd
}
