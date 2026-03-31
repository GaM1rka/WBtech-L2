package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	if flag.NArg() < 2 {
		log.Fatalf("usage: go run . [--timeout=10s] <host> <port>")
	}

	host := flag.Arg(0)
	port := flag.Arg(1)
	socket := net.JoinHostPort(host, port)

	conn, err := net.DialTimeout("tcp", socket, *timeout)
	if err != nil {
		log.Fatalf("failed to connect to %s: %v", socket, err)
	}
	defer conn.Close()

	log.Printf("connected to %s", socket)
	done := make(chan struct{})
	var once sync.Once

	closeDone := func() {
		once.Do(func() {
			close(done)
		})
	}

	go readFromConn(conn, done, closeDone)
	go writeToConn(conn, done, closeDone)

	<-done
	log.Println("connection closed")
}

func readFromConn(conn net.Conn, done <-chan struct{}, closeDone func()) {
	_, err := io.Copy(os.Stdout, conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
	}
	closeDone()
}

func writeToConn(conn net.Conn, done <-chan struct{}, closeDone func()) {
	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-done:
			return
		default:
		}

		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			_, writeErr := conn.Write(line)
			if writeErr != nil {
				fmt.Fprintf(os.Stderr, "write error: %v\n", writeErr)
				closeDone()
				return
			}
		}

		if err != nil {
			if err == io.EOF {
				_ = conn.Close()
				closeDone()
				return
			}

			fmt.Fprintf(os.Stderr, "stdin read error: %v\n", err)
			closeDone()
			return
		}
	}
}
