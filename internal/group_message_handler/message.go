package group_message_handler

type KafkaMessage struct {
	ConversationID        string `json:"conv_id"`
	ConversationMessageID int64  `json:"conv_msg_id"`
	MessageTime           int64  `json:"msg_time"`
	Sender                string `json:"sender"`
	Content               []byte `json:"content"`
}

type Message struct {
	ConversationID        string `json:"conv_id" `
	ConversationMessageID int64  `json:"conv_msg_id"`
	MessageTime           int64  `json:"msg_time"`
	Sender                string `json:"sender"`
	Content               []byte `json:"content"`
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
