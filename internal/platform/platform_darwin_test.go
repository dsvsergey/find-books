package platform

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"golang.org/x/sys/unix"
)

func TestVolumeIDOfRoot(t *testing.T) {
	id, name, err := VolumeID("/")
	if err != nil {
		t.Fatal(err)
	}
	if id == "" || name == "" {
		t.Fatalf("empty id=%q name=%q", id, name)
	}
}

func TestMountPointFindsVolumeOfTempDir(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	id, _, err := VolumeID(dir)
	if err != nil {
		t.Fatal(err)
	}
	mp, ok := MountPoint(id)
	if !ok {
		t.Fatalf("volume %s not found among mounts", id)
	}
	want, err := mountOf(dir)
	if err != nil {
		t.Fatal(err)
	}
	if mp != want {
		t.Fatalf("MountPoint = %q, statfs says %q", mp, want)
	}
}

func TestMountPointUnknownVolume(t *testing.T) {
	if mp, ok := MountPoint("00000000-0000-0000-0000-000000000000"); ok {
		t.Fatalf("unexpected mount point %q", mp)
	}
}

func TestPlistStringsTopLevelOnly(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>VolumeName</key><string>dsvDev</string>
<key>Nested</key><dict><key>VolumeUUID</key><string>WRONG</string></dict>
<key>Writable</key><true/>
<key>VolumeUUID</key><string>BB26BFDB-2128</string>
</dict></plist>`)
	got, err := plistStrings(data)
	if err != nil {
		t.Fatal(err)
	}
	if got["VolumeName"] != "dsvDev" || got["VolumeUUID"] != "BB26BFDB-2128" {
		t.Fatalf("got %v", got)
	}
}

// TestMain keeps the volume cache out of the user's DataDir during tests.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "findbooks-platform-test")
	if err != nil {
		panic(err)
	}
	volumeCachePath = func() (string, error) { return filepath.Join(dir, "volumes.json"), nil }
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// useCacheFile points the volume cache at path, counts diskutil calls and
// starts from an unloaded in-memory cache, as a new process would.
func useCacheFile(t *testing.T, path string) *atomic.Int64 {
	t.Helper()
	oldPath, oldInfo := volumeCachePath, diskutilInfoFunc
	var calls atomic.Int64
	volumeCachePath = func() (string, error) { return path, nil }
	diskutilInfoFunc = func(mnt string) (map[string]string, error) {
		calls.Add(1)
		return oldInfo(mnt)
	}
	resetVolumeCache()
	t.Cleanup(func() {
		volumeCachePath, diskutilInfoFunc = oldPath, oldInfo
		resetVolumeCache()
	})
	return &calls
}

func tempDirVolume(t *testing.T) (dir, id string) {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	id, _, err = VolumeID(dir)
	if err != nil {
		t.Fatal(err)
	}
	return dir, id
}

func TestMountPointWarmCacheSkipsDiskutil(t *testing.T) {
	calls := useCacheFile(t, filepath.Join(t.TempDir(), "volumes.json"))
	_, id := tempDirVolume(t)
	mp1, ok1 := MountPoint(id)
	if !ok1 {
		t.Fatalf("volume %s not found among mounts", id)
	}
	resetVolumeCache() // reload from the file, as a new process would
	before := calls.Load()
	mp2, ok2 := MountPoint(id)
	if n := calls.Load() - before; n != 0 {
		t.Fatalf("warm MountPoint ran diskutil %d times, want 0", n)
	}
	if mp2 != mp1 || ok2 != ok1 {
		t.Fatalf("warm MountPoint = %q,%v, cold = %q,%v", mp2, ok2, mp1, ok1)
	}
}

func TestMountPointOfflineVolumeSkipsDiskutil(t *testing.T) {
	calls := useCacheFile(t, filepath.Join(t.TempDir(), "volumes.json"))
	MountPoint("00000000-0000-0000-0000-000000000000") // warm the cache
	resetVolumeCache()
	before := calls.Load()
	if mp, ok := MountPoint("00000000-0000-0000-0000-000000000000"); ok {
		t.Fatalf("unexpected mount point %q", mp)
	}
	if n := calls.Load() - before; n != 0 {
		t.Fatalf("offline MountPoint ran diskutil %d times, want 0", n)
	}
}

func TestVolumeIDCorruptCacheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "volumes.json")
	if err := os.WriteFile(path, []byte("\x00not json{{{"), 0o644); err != nil {
		t.Fatal(err)
	}
	useCacheFile(t, path)
	dir, id := tempDirVolume(t)
	if id == "" {
		t.Fatal("empty id")
	}
	mp, ok := MountPoint(id)
	want, err := mountOf(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || mp != want {
		t.Fatalf("MountPoint = %q,%v, want %q", mp, ok, want)
	}
}

func TestFingerprintIncludesRootBirthtime(t *testing.T) {
	fsid := [2]int32{1, 2}
	a := fingerprint("/dev/disk4s1", fsid, 1000, 4096, unix.Timespec{Sec: 1700000000, Nsec: 1})
	b := fingerprint("/dev/disk4s1", fsid, 1000, 4096, unix.Timespec{Sec: 1700000000, Nsec: 2})
	c := fingerprint("/dev/disk4s1", fsid, 1000, 4096, unix.Timespec{Sec: 1700000001, Nsec: 1})
	if a == b || a == c || b == c {
		t.Fatalf("fingerprints must differ by root birthtime: %q %q %q", a, b, c)
	}
	if again := fingerprint("/dev/disk4s1", fsid, 1000, 4096, unix.Timespec{Sec: 1700000000, Nsec: 1}); again != a {
		t.Fatalf("fingerprint not deterministic: %q vs %q", again, a)
	}
}

func TestVolumeIDFallsBackToMountWhenDiskutilFails(t *testing.T) {
	useCacheFile(t, filepath.Join(t.TempDir(), "volumes.json"))
	real := diskutilInfoFunc
	diskutilInfoFunc = func(string) (map[string]string, error) { return nil, errors.New("diskutil: boom") }
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	mnt, err := mountOf(dir)
	if err != nil {
		t.Fatal(err)
	}
	id, name, err := VolumeID(dir)
	if err != nil {
		t.Fatal(err)
	}
	if id != "mnt:"+mnt || name != mnt {
		t.Fatalf("VolumeID = %q, %q; want mnt:%s, %s", id, name, mnt, mnt)
	}
	if mp, ok := MountPoint(id); !ok || mp != mnt {
		t.Fatalf("MountPoint(%q) = %q, %v; want %q", id, mp, ok, mnt)
	}
	// The failure is not cached: once diskutil works, the real UUID comes back.
	diskutilInfoFunc = real
	resetVolumeCache()
	if id2, _, err := VolumeID(dir); err != nil || strings.HasPrefix(id2, "mnt:") {
		t.Fatalf("after diskutil recovers: id = %q, %v", id2, err)
	}
}
