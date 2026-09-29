// Package fb2 streams FictionBook 2 files and returns their metadata and the
// tree of section titles of the main body. Parse and ParseFile never keep
// section text; ParseText and ParseTextFile also fill Section.Blocks.
package fb2

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/htmlindex"
)

type Author struct{ First, Middle, Last, Nick string }

// Name returns "First Last", falling back to the nickname.
func (a Author) Name() string {
	parts := make([]string, 0, 2)
	for _, p := range []string{a.First, a.Last} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) == 0 {
		return a.Nick
	}
	return strings.Join(parts, " ")
}

// Section is a <section> of the main body. Title is the text of its <title>
// with paragraphs joined by spaces and whitespace collapsed ("" if none).
// Blocks holds the section's own text (not its subsections') and is filled
// only by ParseText/ParseTextFile; Parse/ParseFile leave it nil.
type Section struct {
	Title    string
	Children []*Section
	Blocks   []Block
}

type Book struct {
	Title    string
	Authors  []Author
	Year     string
	Sections []*Section
}

// AuthorNames joins the authors' names with ", ".
func (b *Book) AuthorNames() string {
	names := make([]string, 0, len(b.Authors))
	for _, a := range b.Authors {
		if n := a.Name(); n != "" {
			names = append(names, n)
		}
	}
	return strings.Join(names, ", ")
}

var yearRe = regexp.MustCompile(`\b(1[5-9]\d\d|20\d\d)\b`)

// Parse reads one FB2 document. Bodies with a name attribute (notes,
// comments) and <binary> payloads are skipped.
func Parse(r io.Reader) (*Book, error) { return parse(r, false) }

func parse(r io.Reader, withText bool) (*Book, error) {
	d := xml.NewDecoder(r)
	d.Strict = false
	d.AutoClose = xml.HTMLAutoClose
	d.Entity = xml.HTMLEntity
	d.CharsetReader = charsetReader
	p := &parser{d: d, withText: withText}
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("fb2: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if err := p.start(t); err != nil {
				return nil, fmt.Errorf("fb2: %w", err)
			}
		case xml.EndElement:
			p.end(t.Name.Local)
		case xml.CharData:
			p.text(string(t))
		}
	}
	if !p.seenRoot {
		return nil, errors.New("fb2: no <FictionBook> root element")
	}
	p.book.Title = collapse(p.bookTitle.String())
	p.book.Year = yearRe.FindString(p.date)
	if p.book.Year == "" {
		p.book.Year = yearRe.FindString(p.pubYear)
	}
	return &p.book, nil
}

// charsetReader decodes known encodings; an unknown label is read as UTF-8,
// so a mislabelled UTF-8 file parses and anything else fails on invalid
// UTF-8 and lands in the error report.
func charsetReader(label string, input io.Reader) (io.Reader, error) {
	enc, err := htmlindex.Get(label)
	if err != nil {
		return input, nil
	}
	return enc.NewDecoder().Reader(input), nil
}

type parser struct {
	d         *xml.Decoder
	book      Book
	path      []string // local names of open elements, innermost last
	author    *Author  // title-info author being read
	open      []*Section
	title     *strings.Builder // text of the section title being read
	titleFor  *Section
	bookTitle strings.Builder
	date      string
	pubYear   string
	inMain    bool
	seenMain  bool
	seenRoot  bool

	withText bool             // keep section text in Section.Blocks
	blk      *strings.Builder // text of the block being read
	blkKind  BlockKind
	blkDepth int // len(path) outside the block's element
}

func (p *parser) parent() string {
	if len(p.path) == 0 {
		return ""
	}
	return p.path[len(p.path)-1]
}

// at reports whether the open elements end with names (innermost last).
func (p *parser) at(names ...string) bool {
	if len(p.path) < len(names) {
		return false
	}
	tail := p.path[len(p.path)-len(names):]
	for i := range names {
		if tail[i] != names[i] {
			return false
		}
	}
	return true
}

func (p *parser) start(t xml.StartElement) error {
	name := t.Name.Local
	switch {
	case name == "FictionBook":
		p.seenRoot = true
	case name == "binary":
		return p.d.Skip()
	case name == "body":
		if !p.seenMain && attr(t, "name") == "" {
			p.inMain, p.seenMain = true, true
		}
	case name == "section" && p.inMain:
		s := &Section{}
		if n := len(p.open); n > 0 {
			p.open[n-1].Children = append(p.open[n-1].Children, s)
		} else {
			p.book.Sections = append(p.book.Sections, s)
		}
		p.open = append(p.open, s)
	case name == "title" && p.inMain && len(p.open) > 0 && p.parent() == "section":
		p.title = &strings.Builder{}
		p.titleFor = p.open[len(p.open)-1]
	case name == "author" && p.parent() == "title-info":
		p.author = &Author{}
	case name == "date" && p.parent() == "title-info":
		p.date += attr(t, "value") + " "
	}
	if p.withText && p.inMain && len(p.open) > 0 && p.blk == nil {
		p.startBlock(name)
	}
	if p.title != nil && (name == "p" || name == "empty-line") {
		p.title.WriteByte(' ')
	}
	p.path = append(p.path, name)
	return nil
}

func (p *parser) end(name string) {
	if len(p.path) > 0 {
		p.path = p.path[:len(p.path)-1]
	}
	if p.withText {
		p.endBlock(name)
	}
	switch {
	case name == "title" && p.title != nil:
		p.titleFor.Title = collapse(p.title.String())
		p.title, p.titleFor = nil, nil
	case name == "section" && p.inMain && len(p.open) > 0:
		p.open = p.open[:len(p.open)-1]
	case name == "body" && p.inMain:
		p.inMain = false
	case name == "author" && p.author != nil:
		a := *p.author
		p.book.Authors = append(p.book.Authors, Author{
			First: collapse(a.First), Middle: collapse(a.Middle),
			Last: collapse(a.Last), Nick: collapse(a.Nick),
		})
		p.author = nil
	}
}

func (p *parser) text(s string) {
	if p.blk != nil {
		p.blk.WriteString(s)
	}
	switch {
	case p.title != nil:
		p.title.WriteString(s)
	case p.at("title-info", "book-title"):
		p.bookTitle.WriteString(s)
	case p.author != nil && p.at("author", "first-name"):
		p.author.First += s
	case p.author != nil && p.at("author", "middle-name"):
		p.author.Middle += s
	case p.author != nil && p.at("author", "last-name"):
		p.author.Last += s
	case p.author != nil && p.at("author", "nickname"):
		p.author.Nick += s
	case p.at("title-info", "date"):
		p.date += s + " "
	case p.at("publish-info", "year"):
		p.pubYear += s
	}
}

func attr(t xml.StartElement, name string) string {
	for _, a := range t.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func collapse(s string) string { return strings.Join(strings.Fields(s), " ") }
