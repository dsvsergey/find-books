package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"findbooks/internal/config"
	"findbooks/internal/i18n"
)

func noEnv(string) string { return "" }

// captureStdout redirects os.Stdout for the duration of fn and returns what
// was written to it. Used only for the one execute() path that writes
// through cobra's default (unset) stdout rather than an injected writer.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = orig
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// TestExecuteBadLangFlag pins the exact first stderr line for a bad --lang
// value, and that no second (source) line follows it.
func TestExecuteBadLangFlag(t *testing.T) {
	t.Cleanup(func() { i18n.Set(i18n.UK) })
	var buf bytes.Buffer
	code := execute([]string{"--lang", "xx", "list"}, noEnv, &buf)
	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	want := "unknown language \"xx\" (available: en, uk)\n"
	if buf.String() != want {
		t.Fatalf("stderr = %q, want exactly %q", buf.String(), want)
	}
}

// TestExecuteBadEnvLang pins the "(from FINDBOOKS_LANG)" second line.
func TestExecuteBadEnvLang(t *testing.T) {
	t.Cleanup(func() { i18n.Set(i18n.UK) })
	getenv := func(k string) string {
		if k == "FINDBOOKS_LANG" {
			return "fr"
		}
		return ""
	}
	var buf bytes.Buffer
	code := execute([]string{"list"}, getenv, &buf)
	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	want := "unknown language \"fr\" (available: en, uk)\n(from FINDBOOKS_LANG)\n"
	if buf.String() != want {
		t.Fatalf("stderr = %q, want exactly %q", buf.String(), want)
	}
}

// TestExecuteBadConfigLang pins the "(from <path>; fix with: ...)" second
// line for a bad value stored in config.json. TestMain already isolates
// config.json under a temp XDG_DATA_HOME; the original content (if any) is
// restored afterward.
func TestExecuteBadConfigLang(t *testing.T) {
	t.Cleanup(func() { i18n.Set(i18n.UK) })
	p, err := config.Path()
	if err != nil {
		t.Fatal(err)
	}
	orig, readErr := os.ReadFile(p)
	t.Cleanup(func() {
		if readErr == nil {
			os.WriteFile(p, orig, 0o644)
		} else {
			os.Remove(p)
		}
	})
	if err := config.Save(config.Config{Lang: "fr"}); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	code := execute([]string{"list"}, noEnv, &buf)
	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	got := buf.String()
	if !strings.HasPrefix(got, "unknown language \"fr\" (available: en, uk)\n") {
		t.Fatalf("stderr = %q, missing the exact first line", got)
	}
	wantSecond := "(from " + p + "; fix with: findbooks --lang en config lang en)\n"
	if !strings.HasSuffix(got, wantSecond) {
		t.Fatalf("stderr = %q, want it to end with %q", got, wantSecond)
	}
}

// TestExecuteLangFlagWithConfigLangShow pins that --lang en, combined with
// `config lang` (no argument, the show form), prints "en".
func TestExecuteLangFlagWithConfigLangShow(t *testing.T) {
	t.Cleanup(func() { i18n.Set(i18n.UK) })
	var code int
	out := captureStdout(t, func() {
		code = execute([]string{"--lang", "en", "config", "lang"}, noEnv, io.Discard)
	})
	if code != 0 {
		t.Fatalf("code = %d, want 0", code)
	}
	if strings.TrimSpace(out) != "en" {
		t.Fatalf("stdout = %q, want \"en\"", out)
	}
}
