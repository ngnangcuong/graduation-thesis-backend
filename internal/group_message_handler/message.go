package group_message_handler

const (
	MESSAGE_TYPE = "message"
	EVENT_TYPE   = "event"
)

type ConversationMessage struct {
	ConversationID        string `json:"conv_id"`
	ConversationMessageID int64  `json:"conv_msg_id"`
	MessageTime           int64  `json:"msg_time"`
	Sender                string `json:"sender"`
	Content               []byte `json:"content"`
}

type Event struct {
	Actor          string `json:"actor"`
	ConversationID string `json:"conversation_id"`
	Action         string `json:"action"`
	Object         string `json:"object"`
	ObjectID       string `json:"objectID"`
}

type KafkaMessage struct {
	UserID         string      `json:"user_id"`
	ConversationID string      `json:"conversation_id"`
	Type           string      `json:"type"`
	Timestamp      int64       `json:"timestamp"`
	Data           interface{} `json:"data"`
}

type Message struct {
	ConversationID        string `json:"conv_id" `
	ConversationMessageID int64  `json:"conv_msg_id"`
	MessageTime           int64  `json:"msg_time"`
	Sender                string `json:"sender"`
	Content               string `json:"content"`
	Receiver              string `json:"receiver"`
}

type Conversation struct {
	ID      string   `json:"id"`
	Members []string `json:"members"`
}

type WebsocketHandler struct {
	ID           string `json:"id"`
	IPAddress    string `json:"ip_address"`
	NumberClient int    `json:"number_client"`
}
