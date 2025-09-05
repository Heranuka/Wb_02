package main

import (
	"os/exec"
	"strings"
	"testing"
)

func runCut(args []string, input string) (string, error) {
	cmd := exec.Command("./main", args...)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestBasicCut(t *testing.T) {
	input := "a\tb\tc\td\te\n"
	expected := "a\tc\td\n"
	out, err := runCut([]string{"-f", "1,3-4"}, input)
	if err != nil {
		t.Fatalf("cut -f 1,3-4 failed: %v", err)
	}
	if out != expected {
		t.Errorf("cut -f 1,3-4 output mismatch:\nexpected: %q\nactual:   %q", expected, out)
	}
}

func TestCustomDelimiter(t *testing.T) {
	input := "a,b,c,d,e\n"
	expected := "a,c,d\n"
	out, err := runCut([]string{"-f", "1,3-4", "-d", ","}, input)
	if err != nil {
		t.Fatalf("cut -f 1,3-4 -d , failed: %v", err)
	}
	if out != expected {
		t.Errorf("cut -f 1,3-4 -d , output mismatch:\nexpected: %q\nactual:   %q", expected, out)
	}
}

func TestOutOfRange(t *testing.T) {
	input := "a\tb\n"
	expected := "\n"
	out, err := runCut([]string{"-f", "3"}, input)
	if err != nil {
		t.Fatalf("cut -f 3 failed: %v", err)
	}
	if out != expected {
		t.Errorf("cut -f 3 output mismatch:\nexpected: %q\nactual:   %q", expected, out)
	}
}

func TestEmptyInput(t *testing.T) {
	input := ""
	expected := ""
	out, err := runCut([]string{"-f", "1"}, input)
	if err != nil {
		t.Fatalf("cut -f 1 failed on empty input: %v", err)
	}
	if out != expected {
		t.Errorf("cut -f 1 output mismatch on empty input:\nexpected: %q\nactual:   %q", expected, out)
	}
}

func TestNoDelimiter(t *testing.T) {
	input := "abcde\n"
	expected := "abcde\n"

	out, err := runCut([]string{"-f", "1"}, input)
	if err != nil {
		t.Fatalf("cut -f 1 failed on no delimiter input: %v", err)
	}
	if out != expected {
		t.Errorf("cut -f 1 output mismatch on no delimiter input:\nexpected: %q\nactual:   %q", expected, out)
	}
}

func TestStdinRedirection(t *testing.T) {
	input := "a\tb\tc\n"
	expected := "b\n"

	cmd := exec.Command("./main", "-f", "2")
	cmd.Stdin = strings.NewReader(input)
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run command: %v", err)
	}

	outStr := string(outBytes)
	if outStr != expected {
		t.Errorf("Output mismatch:\nExpected: %q\nActual:   %q", expected, outStr)
	}
}

func TestStdinMultipleLines(t *testing.T) {
	input := "a\tb\tc\n" +
		"d\te\tf\n"
	expected := "b\ne\n"

	cmd := exec.Command("./main", "-f", "2")
	cmd.Stdin = strings.NewReader(input)
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run command: %v", err)
	}

	outStr := string(outBytes)
	if outStr != expected {
		t.Errorf("Output mismatch:\nExpected: %q\nActual:   %q", expected, outStr)
	}
}

func TestStdinWithCustomDelimiter(t *testing.T) {
	input := "a,b,c\n" +
		"d,e,f\n"
	expected := "b\ne\n"

	cmd := exec.Command("./main", "-f", "2", "-d", ",")
	cmd.Stdin = strings.NewReader(input)
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run command: %v", err)
	}

	outStr := string(outBytes)
	if outStr != expected {
		t.Errorf("Output mismatch:\nExpected: %q\nActual:   %q", expected, outStr)
	}
}

func TestSeparatedOnlyNoDelimiter(t *testing.T) {
	input := "abc\n" +
		"def\n"
	expected := ""

	cmd := exec.Command("./main", "-f", "1", "-s")
	cmd.Stdin = strings.NewReader(input)
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run command: %v", err)
	}

	outStr := string(outBytes)
	if outStr != expected {
		t.Errorf("Output mismatch:\nExpected: %q\nActual:   %q", expected, outStr)
	}
}

func TestEmptyField(t *testing.T) {
	input := "a\t\tc\n"
	expected := "\tc\n"

	out, err := runCut([]string{"-f", "2-3"}, input)

	if err != nil {
		t.Fatalf("cut -f 2-3 failed: %v", err)
	}

	if out != expected {
		t.Errorf("cut -f 2-3 output mismatch:\nexpected: %q\nactual:   %q", expected, out)
	}
}
func TestSeparatedOnly(t *testing.T) {
	input := "a\n" +
		"a\tb\tc\n" +
		"d\n" +
		"e\tf\n"
	expected := "a\tb\ne\tf\n"

	out, err := runCut([]string{"-f", "1-2", "-s"}, input)
	if err != nil {
		t.Fatalf("cut -f 1-2 -s failed: %v", err)
	}
	if out != expected {
		t.Errorf("cut -f 1-2 -s output mismatch:\nexpected: %q\nactual:   %q", expected, out)
	}
}

func TestInvalidFields(t *testing.T) {
	out, err := runCut([]string{"-f", "a"}, "input")
	if err == nil {
		t.Fatalf("cut -f a should have failed")
	}
	if !strings.Contains(out, "invalid field") {
		t.Errorf("cut -f a returned incorrect error in output: %q", out)
	}
}

func TestMissingFields(t *testing.T) {
	out, err := runCut([]string{"-f", ""}, "input")
	if err == nil {
		t.Fatalf("cut -f  should have failed")
	}
	if !strings.Contains(out, "no fields specified") {
		t.Errorf("cut -f returned incorrect error in output: %q", out)
	}
}
