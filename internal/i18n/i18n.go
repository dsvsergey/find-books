// Package i18n holds every user-facing text of findbooks in English and
// Ukrainian. It imports no other findbooks package.
package i18n

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type Lang string

const (
	EN Lang = "en"
	UK Lang = "uk"
)

// Langs lists the supported languages in display order.
var Langs = []Lang{EN, UK}

var current atomic.Value // Lang

func init() { current.Store(EN) }

// Parse accepts "en" or "uk" in any case, ignoring surrounding spaces.
func Parse(s string) (Lang, error) {
	l := Lang(strings.ToLower(strings.TrimSpace(s)))
	for _, known := range Langs {
		if l == known {
			return l, nil
		}
	}
	return "", fmt.Errorf("unknown language %q (available: en, uk)", strings.TrimSpace(s))
}

// Set changes the current language.
func Set(l Lang) { current.Store(l) }

// Current returns the current language.
func Current() Lang { return current.Load().(Lang) }

// T returns the text for k in the current language, formatted with args.
func T(k Key, args ...any) string {
	s := catalog[Current()][k]
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}

type localizedError struct {
	msg      string
	sentinel error
}

func (e *localizedError) Error() string { return e.msg }
func (e *localizedError) Unwrap() error { return e.sentinel }

// Errorf returns an error whose text is T(k, args...) in the current
// language and which wraps sentinel (nil allowed) for errors.Is.
func Errorf(sentinel error, k Key, args ...any) error {
	return &localizedError{msg: T(k, args...), sentinel: sentinel}
}

var catalog = map[Lang]map[Key]string{EN: catalogEN, UK: catalogUK}
