package model

import (
	"sync"
	"time"
)

type SMap struct {
	mu map[string]*sync.RWMutex

	data        map[string]string
	timeExpired map[string]time.Time
	timeout     time.Duration
}

func NewSMap(timeout time.Duration) *SMap {
	return &SMap{
		mu:          make(map[string]*sync.RWMutex),
		data:        make(map[string]string),
		timeExpired: make(map[string]time.Time),
		timeout:     timeout,
	}
}

func (s *SMap) Get(key string) (string, bool) {
	s.mu[key].RLock()
	defer s.mu[key].RUnlock()

	conn, ok := s.data[key]
	if !ok {
		return conn, ok
	}

	if time.Now().After(s.timeExpired[key]) {
		return conn, false
	}

	return conn, ok
}

func (s *SMap) Set(key string, conn string) {
	s.mu[key].Lock()
	defer s.mu[key].Unlock()

	s.data[key] = conn
	s.timeExpired[key] = time.Now().Add(time.Minute)
}

func (s *SMap) Del(key string) {
	s.mu[key].Lock()
	defer s.mu[key].Unlock()

	delete(s.data, key)
}

type Peer struct {
	mu sync.Mutex

	id      string
	expired time.Time
}

func NewPeer(id string, timeout time.Duration) *Peer {
	return &Peer{
		id:      id,
		expired: time.Now().Add(timeout),
	}
}

func (p *Peer) Get() (string, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if time.Now().After(p.expired) {
		return "", false
	}

	return p.id, true
}
