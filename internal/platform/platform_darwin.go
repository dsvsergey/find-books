package platform

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// VolumeID returns the volume UUID and display name of the volume that
// contains path. Volumes without a UUID fall back to "mnt:<mount point>".
func VolumeID(path string) (id, name string, err error) {
	mnt, err := mountOf(path)
	if err != nil {
		return "", "", err
	}
	info, err := diskutilInfo(mnt)
	if err != nil {
		return "", "", err
	}
	id, name = info["VolumeUUID"], info["VolumeName"]
	if id == "" {
		id = "mnt:" + mnt
	}
	if name == "" {
		name = mnt
	}
	return id, name, nil
}

var (
	mountMu  sync.Mutex
	mounts   map[string]string // volume UUID -> mount point
	mountsAt time.Time
)

// mountRescanEvery bounds how often a miss triggers a full diskutil scan,
// so offline volumes in a result list do not cost a scan per lookup.
const mountRescanEvery = 5 * time.Second

// MountPoint reports where the volume is mounted right now.
func MountPoint(volumeID string) (string, bool) {
	if p, ok := strings.CutPrefix(volumeID, "mnt:"); ok {
		m, err := mountOf(p)
		return p, err == nil && m == p
	}
	mountMu.Lock()
	defer mountMu.Unlock()
	if mp, ok := mounts[volumeID]; ok {
		if m, err := mountOf(mp); err == nil && m == mp {
			return mp, true
		}
	}
	if mounts == nil || time.Since(mountsAt) > mountRescanEvery {
		mounts, mountsAt = scanMounts(), time.Now()
	}
	mp, ok := mounts[volumeID]
	return mp, ok
}

// Open opens path in its default application.
func Open(path string) error { return exec.Command("open", path).Run() }

// Reveal selects path in Finder.
func Reveal(path string) error { return exec.Command("open", "-R", path).Run() }

func mountOf(path string) (string, error) {
	var st unix.Statfs_t
	if err := unix.Statfs(path, &st); err != nil {
		return "", fmt.Errorf("statfs %s: %w", path, err)
	}
	return unix.ByteSliceToString(st.Mntonname[:]), nil
}

func scanMounts() map[string]string {
	res := map[string]string{}
	n, err := unix.Getfsstat(nil, unix.MNT_NOWAIT)
	if err != nil {
		return res
	}
	buf := make([]unix.Statfs_t, n)
	n, err = unix.Getfsstat(buf, unix.MNT_NOWAIT)
	if err != nil {
		return res
	}
	for _, st := range buf[:n] {
		if !strings.HasPrefix(unix.ByteSliceToString(st.Mntfromname[:]), "/dev/") {
			continue
		}
		mnt := unix.ByteSliceToString(st.Mntonname[:])
		info, err := diskutilInfo(mnt)
		if err != nil || info["VolumeUUID"] == "" {
			continue
		}
		res[info["VolumeUUID"]] = mnt
	}
	return res
}

func diskutilInfo(mnt string) (map[string]string, error) {
	out, err := exec.Command("diskutil", "info", "-plist", mnt).Output()
	if err != nil {
		return nil, fmt.Errorf("diskutil info %s: %w", mnt, err)
	}
	return plistStrings(out)
}

// plistStrings extracts the <key>/<string> pairs of a plist's top-level dict.
func plistStrings(data []byte) (map[string]string, error) {
	d := xml.NewDecoder(bytes.NewReader(data))
	res := map[string]string{}
	depth := 0
	key := ""
	for {
		tok, err := d.Token()
		if err == io.EOF {
			return res, nil
		}
		if err != nil {
			return nil, fmt.Errorf("plist: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if depth < 2 { // <plist> and the top-level <dict>
				depth++
				continue
			}
			switch t.Name.Local {
			case "key", "string":
				var s string
				if err := d.DecodeElement(&s, &t); err != nil {
					return nil, fmt.Errorf("plist: %w", err)
				}
				if t.Name.Local == "key" {
					key = s
				} else if key != "" {
					res[key], key = s, ""
				}
			default:
				key = ""
				if err := d.Skip(); err != nil {
					return nil, fmt.Errorf("plist: %w", err)
				}
			}
		case xml.EndElement:
			depth--
		}
	}
}
