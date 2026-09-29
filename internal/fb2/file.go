package fb2

import (
	"archive/zip"
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ParseFile parses a .fb2 file or the first .fb2 entry of a .fb2.zip archive.
func ParseFile(path string) (*Book, error) {
	if strings.EqualFold(filepathExt(path), ".zip") {
		return parseZip(path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(bufio.NewReaderSize(f, 64<<10))
}

func parseZip(path string) (*Book, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("fb2: %w", err)
	}
	defer zr.Close()
	for _, f := range zr.File {
		if !strings.EqualFold(filepathExt(f.Name), ".fb2") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("fb2: %w", err)
		}
		defer rc.Close()
		return Parse(bufio.NewReaderSize(rc, 64<<10))
	}
	return nil, fmt.Errorf("fb2: no .fb2 entry in %s", path)
}

func filepathExt(name string) string {
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		return name[i:]
	}
	return ""
}
