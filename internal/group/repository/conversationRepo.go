package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"graduation-thesis/internal/group/model"
	"graduation-thesis/pkg/custom_error"
	"graduation-thesis/pkg/interfaces"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/twinj/uuid"
)

type ConversationRepo struct {
	db    interfaces.DBTX
	redis *redis.Client
}

func NewConversationRepo(db interfaces.DBTX, redis *redis.Client) *ConversationRepo {
	return &ConversationRepo{
		db:    db,
		redis: redis,
	}
}

func (c *ConversationRepo) WithTx(tx *sql.Tx) *ConversationRepo {
	return &ConversationRepo{
		db:    tx,
		redis: c.redis,
	}
}

func (c *ConversationRepo) getFromRedis(ctx context.Context, key string, dest interface{}) error {
	value, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(value), dest)
}

func (c *ConversationRepo) setIntoRedis(ctx context.Context, key string, value interface{}) error {
	v, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, key, v, 5*time.Minute).Err()
}

func (c *ConversationRepo) deleteFromRedis(ctx context.Context, key string) error {
	_, err := c.redis.Del(ctx, key).Result()
	return err
}

func (c *ConversationRepo) GetMembers(ctx context.Context, conversationID string) (*model.Conversation, error) {
	var conversation model.Conversation
	if err := c.getFromRedis(ctx, conversationID, conversation); err == nil {
		return &conversation, nil
	}

	query := `SELECT user_id FROM conversation c INNER JOIN conv_map_user cmu ON c.conv_id = cmu.conv_id
				WHERE cmu.conv_id = $1`
	rows, err := c.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, custom_error.HandlePostgreError(err)
	}

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, custom_error.HandlePostgreError(err)
		}

		conversation.Members = append(conversation.Members, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, custom_error.HandlePostgreError(err)
	}

	_ = c.setIntoRedis(ctx, conversationID, conversation)
	return &conversation, nil
}

func (c *ConversationRepo) Create(ctx context.Context, conversationID string) error {
	query := `INSERT INTO conversations (id,) VALUES ($1)`
	_, err := c.db.ExecContext(ctx, query, conversationID)
	if err != nil {
		return custom_error.HandlePostgreError(err)
	}

	return nil
}

func (c *ConversationRepo) AddMembers(ctx context.Context, conversationID string, members []string) error {
	if len(members) == 0 {
		return nil
	}

	query := `INSERT INTO conv_map_user (id, conv_id, user_id) VALUES ($1, $2, $3)`
	for _, member := range members {
		_, err := c.db.ExecContext(ctx, query, uuid.NewV4().String(), conversationID, member)
		if err != nil {
			return custom_error.HandlePostgreError(err)
		}
	}

	_ = c.deleteFromRedis(ctx, conversationID)
	return nil
}

func (c *ConversationRepo) RemoveMembers(ctx context.Context, conversationID string, members []string) error {
	if len(members) == 0 {
		return nil
	}

	query := `DELETE FROM conv_map_user WHERE conv_id = $1 AND user_id = $2`
	for _, member := range members {
		_, err := c.db.ExecContext(ctx, query, conversationID, member)
		if err != nil {
			return custom_error.HandlePostgreError(err)
		}
	}

	_ = c.deleteFromRedis(ctx, conversationID)
	return nil
}

func (c *ConversationRepo) Update(ctx context.Context, params model.UpdateConversationParams) error {
	query := `UPDATE conversations SET members = $2 WHERE id = $1`
	_, err := c.db.ExecContext(ctx, query, params.ID, params.Members)
	return custom_error.HandlePostgreError(err)
}

func (c *ConversationRepo) GetConversations(ctx context.Context, userID string) ([]string, error) {
	query := `SELECT conv_id FROM conv_map_user WHERE user_id = $1`
	rows, err := c.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, custom_error.HandlePostgreError(err)
	}

	var conversations []string
	for rows.Next() {
		var conversation string
		if err := rows.Scan(&conversation); err != nil {
			return nil, custom_error.HandlePostgreError(err)
		}

		conversations = append(conversations, conversation)
	}

	if err := rows.Err(); err != nil {
		return nil, custom_error.HandlePostgreError(err)
	}

	return conversations, nil
}
