package main

import (
	"fmt"
	"sort"
	"strings"
)

func findAnagrams(words []string) map[string][]string {
	groups := make(map[string][]string)

	for _, w := range words {
		lower := strings.ToLower(w)
		key := sortedRunes(lower)

		groups[key] = append(groups[key], lower)
	}

	result := make(map[string][]string)
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}

		sort.Strings(group)
		keyWord := group[0]
		result[keyWord] = group
	}

	return result
}

func sortedRunes(s string) string {
	r := []rune(s)
	sort.Slice(r, func(i, j int) bool { return r[i] < r[j] })
	return string(r)
}

// пример использования
func main() {
	words := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}
	anagrams := findAnagrams(words)

	for k, group := range anagrams {
		fmt.Printf("– %q: %q\n", k, group)
	}
}
