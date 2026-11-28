
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	MaxCommandsPerSecond = 20    
	MaxConnectionsPerIP  = 100
	QueueSize            = 65536
)

var (
	out   = slog.New(slog.NewJSONHandler(io.Discard, nil))
	queue chan []byte                                   
	done  chan struct{}                                  
)


type queueWriter struct{}

func (queueWriter) Write(p []byte) (int, error) {

	record := make([]byte, len(p))
	copy(record, p)

	select {
	case queue <- record:
	default:
	}

	
	return len(p), nil
}


func Init(filePath string) error {
	dst := io.Writer(os.Stdout)
	if filePath != "" {
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return err
		}
		dst = io.MultiWriter(os.Stdout, f) 
	}
	queue = make(chan []byte, QueueSize)
	done = make(chan struct{})
	go func() { 
		for b := range queue {
			_, _ = dst.Write(b)
		}
		close(done)
	}()
	out = slog.New(slog.NewJSONHandler(queueWriter{}, nil))
	return nil
}


func Close() {
	if queue != nil {
		close(queue)
		<-done
		queue = nil
	}
}

//queste perche senno chi e esterno non puo usare direttament 
func Info(msg string, args ...any)  { out.Info(msg, args...) }
func Warn(msg string, args ...any)  { out.Warn(msg, args...) }
func Error(msg string, args ...any) { out.Error(msg, args...) }

func Connect(connID int, addr string) {
	out.Info("client_connected", "conn_id", connID, "addr", addr)
}

func Disconnect(connID int, addr, player string, d time.Duration) {
	out.Info("client_disconnected", "conn_id", connID, "addr", addr,
		"player", player, "duration_ms", d.Milliseconds())
}

func Command(connID int, player, line string) {
	name, params, _ := strings.Cut(strings.TrimSpace(line), " ")
	if len(params) > 300 {
		params = params[:300]
	}
	out.Info("command", "conn_id", connID, "player", player,
		"cmd", strings.ToUpper(name), "params", params)
}


func Response(player, line string) {
	if len(line) > 300 {
		line = line[:300]
	}
	if f := strings.Fields(line); len(f) > 1 && f[0] == "ERR" {
		out.Warn("response_error", "player", player, "code", f[1], "line", line)
	} else if strings.HasPrefix(line, "OK") {
		out.Info("response", "player", player, "line", line)
	}
}

func World(event string, args ...any) { out.Info("world_"+event, args...) }

func Quest(player, questID, event string, args ...any) {
	out.Info("quest_"+event, append([]any{"player", player, "quest", questID}, args...)...)
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
		out.Warn("command_flood", "conn_id", connID, "player", player,
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
		out.Warn("rapid_connections", "ip", ip, "limit", MaxConnectionsPerIP)
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