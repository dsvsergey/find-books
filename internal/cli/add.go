package cli

import (
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
			return runScan(cmd, st, lib, root)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "назва бібліотеки (за замовчуванням — ім'я теки)")
	return cmd
}
