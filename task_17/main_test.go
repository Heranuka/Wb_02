package main

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"
)

func startTestEchoServer(t *testing.T) (net.Listener, string) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				scanner := bufio.NewScanner(c)
				for scanner.Scan() {
					line := scanner.Text()
					c.Write([]byte(line + "\n"))
				}
			}(conn)
		}
	}()
	return l, l.Addr().String()
}

func TestTelnetEcho(t *testing.T) {
	listener, addr := startTestEchoServer(t)
	defer listener.Close()

	done := make(chan struct{})
	go func() {
		timeout := 3 * time.Second
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			t.Errorf("failed to connect: %v", err)
			close(done)
			return
		}
		defer conn.Close()

		msg := "hello telnet"
		_, err = conn.Write([]byte(msg + "\n"))
		if err != nil {
			t.Errorf("failed to write: %v", err)
			close(done)
			return
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			t.Errorf("failed to read: %v", err)
			close(done)
			return
		}

		got := strings.TrimSpace(string(buf[:n]))
		if got != msg {
			t.Errorf("unexpected echo: got %q, want %q", got, msg)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("test timeout")
	}
}

func TestConnectTimeout(t *testing.T) {
	_, err := net.DialTimeout("tcp", "10.255.255.1:12345", 1*time.Second)
	if err == nil {
		t.Fatal("expected timeout error but got nil")
	}
}
