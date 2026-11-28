package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"tap/internal/handlers"
	"tap/internal/logger"
	"tap/internal/transport"
	"tap/internal/world"
	"time"
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

func handleConnection(conn net.Conn, counter int, registry *transport.ConnRegistry, w *world.World) {
	fmt.Printf("connected client %d: %s\n", counter, conn.RemoteAddr())
	

	defer registry.Remove(counter)
	defer fmt.Printf("disconnected client number %d\n", counter)
	defer conn.Close()
	//remove player from world on ctrl c
	defer func() {
		w.Mu.Lock()
		defer w.Mu.Unlock()
		if p := transport.FindPlayer(w, conn); p != nil {
			handlers.RemovePlayer(w, p)
			close(p.Outbox)
		}
	}()
	// Initial greeting
	fmt.Fprintln(conn, "OK hello proto=1")

	reader := bufio.NewReaderSize(conn, 1024)
	discarding := false

	for {
		// Read incoming message
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

		// Discard oversized message
		if discarding {
			discarding = false
			continue
		}

		
		if !logger.AllowCommand(counter, "") {
			continue
		}

		// Clean and validate message
		line, err := cleanLine(string(raw))

		if err != nil {
			fmt.Fprintln(conn, "ERR 400 BAD_REQUEST")
			continue
		}

	
		if !logger.AllowCommand(counter, "") {
			continue
		}
		// Dispatch request
		transport.DispatchRequest(conn, line, w)
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	port := flag.Int("port", 4242, "TCP port to listen on")
	worldPath := flag.String("world", "world.json", "World file used to build the server's world")
	seed := flag.Int64("seed", 0, "Seed to determine luck and random rolls inside the game")
	host := flag.String("host", "", "Host address")
	logFile := flag.String("log", "server.log", "Log file path")
	flag.Parse()

	if err := logger.Init(*logFile); err != nil {
		return err
	}
	defer logger.Close()

	w, err := world.Load(*worldPath)
	if err != nil {
		logger.Error("world_load_failed", "path", *worldPath, "err", err.Error())
		return err
	}
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}
	w.Rand = rand.New(rand.NewSource(*seed))
	logger.Info("world_loaded", "rooms", len(w.Rooms), "items", len(w.Items),
		"npcs", len(w.Npcs), "quests", len(w.Quests), "seed", *seed)

	ln, err := net.Listen("tcp", net.JoinHostPort(*host, strconv.Itoa(*port)))
	if err != nil {
		logger.Error("listen_failed", "err", err.Error())
		return err
	}
	registry := transport.NewConnRegistry()
	defer ln.Close()
	logger.Info("server_listening", "addr", ln.Addr().String())

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
				logger.Info("server_shutting_down")
				registry.CloseAll()
				return nil
			}
			logger.Error("accept_failed", "err", err.Error())
			continue
		}
		registry.Add(conn, counter)
		go handleConnection(conn, counter, registry, w)
		counter += 1
	}
}