package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

var (
	flagK = flag.Int("k", 0, "field to sort by")
	flagN = flag.Bool("n", false, "numeric sort")
	flagR = flag.Bool("r", false, "reverse sort")
	flagU = flag.Bool("u", false, "unique lines only")
	flagC = flag.Bool("c", false, "check if sorted")
	flagB = flag.Bool("b", false, "ignore trailing blanks")
	flagM = flag.Bool("M", false, "month sort")
	flagH = flag.Bool("h", false, "human sort")
)

var monthMap = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

type sorter struct {
	lines                []string
	key                  int
	numeric              bool
	reverse              bool
	fieldSep             string
	checkOnly            bool
	unique               bool
	ignoreTrailingBlanks bool
	monthSort            bool
	humanSort            bool
}

func (s *sorter) Len() int {
	return len(s.lines)
}

func (s *sorter) Swap(i, j int) {
	s.lines[i], s.lines[j] = s.lines[j], s.lines[i]
}

func (s *sorter) Less(i, j int) bool {
	vi := s.getKeyField(s.lines[i])
	vj := s.getKeyField(s.lines[j])

	if s.ignoreTrailingBlanks {
		vi = trimRight(vi) // Используем trimRight
		vj = trimRight(vj)
	}

	var less bool
	if s.numeric {
		fi, err1 := strconv.ParseFloat(vi, 64)
		fj, err2 := strconv.ParseFloat(vj, 64)
		if err1 == nil && err2 == nil {
			less = fi < fj
		} else {
			less = vi < vj
		}
	} else if s.humanSort {
		fi, err1 := parseHumanReadable(vi)
		fj, err2 := parseHumanReadable(vj)
		if err1 == nil && err2 == nil {
			less = fi < fj
		} else {
			less = vi < vj
		}
	} else if s.monthSort {
		mi, ok1 := monthMap[strings.ToLower(vi)]
		mj, ok2 := monthMap[strings.ToLower(vj)]
		if ok1 && ok2 {
			less = mi < mj
		} else {
			less = vi < vj
		}
	} else {
		less = vi < vj
	}

	if s.reverse {
		return !less
	}
	return less
}

func (s *sorter) getKeyField(line string) string {
	if s.key <= 0 {
		return line
	}
	fields := strings.Split(line, s.fieldSep)
	if s.key-1 < len(fields) {
		return fields[s.key-1]
	}
	return ""
}

func parseHumanReadable(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)

	var multiplier float64 = 1
	var numStr string

	switch {
	case strings.HasSuffix(s, "K"):
		multiplier = 1024
		numStr = strings.TrimSuffix(s, "K")
	case strings.HasSuffix(s, "M"):
		multiplier = 1024 * 1024
		numStr = strings.TrimSuffix(s, "M")
	case strings.HasSuffix(s, "G"):
		multiplier = 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(s, "G")
	default:
		numStr = s
	}

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}

	return num * multiplier, nil
}

func main() {
	flag.Parse()

	var input io.Reader = os.Stdin
	if flag.NArg() > 0 {
		fname := flag.Arg(0)
		f, err := os.Open(fname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", fname, err)
			os.Exit(1)
		}
		defer f.Close()
		input = f
	}

	scanner := bufio.NewScanner(input)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading: %v\n", err)
		os.Exit(1)
	}

	s := &sorter{
		lines:                lines,
		key:                  *flagK,
		numeric:              *flagN,
		reverse:              *flagR,
		fieldSep:             "\t",
		checkOnly:            *flagC,
		unique:               *flagU,
		ignoreTrailingBlanks: *flagB,
		monthSort:            *flagM,
		humanSort:            *flagH,
	}

	isSorted := sort.SliceIsSorted(s.lines, func(i, j int) bool {
		return s.Less(i, j)
	})

	if s.checkOnly {
		if !isSorted {
			fmt.Fprintf(os.Stderr, "File is not sorted\n")
			os.Exit(1)
		}
		return
	}

	if !isSorted {
		sort.Sort(s)
	}

	if s.unique {
		printUniq(s.lines)
	} else {
		for _, line := range s.lines {
			fmt.Println(line)
		}
	}
}

func printUniq(lines []string) {
	if len(lines) == 0 {
		return
	}
	prev := lines[0]
	fmt.Println(prev)
	for _, line := range lines[1:] {
		if line != prev {
			fmt.Println(line)
			prev = line
		}
	}
}

func trimRight(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}
