package model

type Message struct {
	ConversationID        string `json:"conv_id"`
	ConversationMessageID int64  `json:"conv_msg_id"`
	MessageTime           int64  `json:"msg_time"`
	Sender                string `json:"sender"`
	Content               string `json:"content"`
	Receiver              string `json:"receiver"`
}
type SendMessageRequest struct {
	ConversationID string `json:"conv_id" binding:"required"`
	Sender         string `json:"sender" binding:"required"`
	Content        string `json:"content" binding:"required,max=10000"`
	MessageTime    int64  `json:"msg_time"`
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
	Content               string `json:"content"`
}

type UserInboxResponse struct {
	UserID  string  `json:"user_id"`
	Inboxes []Inbox `json:"inboxes"`
}

type KafkaMessage struct {
	WebsocketHandlerID string `json:"websocket_id"`
	UserID             string `json:"user_id"`
	Action             string `json:"action"`
}

type ReadReceiptRequest struct {
	ConversationID string `json:"conv_id"`
	UserID         string `json:"user_id"`
}

type ReadReceipt struct {
	ConversationID string `json:"conv_id" cql:"conv_id"`
	UserID         string `json:"user_id" cql:"user_id"`
	MessageID      int64  `json:"msg_id" cql:"msg_id"`
}

type UserInbox struct {
	UserID                string `json:"user_id" cql:"user_id"`
	InboxMessageID        int64  `json:"inbox_msg_id" cql:"inbox_msg_id"`
	ConversationID        string `json:"conv_id" cql:"conv_id"`
	ConversationMessageID int64  `json:"conv_msg_id" cql:"conv_msg_id"`
	MessageTime           int64  `json:"msg_time" cql:"msg_time"`
	Sender                string `json:"sender" cql:"sender"`
	Content               string `json:"content" cql:"content"`
}
