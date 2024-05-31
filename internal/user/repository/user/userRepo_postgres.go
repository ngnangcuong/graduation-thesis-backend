package user

import (
	"context"
	"database/sql"
	"graduation-thesis/internal/user/model"
	"graduation-thesis/pkg/interfaces"
	"time"
)

type UserRepoPostgres struct {
	db interfaces.DBTX
}

func NewUserRepoPostgres(db interfaces.DBTX) *UserRepoPostgres {
	return &UserRepoPostgres{
		db: db,
	}
}

func (u *UserRepoPostgres) WithTx(tx *sql.Tx) *UserRepoPostgres {
	return &UserRepoPostgres{
		db: tx,
	}
}

func (u *UserRepoPostgres) Create(ctx context.Context, params *model.CreateUserParams) (*model.User, error) {
	query := `INSERT INTO users (id, username, password, first_name, last_name, email, phone_number, created_at, last_updated, avatar)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id, username, password, first_name, last_name, email, phone_number, created_at, last_updated, avatar`
	row := u.db.QueryRowContext(ctx, query, params.ID, params.Username, params.HashPassword, params.FirstName, params.LastName, params.Email, params.PhoneNumber, time.Now(), time.Now(), params.Avatar)

	var result model.User
	err := row.Scan(&result.ID, &result.Username, &result.Password, &result.FirstName, &result.LastName, &result.Email, &result.PhoneNumber, &result.CreatedAt, &result.LastUpdated, &result.Avatar)
	return &result, err
}

func (u *UserRepoPostgres) Update(ctx context.Context, userId string, params model.UpdateUserParams) error {
	query := `UPDATE users SET first_name = $2, last_name = $3, email = $4, phone_number = $5, last_updated = $6, avatar = $7, password = $8
			WHERE id = $1`
	_, err := u.db.ExecContext(ctx, query, userId, params.FirstName, params.LastName, params.Email, params.PhoneNumber, time.Now(), params.Avatar, params.Password)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserRepoPostgres) Get(ctx context.Context, userId string) (*model.User, error) {
	query := `SELECT id, username, password, first_name, last_name, email, phone_number, created_at, last_updated, avatar FROM users WHERE id = $1`
	row := u.db.QueryRowContext(ctx, query, userId)

	var user model.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber, &user.CreatedAt, &user.LastUpdated, &user.Avatar)
	return &user, err
}

func (u *UserRepoPostgres) GetForUpdate(ctx context.Context, userId string) (*model.User, error) {
	query := `SELECT id, username, password, first_name, last_name, email, phone_number, created_at, last_updated, avatar FROM users WHERE id = $1 FOR UPDATE`
	row := u.db.QueryRowContext(ctx, query, userId)

	var user model.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber, &user.CreatedAt, &user.LastUpdated, &user.Avatar)
	return &user, err
}

func (u *UserRepoPostgres) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `SELECT id, username, password, first_name, last_name, email, phone_number, created_at, last_updated, avatar FROM users WHERE username = $1`
	row := u.db.QueryRowContext(ctx, query, username)

	var user model.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber, &user.CreatedAt, &user.LastUpdated, &user.Avatar)
	return &user, err
}

func (u *UserRepoPostgres) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, username, password, first_name, last_name, email, phone_number, created_at, last_updated, avatar FROM users WHERE email = $1`
	row := u.db.QueryRowContext(ctx, query, email)

	var user model.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.FirstName, &user.LastName, &user.Email, &user.PhoneNumber, &user.CreatedAt, &user.LastUpdated, &user.Avatar)
	return &user, err
}

func (u *UserRepoPostgres) Delete(ctx context.Context, userId string) error {
	return nil
}
