package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func runGrep(t *testing.T, args ...string) (string, error) {
	cmd := exec.Command(os.Args[0], args...)
	// Передать переменную окружения чтобы запускать тесты как отдельный процесс
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func TestBasicMatch(t *testing.T) {
	content := `foo
bar
baz
foo
qux`
	filename := "testfile.txt"
	os.WriteFile(filename, []byte(content), 0644)
	defer os.Remove(filename)

	out, err := runGrep(t, "-n", "foo", filename)
	if err != nil {
		t.Fatal(err)
	}
	expectedLines := []string{"1:foo", "4:foo"}
	for _, line := range expectedLines {
		if !strings.Contains(out, line) {
			t.Errorf("Expected output to contain %s, got: %s", line, out)
		}
	}
}

func TestCaseInsensitive(t *testing.T) {
	content := `Hello
world
HELLO
World`
	filename := "testfile2.txt"
	os.WriteFile(filename, []byte(content), 0644)
	defer os.Remove(filename)

	out, err := runGrep(t, "-i", "hello", filename)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "1:Hello") || !strings.Contains(out, "3:HELLO") {
		t.Errorf("Expected matches with -i to contain both lines, got: %s", out)
	}
}

func TestCountOnly(t *testing.T) {
	content := `one
two
three
two
one`
	filename := "testfile3.txt"
	os.WriteFile(filename, []byte(content), 0644)
	defer os.Remove(filename)

	out, err := runGrep(t, "-c", "two", filename)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "2" {
		t.Errorf("Expected count 2, got: %s", out)
	}
}

func TestInvertMatch(t *testing.T) {
	content := `apple
banana
apricot
blueberry`
	filename := "testfile4.txt"
	os.WriteFile(filename, []byte(content), 0644)
	defer os.Remove(filename)

	out, err := runGrep(t, "-v", "ap", filename)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "apple") || strings.Contains(out, "apricot") {
		t.Errorf("Invert match failed, output: %s", out)
	}
	if !strings.Contains(out, "banana") && !strings.Contains(out, "blueberry") {
		t.Errorf("Unexpected output: %s", out)
	}
}

func TestContextLines(t *testing.T) {
	content := `line1
match
line3
line4
match again
line6`
	filename := "testfile5.txt"
	os.WriteFile(filename, []byte(content), 0644)
	defer os.Remove(filename)

	out, err := runGrep(t, "-C", "1", "match", filename)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "line1") || !strings.Contains(out, "line3") || !strings.Contains(out, "line4") || !strings.Contains(out, "line6") {
		t.Errorf("Context lines not included properly, output: %s", out)
	}
}

func TestBeforeAfterFlags(t *testing.T) {
	content := `line1
match
line3
line4
match
line6`
	filename := "testfile6.txt"
	os.WriteFile(filename, []byte(content), 0644)
	defer os.Remove(filename)

	out, err := runGrep(t, "-A", "1", "-B", "1", "match", filename)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "line1") || !strings.Contains(out, "line3") || !strings.Contains(out, "line4") || !strings.Contains(out, "line6") {
		t.Errorf("Expected context lines with -A and -B, got: %s", out)
	}
}

func TestOutputWithoutLineNumber(t *testing.T) {
	content := `foo
bar`
	filename := "testfile7.txt"
	os.WriteFile(filename, []byte(content), 0644)
	defer os.Remove(filename)

	out, err := runGrep(t, "foo", filename)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, ":") {
		t.Errorf("Expected no line numbers, got: %s", out)
	}
}

func TestMultipleIntervalsMerge(t *testing.T) {
	content := `line1
match1
line3
line4
match2
line6`
	filename := "testfile8.txt"
	os.WriteFile(filename, []byte(content), 0644)
	defer os.Remove(filename)

	out, err := runGrep(t, "-C", "0", "match", filename)
	if err != nil {
		t.Fatal(err)
	}
	// Should include all lines with matches and possibly inline context
	if !strings.Contains(out, "match1") || !strings.Contains(out, "match2") {
		t.Errorf("Expected to find both matches, got: %s", out)
	}
}
