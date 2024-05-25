package model

import "time"

type Message struct {
	ID          string    `cql:"id" json:"id"`
	From        string    `cql:"from" json:"from"`
	To          string    `cql:"to" json:"to"`
	Content     []byte    `cql:"content" json:"content"`
	DeliveredAt time.Time `cql:"delivered_at" json:"delivered_at"`
	LastUpdated time.Time `cql:"last_updated" json:"last_updated"`
	ReceivedAt  time.Time `cql:"received_at" json:"received_at,omitempty"`
	ReadAt      time.Time `cql:"read_at" json:"read_at,omitempty"`
	Deleted     bool      `cql:"deleted" json:"deleted"`
	Status      string    `cql:"status" json:"status" binding:"oneof=delivered received read"`
	Type        string    `cql:"type" json:"type" binding:"oneof=image video audio word"`
	LinkTo      string    `cql:"link_to" json:"link_to,omitempty" binding:"filepath"`
}

type SearchMessagesRequest struct {
	From       string `json:"from" binding:"required"`
	To         string `json:"to" binding:"required"`
	OnlyUnread bool   `json:"only_unread" binding:"boolean"`
	Limit      int    `json:"limit" binding:"max=500"`
}

type SearchConversionRequest struct {
	Users    []string  `json:"users" binding:"required,len=2"`
	TimeFrom time.Time `json:"time_from" binding:"ltecsfield=SearchConversionRequest.TimeTo"`
	TimeTo   time.Time `json:"time_to" binding:"gtecsfield=SearchConversionRequest.TimeFrom"`
	Limit    int       `json:"limit" binding:"max=500"`
}

type EvictMessageRequest struct {
	ID string `json:"id" binding:"required"`
}

type ConversationMessage struct {
	ConversationID        string `json:"conv_id" cql:"conv_id"`
	ConversationMessageID int64  `json:"conv_msg_id" cql:"conv_msg_id"`
	MessageTime           int64  `json:"msg_time" cql:"msg_time"`
	Sender                string `json:"sender" cql:"sender"`
	Content               []byte `json:"content" cql:"content"`
}

type UserInbox struct {
	UserID                string `json:"user_id" cql:"user_id"`
	InboxMessageID        int64  `json:"inbox_msg_id" cql:"inbox_msg_id"`
	ConversationID        string `json:"conv_id" cql:"conv_id"`
	ConversationMessageID int64  `json:"conv_msg_id" cql:"conv_msg_id"`
	MessageTime           int64  `json:"msg_time" cql:"msg_time"`
	Sender                string `json:"sender" cql:"sender"`
	Content               []byte `json:"content" cql:"content"`
}

type ReadReceipt struct {
	ConversationID string `json:"conv_id" cql:"conv_id"`
	UserID         string `json:"user_id" cql:"user_id"`
	MessageID      int64  `json:"msg_id" cql:"msg_id"`
}

type Inbox struct {
	Count          int                `json:"count"`
	ConversationID string             `json:"conv_id"`
	Messages       []*MessageResponse `json:"messages"`
}

type MessageResponse struct {
	ConversationMessageID int64  `json:"conv_msg_id"`
	MessageTime           int64  `json:"msg_time"`
	Sender                string `json:"sender"`
	Content               []byte `json:"content"`
}

type ReadReceiptRequest struct {
	ConversationID string `json:"conv_id"`
	UserID         string `json:"user_id"`
}

type UpdateReadReceiptRequest struct {
	ConversationID    string              `json:"conv_id"`
	ReadReceiptUpdate []ReadReceiptUpdate `json:"read_receipt_update"`
}

type ReadReceiptUpdate struct {
	UserID    string `json:"user_id" binding:"required"`
	MessageID int64  `json:"msg_id" binding:"required"`
}

type SendMessageRequest struct {
	ConversationID string `json:"conv_id" binding:"required"`
	Sender         string `json:"sender" binding:"required"`
	Content        []byte `json:"content" binding:"required,max=10000"`
	MessageTime    int64  `json:"msg_time"`
}

type UserInboxResponse struct {
	UserID  string  `json:"user_id"`
	Inboxes []Inbox `json:"inboxes"`
}
