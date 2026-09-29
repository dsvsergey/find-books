// Package preview loads the text of one indexed work for reading in the TUI.
package preview

import (
	"errors"
	"fmt"
	"os"

	"findbooks/internal/fb2"
	"findbooks/internal/index"
)

// Doc is the text of one work. Stale means the file changed after indexing
// or the work was not found where the index put it, so the text shown is
// the closest match.
type Doc struct {
	Title, Author string
	Blocks        []fb2.Block
	Stale         bool
}

var ErrUnsupported = errors.New("preview: format not supported")

// Supported reports whether Load can show books of format.
func Supported(format string) bool { return format == "fb2" || format == "txt" }

// Load reads the work h from the file at absPath.
func Load(absPath string, h index.Hit) (doc Doc, err error) {
	if !Supported(h.Format) {
		return Doc{}, ErrUnsupported
	}
	defer func() {
		if p := recover(); p != nil {
			doc, err = Doc{}, fmt.Errorf("preview: parser panic: %v", p)
		}
	}()
	if h.Format == "fb2" {
		doc, err = loadFB2(absPath, h)
	} else {
		doc, err = loadTXT(absPath, h)
	}
	if err != nil {
		return Doc{}, err
	}
	if changed(absPath, h) {
		doc.Stale = true
	}
	return doc, nil
}

// changed reports whether the file differs from what was indexed. A book
// whose parse failed at index time is stored with size -1 and is not
// compared.
func changed(absPath string, h index.Hit) bool {
	if h.Size < 0 {
		return false
	}
	fi, err := os.Stat(absPath)
	if err != nil {
		return false
	}
	return fi.Size() != h.Size || fi.ModTime().Unix() != h.MTime
}
