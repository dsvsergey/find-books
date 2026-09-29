// Package library maps library folders to (volume, relative path) pairs so
// the index survives remounts and changing mount points or drive letters.
package library

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"findbooks/internal/platform"
)

type Location struct{ VolumeID, VolumeName, RootRel string }

// Locate resolves dir to the volume it lives on and its path on that volume.
func Locate(dir string) (Location, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Location{}, err
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return Location{}, fmt.Errorf("тека %s: %w", dir, err)
	}
	fi, err := os.Stat(real)
	if err != nil {
		return Location{}, err
	}
	if !fi.IsDir() {
		return Location{}, fmt.Errorf("%s — не тека", dir)
	}
	id, name, err := platform.VolumeID(real)
	if err != nil {
		return Location{}, fmt.Errorf("не вдалося визначити диск для %s: %w", dir, err)
	}
	mp, ok := platform.MountPoint(id)
	if !ok {
		return Location{}, fmt.Errorf("диск «%s» не знайдено серед підключених", name)
	}
	rel, err := relToMount(mp, real)
	if err != nil {
		return Location{}, err
	}
	return Location{VolumeID: id, VolumeName: name, RootRel: filepath.ToSlash(rel)}, nil
}

// relToMount returns real relative to the mount point mp. A path that is not
// lexically under mp (macOS firmlinks: /Users and /private/var live on the
// Data volume mounted at /System/Volumes/Data) is accepted when mp+real is
// the same directory.
func relToMount(mp, real string) (string, error) {
	rel, err := filepath.Rel(mp, real)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return rel, nil
	}
	alt := filepath.Join(mp, real)
	if a, err1 := os.Stat(alt); err1 == nil {
		if r, err2 := os.Stat(real); err2 == nil && os.SameFile(a, r) {
			return filepath.Rel(mp, alt)
		}
	}
	return "", fmt.Errorf("%s не лежить на томі %s", real, mp)
}

// Root returns the library's full path, or false if its volume is not mounted.
func Root(volumeID, rootRel string) (string, bool) {
	mp, ok := platform.MountPoint(volumeID)
	if !ok {
		return "", false
	}
	return filepath.Join(mp, filepath.FromSlash(rootRel)), true
}

// FilePath joins a library root with a slash-separated rel_path.
func FilePath(root, relPath string) string {
	return filepath.Join(root, filepath.FromSlash(relPath))
}
