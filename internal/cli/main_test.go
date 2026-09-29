package cli

import (
	"os"
	"testing"

	"findbooks/internal/i18n"
)

// TestMain keeps the platform cache out of the user's home directory during tests.
func TestMain(m *testing.M) {
	tempDir, err := os.MkdirTemp("", "findbooks-test-*")
	if err != nil {
		panic(err)
	}
	oldXDG := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tempDir)
	i18n.Set(i18n.UK)
	code := m.Run()
	os.Setenv("XDG_DATA_HOME", oldXDG)
	os.RemoveAll(tempDir)
	os.Exit(code)
}
