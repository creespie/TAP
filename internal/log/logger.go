package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	MaxCommandsPerSecond = 5
	MaxConnectionsPerIP  = 10
	QueueSize            = 65536
)

var (
	queue chan []byte
	done  chan struct{}
)

// write costruisce una riga JSON (time, level, msg, poi le coppie chiave/valore)
// e la mette in coda. Se la coda è piena o Init non è stato chiamato, la scarta.
func write(level, msg string, args ...any) {
	var b bytes.Buffer
	b.WriteByte('{')
	field(&b, "time", time.Now().Format(time.RFC3339Nano))
	b.WriteByte(',')
	field(&b, "level", level)
	b.WriteByte(',')
	field(&b, "msg", msg)
	for i := 0; i < len(args); i += 2 {
		b.WriteByte(',')
		if i+1 == len(args) { // argomento dispari senza chiave
			field(&b, "!BADKEY", args[i])
			break
		}
		field(&b, fmt.Sprint(args[i]), args[i+1])
	}
	b.WriteString("}\n")

	select {
	case queue <- b.Bytes():
	default:
	}
}

func field(b *bytes.Buffer, key string, val any) {
	if e, ok := val.(error); ok {
		val = e.Error()
	}
	k, _ := json.Marshal(key)
	v, err := json.Marshal(val)
	if err != nil {
		v, _ = json.Marshal(fmt.Sprint(val))
	}
	b.Write(k)
	b.WriteByte(':')
	b.Write(v)
}

func Init(filePath string) error {
	dst := io.Writer(os.Stdout)
	if filePath != "" {
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return err
		}
		dst = io.MultiWriter(os.Stdout, f)
		dst = io.MultiWriter(os.Stdout, f)
	}
	queue = make(chan []byte, QueueSize)
	done = make(chan struct{})
	go func() {
	go func() {
		for b := range queue {
			_, _ = dst.Write(b)
		}
		close(done)
	}()
	return nil
}

func Close() {
	if queue != nil {
		close(queue)
		<-done
		queue = nil
	}
}

// queste perche senno chi e esterno non puo usare direttamente
func Info(msg string, args ...any)  { write("INFO", msg, args...) }
func Warn(msg string, args ...any)  { write("WARN", msg, args...) }
func Error(msg string, args ...any) { write("ERROR", msg, args...) }

func Connect(connID int, addr string) {
	Info("client_connected", "conn_id", connID, "addr", addr)
}

func Disconnect(connID int, addr, player string, d time.Duration) {
	Info("client_disconnected", "conn_id", connID, "addr", addr,
		"player", player, "duration_ms", d.Milliseconds())
}

func Command(connID int, player, line string) {
	name, params, _ := strings.Cut(strings.TrimSpace(line), " ")
	if len(params) > 300 {
		params = params[:300]
	}
	Info("command", "conn_id", connID, "player", player,
		"cmd", strings.ToUpper(name), "params", params)
}

func Response(player, line string) {
	if len(line) > 300 {
		line = line[:300]
	}
	if f := strings.Fields(line); len(f) > 1 && f[0] == "ERR" {
		Warn("response_error", "player", player, "code", f[1], "line", line)
	} else if strings.HasPrefix(line, "OK") {
		Info("response", "player", player, "line", line)
	}
}

func World(event string, args ...any) { Info("world_"+event, args...) }

func Quest(player, questID, event string, args ...any) {
	Info("quest_"+event, append([]any{"player", player, "quest", questID}, args...)...)
}

type counter struct {
	start  time.Time
	n      int
	warned bool
}

var (
	mu    sync.Mutex
	cmds  = map[int]*counter{} // conn id -> commands in the current second
	conns = map[string]int{}   // ip -> open connections
)

func AllowCommand(connID int, player string) bool {
	mu.Lock()
	now := time.Now()
	c := cmds[connID]
	if c == nil || now.Sub(c.start) >= time.Second {
		c = &counter{start: now}
		cmds[connID] = c
	}
	c.n++
	ok := c.n <= MaxCommandsPerSecond
	first := !ok && !c.warned
	if first {
		c.warned = true
	}
	mu.Unlock()
	if first {
		Warn("command_flood", "conn_id", connID, "player", player,
			"limit_per_sec", MaxCommandsPerSecond)
	}
	return ok
}

func ForgetConnection(connID int) {
	mu.Lock()
	delete(cmds, connID)
	mu.Unlock()
}

func ipOf(addr string) string {
	if i := strings.LastIndex(addr, ":"); i >= 0 {
		return addr[:i]
	}
	return addr
}

func AllowConnection(addr string) bool {
	ip := ipOf(addr)
	mu.Lock()
	defer mu.Unlock()
	if conns[ip] >= MaxConnectionsPerIP {
		Warn("rapid_connections", "ip", ip, "limit", MaxConnectionsPerIP)
		return false
	}
	conns[ip]++
	return true
}

func ReleaseConnection(addr string) {
	ip := ipOf(addr)
	mu.Lock()
	defer mu.Unlock()
	if conns[ip] <= 1 {
		delete(conns, ip)
	} else {
		conns[ip]--
	}
}
