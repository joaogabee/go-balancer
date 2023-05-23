package ServerPool

import (
	"GOBalancer/src/pkg/Backend"
	"log"
	"net"
	"net/url"
	"sync/atomic"
	"time"
)

type ServerPool struct {
	backends []*Backend.Backend
	current  uint64
}

func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}
func (s *ServerPool) GetNextPeer() *Backend.Backend {
	next := s.NextIndex()
	l := len(s.backends) + next

	for i := next; i < l; i++ {
		idx := i % len(s.backends)

		if s.backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.backends[idx]
		}
	}
	return nil
}
func (s *ServerPool) IsBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	_ = conn.Close() // close it, we dont need to maintain this connection
	return true
}
func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		status := "up"
		alive := s.IsBackendAlive(b.URL)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}
func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if b.URL.String() == backendUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}
func (s *ServerPool) AddBackend(backend *Backend.Backend) {
	s.backends = append(s.backends, backend)
}
