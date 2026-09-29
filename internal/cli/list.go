package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"findbooks/internal/i18n"
	"findbooks/internal/library"
)

func onlineMark(volumeID, rootRel string) string {
	if _, ok := library.Root(volumeID, rootRel); ok {
		return "●"
	}
	return "○"
}

func newListCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: i18n.T(i18n.KeyCmdListShort),
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer st.Close()
			libs, err := st.Libraries()
			if err != nil {
				return err
			}
			if len(libs) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), noLibraries())
				return nil
			}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, i18n.T(i18n.KeyListHeader))
			for _, l := range libs {
				updated := "—"
				if !l.LastScan.IsZero() {
					updated = l.LastScan.Format("2006-01-02 15:04")
				}
				fmt.Fprintf(tw, "%s\t%s %s\t%d\t%d\t%s\n", l.Name, l.VolumeName, onlineMark(l.VolumeID, l.RootRel), l.Books, l.Works, updated)
			}
			return tw.Flush()
		},
	}
}
