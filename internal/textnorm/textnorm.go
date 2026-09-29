// Package textnorm folds titles and queries into one comparable form.
package textnorm

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Normalize composes s to NFC, lower-cases it, folds ё into е, drops
// combining marks (stress accents), and turns every run of characters that
// are not letters or digits into a single space, trimming both ends.
func Normalize(s string) string {
	s = norm.NFC.String(s)
	var b strings.Builder
	b.Grow(len(s))
	pendingSpace := false
	for _, r := range s {
		switch {
		case unicode.Is(unicode.Mn, r):
			continue
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if pendingSpace && b.Len() > 0 {
				b.WriteByte(' ')
			}
			pendingSpace = false
			r = unicode.ToLower(r)
			if r == 'ё' {
				r = 'е'
			}
			b.WriteRune(r)
		default:
			pendingSpace = true
		}
	}
	return b.String()
}
