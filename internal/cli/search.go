package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"findbooks/internal/i18n"
	"findbooks/internal/index"
	"findbooks/internal/library"
)

type jsonHit struct {
	Title      string `json:"title"`
	Author     string `json:"author"`
	TreePath   string `json:"tree_path"`
	Book       string `json:"book"`
	Year       string `json:"year"`
	Collection bool   `json:"collection"`
	Library    string `json:"library"`
	Volume     string `json:"volume"`
	RelPath    string `json:"rel_path"`
	Path       string `json:"path,omitempty"`
	Online     bool   `json:"online"`
}

func newSearchCmd(a *app) *cobra.Command {
	var (
		author string
		limit  int
		asJSON bool
	)
	cmd := &cobra.Command{
		Use:   i18n.T(i18n.KeyCmdSearchUse),
		Short: i18n.T(i18n.KeyCmdSearchShort),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := strings.Join(args, " ")
			if strings.TrimSpace(text) == "" && strings.TrimSpace(author) == "" {
				return i18n.Errorf(nil, i18n.KeyNeedQuery)
			}
			st, err := a.openStore()
			if err != nil {
				return err
			}
			defer st.Close()
			total, err := st.TotalWorks()
			if err != nil {
				return err
			}
			if total == 0 {
				return noLibraries()
			}
			hits, err := st.Search(index.Query{Text: text, Author: author, Limit: limit})
			if err != nil {
				return err
			}
			if asJSON {
				return writeJSON(cmd.OutOrStdout(), hits)
			}
			if len(hits) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), i18n.T(i18n.KeyNothingFound))
				return nil
			}
			return writeTable(cmd.OutOrStdout(), hits)
		},
	}
	cmd.Flags().StringVar(&author, "author", "", i18n.T(i18n.KeyFlagAuthor))
	cmd.Flags().IntVar(&limit, "limit", 20, i18n.T(i18n.KeyFlagLimit))
	cmd.Flags().BoolVar(&asJSON, "json", false, i18n.T(i18n.KeyFlagJSON))
	return cmd
}

func writeJSON(w io.Writer, hits []index.Hit) error {
	out := make([]jsonHit, 0, len(hits))
	for _, h := range hits {
		j := jsonHit{
			Title: h.Title, Author: h.Author, TreePath: h.TreePath, Book: h.BookTitle, Year: h.BookYear,
			Collection: h.IsCollection, Library: h.LibraryName, Volume: h.VolumeName, RelPath: h.RelPath,
		}
		if root, ok := library.Root(h.VolumeID, h.RootRel); ok {
			j.Path, j.Online = library.FilePath(root, h.RelPath), true
		}
		out = append(out, j)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(out)
}

func writeTable(w io.Writer, hits []index.Hit) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, i18n.T(i18n.KeySearchHeader))
	for _, h := range hits {
		book := h.BookTitle
		if h.BookYear != "" {
			book += " (" + h.BookYear + ")"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s: %s %s\n", h.Title, h.Author, book, h.LibraryName, h.RelPath, onlineMark(h.VolumeID, h.RootRel))
	}
	return tw.Flush()
}
