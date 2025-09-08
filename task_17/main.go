package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	timeoutFlag := flag.Int("timeout", 10, "connection timeout in seconds")
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--timeout seconds] host port\n", os.Args[0])
		os.Exit(1)
	}

	host := flag.Arg(0)
	port := flag.Arg(1)
	addr := net.JoinHostPort(host, port)

	timeout := time.Duration(*timeoutFlag) * time.Second

	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to %s: %v\n", addr, err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Connected to %s\n", addr)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(2)

	done := make(chan struct{})

	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "\nConnection error: %v\n", err)
				}
				close(done)
				return
			}
			if n > 0 {
				_, _ = os.Stdout.Write(buf[:n])
			}
		}
	}()

	go func() {
		defer wg.Done()
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					close(done)
				} else {
					fmt.Fprintf(os.Stderr, "\nInput error: %v\n", err)
					close(done)
				}
				return
			}
			_, err = conn.Write(line)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nWrite error: %v\n", err)
				close(done)
				return
			}
		}
	}()

	select {
	case <-done:
	case <-sigCh:
		fmt.Fprintln(os.Stderr, "\nInterrupted")
	}

	conn.Close()

	wg.Wait()

	fmt.Println("\nConnection closed, exiting.")
}
