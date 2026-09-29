package preview

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/text/encoding/charmap"

	"findbooks/internal/fb2"
	"findbooks/internal/index"
)

const collection = `<?xml version="1.0" encoding="utf-8"?>
<FictionBook><description><title-info>
<author><first-name>Евгений</first-name><last-name>Носов</last-name></author>
<author><first-name>Игорь</first-name><last-name>Пидоренко</last-name></author>
<book-title>Румбы</book-title></title-info></description>
<body>
<section><title><p>Евгений Носов</p></title>
 <section><title><p>ЗЕМЛЕЙ РОЖДЕННЫЕ</p></title><p>Текст Носова.</p></section></section>
<section><title><p>Игорь Пидоренко</p></title>
 <section><title><p>ЧУЖИЕ ДЕТИ</p></title>
  <section><title><p>1</p></title><p>Начало.</p></section>
  <section><title><p>2</p></title><p>Конец.</p></section></section></section>
</body></FictionBook>`

// write puts body into dir/name and returns the path and a hit stamped
// with the file's real size and mtime.
func write(t *testing.T, name, body string) (string, index.Hit) {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	return p, index.Hit{Size: fi.Size(), MTime: fi.ModTime().Unix()}
}

func texts(bs []fb2.Block) []string {
	var out []string
	for _, b := range bs {
		out = append(out, b.Text)
	}
	return out
}

func chuzhieHit(h index.Hit) index.Hit {
	h.Format, h.IsCollection, h.Title, h.Author, h.WorkOrd = "fb2", true, "ЧУЖИЕ ДЕТИ", "Игорь Пидоренко", 1
	return h
}

var chuzhieText = []string{"ЧУЖИЕ ДЕТИ", "1", "Начало.", "2", "Конец."}

func TestLoadWorkByOrd(t *testing.T) {
	p, h := write(t, "rumby.fb2", collection)
	doc, err := Load(p, chuzhieHit(h))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Title != "ЧУЖИЕ ДЕТИ" || doc.Author != "Игорь Пидоренко" || doc.Stale {
		t.Fatalf("doc = %+v", doc)
	}
	if got := texts(doc.Blocks); !reflect.DeepEqual(got, chuzhieText) {
		t.Fatalf("text = %q, want %q", got, chuzhieText)
	}
}

func TestLoadWorkByTitleWhenOrdMoved(t *testing.T) {
	p, h := write(t, "rumby.fb2", collection)
	h = chuzhieHit(h)
	h.WorkOrd = 0
	doc, err := Load(p, h)
	if err != nil {
		t.Fatal(err)
	}
	if !doc.Stale || !reflect.DeepEqual(texts(doc.Blocks), chuzhieText) {
		t.Fatalf("stale = %v, text = %q", doc.Stale, texts(doc.Blocks))
	}
}

func TestLoadWholeBookWhenWorkGone(t *testing.T) {
	p, h := write(t, "rumby.fb2", collection)
	h = chuzhieHit(h)
	h.Title, h.WorkOrd = "НЕТ ТАКОГО", 5
	doc, err := Load(p, h)
	if err != nil {
		t.Fatal(err)
	}
	if got := texts(doc.Blocks); !doc.Stale || len(got) == 0 || got[0] != "Евгений Носов" {
		t.Fatalf("stale = %v, text = %q", doc.Stale, got)
	}
}

func TestLoadNotCollectionShowsWholeBook(t *testing.T) {
	p, h := write(t, "rumby.fb2", collection)
	h.Format, h.Title = "fb2", "Румбы"
	doc, err := Load(p, h)
	if err != nil {
		t.Fatal(err)
	}
	if got := texts(doc.Blocks); doc.Stale || got[0] != "Евгений Носов" || got[len(got)-1] != "Конец." {
		t.Fatalf("stale = %v, text = %q", doc.Stale, got)
	}
}

func TestLoadChangedFileIsStale(t *testing.T) {
	p, h := write(t, "rumby.fb2", collection)
	h = chuzhieHit(h)
	h.Size++
	if doc, err := Load(p, h); err != nil || !doc.Stale {
		t.Fatalf("stale = %v, err = %v", doc.Stale, err)
	}
	h.Size = -1 // parse failed at index time: size is not comparable
	if doc, err := Load(p, h); err != nil || doc.Stale {
		t.Fatalf("failed-parse size: stale = %v, err = %v", doc.Stale, err)
	}
}

func TestLoadEmptyBookTitleUsesFileStem(t *testing.T) {
	const book = `<?xml version="1.0" encoding="utf-8"?>
<FictionBook><description><title-info>
<author><first-name>Роберт</first-name><last-name>Шекли</last-name></author>
</title-info></description>
<body><section><title><p>Сборник</p></title>
 <section><title><p>Рассказ А</p></title><p>Текст А.</p></section>
 <section><title><p>Рассказ Б</p></title><p>Текст Б.</p></section>
</section></body></FictionBook>`
	p, h := write(t, "Сборник.fb2", book)
	h.Format, h.IsCollection, h.Title, h.WorkOrd = "fb2", true, "Рассказ Б", 1
	doc, err := Load(p, h)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"Рассказ Б", "Текст Б."}; doc.Stale || !reflect.DeepEqual(texts(doc.Blocks), want) {
		t.Fatalf("stale = %v, text = %q, want %q", doc.Stale, texts(doc.Blocks), want)
	}
}

func TestLoadRecoversParserPanic(t *testing.T) {
	p, h := write(t, "rumby.fb2", collection)
	old := parseFB2
	parseFB2 = func(string) (*fb2.Book, error) { panic("boom") }
	t.Cleanup(func() { parseFB2 = old })
	doc, err := Load(p, chuzhieHit(h))
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v, want it to contain %q", err, "boom")
	}
	if !reflect.DeepEqual(doc, Doc{}) {
		t.Fatalf("doc = %+v, want zero value", doc)
	}
}

func TestLoadTXT(t *testing.T) {
	body := "\xef\xbb\xbfПервый абзац\r\nпродолжение.\r\n\r\n    Второй абзац.\r\n    Третий абзац.\r\n"
	p, h := write(t, "book.txt", body)
	h.Format, h.Title = "txt", "book"
	doc, err := Load(p, h)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Первый абзац продолжение.", "Второй абзац.", "Третий абзац."}
	if got := texts(doc.Blocks); !reflect.DeepEqual(got, want) || doc.Blocks[0].Kind != fb2.BlockPara {
		t.Fatalf("text = %q, want %q", got, want)
	}

	cp, _ := charmap.Windows1251.NewEncoder().String("Привет, мир.")
	p, h = write(t, "cp.txt", cp)
	h.Format = "txt"
	doc, err = Load(p, h)
	if err != nil || !reflect.DeepEqual(texts(doc.Blocks), []string{"Привет, мир."}) {
		t.Fatalf("cp1251 text = %q, err = %v", texts(doc.Blocks), err)
	}
}

func TestParagraphs(t *testing.T) {
	got := texts(paragraphs("\tОдин.\n\tДва\nдва.\n\n\n Три."))
	if want := []string{"Один.", "Два два.", "Три."}; !reflect.DeepEqual(got, want) {
		t.Fatalf("paragraphs = %q, want %q", got, want)
	}
	if got := paragraphs(" \n\n"); len(got) != 0 {
		t.Fatalf("blank text = %q", texts(got))
	}
}

func TestLoadUnsupportedAndMissing(t *testing.T) {
	if _, err := Load("/nonexistent.pdf", index.Hit{Format: "pdf"}); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("pdf err = %v", err)
	}
	if Supported("djvu") || !Supported("fb2") || !Supported("txt") {
		t.Fatal("Supported mismatch")
	}
	_, err := Load(filepath.Join(t.TempDir(), "gone.fb2"), index.Hit{Format: "fb2"})
	if err == nil || !strings.Contains(err.Error(), "gone.fb2") {
		t.Fatalf("missing file err = %v", err)
	}
}
