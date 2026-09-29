package cli

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"findbooks/internal/config"
	"findbooks/internal/i18n"
)

func TestResolveLang(t *testing.T) {
	env := func(v string) func(string) string {
		return func(k string) string {
			if k == "FINDBOOKS_LANG" {
				return v
			}
			return ""
		}
	}
	cfg := func(v string) func() string { return func() string { return v } }
	tests := []struct {
		name string
		args []string
		env  string
		cfg  string
		want i18n.Lang
	}{
		{"default", nil, "", "", i18n.EN},
		{"config", nil, "", "uk", i18n.UK},
		{"env beats config", nil, "en", "uk", i18n.EN},
		{"flag beats env", []string{"--lang", "uk"}, "en", "en", i18n.UK},
		{"flag with =", []string{"--lang=uk"}, "", "", i18n.UK},
		{"flag after subcommand", []string{"search", "--lang", "uk", "x"}, "", "", i18n.UK},
		{"after -- is not a flag", []string{"search", "--", "--lang", "uk"}, "", "", i18n.EN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveLang(tt.args, env(tt.env), cfg(tt.cfg))
			if err != nil || got != tt.want {
				t.Fatalf("resolveLang = %q, %v; want %q", got, err, tt.want)
			}
		})
	}
	if _, err := resolveLang([]string{"--lang", "xx"}, env(""), cfg("")); err == nil ||
		err.Error() != `unknown language "xx" (available: en, uk)` {
		t.Fatalf("bad flag err = %v", err)
	}
	if _, err := resolveLang(nil, env("fr"), cfg("")); err == nil {
		t.Fatal("bad env must fail")
	}
}

func TestEnglishByDefault(t *testing.T) {
	i18n.Set(i18n.EN)
	t.Cleanup(func() { i18n.Set(i18n.UK) })
	db := filepath.Join(t.TempDir(), "index.db")
	_, _, err := run(t, db, "search", "x")
	if !errors.Is(err, errNoLibraries) || !strings.Contains(err.Error(), "the index is empty") {
		t.Fatalf("err = %v", err)
	}
}

func TestLangFlagSwitchesOutput(t *testing.T) {
	i18n.Set(i18n.UK)
	t.Cleanup(func() { i18n.Set(i18n.UK) })
	db := filepath.Join(t.TempDir(), "index.db")
	_, _, err := run(t, db, "--lang", "en", "search", "x")
	if err == nil || !strings.Contains(err.Error(), "the index is empty") {
		t.Fatalf("err = %v", err)
	}
}

func TestConfigLang(t *testing.T) {
	t.Cleanup(func() {
		i18n.Set(i18n.UK)
		config.Save(config.Config{})
	})
	db := filepath.Join(t.TempDir(), "index.db")
	out, _, err := run(t, db, "config", "lang", "uk")
	if err != nil || !strings.Contains(out, "Мову інтерфейсу змінено: українська") {
		t.Fatalf("set uk: %q, %v", out, err)
	}
	if c, _ := config.Load(); c.Lang != "uk" {
		t.Fatalf("config.json lang = %q", c.Lang)
	}
	out, _, err = run(t, db, "config", "lang")
	if err != nil || strings.TrimSpace(out) != "uk" {
		t.Fatalf("show: %q, %v", out, err)
	}
	out, _, err = run(t, db, "config", "lang", "en")
	if err != nil || !strings.Contains(out, "Interface language set: English") {
		t.Fatalf("set en: %q, %v", out, err)
	}
	if _, _, err = run(t, db, "config", "lang", "xx"); err == nil {
		t.Fatal("bad language must fail")
	}
}
