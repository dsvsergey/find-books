package preview

import (
	"bytes"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"

	"findbooks/internal/fb2"
	"findbooks/internal/index"
)

func loadTXT(absPath string, h index.Hit) (Doc, error) {
	raw, err := os.ReadFile(absPath)
	if err != nil {
		return Doc{}, err
	}
	return Doc{Title: h.Title, Author: h.Author, Blocks: paragraphs(decode(raw))}, nil
}

// decode reads UTF-8 (BOM dropped), falling back to windows-1251.
func decode(raw []byte) string {
	raw = bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf"))
	if utf8.Valid(raw) {
		return string(raw)
	}
	s, err := charmap.Windows1251.NewDecoder().Bytes(raw)
	if err != nil {
		return string(raw)
	}
	return string(s)
}

// paragraphs splits plain text into paragraphs: a blank line ends one, and
// a line starting with a space or tab begins one (books typed one indented
// line per paragraph); other line breaks are joined with a space.
func paragraphs(s string) []fb2.Block {
	var out []fb2.Block
	var cur []string
	flush := func() {
		if t := strings.Join(strings.Fields(strings.Join(cur, " ")), " "); t != "" {
			out = append(out, fb2.Block{Kind: fb2.BlockPara, Text: t})
		}
		cur = cur[:0]
	}
	for _, line := range strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n") {
		switch {
		case strings.TrimSpace(line) == "":
			flush()
		case line[0] == ' ' || line[0] == '\t':
			flush()
			cur = append(cur, line)
		default:
			cur = append(cur, line)
		}
	}
	flush()
	return out
}
