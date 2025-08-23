package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	fieldsFlag    = flag.String("f", "", "fields to extract (comma-separated, ranges allowed)")
	delimiterFlag = flag.String("d", "\t", "delimiter character (default: tab)")
	separatedFlag = flag.Bool("s", false, "only lines containing delimiter")
)

func main() {
	flag.Parse()

	fields, err := parseFields(*fieldsFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// Для универсальности удаляем и \r и \n
		line = strings.TrimRight(line, "\r\n")

		result := processLine(line, *delimiterFlag, fields)
		output := strings.Join(result, *delimiterFlag)

		if *separatedFlag && !strings.Contains(line, *delimiterFlag) {
			continue
		}

		if *separatedFlag && output == "" {
			continue
		}

		fmt.Println(output)
	}
}

func parseFields(fieldsStr string) ([]int, error) {
	if fieldsStr == "" {
		return nil, fmt.Errorf("no fields specified")
	}

	var fields []int
	ranges := strings.Split(fieldsStr, ",")

	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if strings.Contains(r, "-") {
			parts := strings.Split(r, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", r)
			}

			start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid range start: %s", parts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid range end: %s", parts[1])
			}

			if start > end {
				return nil, fmt.Errorf("invalid range: start > end")
			}

			for i := start; i <= end; i++ {
				fields = append(fields, i)
			}
		} else {
			field, err := strconv.Atoi(r)
			if err != nil {
				return nil, fmt.Errorf("invalid field: %s", r)
			}
			fields = append(fields, field)
		}
	}

	return fields, nil
}

func processLine(line string, delimiter string, fields []int) []string {
	parts := strings.Split(line, delimiter)
	result := []string{}

	for _, field := range fields {
		if field > 0 && field <= len(parts) {
			result = append(result, parts[field-1])
		}
	}

	return result
}
