package main

import (
	"fmt"
	"sort"
	"strings"
)

// findAnagrams принимает срез слов и возвращает map,
// где ключ — первое встретившееся слово множества анаграмм,
// значение — срез всех анаграмм (включая ключ), отсортированных.
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
			continue // пропускаем одиночки
		}

		sort.Strings(group)
		keyWord := group[0] // ключ — первое слово по алфавиту среди анаграмм
		result[keyWord] = group
	}

	return result
}

// sortedRunes возвращает строку с буквами исходной, отсортированными по возрастанию
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
