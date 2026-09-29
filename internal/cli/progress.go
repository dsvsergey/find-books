package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"charm.land/bubbles/v2/progress"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"findbooks/internal/i18n"
	"findbooks/internal/index"
	"findbooks/internal/scan"
)

const maxReportedErrors = 20

type progressBar struct {
	w     io.Writer
	bar   progress.Model
	on    bool
	drawn bool
	last  time.Time
}

// newProgressBar draws only when w is a terminal, so pipes and tests stay clean.
func newProgressBar(w io.Writer) *progressBar {
	f, ok := w.(*os.File)
	return &progressBar{
		w:   w,
		on:  ok && term.IsTerminal(int(f.Fd())),
		bar: progress.New(progress.WithDefaultBlend(), progress.WithWidth(40)),
	}
}

func (p *progressBar) Update(pr scan.Progress) {
	if !p.on || pr.Total == 0 {
		return
	}
	if pr.Done < pr.Total && time.Since(p.last) < 100*time.Millisecond {
		return
	}
	p.last = time.Now()
	fmt.Fprintf(p.w, "\r%s %d/%d", p.bar.ViewAs(float64(pr.Done)/float64(pr.Total)), pr.Done, pr.Total)
	p.drawn = true
}

func (p *progressBar) Done() {
	if p.drawn {
		fmt.Fprintln(p.w)
	}
}

// runScan indexes one library with a progress bar and prints the report.
func runScan(cmd *cobra.Command, st *index.Store, lib index.Library, root string) error {
	errOut := cmd.ErrOrStderr()
	fmt.Fprint(errOut, i18n.T(i18n.KeyIndexing, lib.Name, root)+"\n")
	bar := newProgressBar(errOut)
	rep, err := scan.Run(cmd.Context(), st, lib.ID, root, bar.Update)
	bar.Done()
	fmt.Fprint(errOut, i18n.T(i18n.KeyScanReport, rep.Added, rep.Updated, rep.Removed, rep.Unchanged)+"\n")
	if n := len(rep.Errors); n > 0 {
		fmt.Fprint(errOut, i18n.T(i18n.KeyProblemFiles, n)+"\n")
		for i, e := range rep.Errors {
			if i == maxReportedErrors {
				fmt.Fprint(errOut, i18n.T(i18n.KeyAndMore, n-i)+"\n")
				break
			}
			fmt.Fprintf(errOut, "  %s: %v\n", e.RelPath, e.Err)
		}
	}
	return err
}
