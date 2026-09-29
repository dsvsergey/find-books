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
		name       string
		args       []string
		env        string
		cfg        string
		want       i18n.Lang
		wantSource langSource
	}{
		{"default", nil, "", "", i18n.EN, sourceNone},
		{"config", nil, "", "uk", i18n.UK, sourceConfig},
		{"env beats config", nil, "en", "uk", i18n.EN, sourceEnv},
		{"flag beats env", []string{"--lang", "uk"}, "en", "en", i18n.UK, sourceFlag},
		{"flag with =", []string{"--lang=uk"}, "", "", i18n.UK, sourceFlag},
		{"flag after subcommand", []string{"search", "--lang", "uk", "x"}, "", "", i18n.UK, sourceFlag},
		{"after -- is not a flag", []string{"search", "--", "--lang", "uk"}, "", "", i18n.EN, sourceNone},
		{"last --lang wins", []string{"--lang", "en", "--lang", "uk"}, "", "", i18n.UK, sourceFlag},
		{"last --lang= wins over plain form", []string{"--lang", "en", "--lang=uk"}, "", "", i18n.UK, sourceFlag},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, source, err := resolveLang(tt.args, env(tt.env), cfg(tt.cfg))
			if err != nil || got != tt.want {
				t.Fatalf("resolveLang = %q, %v; want %q", got, err, tt.want)
			}
			if source != tt.wantSource {
				t.Fatalf("source = %q, want %q", source, tt.wantSource)
			}
		})
	}
	if _, source, err := resolveLang([]string{"--lang", "xx"}, env(""), cfg("")); err == nil ||
		err.Error() != `unknown language "xx" (available: en, uk)` || source != sourceFlag {
		t.Fatalf("bad flag err = %v, source = %q", err, source)
	}
	if _, source, err := resolveLang(nil, env("fr"), cfg("")); err == nil || source != sourceEnv {
		t.Fatalf("bad env must fail with source env; err = %v, source = %q", err, source)
	}
	if _, source, err := resolveLang(nil, env(""), cfg("fr")); err == nil || source != sourceConfig {
		t.Fatalf("bad config must fail with source config; err = %v, source = %q", err, source)
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

// TestConfigNoSubcommandShowsHelp pins that `findbooks config` alone (no
// subcommand, no args) prints help and succeeds, same as before this
// package's config command grew a RunE.
func TestConfigNoSubcommandShowsHelp(t *testing.T) {
	db := filepath.Join(t.TempDir(), "index.db")
	out, _, err := run(t, db, "config")
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	if !strings.Contains(out, "Usage:") {
		t.Fatalf("out = %q, want help text", out)
	}
}

// TestConfigUnknownSubcommandFails pins the controller ruling that an
// unrecognized config subcommand (e.g. a typo like "lnag") must fail rather
// than silently print help and exit 0.
func TestConfigUnknownSubcommandFails(t *testing.T) {
	db := filepath.Join(t.TempDir(), "index.db")
	if _, _, err := run(t, db, "config", "lnag", "uk"); err == nil {
		t.Fatal("config with an unknown subcommand must return an error")
	}
}

// TestConfigLangEnvOverrideNote pins that saving a language via `config
// lang` warns on stderr when FINDBOOKS_LANG is set to something else: the
// saved value will not take effect on the next run without overriding the
// environment too.
func TestConfigLangEnvOverrideNote(t *testing.T) {
	t.Cleanup(func() {
		i18n.Set(i18n.UK)
		config.Save(config.Config{})
	})
	t.Setenv("FINDBOOKS_LANG", "en")
	db := filepath.Join(t.TempDir(), "index.db")
	_, stderr, err := run(t, db, "config", "lang", "uk")
	if err != nil {
		t.Fatalf("config lang uk: %v", err)
	}
	if !strings.Contains(stderr, "FINDBOOKS_LANG=en") || !strings.Contains(stderr, "пріоритет") {
		t.Fatalf("stderr = %q, want the UK env-override note", stderr)
	}
}

// TestConfigLangEnvOverrideNoteSuppressedWhenMatching pins that no note is
// printed when FINDBOOKS_LANG already agrees with the language just saved.
func TestConfigLangEnvOverrideNoteSuppressedWhenMatching(t *testing.T) {
	t.Cleanup(func() {
		i18n.Set(i18n.UK)
		config.Save(config.Config{})
	})
	t.Setenv("FINDBOOKS_LANG", "uk")
	db := filepath.Join(t.TempDir(), "index.db")
	_, stderr, err := run(t, db, "config", "lang", "uk")
	if err != nil {
		t.Fatalf("config lang uk: %v", err)
	}
	if strings.Contains(stderr, "FINDBOOKS_LANG") {
		t.Fatalf("stderr = %q, want no env-override note", stderr)
	}
}
