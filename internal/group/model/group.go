package model

import "time"

type Group struct {
	ID             string
	GroupName      string
	CreatedAt      time.Time
	LastUpdated    time.Time
	Members        []string `json:"members"`
	Admins         []string
	ConversationID string
	Deleted        bool
}

type Conversation struct {
	ID      string   `json:"id"`
	Members []string `json:"members"`
}

type CreateGroupRequest struct {
	GroupName string   `json:"group_name" binding:"required"`
	Members   []string `json:"members" binding:"required,gte=3"`
}

type CreateGroupResponse struct {
	GroupID        string `json:"group_id"`
	ConversationID string `json:"conv_id"`
}

type UpdateGroupRequest struct {
	GroupName string       `json:"group_name"`
	Members   []ChangeUser `json:"members" binding:"lte=2"`
	Admins    []ChangeUser `json:"admins" binding:"lte=2"`
}

type ChangeUser struct {
	Action string   `json:"action" binding:"oneof=add remove"`
	Users  []string `json:"users"`
}

type UpdateGroupParams struct {
	ID          string
	GroupName   string
	LastUpdated time.Time
	Deleted     bool
	Admins      []string
}

type UpdateConversationParams struct {
	ID      string
	Members []string
}

type CreateConversationRequest struct {
	Members []string `json:"members" binding:"required,gte=2"`
}
