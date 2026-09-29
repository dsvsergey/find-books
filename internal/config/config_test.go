package config

import (
	"os"
	"path/filepath"
	"testing"
)

func reset(t *testing.T) string {
	t.Helper()
	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	os.Remove(p)
	t.Cleanup(func() { os.Remove(p) })
	return p
}

func TestLoadMissing(t *testing.T) {
	reset(t)
	c, err := Load()
	if err != nil || c != (Config{}) {
		t.Fatalf("Load() = %+v, %v", c, err)
	}
}

func TestLoadCorrupt(t *testing.T) {
	p := reset(t)
	for _, data := range []string{"", "{not json", `{"lang": 7}`} {
		if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
		if c, err := Load(); err != nil || c != (Config{}) {
			t.Errorf("Load(%q) = %+v, %v", data, c, err)
		}
	}
}

func TestSaveLoad(t *testing.T) {
	p := reset(t)
	if err := Save(Config{Lang: "uk"}); err != nil {
		t.Fatal(err)
	}
	c, err := Load()
	if err != nil || c.Lang != "uk" {
		t.Fatalf("Load() = %+v, %v", c, err)
	}
	entries, _ := os.ReadDir(filepath.Dir(p))
	for _, e := range entries {
		if e.Name() != "config.json" && filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("temporary file left behind: %s", e.Name())
		}
	}
}
