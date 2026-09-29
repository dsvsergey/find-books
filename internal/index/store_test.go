package index

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	st, err := Open(filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func addLib(t *testing.T, st *Store, name string) Library {
	t.Helper()
	lib, err := st.AddLibrary(name, "VOL-"+name, "Disk "+name, "Бібліотека/"+name)
	if err != nil {
		t.Fatal(err)
	}
	return lib
}

var (
	rumby = BookRecord{
		RelPath: "Сборники/Румбы/(1988) Румбы фантастики. Том II.fb2", Format: "fb2", Size: 100, MTime: 1000,
		Title: "Румбы фантастики. 1988 год. Том II", Authors: "Евгений Носов, Игорь Пидоренко", Year: "1988",
		IsCollection: true,
		Works: []WorkRecord{
			{"ЗЕМЛЕЙ РОЖДЕННЫЕ", "Евгений Носов", "Евгений Носов › ЗЕМЛЕЙ РОЖДЕННЫЕ"},
			{"ЧУЖИЕ ДЕТИ", "Игорь Пидоренко", "Игорь Пидоренко › ЧУЖИЕ ДЕТИ"},
		},
	}
	belyaev = BookRecord{
		RelPath: "Авторы/Беляев Александр/Звезда КЭЦ.pdf", Format: "pdf", Size: 5, MTime: 1,
		Title: "Звезда КЭЦ", Works: []WorkRecord{{Title: "Звезда КЭЦ", TreePath: "Звезда КЭЦ"}},
	}
	yolka = BookRecord{
		RelPath: "Ёлка.fb2", Format: "fb2", Size: 7, MTime: 7, Title: "Ёлка", Authors: "Иван Петров",
		Works: []WorkRecord{{"Ёлка", "Иван Петров", "Ёлка"}},
	}
)

func TestAddLibraryDuplicate(t *testing.T) {
	st := newStore(t)
	addLib(t, st, "A")
	if _, err := st.AddLibrary("A", "other", "x", "y"); !errors.Is(err, ErrExists) {
		t.Fatalf("same name: err = %v, want ErrExists", err)
	}
	if _, err := st.AddLibrary("B", "VOL-A", "x", "Бібліотека/A"); !errors.Is(err, ErrExists) {
		t.Fatalf("same volume+root: err = %v, want ErrExists", err)
	}
}

func TestLibrariesCountsAndScanTime(t *testing.T) {
	st := newStore(t)
	lib := addLib(t, st, "A")
	if err := st.WriteBatch(lib.ID, []BookRecord{rumby, belyaev}, nil); err != nil {
		t.Fatal(err)
	}
	when := time.Unix(1_800_000_000, 0)
	if err := st.MarkScanned(lib.ID, when); err != nil {
		t.Fatal(err)
	}
	libs, err := st.Libraries()
	if err != nil {
		t.Fatal(err)
	}
	if len(libs) != 1 || libs[0].Books != 2 || libs[0].Works != 3 || !libs[0].LastScan.Equal(when) {
		t.Fatalf("libraries = %+v", libs)
	}
	got, err := st.Library("A")
	if err != nil || got.RootRel != "Бібліотека/A" || got.VolumeID != "VOL-A" {
		t.Fatalf("Library(A) = %+v, %v", got, err)
	}
	if _, err := st.Library("nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	if n, err := st.TotalWorks(); err != nil || n != 3 {
		t.Fatalf("TotalWorks = %d, %v", n, err)
	}
}

func TestFilesReflectsWrites(t *testing.T) {
	st := newStore(t)
	lib := addLib(t, st, "A")
	if err := st.WriteBatch(lib.ID, []BookRecord{rumby}, nil); err != nil {
		t.Fatal(err)
	}
	files, err := st.Files(lib.ID)
	if err != nil {
		t.Fatal(err)
	}
	f, ok := files[rumby.RelPath]
	if !ok || f.Size != 100 || f.MTime != 1000 || f.BookID == 0 {
		t.Fatalf("files = %+v", files)
	}
}

func TestWriteBatchReplacesBook(t *testing.T) {
	st := newStore(t)
	lib := addLib(t, st, "A")
	if err := st.WriteBatch(lib.ID, []BookRecord{rumby}, nil); err != nil {
		t.Fatal(err)
	}
	changed := rumby
	changed.Size = 200
	changed.Works = []WorkRecord{{"НОВАЯ ПОВЕСТЬ", "Игорь Пидоренко", "Игорь Пидоренко › НОВАЯ ПОВЕСТЬ"}}
	if err := st.WriteBatch(lib.ID, []BookRecord{changed}, nil); err != nil {
		t.Fatal(err)
	}
	mustHits(t, st, Query{Text: "чужие"}, 0)
	hits := mustHits(t, st, Query{Text: "новая"}, 1)
	titles, err := st.BookWorks(hits[0].BookID)
	if err != nil || len(titles) != 1 || titles[0] != "НОВАЯ ПОВЕСТЬ" {
		t.Fatalf("BookWorks = %q, %v", titles, err)
	}
	if n, _ := st.TotalWorks(); n != 1 {
		t.Fatalf("TotalWorks = %d, want 1 (old works must be gone)", n)
	}
}

func TestDeleteAndRemoveLibrary(t *testing.T) {
	st := newStore(t)
	lib := addLib(t, st, "A")
	if err := st.WriteBatch(lib.ID, []BookRecord{rumby, yolka}, nil); err != nil {
		t.Fatal(err)
	}
	files, _ := st.Files(lib.ID)
	if err := st.WriteBatch(lib.ID, nil, []int64{files[rumby.RelPath].BookID}); err != nil {
		t.Fatal(err)
	}
	mustHits(t, st, Query{Text: "чужие"}, 0)
	mustHits(t, st, Query{Text: "елка"}, 1)
	if err := st.RemoveLibrary("A"); err != nil {
		t.Fatal(err)
	}
	mustHits(t, st, Query{Text: "елка"}, 0)
	if libs, _ := st.Libraries(); len(libs) != 0 {
		t.Fatalf("libraries left: %+v", libs)
	}
	if err := st.RemoveLibrary("A"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
