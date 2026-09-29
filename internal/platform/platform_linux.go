package platform

// Planned: UUID from /dev/disk/by-uuid, /proc/self/mounts, xdg-open.

func VolumeID(path string) (id, name string, err error) { return "", "", ErrUnsupported }
func MountPoint(volumeID string) (string, bool)         { return "", false }
func Open(path string) error                            { return ErrUnsupported }
func Reveal(path string) error                          { return ErrUnsupported }
