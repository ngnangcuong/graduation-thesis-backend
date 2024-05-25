package model

import "sync"

type MapUserPeer struct {
	mu   sync.RWMutex
	data map[string]*Peer
}

func NewMapUserPeer() *MapUserPeer {
	return &MapUserPeer{
		data: make(map[string]*Peer),
	}
}

func (m *MapUserPeer) Get(key string) *Peer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[key]
}

func (m *MapUserPeer) Set(key string, value *Peer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}
