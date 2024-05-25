package model

type KafkaMessage struct {
	WebsocketHandlerID string `json:"websocket_id"`
	UserID             string `json:"user_id"`
	Action             string `json:"action"`
}
