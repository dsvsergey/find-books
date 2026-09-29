package platform

// Planned: volume serial via GetVolumeInformationW, drive-letter scan,
// `cmd /c start ""`, `explorer /select,`.

func VolumeID(path string) (id, name string, err error) { return "", "", ErrUnsupported }
func MountPoint(volumeID string) (string, bool)         { return "", false }
func Open(path string) error                            { return ErrUnsupported }
func Reveal(path string) error                          { return ErrUnsupported }
