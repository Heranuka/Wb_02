package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func runGrep(t *testing.T, args ...string) (string, error) {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	cmd.Args = append(cmd.Args, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	filename := t.TempDir() + "/testfile.txt"
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	return filename
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	os.Args = append([]string{os.Args[0]}, os.Args[3:]...)

	main()

	os.Exit(0)
}

func TestBasicMatch(t *testing.T) {
	t.Run("BasicMatch", func(t *testing.T) {
		content := `foo
bar
baz
foo
qux`
		filename := writeTempFile(t, content)

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
	})
}

func TestCaseInsensitive(t *testing.T) {
	t.Run("CaseInsensitive", func(t *testing.T) {
		content := `Hello
world
HELLO
World`
		filename := writeTempFile(t, content)

		out, err := runGrep(t, "-i", "-n", "hello", filename)

		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "1:Hello") || !strings.Contains(out, "3:HELLO") {
			t.Errorf("Expected matches with -i to contain both lines, got: %s", out)
		}
	})
}

func TestCountOnly(t *testing.T) {
	t.Run("CountOnly", func(t *testing.T) {
		content := `one
two
three
two
one`
		filename := writeTempFile(t, content)

		out, err := runGrep(t, "-c", "two", filename)
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(out) != "2" {
			t.Errorf("Expected count 2, got: %s", out)
		}
	})
}

func TestInvertMatch(t *testing.T) {
	t.Run("InvertMatch", func(t *testing.T) {
		content := `apple
banana
apricot
blueberry`
		filename := writeTempFile(t, content)

		out, err := runGrep(t, "-v", "ap", filename)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(out, "apple") || strings.Contains(out, "apricot") {
			t.Errorf("Invert match failed, output: %s", out)
		}
		if !strings.Contains(out, "banana") || !strings.Contains(out, "blueberry") {
			t.Errorf("Expected lines 'banana' and 'blueberry' in output, got: %s", out)
		}
	})
}

func TestContextLines(t *testing.T) {
	t.Run("ContextLines", func(t *testing.T) {
		content := `line1
match
line3
line4
match again
line6`
		filename := writeTempFile(t, content)

		out, err := runGrep(t, "-C", "1", "match", filename)
		if err != nil {
			t.Fatal(err)
		}
		expected := []string{"line1", "line3", "line4", "line6"}
		for _, line := range expected {
			if !strings.Contains(out, line) {
				t.Errorf("Expected context lines %s in output, got: %s", line, out)
			}
		}
	})
}

func TestBeforeAfterFlags(t *testing.T) {
	t.Run("BeforeAfterFlags", func(t *testing.T) {
		content := `line1
match
line3
line4
match
line6`
		filename := writeTempFile(t, content)

		out, err := runGrep(t, "-A", "1", "-B", "1", "match", filename)
		if err != nil {
			t.Fatal(err)
		}
		expected := []string{"line1", "line3", "line4", "line6"}
		for _, line := range expected {
			if !strings.Contains(out, line) {
				t.Errorf("Expected context lines with -A and -B %s in output, got: %s", line, out)
			}
		}
	})
}

func TestOutputWithoutLineNumber(t *testing.T) {
	t.Run("OutputWithoutLineNumber", func(t *testing.T) {
		content := `foo
bar`
		filename := writeTempFile(t, content)

		out, err := runGrep(t, "foo", filename)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(out, ":") {
			t.Errorf("Expected no line numbers, got: %s", out)
		}
	})
}

func TestMultipleIntervalsMerge(t *testing.T) {
	t.Run("MultipleIntervalsMerge", func(t *testing.T) {
		content := `line1
match1
line3
line4
match2
line6`
		filename := writeTempFile(t, content)

		out, err := runGrep(t, "-C", "0", "match", filename)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "match1") || !strings.Contains(out, "match2") {
			t.Errorf("Expected to find both matches, got: %s", out)
		}
	})
}

func TestFixedStringFlag(t *testing.T) {
	t.Run("FixedStringFlag", func(t *testing.T) {
		content := `foo.bar
foo?bar
foobar`
		filename := writeTempFile(t, content)

		outRegexp, err := runGrep(t, "foo.bar", filename)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(outRegexp, "foo.bar") {
			t.Errorf("Expected match with regexp, got: %s", outRegexp)
		}

		outFixed, err := runGrep(t, "-F", "foo.bar", filename)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(outFixed, "foo.bar") || strings.Contains(outFixed, "foo?bar") {
			t.Errorf("Fixed string match failed, got: %s", outFixed)
		}
	})
}

func TestEmptyFile(t *testing.T) {
	t.Run("EmptyFile", func(t *testing.T) {
		filename := writeTempFile(t, "")

		out, err := runGrep(t, "anything", filename)
		if err != nil {
			t.Fatal(err)
		}
		if out != "" {
			t.Errorf("Expected empty output for empty file, got: %s", out)
		}
	})
}

func TestInvalidRegexp(t *testing.T) {
	t.Run("InvalidRegexp", func(t *testing.T) {
		content := `something`
		filename := writeTempFile(t, content)

		_, err := runGrep(t, "[", filename)
		if err == nil {
			t.Errorf("Expected error for invalid regexp, got nil")
		}
	})
}
