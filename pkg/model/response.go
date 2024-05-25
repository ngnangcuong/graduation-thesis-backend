package model

type SuccessResponse struct {
	Result       interface{} `json:"result"`
	Status       int         `json:"status"`
	InfoMessages []string    `json:"info_messages"`
}

type ErrorResponse struct {
	Status       int      `json:"status"`
	ErrorMessage string   `json:"error_message"`
	InfoMessages []string `json:"info_messages"`
}
