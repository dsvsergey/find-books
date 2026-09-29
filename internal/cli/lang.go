package cli

import (
	"strings"

	"findbooks/internal/config"
	"findbooks/internal/i18n"
)

// resolveLang picks the interface language: --lang (anywhere before "--"),
// then FINDBOOKS_LANG, then config.json, then English.
func resolveLang(args []string, getenv func(string) string, cfgLang func() string) (i18n.Lang, error) {
	for i, a := range args {
		if a == "--" {
			break
		}
		if v, ok := strings.CutPrefix(a, "--lang="); ok {
			return i18n.Parse(v)
		}
		if a == "--lang" && i+1 < len(args) {
			return i18n.Parse(args[i+1])
		}
	}
	if v := getenv("FINDBOOKS_LANG"); v != "" {
		return i18n.Parse(v)
	}
	if v := cfgLang(); v != "" {
		return i18n.Parse(v)
	}
	return i18n.EN, nil
}

func configLang() string {
	c, _ := config.Load()
	return c.Lang
}
