// Package platform isolates every OS-specific call used by findbooks.
// Nothing outside this package may shell out to OS tools or use build tags.
package platform

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// ErrUnsupported is returned by functions not yet implemented for this OS.
var ErrUnsupported = errors.New("platform: not supported on this OS yet")

// DataDir returns (and creates) the directory that holds index.db:
// %LOCALAPPDATA%\findbooks on Windows, $XDG_DATA_HOME/findbooks or
// ~/.local/share/findbooks elsewhere.
func DataDir() (string, error) {
	var base string
	if runtime.GOOS == "windows" {
		base = os.Getenv("LOCALAPPDATA")
		if base == "" {
			return "", errors.New("LOCALAPPDATA is not set")
		}
	} else {
		base = os.Getenv("XDG_DATA_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			base = filepath.Join(home, ".local", "share")
		}
	}
	dir := filepath.Join(base, "findbooks")
	return dir, os.MkdirAll(dir, 0o755)
}
