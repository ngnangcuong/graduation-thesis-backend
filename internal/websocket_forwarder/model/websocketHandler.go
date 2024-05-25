package model

type WebsocketHandler struct {
	ID           string `json:"id"`
	IPAddress    string `json:"ip_address"`
	NumberClient int    `json:"number_client"`
}

type HandleRequestResponse struct {
	IPAddress string `json:"ip_address"`
}
