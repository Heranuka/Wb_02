package main

import (
	"reflect"
	"testing"
)

func TestFindAnagrams(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want map[string][]string
	}{
		{
			name: "Basic example",
			in:   []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"},
			want: map[string][]string{
				"пятак":  {"пятак", "пятка", "тяпка"},
				"листок": {"листок", "слиток", "столик"},
			},
		},
		{
			name: "No anagrams",
			in:   []string{"машина", "автомобиль", "самолет"},
			want: map[string][]string{},
		},
		{
			name: "Case insensitivity",
			in:   []string{"Пятак", "ПЯТКА", "Тяпка"},
			want: map[string][]string{
				"пятак": {"пятак", "пятка", "тяпка"},
			},
		},
		{
			name: "Single word input",
			in:   []string{"один"},
			want: map[string][]string{},
		},
		{
			name: "Empty input",
			in:   []string{},
			want: map[string][]string{},
		},
		{
			name: "Mixed group sizes",
			in:   []string{"кино", "инок", "ник", "кин", "ник", "кин"},
			want: map[string][]string{
				"инок": {"инок", "кино"},
				"кин":  {"кин", "кин", "ник", "ник"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findAnagrams(tt.in)

			// Проверяем совпадение ключей
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d groups, got %d", len(tt.want), len(got))
			}

			for k, wantGroup := range tt.want {
				gotGroup, ok := got[k]
				if !ok {
					t.Errorf("key %q not found in result", k)
					continue
				}

				if !reflect.DeepEqual(wantGroup, gotGroup) {
					t.Errorf("for key %q: expected group %v, got %v", k, wantGroup, gotGroup)
				}
			}
		})
	}
}
