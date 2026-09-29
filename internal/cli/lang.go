package cli

import (
	"strings"

	"findbooks/internal/config"
	"findbooks/internal/i18n"
)

// langSource names where a resolved language value came from, for the
// second diagnostic line Execute prints when the value is invalid.
type langSource string

const (
	sourceNone   langSource = ""
	sourceFlag   langSource = "flag"
	sourceEnv    langSource = "env"
	sourceConfig langSource = "config"
)

// resolveLang picks the interface language: --lang (anywhere before "--";
// the last occurrence wins, as pflag does), then FINDBOOKS_LANG, then
// config.json, then English. The returned source identifies where the value
// (valid or not) came from, so callers can explain a bad value.
func resolveLang(args []string, getenv func(string) string, cfgLang func() string) (i18n.Lang, langSource, error) {
	var flagVal string
	haveFlag := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			break
		}
		if v, ok := strings.CutPrefix(a, "--lang="); ok {
			flagVal, haveFlag = v, true
			continue
		}
		if a == "--lang" && i+1 < len(args) {
			flagVal, haveFlag = args[i+1], true
			i++
			continue
		}
	}
	if haveFlag {
		l, err := i18n.Parse(flagVal)
		return l, sourceFlag, err
	}
	if v := getenv("FINDBOOKS_LANG"); v != "" {
		l, err := i18n.Parse(v)
		return l, sourceEnv, err
	}
	if v := cfgLang(); v != "" {
		l, err := i18n.Parse(v)
		return l, sourceConfig, err
	}
	return i18n.EN, sourceNone, nil
}

func configLang() string {
	c, _ := config.Load()
	return c.Lang
}
