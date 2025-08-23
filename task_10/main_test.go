package main

import (
	"os/exec"
	"sort"
	"strings"
	"testing"
)

func runSort(args []string, input string) (string, error) {
	cmd := exec.Command("./main", args...)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestBasicSort(t *testing.T) {
	input := "c\nb\na\n"
	expected := "a\nb\nc\n"
	out, err := runSort([]string{}, input)
	if err != nil {
		t.Fatalf("sort failed: %v", err)
	}
	if out != expected {
		t.Errorf("sort output mismatch:\nexpected: %q\nactual:   %q", expected, out)
	}
}

func TestReverseSort(t *testing.T) {
	input := "c\nb\na\n"
	expected := "c\nb\na\n"
	out, err := runSort([]string{"-r"}, input)
	if err != nil {
		t.Fatalf("sort -r failed: %v", err)
	}
	if out != expected {
		t.Errorf("sort -r output mismatch:\nexpected: %q\nactual:   %q", expected, out)
	}
}

func TestNumericSort(t *testing.T) {
	input := "10\n2\n1\n"
	expected := "1\n2\n10\n"
	out, err := runSort([]string{"-n"}, input)
	if err != nil {
		t.Fatalf("sort -n failed: %v", err)
	}
	if out != expected {
		t.Errorf("sort -n output mismatch:\nexpected: %q\nactual:   %q", expected, out)
	}
}

func TestReverseNumericSort(t *testing.T) {
	input := "10\n2\n1\n"
	expected := "10\n2\n1\n"
	out, err := runSort([]string{"-n", "-r"}, input)
	if err != nil {
		t.Fatalf("sort -nr failed: %v", err)
	}
	if out != expected {
		t.Errorf("sort -nr output mismatch:\nexpected: %q\nactual:   %q", expected, out)
	}
}

func TestUniqueSort(t *testing.T) {
	input := "a\nb\na\nc\nb\n"
	expected := "a\nb\nc\n"
	out, err := runSort([]string{"-u"}, input)
	if err != nil {
		t.Fatalf("sort -u failed: %v", err)
	}
	if out != expected {
		t.Errorf("sort -u output mismatch:\nexpected: %q\nactual:   %q", expected, out)
	}
}

func TestColumnSort(t *testing.T) {
	input := "a\t3\nb\t1\nc\t2\n"
	expected := "b\t1\nc\t2\na\t3\n"
	out, err := runSort([]string{"-k", "2"}, input)
	if err != nil {
		t.Fatalf("sort -k 2 failed: %v", err)
	}
	if out != expected {
		t.Errorf("sort -k 2 output mismatch:\nexpected: %q\nactual:   %q", expected, out)
	}
}

func TestNumericColumnSort(t *testing.T) {
	input := "a\t10\nb\t2\nc\t1\n"
	expected := "c\t1\nb\t2\na\t10\n"
	out, err := runSort([]string{"-k", "2", "-n"}, input)
	if err != nil {
		t.Fatalf("sort -k 2 -n failed: %v", err)
	}
	if out != expected {
		t.Errorf("sort -k 2 -n output mismatch:\nexpected: %q\nactual:   %q", expected, out)
	}

}
func TestIgnoreBlanks(t *testing.T) {
	input := "a   \nb\nc \n"
	expected := "b\na   \nc \n"

	// Split the input into lines
	lines := strings.Split(input, "\n")

	//Filter out empty strings
	var filteredLines []string
	for _, line := range lines {
		if line != "" {
			filteredLines = append(filteredLines, line)
		}
	}

	// Join the lines back with newlines after sorting
	expected = strings.Join(sortLines(filteredLines, "-b"), "\n") + "\n"

	out, err := runSort([]string{"-b"}, input)
	if err != nil {
		t.Fatalf("sort -b failed: %v", err)
	}
	if out != expected {
		t.Errorf("sort -b output mismatch:\nexpected: %q\nactual:   %q", expected, out)
	}
}

func sortLines(lines []string, flag string) []string {
	// Sort the lines according to the -b flag
	sort.Slice(lines, func(i, j int) bool {
		a := lines[i]
		b := lines[j]

		if flag == "-b" {
			a = trimRight(a)
			b = trimRight(b)
		}
		return a < b
	})
	return lines
}

func TestMonthSort(t *testing.T) {
	input := "Feb\nJan\nDec\n"
	expected := "Jan\nFeb\nDec\n"
	out, err := runSort([]string{"-M"}, input)
	if err != nil {
		t.Fatalf("sort -M failed: %v", err)
	}
	if out != expected {
		t.Errorf("sort -M output mismatch:\nexpected: %q\nactual:   %q", expected, out)
	}
}

func TestHumanSort(t *testing.T) {
	input := "10K\n2M\n1G\n"
	expected := "10K\n2M\n1G\n"
	out, err := runSort([]string{"-h"}, input)
	if err != nil {
		t.Fatalf("sort -h failed: %v", err)
	}
	if out != expected {
		t.Errorf("sort -h output mismatch:\nexpected: %q\nactual:   %q", expected, out)
	}
}

func TestCheckSorted(t *testing.T) {
	input := "a\nb\nc\n"
	_, err := runSort([]string{"-c"}, input)
	if err != nil {
		t.Fatalf("sort -c failed on sorted input: %v", err)
	}

	input = "b\na\nc\n"
	_, err = runSort([]string{"-c"}, input)
	if err == nil {
		t.Fatalf("sort -c should have failed on unsorted input")
	}

	expectedError := "exit status 1"
	if err.Error() != expectedError {
		t.Errorf("sort -c returned incorrect error: expected %q, got %q", expectedError, err.Error())
	}
}

func TestCheckColumnSort(t *testing.T) {
	input := "a\t1\nb\t2\nc\t3\n"
	_, err := runSort([]string{"-c", "-k", "2", "-n"}, input)
	if err != nil {
		t.Fatalf("sort -c -k 2 -n failed on sorted input: %v", err)
	}

	input = "a\t2\nb\t1\nc\t3\n"
	_, err = runSort([]string{"-c", "-k", "2", "-n"}, input)
	if err == nil {
		t.Fatalf("sort -c -k 2 -n should have failed on unsorted input")
	}

	expectedError := "exit status 1"
	if err.Error() != expectedError {
		t.Errorf("sort -c -k 2 -n returned incorrect error: expected %q, got %q", expectedError, err.Error())
	}
}

func TestMain_MissingFile(t *testing.T) {
	_, err := runSort([]string{"missing_file.txt"}, "")
	if err == nil {
		t.Fatalf("sort should have failed on missing file")
	}

	expectedError := "exit status 1"
	if err.Error() != expectedError {
		t.Errorf("sort missing file returned incorrect error: expected %q, got %q", expectedError, err.Error())
	}
}

func TestEmptyInput(t *testing.T) {
	input := ""
	expected := ""
	out, err := runSort([]string{}, input)
	if err != nil {
		t.Fatalf("sort failed on empty input: %v", err)
	}
	if out != expected {
		t.Errorf("sort output mismatch on empty input:\nexpected: %q\nactual:   %q", expected, out)
	}
}
