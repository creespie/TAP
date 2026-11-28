package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

const (
	maxInputLength = 4096
	dialTimeout    = 10 * time.Second
	quitWait       = 1 * time.Second
)

func readInput(scanner *bufio.Scanner) (string, error) {
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("input error: %w", err)
		}
		return "", fmt.Errorf("input closed")
	}

	input := strings.TrimSpace(scanner.Text())

	if !utf8.ValidString(input) {
		return "", fmt.Errorf("invalid UTF-8 input")
	}

	if len(input) > maxInputLength {
		return "", fmt.Errorf("input too long")
	}

	return input, nil
}

func getServerIP(scanner *bufio.Scanner) (string, error) {
	fmt.Print("Server IP: ")

	ip, err := readInput(scanner)
	if err != nil {
		return "", err
	}

	if strings.EqualFold(ip, "localhost") {
		return "localhost", nil
	}

	if ip == "" || net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid server IP")
	}

	return ip, nil
}

func getServerPort(scanner *bufio.Scanner) (int, error) {
	fmt.Print("Server Port: ")

	input, err := readInput(scanner)
	if err != nil {
		return 0, err
	}

	port, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("port must be a number")
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port must be between 1 and 65535")
	}

	return port, nil
}

func connectToServer(ip string, port int) (net.Conn, error) {
	address := net.JoinHostPort(ip, strconv.Itoa(port))

	conn, err := net.DialTimeout("tcp", address, dialTimeout)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	return conn, nil
}

// authenticate asks for a username until the server accepts it.
func authenticate(scanner *bufio.Scanner, conn net.Conn, serverReader *bufio.Reader) error {
	for {
		fmt.Println()
		fmt.Print("Username: ")

		username, err := readInput(scanner)
		if err != nil {
			return err
		}

		if username == "" || strings.ContainsAny(username, " \t\r\n") {
			fmt.Println("Invalid username")
			continue
		}

		if _, err := fmt.Fprintf(conn, "CONNECT %s\n", username); err != nil {
			return fmt.Errorf("send error: %w", err)
		}

		response, err := serverReader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}

		response = strings.TrimSpace(response)
		fmt.Println("<", response)

		if strings.HasPrefix(response, "OK") {
			return nil
		}
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024), maxInputLength)

	fmt.Println("Welcome to TAP!")
	fmt.Println()

	serverIP, err := getServerIP(scanner)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	serverPort, err := getServerPort(scanner)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	conn, err := connectToServer(serverIP, serverPort)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer conn.Close()

	fmt.Println()
	fmt.Println("Connected!")

	serverReader := bufio.NewReader(conn)

	// Read server greeting
	greeting, err := serverReader.ReadString('\n')
	if err != nil {
		fmt.Println("Server error:", err)
		return
	}

	fmt.Println("<", strings.TrimSpace(greeting))

	if err := authenticate(scanner, conn, serverReader); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Go 1.18 compatible atomic flag
	var quitting int32

	done := make(chan struct{})

	go serverListener(serverReader, &quitting, done)

	fmt.Println()

	// Main command loop
	for {
		fmt.Print("> ")

		message, err := readInput(scanner)
		if err != nil {
			if err.Error() == "input closed" {
				fmt.Println("\nInput closed")
			} else {
				fmt.Println("Input error:", err)
			}
			return
		}

		if message == "" {
			continue
		}

		if strings.ContainsAny(message, "\r\n") {
			fmt.Println("Invalid command")
			continue
		}

		if strings.EqualFold(message, "QUIT") {
			atomic.StoreInt32(&quitting, 1)
		}

		if _, err := fmt.Fprintln(conn, message); err != nil {
			fmt.Println("Send error:", err)
			return
		}

		if atomic.LoadInt32(&quitting) == 1 {
			// Wait for "OK bye" and server disconnection
			select {
			case <-done:
			case <-time.After(quitWait):
			}

			fmt.Println("Disconnected.")
			return
		}
	}
}

func serverListener(serverReader *bufio.Reader, quitting *int32, done chan<- struct{}) {
	defer close(done)

	for {
		response, err := serverReader.ReadString('\n')
		if err != nil {
			if atomic.LoadInt32(quitting) == 0 {
				fmt.Println("\nServer disconnected:", err)
				os.Exit(0)
			}
			return
		}

		line := strings.TrimSpace(response)

		if atomic.LoadInt32(quitting) == 1 {
			fmt.Println("\r< " + line)
		} else {
			fmt.Print("\r< " + line + "\n> ")
		}
	}
}
