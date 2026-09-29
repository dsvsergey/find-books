package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"findbooks/internal/fb2"
	"findbooks/internal/i18n"
)

func TestWrapCutsLongWord(t *testing.T) {
	if got, want := wrap("aaaaaaaaaa", 4, 4), []string{"aaaa", "aaaa", "aa"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("wrap = %q, want %q", got, want)
	}
	if got, want := wrap("один два три", 6, 10), []string{"один", "два три"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("wrap = %q, want %q", got, want)
	}
	if got := wrap("", 10, 10); !reflect.DeepEqual(got, []string{""}) {
		t.Fatalf("wrap empty = %q", got)
	}
}

// TestWrapLongMultiWordStringStaysWithinLimit guards the incremental-width
// change to wrap: no line may exceed the limit even when many words in a
// row keep extending the current line.
func TestWrapLongMultiWordStringStaysWithinLimit(t *testing.T) {
	long := strings.Repeat("слово ", 500)
	for _, l := range wrap(long, 30, 30) {
		if w := ansi.StringWidth(l); w > 30 {
			t.Fatalf("line wider than limit 30: %d %q", w, l)
		}
	}
}

func TestRenderBlocksWrapsToWidth(t *testing.T) {
	long := strings.Repeat("слово ", 40)
	lines := strings.Split(ansi.Strip(renderBlocks([]fb2.Block{{Kind: fb2.BlockPara, Text: long}}, 40)), "\n")
	if len(lines) < 3 {
		t.Fatalf("lines = %q", lines)
	}
	for _, l := range lines {
		if ansi.StringWidth(l) > 40 {
			t.Fatalf("line wider than 40: %q", l)
		}
	}
	if !strings.HasPrefix(lines[0], "    слово") || strings.HasPrefix(lines[1], " ") {
		t.Fatalf("first-line indent only: %q / %q", lines[0], lines[1])
	}
}

func TestRenderBlocksKinds(t *testing.T) {
	blocks := []fb2.Block{
		{Kind: fb2.BlockPara, Text: "Абзац."},
		{Kind: fb2.BlockTitle, Text: "Заголовок"},
		{Kind: fb2.BlockTitle, Text: "Второй"},
		{Kind: fb2.BlockEpigraph, Text: "Эпиграф"},
		{Kind: fb2.BlockPoemLine, Text: "Строка"},
		{Kind: fb2.BlockEmpty},
		{Kind: fb2.BlockImage},
	}
	got := strings.Split(ansi.Strip(renderBlocks(blocks, 60)), "\n")
	want := []string{"    Абзац.", "", "Заголовок", "Второй", "    Эпиграф", "    Строка", "", "    " + i18n.T(i18n.KeyIllustration)}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("render =\n%q\nwant\n%q", got, want)
	}
}
