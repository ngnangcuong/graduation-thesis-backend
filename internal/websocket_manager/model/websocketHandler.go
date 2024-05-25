package model

type WebsocketHandlerClient struct {
	ID           string `json:"id" redis:"id"`
	IPAddress    string `json:"ip_address" redis:"ip_address"`
	NumberClient int    `json:"number_client" redis:"-"`
}

type UserID string

type AddNewWebsocketHandlerRequest struct {
	ID        string `json:"id"`
	IPAddress string `json:"ip_address"`
}

type AddNewUserRequest struct {
	WebsocketID string `json:"websocket_id"`
	UserID      string `json:"user_id"`
}

type WebsocketHandlerMonitoring struct {
	ID        string
	IPAddress string
	Hearbeat  chan struct{}
}

type PingRequest struct {
	ID        string `json:"id"`
	IPAddress string `json:"ip_address"`
}
