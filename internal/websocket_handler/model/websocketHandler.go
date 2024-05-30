package model

import "sync"

type WebsocketHandlerClient struct {
	ID           string `json:"id"`
	IPAddress    string `json:"ip_address"`
	NumberClient int    `json:"number_client,omitempty"`
}

type AddNewWebsocketHandlerRequest struct {
	ID        string `json:"id"`
	IPAddress string `json:"ip_address"`
}

type PingRequest struct {
	ID        string `json:"id"`
	IPAddress string `json:"ip_address"`
}

type Connection struct {
	Mu                 sync.RWMutex
	WebsocketHandlerID string
	WriteChannel       chan Message
	IsDeleted          bool // For coodinating concurrent reads and writes
}

func (c *Connection) CheckDeleted() bool {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	return c.IsDeleted
}

func (c *Connection) Delete() {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if !c.IsDeleted {
		c.IsDeleted = true
		close(c.WriteChannel)
	}
}

func (c *Connection) Write(message Message) bool {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	if !c.IsDeleted {
		c.WriteChannel <- message
		return true
	}
	return false
}
