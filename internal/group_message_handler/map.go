package group_message_handler

import "sync"

type MapConnection struct {
	mu   sync.RWMutex
	data map[string]*ChanMessage
}

func (m *MapConnection) Get(key string) *ChanMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[key]
}

func (m *MapConnection) Set(key string, size int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = NewChanMessage(size)
}

func (m *MapConnection) Del(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

type MapMu struct {
	mu   sync.RWMutex
	data map[string]*sync.Mutex
}

func (m *MapMu) Get(key string) *sync.Mutex {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[key]
}

func (m *MapMu) Set(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = &sync.Mutex{}
}

func (m *MapMu) Del(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}
