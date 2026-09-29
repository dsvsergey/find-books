// Package extract decides which sections of an FB2 book are separate works.
package extract

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"findbooks/internal/fb2"
	"findbooks/internal/textnorm"
)

type Work struct{ Title, Author, TreePath string }

const treeSep = " › "

// Works walks the section tree. Untitled, noise (chapters, notes, prefaces)
// and book-title sections are transparent; a section titled with one of the
// book's authors sets the author for its children; any other titled section
// is a work and its subsections are not visited. With fewer than two works
// the book is not a collection and a single work for the whole book is
// returned.
func Works(b *fb2.Book) ([]Work, bool) {
	bookTitle := textnorm.Normalize(b.Title)
	var found []Work
	var walk func(secs []*fb2.Section, ctxAuthor string)
	walk = func(secs []*fb2.Section, ctxAuthor string) {
		for _, s := range secs {
			norm := textnorm.Normalize(s.Title)
			if norm == "" || norm == bookTitle || isNoise(norm) {
				walk(s.Children, ctxAuthor)
				continue
			}
			if name := matchAuthor(norm, b.Authors); name != "" {
				walk(s.Children, name)
				continue
			}
			found = append(found, newWork(s.Title, ctxAuthor, b.Authors))
		}
	}
	walk(b.Sections, "")
	if len(found) < 2 {
		return []Work{{Title: b.Title, Author: b.AuthorNames(), TreePath: b.Title}}, false
	}
	return found, true
}

func newWork(title, ctxAuthor string, authors []fb2.Author) Work {
	author := ctxAuthor
	if author == "" && len(authors) == 1 {
		author = authors[0].Name()
	}
	tree := title
	if ctxAuthor != "" {
		tree = ctxAuthor + treeSep + title
	}
	return Work{Title: title, Author: author, TreePath: tree}
}

var (
	noiseExact = map[string]bool{
		"пролог": true, "эпилог": true, "примечания": true, "примечание": true,
		"комментарии": true, "содержание": true, "оглавление": true,
		"от составителя": true, "от автора": true, "от редакции": true,
		"предисловие": true, "послесловие": true, "вместо предисловия": true,
		"вместо послесловия": true, "аннотация": true, "notes": true, "contents": true,
	}
	noiseContains = []string{"об авторах", "об авторе", "библиограф"}
	numberRe      = regexp.MustCompile(`^[0-9]+$`)
	romanRe       = regexp.MustCompile(`^m{0,3}(cm|cd|d?c{0,3})(xc|xl|l?x{0,3})(ix|iv|v?i{0,3})$`)
	chapterRe     = regexp.MustCompile(`^(глава|часть|книга|том|раздел|chapter|part) (([0-9]+|[ivxlcdm]+)( |$)|перв|втор|трет|четв|пят|шест|седьм|восьм|девят|десят|one|two|three)`)
)

// isNoise reports whether a normalized, non-empty title is structural or
// back matter rather than a work.
func isNoise(norm string) bool {
	if noiseExact[norm] || numberRe.MatchString(norm) || romanRe.MatchString(norm) || chapterRe.MatchString(norm) {
		return true
	}
	for _, s := range noiseContains {
		if strings.Contains(norm, s) {
			return true
		}
	}
	return false
}

// matchAuthor returns the author's display name when the normalized title
// consists only of that author's name parts (initials allowed) and contains
// the full last name and at least one given name token.
func matchAuthor(norm string, authors []fb2.Author) string {
	words := strings.Fields(norm)
	if len(words) == 0 || len(words) > 5 {
		return ""
	}
	for _, a := range authors {
		last := strings.Fields(textnorm.Normalize(a.Last))
		if len(last) == 0 {
			continue
		}
		given := strings.Fields(textnorm.Normalize(a.First + " " + a.Middle))
		if containsAll(words, last) && hasGivenName(words, given) && allKnown(words, last, given) {
			return a.Name()
		}
	}
	return ""
}

func containsAll(words, need []string) bool {
	for _, n := range need {
		if !slices.Contains(words, n) {
			return false
		}
	}
	return true
}

func allKnown(words, last, given []string) bool {
	for _, w := range words {
		if slices.Contains(last, w) || slices.Contains(given, w) {
			continue
		}
		if utf8.RuneCountInString(w) == 1 && slices.ContainsFunc(given, func(g string) bool { return strings.HasPrefix(g, w) }) {
			continue
		}
		return false
	}
	return true
}

func hasGivenName(words, given []string) bool {
	for _, w := range words {
		// Check full given name
		if slices.Contains(given, w) {
			return true
		}
		// Check single-letter initial
		if utf8.RuneCountInString(w) == 1 && slices.ContainsFunc(given, func(g string) bool { return strings.HasPrefix(g, w) }) {
			return true
		}
	}
	return false
}
