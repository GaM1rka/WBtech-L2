package main

import (
	"fmt"
	"sort"
	"strings"
)

// функция для сортировки букв в слове
func sortRunes(s string) string {
	runes := []rune(s)
	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})
	return string(runes)
}

// основная функция
func FindAnagramSets(words []string) map[string][]string {
	anagramGroups := make(map[string][]string)

	for _, word := range words {
		w := strings.ToLower(word) // в нижний регистр
		key := sortRunes(w)        // отсортированные буквы — "каноническая форма"
		anagramGroups[key] = append(anagramGroups[key], w)
	}

	result := make(map[string][]string)

	for _, group := range anagramGroups {
		if len(group) > 1 { // игнорируем множества из одного слова
			sort.Strings(group)      // сортировка слов внутри группы
			result[group[0]] = group // ключом берем первое слово после сортировки
		}
	}

	return result
}

func main() {
	words := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}

	res := FindAnagramSets(words)

	for key, group := range res {
		fmt.Printf("%q: %v\n", key, group)
	}
}
