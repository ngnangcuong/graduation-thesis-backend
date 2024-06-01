package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"graduation-thesis/internal/group/model"
	"graduation-thesis/pkg/custom_error"
	"graduation-thesis/pkg/interfaces"
	"time"

	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type GroupRepo struct {
	db    interfaces.DBTX
	redis *redis.Client
}

func NewGroupRepo(db interfaces.DBTX, redis *redis.Client) *GroupRepo {
	return &GroupRepo{
		db:    db,
		redis: redis,
	}
}

func (g *GroupRepo) WithTx(tx *sql.Tx) *GroupRepo {
	return &GroupRepo{
		db:    tx,
		redis: g.redis,
	}
}

func (g *GroupRepo) getFromRedis(ctx context.Context, key string, dest interface{}) error {
	value, err := g.redis.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(value), dest)
}

func (g *GroupRepo) setIntoRedis(ctx context.Context, key string, value interface{}) error {
	v, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return g.redis.Set(ctx, key, v, 5*time.Minute).Err()
}

func (g *GroupRepo) deleteFromRedis(ctx context.Context, key string) error {
	_, err := g.redis.Del(ctx, key).Result()
	return err
}

func (g *GroupRepo) Get(ctx context.Context, groupID string) (*model.Group, error) {
	var group model.Group
	if err := g.getFromRedis(ctx, groupID, group); err == nil {
		return &group, nil
	}

	query := `SELECT id, group_name, created_at, last_updated, conv_id, admins FROM groups WHERE id = $1 AND deleted = false`
	row := g.db.QueryRowContext(ctx, query, groupID)

	if err := row.Scan(&group.ID, &group.GroupName, &group.CreatedAt, &group.LastUpdated, &group.ConversationID, (*pq.StringArray)(&group.Admins)); err != nil {
		return nil, custom_error.HandlePostgreError(err)
	}

	_ = g.setIntoRedis(ctx, groupID, group)

	return &group, nil
}

func (g *GroupRepo) GetByName(ctx context.Context, groupName string) (*model.Group, error) {
	query := `SELECT id, group_name, created_at, last_updated, conv_id, admins FROM groups WHERE group_name = $1 AND deleted = false LIMIT 1`
	row := g.db.QueryRowContext(ctx, query, groupName)

	var group model.Group
	err := row.Scan(&group.ID, &group.GroupName, &group.CreatedAt, &group.LastUpdated, &group.ConversationID, (*pq.StringArray)(&group.Admins))
	if err != nil {
		return nil, custom_error.HandlePostgreError(err)
	}

	return &group, nil
}

func (g *GroupRepo) GetByConversationID(ctx context.Context, conversationID string) (*model.Group, error) {
	query := `SELECT id, group_name, created_at, last_updated, conv_id, admins FROM groups WHERE conv_id = $1 AND deleted = false LIMIT 1`
	row := g.db.QueryRowContext(ctx, query, conversationID)

	var group model.Group
	err := row.Scan(&group.ID, &group.GroupName, &group.CreatedAt, &group.LastUpdated, &group.ConversationID, (*pq.StringArray)(&group.Admins))
	if err != nil {
		return nil, custom_error.HandlePostgreError(err)
	}

	return &group, nil
}

func (g *GroupRepo) Create(ctx context.Context, group model.Group) error {
	query := `INSERT INTO groups (id, group_name, created_at, last_updated, conv_id, deleted, admins) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := g.db.ExecContext(ctx, query, group.ID, group.GroupName, group.CreatedAt, group.LastUpdated,
		group.ConversationID, group.Deleted, pq.StringArray(group.Admins))
	return custom_error.HandlePostgreError(err)
}

func (g *GroupRepo) Update(ctx context.Context, params model.UpdateGroupParams) (*model.Group, error) {
	query := `UPDATE groups SET group_name = $2, last_updated = $3, deleted = $4, admins = $5 WHERE id = $1 RETURNING id, group_name, created_at, last_updated, conv_id, admins`
	row := g.db.QueryRowContext(ctx, query, params.ID, params.GroupName, params.LastUpdated, params.Deleted, pq.StringArray(params.Admins))

	var group model.Group
	err := row.Scan(&group.ID, &group.GroupName, &group.CreatedAt, &group.LastUpdated, &group.ConversationID, (*pq.StringArray)(&group.Admins))
	if err != nil {
		return nil, custom_error.HandlePostgreError(err)
	}

	_ = g.deleteFromRedis(ctx, group.ID)
	return &group, nil
}
