package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

type command struct {
	args        []string
	redirectIn  string
	redirectOut string
	pipeNext    bool
	andNext     bool
	orNext      bool
}

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("maxishell> ")
		if !scanner.Scan() {
			fmt.Println()
			break
		}

		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		select {
		case <-sigChan:
			fmt.Println("\nInterrupted")
			continue
		default:
		}

		line = substituteEnvVars(line)

		cmds, err := parseLine(line)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Parse error:", err)
			continue
		}

		i := 0
		for i < len(cmds) {
			pipeline := []command{cmds[i]}
			for i < len(cmds)-1 && cmds[i].pipeNext {
				i++
				pipeline = append(pipeline, cmds[i])
			}

			if len(pipeline) == 1 {
				c := pipeline[0]
				if isBuiltin(c.args[0]) {
					exitCode := runBuiltin(c)
					if c.andNext && exitCode != 0 {
						break
					}
					if c.orNext && exitCode == 0 {
						break
					}
					i++
					continue
				}
				cmd, err := runExternal(c, nil)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Execution error:", err)
					break
				}
				exitCode := waitCommand(cmd)
				if c.andNext && exitCode != 0 {
					break
				}
				if c.orNext && exitCode == 0 {
					break
				}
				i++
			} else {
				exitCode := runFullPipeline(pipeline)
				lastCmd := pipeline[len(pipeline)-1]
				if lastCmd.andNext && exitCode != 0 {
					break
				}
				if lastCmd.orNext && exitCode == 0 {
					break
				}
				i++
			}
		}
	}
}

func substituteEnvVars(line string) string {
	var buf bytes.Buffer
	for i := 0; i < len(line); i++ {
		if line[i] == '$' {
			varName := ""
			j := i + 1
			for ; j < len(line); j++ {
				if !(line[j] == '_' || ('a' <= line[j] && line[j] <= 'z') || ('A' <= line[j] && line[j] <= 'Z') || ('0' <= line[j] && line[j] <= '9')) {
					break
				}
				varName += string(line[j])
			}
			if val := os.Getenv(varName); val != "" {
				buf.WriteString(val)
			}
			i += len(varName)
		} else {
			buf.WriteByte(line[i])
		}
	}
	return buf.String()
}

func parseLine(line string) ([]command, error) {
	tokens := tokenize(line)
	if len(tokens) == 0 {
		return nil, errors.New("empty command")
	}

	var cmds []command
	var cur command
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]

		switch t {
		case "|":
			if len(cur.args) == 0 {
				return nil, errors.New("'|' unexpected")
			}
			cur.pipeNext = true
			cmds = append(cmds, cur)
			cur = command{}
		case "&&":
			if len(cur.args) == 0 {
				return nil, errors.New("'&&' unexpected")
			}
			cur.andNext = true
			cmds = append(cmds, cur)
			cur = command{}
		case "||":
			if len(cur.args) == 0 {
				return nil, errors.New("'||' unexpected")
			}
			cur.orNext = true
			cmds = append(cmds, cur)
			cur = command{}
		case ">":
			if i+1 >= len(tokens) {
				return nil, errors.New("expected filename after '>'")
			}
			i++
			cur.redirectOut = tokens[i]
		case "<":
			if i+1 >= len(tokens) {
				return nil, errors.New("expected filename after '<'")
			}
			i++
			cur.redirectIn = tokens[i]
		default:
			cur.args = append(cur.args, t)
		}
	}
	if len(cur.args) > 0 {
		cmds = append(cmds, cur)
	}
	return cmds, nil
}

func tokenize(line string) []string {
	var tokens []string
	buf := ""
	inQuote := false
	quoteChar := byte(0)
	for i := 0; i < len(line); i++ {
		ch := line[i]

		if inQuote {
			if ch == quoteChar {
				inQuote = false
			} else {
				buf += string(ch)
			}
			continue
		}

		if ch == '\'' || ch == '"' {
			inQuote = true
			quoteChar = ch
			continue
		}

		if ch == ' ' || ch == '\t' {
			if buf != "" {
				tokens = append(tokens, buf)
				buf = ""
			}
			continue
		}

		if ch == '&' && i+1 < len(line) && line[i+1] == '&' {
			if buf != "" {
				tokens = append(tokens, buf)
				buf = ""
			}
			tokens = append(tokens, "&&")
			i++
			continue
		}

		if ch == '|' && i+1 < len(line) && line[i+1] == '|' {
			if buf != "" {
				tokens = append(tokens, buf)
				buf = ""
			}
			tokens = append(tokens, "||")
			i++
			continue
		}

		if ch == '|' || ch == '>' || ch == '<' {
			if buf != "" {
				tokens = append(tokens, buf)
				buf = ""
			}
			tokens = append(tokens, string(ch))
			continue
		}

		buf += string(ch)
	}

	if buf != "" {
		tokens = append(tokens, buf)
	}

	return tokens
}

func isBuiltin(cmd string) bool {
	switch cmd {
	case "cd", "pwd", "echo", "kill", "ps", "exit":
		return true
	default:
		return false
	}
}

func runBuiltin(cmd command) int {
	switch cmd.args[0] {
	case "cd":
		if len(cmd.args) < 2 {
			fmt.Fprintln(os.Stderr, "cd: missing argument")
			return 1
		}
		if err := os.Chdir(cmd.args[1]); err != nil {
			fmt.Fprintln(os.Stderr, "cd:", err)
			return 1
		}
		return 0
	case "pwd":
		dir, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "pwd:", err)
			return 1
		}
		fmt.Println(dir)
		return 0
	case "echo":
		fmt.Println(strings.Join(cmd.args[1:], " "))
		return 0
	case "kill":
		if len(cmd.args) < 2 {
			fmt.Fprintln(os.Stderr, "kill: missing pid")
			return 1
		}
		pid, err := strconv.Atoi(cmd.args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "kill: invalid pid")
			return 1
		}
		err = syscall.Kill(pid, syscall.SIGTERM)
		if err != nil {
			fmt.Fprintln(os.Stderr, "kill:", err)
			return 1
		}
		return 0
	case "ps":
		psCmd := exec.Command("ps", "aux")
		psCmd.Stdout = os.Stdout
		psCmd.Stderr = os.Stderr
		if err := psCmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "ps error:", err)
			return 1
		}
		return 0
	case "exit":
		os.Exit(0)
	}
	return 0
}

func runFullPipeline(cmds []command) int {
	n := len(cmds)
	if n == 0 {
		return 0
	}

	procs := make([]*exec.Cmd, n)
	var lastStdout io.ReadCloser = nil

	for i, cmd := range cmds {
		procs[i] = exec.Command(cmd.args[0], cmd.args[1:]...)

		if i == 0 {
			if cmd.redirectIn != "" {
				f, err := os.Open(cmd.redirectIn)
				if err != nil {
					fmt.Fprintln(os.Stderr, "open input file error:", err)
					return 1
				}
				procs[i].Stdin = f
				defer f.Close()
			} else {
				procs[i].Stdin = os.Stdin
			}
		} else {
			procs[i].Stdin = lastStdout
		}

		if i == n-1 {
			if cmd.redirectOut != "" {
				f, err := os.Create(cmd.redirectOut)
				if err != nil {
					fmt.Fprintln(os.Stderr, "create output file error:", err)
					return 1
				}
				procs[i].Stdout = f
				procs[i].Stderr = f
				defer f.Close()
			} else {
				procs[i].Stdout = os.Stdout
				procs[i].Stderr = os.Stderr
			}
		} else {
			stdoutPipe, err := procs[i].StdoutPipe()
			if err != nil {
				fmt.Fprintln(os.Stderr, "stdout pipe error:", err)
				return 1
			}
			procs[i].Stderr = os.Stderr
			lastStdout = stdoutPipe
		}
	}

	for _, proc := range procs {
		if err := proc.Start(); err != nil {
			fmt.Fprintln(os.Stderr, "start error:", err)
			return 1
		}
	}

	var exitCode int
	for _, proc := range procs {
		if err := proc.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = 1
			}
		} else {
			exitCode = 0
		}
	}

	return exitCode
}

func runExternal(cmd command, prevCmd *exec.Cmd) (*exec.Cmd, error) {
	execCmd := exec.Command(cmd.args[0], cmd.args[1:]...)
	var stderr bytes.Buffer
	execCmd.Stderr = &stderr
	execCmd.Stdout = os.Stdout
	execCmd.Stdin = os.Stdin

	err := execCmd.Start()
	if err != nil {
		return nil, fmt.Errorf("start error: %v, stderr: %s", err, stderr.String())
	}
	return execCmd, nil
}

func waitCommand(cmd *exec.Cmd) int {
	if cmd == nil {
		return 0
	}
	err := cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}
