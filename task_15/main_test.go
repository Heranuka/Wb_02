package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseLine(t *testing.T) {
	input := `echo "hello world" && ls -l | grep main > out.txt`
	cmds, err := parseLine(input)
	if err != nil {
		t.Fatal("parseLine error:", err)
	}
	if len(cmds) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(cmds))
	}
	if cmds[0].args[0] != "echo" || !cmds[0].andNext {
		t.Error("cmd1 parsing incorrect")
	}
	if cmds[1].args[0] != "ls" || !cmds[1].pipeNext {
		t.Error("cmd2 parsing incorrect")
	}
	if cmds[2].args[0] != "grep" || cmds[2].redirectOut != "out.txt" {
		t.Error("cmd3 parsing or redirect incorrect")
	}
}

func TestTokenize(t *testing.T) {
	line := `echo "hello world" && ps aux | grep "go run" || echo fail`
	tokens := tokenize(line)
	expected := []string{"echo", "hello world", "&&", "ps", "aux", "|", "grep", "go run", "||", "echo", "fail"}
	if len(tokens) != len(expected) {
		t.Fatalf("token count mismatch: expected %d, got %d", len(expected), len(tokens))
	}
	for i := range tokens {
		if tokens[i] != expected[i] {
			t.Errorf("token %d: expected '%s', got '%s'", i, expected[i], tokens[i])
		}
	}
}

func TestSubstituteEnvVars(t *testing.T) {
	os.Setenv("TESTVAR", "value123")
	defer os.Unsetenv("TESTVAR")

	line := "echo $TESTVAR and $UNSETVAR"
	result := substituteEnvVars(line)
	if !strings.Contains(result, "value123") {
		t.Error("substituteEnvVars: expected to contain 'value123'")
	}
	if strings.Contains(result, "$UNSETVAR") {
		t.Error("substituteEnvVars: expected unset var to be empty")
	}
}

func TestRunBuiltinEcho(t *testing.T) {
	cmd := command{args: []string{"echo", "hello", "world"}}
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exit := runBuiltin(cmd)

	w.Close()
	os.Stdout = old

	outBuf := new(bytes.Buffer)
	io.Copy(outBuf, r)

	if exit != 0 {
		t.Errorf("runBuiltin echo exit code = %d, want 0", exit)
	}
	if !strings.Contains(outBuf.String(), "hello world") {
		t.Errorf("runBuiltin echo output = %q, want to contain 'hello world'", outBuf.String())
	}
}

func TestRunBuiltinCd(t *testing.T) {
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)

	tempDir := os.TempDir()

	err := os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	cmd := command{args: []string{"cd", oldDir}}
	exit := runBuiltin(cmd)
	if exit != 0 {
		t.Fatal("cd command failed")
	}

	cwd, _ := os.Getwd()
	realCwd, _ := filepath.EvalSymlinks(cwd)
	realOldDir, _ := filepath.EvalSymlinks(oldDir)

	if realCwd != realOldDir {
		t.Errorf("expected cwd %q, got %q", realOldDir, realCwd)
	}
}

func TestRunExternal(t *testing.T) {
	path, err := exec.LookPath("echo")
	if err != nil {
		t.Skip("echo command not found")
	}
	cmd := command{args: []string{path, "testexternal"}}
	execCmd, err := runExternal(cmd, nil)
	if err != nil {
		t.Fatalf("runExternal error: %v", err)
	}
	exitCode := waitCommand(execCmd)
	t.Logf("exit code: %d", exitCode)
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}

func TestRunFullPipeline(t *testing.T) {
	cmds := []command{
		{args: []string{"echo", "hello"}},
		{args: []string{"grep", "he"}},
	}

	exitCode := runFullPipeline(cmds)
	if exitCode != 0 {
		t.Errorf("runFullPipeline exit code = %d; want 0", exitCode)
	}
}

func TestRunBuiltinKillInvalid(t *testing.T) {
	cmd := command{args: []string{"kill", "abc"}}
	exit := runBuiltin(cmd)
	if exit == 0 {
		t.Error("kill with invalid pid should fail")
	}
}

func TestRunBuiltinPs(t *testing.T) {
	cmd := command{args: []string{"ps"}}
	exit := runBuiltin(cmd)
	if exit != 0 {
		t.Error("ps command failed")
	}
}
