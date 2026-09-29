package fb2

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/text/encoding/charmap"
)

const textDoc = `<?xml version="1.0" encoding="utf-8"?>
<FictionBook xmlns="http://www.gribuser.ru/xml/fictionbook/2.0" xmlns:l="http://www.w3.org/1999/xlink">
<description><title-info><book-title>Книга</book-title></title-info></description>
<body>
<section>
 <title><p>Рассказ</p></title>
 <epigraph><p>Эпиграф</p><text-author>Автор</text-author></epigraph>
 <p>Первый   <emphasis>абзац</emphasis><a l:href="#n1" type="note">[1]</a>.</p>
 <image l:href="#pic"/>
 <empty-line/>
 <subtitle>Подзаголовок</subtitle>
 <poem><stanza><v>Строка раз</v><v>Строка два</v></stanza><stanza><v>Строка три</v></stanza></poem>
 <cite><p>Цитата</p></cite>
 <p>  </p>
 <section><title><p>Вложенная</p></title><p>Текст вложенной.</p></section>
</section>
</body>
<body name="notes"><section><title><p>1</p></title><p>Сноска</p></section></body>
<binary id="pic" content-type="image/png">AAAA</binary>
</FictionBook>`

var wantTop = []Block{
	{BlockTitle, "Рассказ"},
	{BlockEpigraph, "Эпиграф"},
	{BlockEpigraph, "Автор"},
	{BlockPara, "Первый абзац[1]."},
	{BlockImage, ""},
	{BlockEmpty, ""},
	{BlockSubtitle, "Подзаголовок"},
	{BlockPoemLine, "Строка раз"},
	{BlockPoemLine, "Строка два"},
	{BlockEmpty, ""},
	{BlockPoemLine, "Строка три"},
	{BlockEmpty, ""},
	{BlockCite, "Цитата"},
}

func TestParseTextBlocks(t *testing.T) {
	b, err := ParseText(strings.NewReader(textDoc))
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Sections) != 1 {
		t.Fatalf("sections = %d, want 1 (notes body must be skipped)", len(b.Sections))
	}
	s := b.Sections[0]
	if s.Title != "Рассказ" {
		t.Errorf("Title = %q", s.Title)
	}
	if !reflect.DeepEqual(s.Blocks, wantTop) {
		t.Errorf("blocks =\n%v\nwant\n%v", s.Blocks, wantTop)
	}
}

func TestParseTextNested(t *testing.T) {
	b, err := ParseText(strings.NewReader(textDoc))
	if err != nil {
		t.Fatal(err)
	}
	want := []Block{{BlockTitle, "Вложенная"}, {BlockPara, "Текст вложенной."}}
	if got := b.Sections[0].Children[0].Blocks; !reflect.DeepEqual(got, want) {
		t.Errorf("nested blocks = %v, want %v", got, want)
	}
}

func TestParseKeepsNoText(t *testing.T) {
	b, err := Parse(strings.NewReader(textDoc))
	if err != nil {
		t.Fatal(err)
	}
	if b.Sections[0].Blocks != nil || b.Sections[0].Children[0].Blocks != nil {
		t.Fatal("Parse must not keep section text")
	}
}

func TestParseTextFileZipAndCP1251(t *testing.T) {
	dir := t.TempDir()
	zpath := filepath.Join(dir, "book.fb2.zip")
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("book.fb2")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte(textDoc))
	zw.Close()
	if err := os.WriteFile(zpath, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := ParseTextFile(zpath)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(b.Sections[0].Blocks, wantTop) {
		t.Errorf("zip blocks = %v", b.Sections[0].Blocks)
	}

	cp, err := charmap.Windows1251.NewEncoder().String(strings.Replace(textDoc, `encoding="utf-8"`, `encoding="windows-1251"`, 1))
	if err != nil {
		t.Fatal(err)
	}
	b, err = ParseText(strings.NewReader(cp))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(b.Sections[0].Blocks, wantTop) {
		t.Errorf("cp1251 blocks = %v", b.Sections[0].Blocks)
	}
}
