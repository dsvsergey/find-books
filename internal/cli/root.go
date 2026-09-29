// Package cli defines the findbooks commands.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"findbooks/internal/i18n"
	"findbooks/internal/index"
	"findbooks/internal/platform"
	"findbooks/internal/tui"
)

var errNoLibraries = errors.New("the index is empty")

// noLibraries returns errNoLibraries with its message in the current language.
func noLibraries() error { return i18n.Errorf(errNoLibraries, i18n.KeyNoLibraries) }

// runTUI is replaced in tests.
var runTUI = func(st *index.Store, total int) error { return tui.Run(st, total) }

type app struct{ dbPath string }

func (a *app) openStore() (*index.Store, error) {
	p := a.dbPath
	if p == "" {
		dir, err := platform.DataDir()
		if err != nil {
			return nil, err
		}
		p = filepath.Join(dir, "index.db")
	}
	return index.Open(p)
}

// closeStore closes st and, if the command otherwise succeeded, surfaces
// any close error via the named return *err.
func closeStore(st *index.Store, err *error) {
	if cerr := st.Close(); *err == nil {
		*err = cerr
	}
}

// noArgs, exactArgs and maxArgs are cobra.PositionalArgs validators with
// localized error messages (cobra's built-ins are English-only).
func noArgs(_ *cobra.Command, args []string) error {
	if len(args) > 0 {
		return i18n.Errorf(nil, i18n.KeyArgsNone, strings.Join(args, " "))
	}
	return nil
}

func exactArgs(n int) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if len(args) != n {
			return i18n.Errorf(nil, i18n.KeyArgsExact, n, len(args))
		}
		return nil
	}
}

func maxArgs(n int) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if len(args) > n {
			return i18n.Errorf(nil, i18n.KeyArgsMax, n, len(args))
		}
		return nil
	}
}

func NewRootCmd() *cobra.Command {
	a := &app{}
	root := &cobra.Command{
		Use:           "findbooks",
		Short:         i18n.T(i18n.KeyCmdRootShort),
		Args:          noArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer st.Close()
			total, err := st.TotalWorks()
			if err != nil {
				return err
			}
			return runTUI(st, total)
		},
	}
	root.PersistentFlags().StringVar(&a.dbPath, "db", "", i18n.T(i18n.KeyFlagDB))
	var lang string
	root.PersistentFlags().StringVar(&lang, "lang", "", i18n.T(i18n.KeyFlagLang))
	root.PersistentPreRunE = func(*cobra.Command, []string) error {
		if lang == "" {
			return nil
		}
		l, err := i18n.Parse(lang)
		if err != nil {
			return err
		}
		i18n.Set(l)
		return nil
	}
	root.AddCommand(newAddCmd(a), newUpdateCmd(a), newListCmd(a), newRemoveCmd(a), newSearchCmd(a), newConfigCmd())
	return root
}

// Execute runs the CLI and returns the process exit code.
func Execute() int {
	lang, err := resolveLang(os.Args[1:], os.Getenv, configLang)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	i18n.Set(lang)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return exitCode(NewRootCmd().ExecuteContext(ctx), os.Stderr)
}

// exitCode prints err to stderr and returns the process exit code:
// 130 (as shells report SIGINT) for Ctrl+C, 1 for other errors.
func exitCode(err error, stderr io.Writer) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, context.Canceled):
		fmt.Fprintln(stderr, i18n.T(i18n.KeyInterrupted))
		if si, ok := errors.AsType[*scanIncomplete](err); ok {
			fmt.Fprintln(stderr, si)
		}
		return 130
	default:
		fmt.Fprintln(stderr, i18n.T(i18n.KeyErrorPrefix), err)
		return 1
	}
}
