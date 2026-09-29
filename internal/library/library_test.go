package library

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func skipUnlessDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("platform is implemented only for macOS in v1")
	}
}

// TempDir lives under /private/var/folders on the Data volume, which is
// mounted at /System/Volumes/Data — this exercises the firmlink fallback.
func TestLocateAndRootRoundTrip(t *testing.T) {
	skipUnlessDarwin(t)
	dir := t.TempDir()
	loc, err := Locate(dir)
	if err != nil {
		t.Fatal(err)
	}
	if loc.VolumeID == "" || loc.VolumeName == "" || strings.Contains(loc.RootRel, `\`) || strings.HasPrefix(loc.RootRel, "..") {
		t.Fatalf("location = %+v", loc)
	}
	root, ok := Root(loc.VolumeID, loc.RootRel)
	if !ok {
		t.Fatalf("Root(%+v) not mounted", loc)
	}
	a, err1 := os.Stat(root)
	b, err2 := os.Stat(dir)
	if err1 != nil || err2 != nil || !os.SameFile(a, b) {
		t.Fatalf("Root = %q is not %q (%v, %v)", root, dir, err1, err2)
	}
}

func TestLocateRejectsFileAndMissing(t *testing.T) {
	skipUnlessDarwin(t)
	file := filepath.Join(t.TempDir(), "x.fb2")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Locate(file); err == nil {
		t.Error("expected error for a file")
	}
	if _, err := Locate(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("expected error for a missing dir")
	}
}

func TestRootOfflineVolume(t *testing.T) {
	if root, ok := Root("00000000-0000-0000-0000-000000000000", "books"); ok {
		t.Fatalf("unexpected root %q", root)
	}
}

func TestFilePathJoinsSlashPath(t *testing.T) {
	if got, want := FilePath("/a", "b/c.fb2"), filepath.Join("/a", "b", "c.fb2"); got != want {
		t.Fatalf("FilePath = %q, want %q", got, want)
	}
}
