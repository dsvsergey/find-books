package platform

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sys/unix"
)

// VolumeID returns the volume UUID and display name of the volume that
// contains path. Volumes without a UUID fall back to "mnt:<mount point>".
func VolumeID(path string) (id, name string, err error) {
	var st unix.Statfs_t
	if err := unix.Statfs(path, &st); err != nil {
		return "", "", fmt.Errorf("statfs %s: %w", path, err)
	}
	mnt := unix.ByteSliceToString(st.Mntonname[:])
	volMu.Lock()
	defer volMu.Unlock()
	loadVolumesLocked()
	defer saveVolumesLocked()
	v, err := volumeLocked(&st)
	if err != nil {
		return "", "", err
	}
	id, name = v.UUID, v.Name
	if id == "" {
		id = "mnt:" + mnt
	}
	if name == "" {
		name = mnt
	}
	return id, name, nil
}

// MountPoint reports where the volume is mounted right now. Mount points are
// always read live; only the (slow) diskutil UUID lookup is cached, keyed by
// a fingerprint of the mounted filesystem.
func MountPoint(volumeID string) (string, bool) {
	if p, ok := strings.CutPrefix(volumeID, "mnt:"); ok {
		m, err := mountOf(p)
		return p, err == nil && m == p
	}
	if volumeID == "" {
		return "", false
	}
	n, err := unix.Getfsstat(nil, unix.MNT_NOWAIT)
	if err != nil {
		return "", false
	}
	buf := make([]unix.Statfs_t, n)
	n, err = unix.Getfsstat(buf, unix.MNT_NOWAIT)
	if err != nil {
		return "", false
	}
	volMu.Lock()
	defer volMu.Unlock()
	loadVolumesLocked()
	defer saveVolumesLocked()
	for i := range buf[:n] {
		st := &buf[i]
		if !strings.HasPrefix(unix.ByteSliceToString(st.Mntfromname[:]), "/dev/") {
			continue
		}
		if v, err := volumeLocked(st); err == nil && v.UUID == volumeID {
			return unix.ByteSliceToString(st.Mntonname[:]), true
		}
	}
	return "", false
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

// volume is what diskutil tells us about a mounted filesystem. UUID is empty
// for volumes that have none; they are cached too so diskutil is not re-run.
type volume struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

var (
	// diskutilInfoFunc and volumeCachePath are variables so tests can count
	// diskutil calls and use a temporary cache file.
	diskutilInfoFunc = diskutilInfo
	volumeCachePath  = func() (string, error) {
		dir, err := DataDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, "volumes.json"), nil
	}

	volMu     sync.Mutex
	volumes   map[string]volume // fingerprint -> volume; nil until loaded
	volPath   string            // "" means in-memory only
	volsDirty bool
)

// fingerprint identifies a mounted filesystem cheaply, without a subprocess.
// The mount point is deliberately not part of it.
func fingerprint(st *unix.Statfs_t) string {
	return fmt.Sprintf("%s|%d:%d|%d|%d", unix.ByteSliceToString(st.Mntfromname[:]),
		st.Fsid.Val[0], st.Fsid.Val[1], st.Blocks, st.Bsize)
}

// volumeLocked returns the cached volume for st, asking diskutil on a miss.
func volumeLocked(st *unix.Statfs_t) (volume, error) {
	fp := fingerprint(st)
	if v, ok := volumes[fp]; ok {
		return v, nil
	}
	info, err := diskutilInfoFunc(unix.ByteSliceToString(st.Mntonname[:]))
	if err != nil {
		return volume{}, err
	}
	v := volume{UUID: info["VolumeUUID"], Name: info["VolumeName"]}
	volumes[fp], volsDirty = v, true
	return v, nil
}

// loadVolumesLocked reads the cache file once per process. A missing or
// corrupt file is treated as empty.
func loadVolumesLocked() {
	if volumes != nil {
		return
	}
	volumes = map[string]volume{}
	path, err := volumeCachePath()
	if err != nil {
		return
	}
	volPath = path
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var m map[string]volume
	if json.Unmarshal(data, &m) == nil {
		for k, v := range m {
			volumes[k] = v
		}
	}
}

// saveVolumesLocked writes the cache atomically if it gained entries.
// Failures are ignored: the cache is only an optimisation.
func saveVolumesLocked() {
	if !volsDirty || volPath == "" {
		return
	}
	data, err := json.MarshalIndent(volumes, "", "  ")
	if err != nil {
		return
	}
	f, err := os.CreateTemp(filepath.Dir(volPath), ".volumes-*.json")
	if err != nil {
		return
	}
	tmp := f.Name()
	_, werr := f.Write(data)
	cerr := f.Close()
	if werr != nil || cerr != nil || os.Rename(tmp, volPath) != nil {
		os.Remove(tmp)
		return
	}
	volsDirty = false
}

// resetVolumeCache forgets the in-memory cache so the next lookup reloads
// the file, as a new process would.
func resetVolumeCache() {
	volMu.Lock()
	defer volMu.Unlock()
	volumes, volPath, volsDirty = nil, "", false
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
