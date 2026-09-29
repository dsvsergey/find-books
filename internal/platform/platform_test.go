package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDataDirHonorsXDG(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG_DATA_HOME is not used on Windows")
	}
	base := t.TempDir()
	t.Setenv("XDG_DATA_HOME", base)
	dir, err := DataDir()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(base, "findbooks"); dir != want {
		t.Fatalf("DataDir() = %q, want %q", dir, want)
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		t.Fatalf("DataDir did not create %s: %v", dir, err)
	}
}
