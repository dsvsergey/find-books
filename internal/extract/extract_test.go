package extract

import (
	"reflect"
	"testing"

	"findbooks/internal/fb2"
	"findbooks/internal/textnorm"
)

func sec(title string, children ...*fb2.Section) *fb2.Section {
	return &fb2.Section{Title: title, Children: children}
}

func TestWorks(t *testing.T) {
	nosov := fb2.Author{First: "Евгений", Middle: "Валентинович", Last: "Носов"}
	pidorenko := fb2.Author{First: "Игорь", Last: "Пидоренко"}
	charushnikov := fb2.Author{First: "Олег", Last: "Чарушников"}
	sheckley := fb2.Author{First: "Роберт", Last: "Шекли"}

	tests := []struct {
		name     string
		book     fb2.Book
		want     []Work
		wantColl bool
	}{
		{
			name: "anthology with author sections",
			book: fb2.Book{
				Title:   "Румбы фантастики. 1988 год. Том II",
				Authors: []fb2.Author{nosov, pidorenko, charushnikov},
				Sections: []*fb2.Section{
					sec("Евгений Носов", sec("ЗЕМЛЕЙ РОЖДЕННЫЕ")),
					sec("Игорь Пидоренко", sec("ЧУЖИЕ ДЕТИ", sec("1"), sec("2"))),
					sec("Олег Чарушников", sec("НА «ОЛИМПЕ» ВСЕ СПОКОЙНО", sec("История первая ТРУД СИЗИФА"))),
					sec("Александр Каширин БИБЛИОГРАФИЯ ФАНТАСТИКИ [8]"),
					sec("ОБ АВТОРАХ ЭТОГО СБОРНИКА"),
				},
			},
			want: []Work{
				{"ЗЕМЛЕЙ РОЖДЕННЫЕ", "Евгений Носов", "Евгений Носов › ЗЕМЛЕЙ РОЖДЕННЫЕ"},
				{"ЧУЖИЕ ДЕТИ", "Игорь Пидоренко", "Игорь Пидоренко › ЧУЖИЕ ДЕТИ"},
				{"НА «ОЛИМПЕ» ВСЕ СПОКОЙНО", "Олег Чарушников", "Олег Чарушников › НА «ОЛИМПЕ» ВСЕ СПОКОЙНО"},
			},
			wantColl: true,
		},
		{
			name: "novel with parts and chapters",
			book: fb2.Book{
				Title:   "Человек-амфибия",
				Authors: []fb2.Author{{First: "Александр", Last: "Беляев"}},
				Sections: []*fb2.Section{
					sec("Часть первая", sec("Глава 1 «Морской дьявол»"), sec("Глава 2 Доктор Сальватор")),
					sec("Часть вторая", sec("Глава 1 Ихтиандр")),
				},
			},
			want:     []Work{{"Человек-амфибия", "Александр Беляев", "Человек-амфибия"}},
			wantColl: false,
		},
		{
			name: "single-author story collection",
			book: fb2.Book{
				Title:    "Рассказы",
				Authors:  []fb2.Author{sheckley},
				Sections: []*fb2.Section{sec("Запах мысли"), sec("Страж-птица"), sec("Примечания")},
			},
			want: []Work{
				{"Запах мысли", "Роберт Шекли", "Запах мысли"},
				{"Страж-птица", "Роберт Шекли", "Страж-птица"},
			},
			wantColl: true,
		},
		{
			name: "wrapper section equal to book title is transparent",
			book: fb2.Book{
				Title:    "Сборник X",
				Authors:  []fb2.Author{sheckley},
				Sections: []*fb2.Section{sec("Сборник «X»", sec("Рассказ А"), sec("Рассказ Б"))},
			},
			want: []Work{
				{"Рассказ А", "Роберт Шекли", "Рассказ А"},
				{"Рассказ Б", "Роберт Шекли", "Рассказ Б"},
			},
			wantColl: true,
		},
		{
			name: "multi-author anthology without author sections",
			book: fb2.Book{
				Title:    "Антология",
				Authors:  []fb2.Author{nosov, pidorenko},
				Sections: []*fb2.Section{sec("Рассказ А"), sec("Рассказ Б")},
			},
			want:     []Work{{"Рассказ А", "", "Рассказ А"}, {"Рассказ Б", "", "Рассказ Б"}},
			wantColl: true,
		},
		{
			name: "untitled and asterisk sections are transparent",
			book: fb2.Book{
				Title:    "Сборник",
				Authors:  []fb2.Author{sheckley},
				Sections: []*fb2.Section{sec("", sec("Рассказ А")), sec("* * *"), sec("Рассказ Б")},
			},
			want: []Work{
				{"Рассказ А", "Роберт Шекли", "Рассказ А"},
				{"Рассказ Б", "Роберт Шекли", "Рассказ Б"},
			},
			wantColl: true,
		},
		{
			name: "author section with initials",
			book: fb2.Book{
				Title:    "Альманах",
				Authors:  []fb2.Author{nosov, pidorenko},
				Sections: []*fb2.Section{sec("Е. Носов", sec("Повесть")), sec("И. Пидоренко", sec("Рассказ"))},
			},
			want: []Work{
				{"Повесть", "Евгений Носов", "Евгений Носов › Повесть"},
				{"Рассказ", "Игорь Пидоренко", "Игорь Пидоренко › Рассказ"},
			},
			wantColl: true,
		},
		{
			name: "one work plus back matter is not a collection",
			book: fb2.Book{
				Title:    "Повесть",
				Authors:  []fb2.Author{sheckley},
				Sections: []*fb2.Section{sec("Предисловие"), sec("Повесть о странном"), sec("Примечания")},
			},
			want:     []Work{{"Повесть", "Роберт Шекли", "Повесть"}},
			wantColl: false,
		},
		{
			name: "bare surname section is not an author section",
			book: fb2.Book{
				Title:   "Антология",
				Authors: []fb2.Author{nosov, pidorenko},
				Sections: []*fb2.Section{
					sec("Носов", sec("Повесть")),
					sec("Рассказ Б"),
				},
			},
			want: []Work{
				{"Носов", "", "Носов"},
				{"Рассказ Б", "", "Рассказ Б"},
			},
			wantColl: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, coll := Works(&tt.book)
			if coll != tt.wantColl {
				t.Errorf("isCollection = %v, want %v", coll, tt.wantColl)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("works =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

func TestMatchAuthor(t *testing.T) {
	nosov := fb2.Author{First: "Евгений", Middle: "Валентинович", Last: "Носов"}
	authors := []fb2.Author{nosov}

	tests := []struct {
		input string
		want  string
	}{
		{"Носов", ""},                        // bare surname, no match
		{"Евгений Носов", "Евгений Носов"},    // first + last
		{"Носов Евгений", "Евгений Носов"},    // last + first
		{"Е. Носов", "Евгений Носов"},         // initial + last
		{"Евгений Валентинович Носов", "Евгений Носов"}, // full name
		{"Евгений В. Носов", "Евгений Носов"}, // first + middle initial + last
	}

	for _, tt := range tests {
		norm := textnorm.Normalize(tt.input)
		got := matchAuthor(norm, authors)
		if got != tt.want {
			t.Errorf("matchAuthor(normalize(%q)) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsNoise(t *testing.T) {
	tests := map[string]bool{
		"1": true, "iv": true, "глава 5": true, "глава первая встреча": true,
		"часть вторая": true, "пролог": true, "примечания": true, "об авторе": true,
		"предисловие": true, "книга пятая": true,
		"книга мертвых": false, "чужие дети": false, "civil": false, "часть тела": false,
	}
	for in, want := range tests {
		if got := isNoise(in); got != want {
			t.Errorf("isNoise(%q) = %v, want %v", in, got, want)
		}
	}
}
