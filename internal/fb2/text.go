package fb2

import (
	"io"
	"strings"
)

// BlockKind tells how a block of section text is shown.
type BlockKind int

const (
	BlockPara BlockKind = iota
	BlockTitle
	BlockSubtitle
	BlockPoemLine
	BlockEpigraph
	BlockCite
	BlockEmpty
	BlockImage
)

// Block is one paragraph-level piece of section text. Text has its
// whitespace collapsed; it is empty for BlockEmpty and BlockImage.
type Block struct {
	Kind BlockKind
	Text string
}

// ParseText is Parse that also keeps the text of every main-body section
// in Section.Blocks (the section's own content, not its subsections').
func ParseText(r io.Reader) (*Book, error) { return parse(r, true) }

// ParseTextFile is ParseFile that also keeps section text.
func ParseTextFile(path string) (*Book, error) { return parseFile(path, true) }

// blockKind returns the kind of block that element name opens; ctx holds
// the open elements between the innermost section and name, innermost last.
func blockKind(name string, ctx []string) (BlockKind, bool) {
	switch name {
	case "p", "v", "subtitle", "text-author", "td", "th":
	default:
		return 0, false
	}
	for i := len(ctx) - 1; i >= 0; i-- {
		switch ctx[i] {
		case "title":
			return BlockTitle, true
		case "epigraph":
			return BlockEpigraph, true
		case "cite":
			return BlockCite, true
		}
	}
	switch name {
	case "subtitle":
		return BlockSubtitle, true
	case "v":
		return BlockPoemLine, true
	}
	return BlockPara, true
}

// sectionPath returns the open elements inside the innermost section.
func (p *parser) sectionPath() []string {
	for i := len(p.path) - 1; i >= 0; i-- {
		if p.path[i] == "section" {
			return p.path[i+1:]
		}
	}
	return nil
}

func (p *parser) addBlock(b Block) {
	s := p.open[len(p.open)-1]
	s.Blocks = append(s.Blocks, b)
}

// startBlock is called for every element opened inside a main-body section
// while no block is being collected.
func (p *parser) startBlock(name string) {
	switch name {
	case "empty-line":
		p.addBlock(Block{Kind: BlockEmpty})
		return
	case "image":
		p.addBlock(Block{Kind: BlockImage})
		return
	}
	if kind, ok := blockKind(name, p.sectionPath()); ok {
		p.blk, p.blkKind, p.blkDepth = &strings.Builder{}, kind, len(p.path)
	}
}

// endBlock is called after an element closed (p.path already popped).
func (p *parser) endBlock(name string) {
	if p.blk != nil && len(p.path) == p.blkDepth {
		if t := collapse(p.blk.String()); t != "" {
			p.addBlock(Block{Kind: p.blkKind, Text: t})
		}
		p.blk = nil
	}
	if name == "stanza" && p.inMain && len(p.open) > 0 {
		p.addBlock(Block{Kind: BlockEmpty})
	}
}
