package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"findbooks/internal/config"
	"findbooks/internal/i18n"
)

func newConfigCmd() *cobra.Command {
	cfg := &cobra.Command{Use: "config", Short: i18n.T(i18n.KeyCmdConfigShort), Args: noArgs}
	lang := &cobra.Command{
		Use:   "lang [en|uk]",
		Short: i18n.T(i18n.KeyCmdConfigLangShort),
		Args:  maxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), i18n.Current())
				return nil
			}
			l, err := i18n.Parse(args[0])
			if err != nil {
				return err
			}
			c, _ := config.Load()
			c.Lang = string(l)
			if err := config.Save(c); err != nil {
				return err
			}
			i18n.Set(l)
			fmt.Fprintln(cmd.OutOrStdout(), i18n.T(i18n.KeyLangSaved))
			return nil
		},
	}
	cfg.AddCommand(lang)
	return cfg
}
