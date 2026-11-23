package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"bufio"
)

func handleConnect(conn net.Conn, counter int) {
	fmt.Println("connected:", conn.RemoteAddr())
	defer fmt.Printf("disconnected client number %d\n", counter)
	defer conn.Close()
	fmt.Fprintln(conn, "OK hello client number", counter)
	reader := bufio.NewReader(conn)
	for{
		line, err := reader.ReadString('\n')
		if err != nil{
			return 
		}
		fmt.Fprintf(conn, "ECHO %s", line)
	}
}

func main() {
	port := flag.Int("port", 4242, "TCP pport to listen on")
	world := flag.String("world", "world.json", "World file used to build the server's world")
	seed := flag.Int64("seed", 1, "Seed to determine luck and random rolls inside the game")
	host := flag.String("host", "", "Host address")

	flag.Parse()
	fmt.Printf("%d %s %d\n", *port, *world, *seed)
	ln, err := net.Listen("tcp", net.JoinHostPort(*host, strconv.Itoa(*port)))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Printf("Listening on %s\n", ln.Addr())
	counter := 1
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				fmt.Fprintln(os.Stderr, "Server shutting down")
				return
			}
			fmt.Fprintln(os.Stderr, "error:", err)
			continue
		}
		go handleConnect(conn, counter)
		counter += 1
	}

}
