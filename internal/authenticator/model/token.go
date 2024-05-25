package model

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenDetails struct {
	AccessToken  string `json:"access_token,omitempty"`
	AccessUuid   string `json:"access_uuid,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	RefreshUuid  string `json:"refresh_uuid,omitempty"`
	AtExpires    int64  `json:"at_expires,omitempty"`
	RtExpires    int64  `json:"rt_expires,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
