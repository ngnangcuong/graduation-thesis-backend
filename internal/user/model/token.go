package model

type AccessDetails struct {
	AccessUuid string `json:"access_uuid,omitempty"`
	UserID     string `json:"user_id,omitempty"`
}

type TokenDetails struct {
	AccessToken  string `json:"access_token,omitempty"`
	AccessUuid   string `json:"access_uuid,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	RefreshUuid  string `json:"refresh_uuid,omitempty"`
	AtExpires    int64  `json:"at_expires,omitempty"`
	RtExpires    int64  `json:"rt_expires,omitempty"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
