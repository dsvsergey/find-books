package i18n

import (
	"errors"
	"regexp"
	"slices"
	"testing"
)

var verbRe = regexp.MustCompile(`%[-+# 0-9.]*[a-zA-Z%]`)

func TestCatalogComplete(t *testing.T) {
	for k := Key(1); k < keyCount; k++ {
		en, uk := catalog[EN][k], catalog[UK][k]
		if en == "" || uk == "" {
			t.Errorf("key %d: missing translation (en=%q, uk=%q)", k, en, uk)
			continue
		}
		if ve, vu := verbRe.FindAllString(en, -1), verbRe.FindAllString(uk, -1); !slices.Equal(ve, vu) {
			t.Errorf("key %d: format verbs differ: en %v vs uk %v (%q / %q)", k, ve, vu, en, uk)
		}
	}
	for _, l := range Langs {
		if len(catalog[l]) != int(keyCount)-1 {
			t.Errorf("%s catalog has %d entries, want %d", l, len(catalog[l]), keyCount-1)
		}
	}
}

func TestParse(t *testing.T) {
	for in, want := range map[string]Lang{"en": EN, "EN": EN, " uk ": UK, "Uk": UK} {
		if got, err := Parse(in); err != nil || got != want {
			t.Errorf("Parse(%q) = %q, %v", in, got, err)
		}
	}
	_, err := Parse("xx")
	if err == nil || err.Error() != `unknown language "xx" (available: en, uk)` {
		t.Fatalf("Parse(xx) err = %v", err)
	}
}

func TestSetAndT(t *testing.T) {
	t.Cleanup(func() { Set(EN) })
	if Current() != EN {
		t.Fatalf("default language = %q, want en", Current())
	}
	if got := T(KeyHitsCount, 2, 5); got != "2 of 5" {
		t.Fatalf("en T = %q", got)
	}
	Set(UK)
	if got := T(KeyHitsCount, 2, 5); got != "2 з 5" {
		t.Fatalf("uk T = %q", got)
	}
	if got := T(KeyNothingFound); got != "Нічого не знайдено" {
		t.Fatalf("uk T without args = %q", got)
	}
}

func TestErrorf(t *testing.T) {
	t.Cleanup(func() { Set(EN) })
	sentinel := errors.New("library not found")
	Set(UK)
	err := Errorf(sentinel, KeyLibraryNotFound, "Фантастика")
	if !errors.Is(err, sentinel) {
		t.Fatal("Errorf must wrap the sentinel")
	}
	if err.Error() != "бібліотеку не знайдено: Фантастика" {
		t.Fatalf("Error() = %q (must be the localized text only)", err.Error())
	}
	if Errorf(nil, KeyNothingFound).Error() != "Нічого не знайдено" {
		t.Fatal("Errorf with nil sentinel")
	}
}
