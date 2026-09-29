// Package libman is the library management shared by the CLI and the TUI:
// registering a folder as a library and scanning a registered library.
package libman

import (
	"context"
	"errors"
	"path/filepath"

	"findbooks/internal/i18n"
	"findbooks/internal/index"
	"findbooks/internal/library"
	"findbooks/internal/scan"
)

// ErrOffline means the library's disk is not mounted.
var ErrOffline = errors.New("library disk is not mounted")

// Register locates dir on its volume and records it as a library named name
// (the folder name when name is empty). It returns the library and the
// full path of its root.
func Register(st *index.Store, dir, name string) (index.Library, string, error) {
	loc, err := library.Locate(dir)
	if err != nil {
		return index.Library{}, "", err
	}
	root, ok := library.Root(loc.VolumeID, loc.RootRel)
	if !ok {
		return index.Library{}, "", i18n.Errorf(nil, i18n.KeyDiskNotMounted, loc.VolumeName)
	}
	if name == "" {
		name = filepath.Base(root)
	}
	lib, err := st.AddLibrary(name, loc.VolumeID, loc.VolumeName, loc.RootRel)
	if err != nil {
		return index.Library{}, "", err
	}
	return lib, root, nil
}

// Scan brings the index of lib in line with its folder. It fails with
// ErrOffline when the library's disk is not mounted.
func Scan(ctx context.Context, st *index.Store, lib index.Library, onProgress func(scan.Progress)) (scan.Report, error) {
	root, ok := library.Root(lib.VolumeID, lib.RootRel)
	if !ok {
		return scan.Report{}, i18n.Errorf(ErrOffline, i18n.KeyLibraryDiskOffline, lib.VolumeName)
	}
	return scan.Run(ctx, st, lib.ID, root, onProgress)
}
