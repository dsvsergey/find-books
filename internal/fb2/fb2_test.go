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

func titles(secs []*Section) []string {
	var out []string
	for _, s := range secs {
		out = append(out, s.Title)
	}
	return out
}

func TestParseCollection(t *testing.T) {
	b, err := ParseFile("testdata/rumby.fb2")
	if err != nil {
		t.Fatal(err)
	}
	if b.Title != "Румбы фантастики. 1988 год. Том II" {
		t.Errorf("Title = %q", b.Title)
	}
	if b.Year != "1988" {
		t.Errorf("Year = %q", b.Year)
	}
	if got, want := b.AuthorNames(), "Евгений Носов, Игорь Пидоренко, Олег Чарушников"; got != want {
		t.Errorf("AuthorNames = %q, want %q", got, want)
	}
	wantTop := []string{"Евгений Носов", "Игорь Пидоренко", "Олег Чарушников", "ОБ АВТОРАХ ЭТОГО СБОРНИКА"}
	if got := titles(b.Sections); !reflect.DeepEqual(got, wantTop) {
		t.Fatalf("top sections = %q, want %q (notes body must be skipped)", got, wantTop)
	}
	chuzhie := b.Sections[1].Children[0]
	if chuzhie.Title != "ЧУЖИЕ ДЕТИ" || !reflect.DeepEqual(titles(chuzhie.Children), []string{"1", "2"}) {
		t.Errorf("ЧУЖИЕ ДЕТИ subtree = %q %q", chuzhie.Title, titles(chuzhie.Children))
	}
	if got := b.Sections[2].Children[0].Children[0].Title; got != "История первая ТРУД СИЗИФА" {
		t.Errorf("multi-paragraph title = %q", got)
	}
}

func TestParseNovelYearFromPublishInfo(t *testing.T) {
	b, err := ParseFile("testdata/novel.fb2")
	if err != nil {
		t.Fatal(err)
	}
	if b.Title != "Человек-амфибия" || b.Year != "1928" {
		t.Errorf("Title=%q Year=%q", b.Title, b.Year)
	}
	if got := b.Sections[0].Children[0].Title; got != "Глава 1 «Морской дьявол»" {
		t.Errorf("chapter title = %q", got)
	}
}

func TestParseLegacyEncodings(t *testing.T) {
	src, err := os.ReadFile("testdata/novel.fb2")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		label string
		enc   *charmap.Charmap
	}{{"windows-1251", charmap.Windows1251}, {"koi8-r", charmap.KOI8R}} {
		doc := strings.Replace(string(src), `encoding="utf-8"`, `encoding="`+tc.label+`"`, 1)
		doc = strings.NewReplacer("«", `"`, "»", `"`).Replace(doc) // KOI8-R has no guillemets
		encoded, err := tc.enc.NewEncoder().String(doc)
		if err != nil {
			t.Fatal(err)
		}
		b, err := Parse(strings.NewReader(encoded))
		if err != nil {
			t.Fatalf("%s: %v", tc.label, err)
		}
		if b.Title != "Человек-амфибия" {
			t.Errorf("%s: Title = %q", tc.label, b.Title)
		}
	}
}

func TestParseUnknownEncoding(t *testing.T) {
	_, err := Parse(strings.NewReader(`<?xml version="1.0" encoding="x-klingon"?><FictionBook/>`))
	if err == nil {
		t.Fatal("expected error for unknown encoding")
	}
}

func TestParseGarbage(t *testing.T) {
	if _, err := Parse(strings.NewReader("not xml at all")); err == nil {
		t.Fatal("expected error for non-XML input")
	}
}

func TestParseTruncated(t *testing.T) {
	src, err := os.ReadFile("testdata/rumby.fb2")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(bytes.NewReader(src[:len(src)/2])); err == nil {
		t.Fatal("expected error for truncated XML")
	}
}

func TestParseSkipsLargeBinary(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="utf-8"?><FictionBook><description><title-info>` +
		`<book-title>Big</book-title></title-info></description>` +
		`<body><section><title><p>Рассказ</p></title></section></body>` +
		`<binary id="c" content-type="image/jpeg">`)
	buf.WriteString(strings.Repeat("QUFB", 2<<20)) // 8 MiB of base64
	buf.WriteString(`</binary></FictionBook>`)
	b, err := Parse(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if b.Title != "Big" || len(b.Sections) != 1 || b.Sections[0].Title != "Рассказ" {
		t.Fatalf("got %+v", b)
	}
}

func TestParseFileZip(t *testing.T) {
	src, err := os.ReadFile("testdata/rumby.fb2")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	withBook := filepath.Join(dir, "rumby.fb2.zip")
	writeZip(t, withBook, "Румбы.fb2", src)
	b, err := ParseFile(withBook)
	if err != nil {
		t.Fatal(err)
	}
	if b.Title != "Румбы фантастики. 1988 год. Том II" {
		t.Errorf("Title = %q", b.Title)
	}
	empty := filepath.Join(dir, "empty.fb2.zip")
	writeZip(t, empty, "readme.txt", []byte("hi"))
	if _, err := ParseFile(empty); err == nil {
		t.Fatal("expected error for zip without .fb2")
	}
}

func writeZip(t *testing.T, path, name string, data []byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestAuthorNameFallsBackToNick(t *testing.T) {
	if got := (Author{Nick: "Tanja45"}).Name(); got != "Tanja45" {
		t.Errorf("Name() = %q", got)
	}
	if got := (Author{First: "Кир", Middle: "Булычёв", Last: "Булычев"}).Name(); got != "Кир Булычев" {
		t.Errorf("Name() = %q", got)
	}
}
