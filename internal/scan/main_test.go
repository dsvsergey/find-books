package scan

import (
	"os"
	"testing"

	"findbooks/internal/i18n"
)

func TestMain(m *testing.M) {
	i18n.Set(i18n.UK)
	os.Exit(m.Run())
}
