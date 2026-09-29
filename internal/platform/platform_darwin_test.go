package platform

import (
	"path/filepath"
	"testing"
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
