package model

import "sync"

type MapConnection struct {
	mu   sync.RWMutex
	data map[string]*Connection
}

func NewMapConnection() *MapConnection {
	return &MapConnection{
		data: make(map[string]*Connection),
	}
}

func (m *MapConnection) Get(key string) *Connection {
	m.mu.RLock()
	defer m.mu.Unlock()
	return m.data[key]
}

func (m *MapConnection) Set(key string, value *Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *MapConnection) Del(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}
