package group_message_handler

import (
	"graduation-thesis/pkg/custom_error"
	"sync"
)

type ChanMessage struct {
	mu       sync.RWMutex
	isClosed bool
	Channel  chan Message
}

func NewChanMessage(size int) *ChanMessage {
	return &ChanMessage{
		Channel: make(chan Message, size),
	}
}

func (c *ChanMessage) Send(message Message) error {
	c.mu.RLock()
	defer c.mu.Unlock()
	if !c.isClosed {
		c.Channel <- message
		return nil
	}
	return custom_error.ErrChannelHasClosed
}

func (c *ChanMessage) Close() {
	c.mu.Lock()
	c.mu.Unlock()
	if !c.isClosed {
		c.isClosed = true
		close(c.Channel)
	}
}
