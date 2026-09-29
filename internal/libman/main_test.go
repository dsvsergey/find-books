package libman

import (
	"os"
	"testing"
)

// TestMain keeps platform's volume cache out of the user's home directory.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "findbooks-test-*")
	if err != nil {
		panic(err)
	}
	os.Setenv("XDG_DATA_HOME", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
