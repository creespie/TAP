package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"tap/internal/transport"
	"unicode/utf8"
)

var ErrInvalidUTF8 = errors.New("invalid UTF-8")

func cleanLine(raw string) (string, error) {
	line := strings.TrimSuffix(raw, "\n")
	line = strings.TrimSuffix(line, "\r")
	if !utf8.ValidString(line) {
		return "", ErrInvalidUTF8
	}
	return line, nil
}

func handleConnect(conn net.Conn, counter int, registry *transport.ConnRegistry) {
	fmt.Printf("connected client %d: %s\n", counter, conn.RemoteAddr())
	defer registry.Remove(counter)
	defer fmt.Printf("disconnected client number %d\n", counter)
	defer conn.Close()
	fmt.Fprintln(conn, "OK hello proto=1")
	reader := bufio.NewReaderSize(conn, 1024)
	discarding := false
	for {
		raw, err := reader.ReadSlice('\n')
		if err != nil {
			if errors.Is(err, bufio.ErrBufferFull) && !discarding {
				discarding = true
				fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
				continue
			} else if errors.Is(err, bufio.ErrBufferFull) {
				continue
			} else {
				return
			}
		}
		if discarding {
			discarding = false
			continue
		}
		line, err := cleanLine(string(raw))
		if err != nil {
			fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
			continue
		}
		if line == "" {
			continue
		}
		fmt.Fprintf(conn, "ECHO %s\n", line)
	}
}

func main() {
	port := flag.Int("port", 4242, "TCP port to listen on")
	world := flag.String("world", "world.json", "World file used to build the server's world")
	seed := flag.Int64("seed", 0, "Seed to determine luck and random rolls inside the game")
	host := flag.String("host", "", "Host address")

	flag.Parse()
	fmt.Printf("%d %s %d\n", *port, *world, *seed)
	ln, err := net.Listen("tcp", net.JoinHostPort(*host, strconv.Itoa(*port)))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	registry := transport.NewConnRegistry()
	defer ln.Close()
	fmt.Printf("Listening on %s\n", ln.Addr())
	counter := 1
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		ln.Close()
	}()
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				fmt.Fprintln(os.Stderr, "Server shutting down")
				registry.CloseAll()
				return
			}
			fmt.Fprintln(os.Stderr, "error:", err)
			continue
		}
		registry.Add(conn, counter)
		go handleConnect(conn, counter, registry)
		counter += 1
	}

}
