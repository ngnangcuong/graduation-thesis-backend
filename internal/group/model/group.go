package model

import "time"

type Group struct {
	ID             string    `json:"id"`
	GroupName      string    `json:"group_name"`
	CreatedAt      time.Time `json:"created_at"`
	LastUpdated    time.Time `json:"last_updated"`
	Members        []string  `json:"members"`
	Admins         []string  `json:"admins"`
	ConversationID string    `json:"conv_id"`
	Deleted        bool      `json:"deleted"`
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

type GetConversationsContainUserResponse struct {
	ConversationID string `json:"conv_id"`
	MemberCount    int    `json:"member_count"`
}
