package transport

import (
	"net"
	"sync"

	"tap/internal/world"
)

type ConnRegistry struct {
	mu    sync.Mutex
	conns map[int]net.Conn
	wg    sync.WaitGroup
}

func NewConnRegistry() *ConnRegistry {
	return &ConnRegistry{conns: make(map[int]net.Conn)}
}

func (r *ConnRegistry) Add(conn net.Conn, id int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conns[id] = conn
	r.wg.Add(1)
}

func (r *ConnRegistry) Remove(id int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.conns, id)
	r.wg.Done()
}

func (r *ConnRegistry) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.conns)
}

func (r *ConnRegistry) CloseAll() {
	r.mu.Lock()

	for _, c := range r.conns {
		c.Close()
	}
	r.mu.Unlock()
	defer r.wg.Wait()
}

// FindPlayer returns the player using conn, or nil if conn has not
// sent a successful CONNECT yet. The caller must hold w.Mu.
func FindPlayer(w *world.World, conn net.Conn) *world.Player {
	for _, player := range w.Players {
		if player.Connection == conn {
			return player
		}
	}
	return nil
}
