package textnorm

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct{ in, want string }{
		{"ЧУЖИЕ ДЕТИ", "чужие дети"},
		{"  Ёлка  ", "елка"},
		{"«Румбы фантастики»", "румбы фантастики"},
		{"Операция «ПРОГРЕССОР»", "операция прогрессор"},
		{"…ТОЖЕ РЕЗУЛЬТАТ", "тоже результат"},
		{"13-Й ПОДВИГ ГЕРАКЛА", "13 й подвиг геракла"},
		{"за́мок", "замок"},             // stress accent
		{"Бойцов", "бойцов"},            // NFD й from macOS file names
		{"Її ґанок", "її ґанок"},              // Ukrainian letters survive
		{"* * *", ""},
		{"", ""},
		{"Tom's  Diner", "tom s diner"},
	}
	for _, tt := range tests {
		if got := Normalize(tt.in); got != tt.want {
			t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
