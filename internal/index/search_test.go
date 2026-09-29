package index

import "testing"

func mustHits(t *testing.T, st *Store, q Query, want int) []Hit {
	t.Helper()
	hits, err := st.Search(q)
	if err != nil {
		t.Fatalf("Search(%+v): %v", q, err)
	}
	if len(hits) != want {
		t.Fatalf("Search(%+v) = %d hits %+v, want %d", q, len(hits), hits, want)
	}
	return hits
}

func seeded(t *testing.T) *Store {
	st := newStore(t)
	lib := addLib(t, st, "Фантастика")
	if err := st.WriteBatch(lib.ID, []BookRecord{rumby, belyaev, yolka}, nil); err != nil {
		t.Fatal(err)
	}
	return st
}

func TestSearchFindsWorkInsideCollection(t *testing.T) {
	st := seeded(t)
	h := mustHits(t, st, Query{Text: "чужие дети"}, 1)[0]
	if h.Title != "ЧУЖИЕ ДЕТИ" || h.Author != "Игорь Пидоренко" || h.TreePath != "Игорь Пидоренко › ЧУЖИЕ ДЕТИ" ||
		h.BookTitle != rumby.Title || h.BookYear != "1988" || !h.IsCollection || h.RelPath != rumby.RelPath ||
		h.Format != "fb2" || h.LibraryName != "Фантастика" || h.VolumeID != "VOL-Фантастика" ||
		h.VolumeName != "Disk Фантастика" || h.RootRel != "Бібліотека/Фантастика" {
		t.Fatalf("hit = %+v", h)
	}
}

func TestSearchIgnoresCaseYoAndPunctuation(t *testing.T) {
	st := seeded(t)
	for _, q := range []string{"ЧУЖИЕ", "чуж дет", "  Чужие,  дети! "} {
		mustHits(t, st, Query{Text: q}, 1)
	}
	mustHits(t, st, Query{Text: "елка"}, 1)
	mustHits(t, st, Query{Text: "Ёлка"}, 1)
}

func TestSearchByBookTitleAndFolder(t *testing.T) {
	st := seeded(t)
	mustHits(t, st, Query{Text: "румбы"}, 2)
	mustHits(t, st, Query{Text: "беляев"}, 1)
}

func TestSearchAuthorFilter(t *testing.T) {
	st := seeded(t)
	h := mustHits(t, st, Query{Author: "пидоренко"}, 1)[0]
	if h.Title != "ЧУЖИЕ ДЕТИ" {
		t.Fatalf("hit = %+v", h)
	}
	mustHits(t, st, Query{Text: "земле", Author: "пидоренко"}, 0)
	mustHits(t, st, Query{Text: "земле", Author: "носов"}, 1)
}

func TestSearchHostileInput(t *testing.T) {
	st := seeded(t)
	for _, q := range []string{`"`, `*`, `чуж*"`, `NEAR(a b)`, `title:x`, `-дети`, `AND`, `OR NOT`, `')--`, `{title}: x`} {
		if _, err := st.Search(Query{Text: q}); err != nil {
			t.Errorf("Search(%q): %v", q, err)
		}
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	st := seeded(t)
	hits, err := st.Search(Query{Text: " !! "})
	if err != nil || hits != nil {
		t.Fatalf("got %v, %v", hits, err)
	}
}

func TestMatchExpr(t *testing.T) {
	tests := map[string]string{
		"Чужие дети": `"чужие"* "дети"*`,
		`"; DROP`:    `"drop"*`,
		"":           "",
		"«Ёлка»-2":   `"елка"* "2"*`,
	}
	for in, want := range tests {
		if got := MatchExpr(in); got != want {
			t.Errorf("MatchExpr(%q) = %q, want %q", in, got, want)
		}
	}
}
