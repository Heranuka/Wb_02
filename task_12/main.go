package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
)

func main() {
	// Флаги
	afterLines := flag.Int("A", 0, "вывести N строк после найденной")
	beforeLines := flag.Int("B", 0, "вывести N строк до найденной")
	contextLines := flag.Int("C", 0, "вывести N строк контекста вокруг найденной (эквивалент -A N -B N)")
	countOnly := flag.Bool("c", false, "вывести только количество совпадающих строк")
	ignoreCase := flag.Bool("i", false, "игнорировать регистр")
	invertMatch := flag.Bool("v", false, "инвертировать совпадения")
	fixedString := flag.Bool("F", false, "шаблон — фиксированная строка, а не регулярное выражение")
	lineNumber := flag.Bool("n", false, "выводить номер строки перед выводом")

	flag.Parse()

	// Контексты -C задаёт оба
	if *contextLines > 0 {
		*afterLines = *contextLines
		*beforeLines = *contextLines
	}

	args := flag.Args()
	var inputFile io.Reader
	if len(args) > 0 {
		f, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка открытия файла: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		inputFile = f
	} else {
		inputFile = os.Stdin
	}

	if len(args) == 0 {
		fmt.Fprint(os.Stderr, "Необходимо указать шаблон и опционально файл\nПример: grep -i -n -C 2 \"pattern\" file.txt\n")
		if len(args) == 0 && flag.NArg() == 0 {
			os.Exit(1)
		}
	}

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Ошибка: не задан шаблон поиска")
		os.Exit(1)
	}

	pattern := args[0]

	var re *regexp.Regexp
	var err error

	if *fixedString {
		// Экранируем спецсимволы если fixed string
		pattern = regexp.QuoteMeta(pattern)
	}
	if *ignoreCase {
		pattern = "(?i)" + pattern
	}

	re, err = regexp.Compile(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка компиляции регулярного выражения: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(inputFile)

	type line struct {
		number int
		text   string
	}

	// Считываем все строки в память (для поддержки контекста)
	var lines []line
	for scanner.Scan() {
		lines = append(lines, line{number: len(lines) + 1, text: scanner.Text()})
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения: %v\n", err)
		os.Exit(1)
	}

	// Найдём номера строк, которые соответствуют (с учётом флага -v)
	var matchedLines []int
	for i, l := range lines {
		matches := re.MatchString(l.text)
		if *invertMatch {
			matches = !matches
		}
		if matches {
			matchedLines = append(matchedLines, i)
		}
	}

	if *countOnly {
		// Просто вывести количество
		fmt.Println(len(matchedLines))
		return
	}

	// Определяем диапазоны строк для вывода с контекстом
	type interval struct{ start, end int }
	var intervals []interval

	// Добавляем каждое совпадение с контекстом
	for _, idx := range matchedLines {
		start := idx - *beforeLines
		if start < 0 {
			start = 0
		}
		end := idx + *afterLines
		if end >= len(lines) {
			end = len(lines) - 1
		}
		intervals = append(intervals, interval{start, end})
	}

	// Объединяем пересекающиеся и соседние интервалы
	if len(intervals) == 0 {
		// Нет совпадений — ничего не выводим
		return
	}
	merged := []interval{intervals[0]}
	for i := 1; i < len(intervals); i++ {
		last := &merged[len(merged)-1]
		if intervals[i].start <= last.end+1 {
			// Пересечение или смежные интервалы — объединяем
			if intervals[i].end > last.end {
				last.end = intervals[i].end
			}
		} else {
			merged = append(merged, intervals[i])
		}
	}

	// Выводим строки из объединённых интервалов
	for _, iv := range merged {
		for i := iv.start; i <= iv.end; i++ {
			l := lines[i]
			if *lineNumber {
				fmt.Printf("%d:%s\n", l.number, l.text)
			} else {
				fmt.Println(l.text)
			}
		}
		// Между интервалами выводим символ отделителя (--) если есть более одного интервала
		if len(merged) > 1 && iv != merged[len(merged)-1] {
			fmt.Println("--")
		}
	}
}
