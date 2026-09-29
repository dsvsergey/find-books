package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"findbooks/internal/config"
	"findbooks/internal/i18n"
)

func newConfigCmd() *cobra.Command {
	cfg := &cobra.Command{
		Use:   "config",
		Short: i18n.T(i18n.KeyCmdConfigShort),
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
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
			if envVal := os.Getenv("FINDBOOKS_LANG"); envVal != "" {
				if envLang, perr := i18n.Parse(envVal); perr != nil || envLang != l {
					fmt.Fprintln(cmd.ErrOrStderr(), i18n.T(i18n.KeyEnvOverridesLang, envVal))
				}
			}
			return nil
		},
	}
	cfg.AddCommand(lang)
	return cfg
}
