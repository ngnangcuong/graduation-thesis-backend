package user

import (
	"context"
	"database/sql"
	"graduation-thesis/internal/user/model"
)

type IUserRepo interface {
	WithTx(tx *sql.Tx) IUserRepo
	Create(ctx context.Context, params *model.CreateUserParams) (*model.User, error)
	Update(ctx context.Context, userId string, params model.UpdateUserParams) error
	Get(ctx context.Context, userId string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetForUpdate(ctx context.Context, userId string) (*model.User, error)
	Delete(ctx context.Context, userId string) error
}
