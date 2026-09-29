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
// line per paragraph); other line breaks are joined with a space. When the
// text has neither marker (no blank line, no indented line) but more than
// one non-empty line, it is instead treated as one paragraph per line
// (books typed that way too, with no blank lines between paragraphs).
func paragraphs(s string) []fb2.Block {
	norm := strings.ReplaceAll(s, "\r\n", "\n")
	if lines := strings.Split(strings.Trim(norm, "\n"), "\n"); onePerLine(lines) {
		out := make([]fb2.Block, 0, len(lines))
		for _, line := range lines {
			out = append(out, fb2.Block{Kind: fb2.BlockPara, Text: strings.Join(strings.Fields(line), " ")})
		}
		return out
	}
	var out []fb2.Block
	var cur []string
	flush := func() {
		if t := strings.Join(strings.Fields(strings.Join(cur, " ")), " "); t != "" {
			out = append(out, fb2.Block{Kind: fb2.BlockPara, Text: t})
		}
		cur = cur[:0]
	}
	for _, line := range strings.Split(norm, "\n") {
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

// onePerLine reports whether lines (already stripped of any leading or
// trailing blank line from the text's own leading/trailing newlines) has
// more than one line and none of them is blank or starts with a space or
// tab: the text has neither paragraph marker, so each line is its own
// paragraph.
func onePerLine(lines []string) bool {
	if len(lines) < 2 {
		return false
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || line[0] == ' ' || line[0] == '\t' {
			return false
		}
	}
	return true
}
