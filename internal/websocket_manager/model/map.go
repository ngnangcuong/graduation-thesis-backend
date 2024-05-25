package model

import "sync"

type MapWebsocketHandlerMonitoring struct {
	mu   sync.RWMutex
	data map[string]*WebsocketHandlerMonitoring
}

func NewMapWebsocketHandlerMonitoring() *MapWebsocketHandlerMonitoring {
	return &MapWebsocketHandlerMonitoring{
		data: make(map[string]*WebsocketHandlerMonitoring),
	}
}

func (m *MapWebsocketHandlerMonitoring) Get(key string) *WebsocketHandlerMonitoring {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[key]
}

func (m *MapWebsocketHandlerMonitoring) Set(key string, value WebsocketHandlerMonitoring) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = &value
}

func (m *MapWebsocketHandlerMonitoring) Del(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}
