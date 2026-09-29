package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"findbooks/internal/i18n"
	"findbooks/internal/index"
	"findbooks/internal/library"
)

func newUpdateCmd(a *app) *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   i18n.T(i18n.KeyCmdUpdateUse),
		Short: i18n.T(i18n.KeyCmdUpdateShort),
		Args:  maxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if all == (len(args) == 1) {
				return i18n.Errorf(nil, i18n.KeyNeedNameOrAll)
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
				return noLibraries()
			}
			for _, lib := range libs {
				root, ok := library.Root(lib.VolumeID, lib.RootRel)
				if !ok {
					fmt.Fprint(cmd.ErrOrStderr(), i18n.T(i18n.KeySkipOffline, lib.Name, lib.VolumeName)+"\n")
					continue
				}
				if err := runScan(cmd, st, lib, root); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, i18n.T(i18n.KeyFlagAll))
	return cmd
}
