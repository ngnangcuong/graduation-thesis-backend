package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"graduation-thesis/internal/user/helper/argon2"
	"graduation-thesis/internal/user/model"
	user "graduation-thesis/internal/user/repository/user"
	"graduation-thesis/pkg/custom_error"
	responseModel "graduation-thesis/pkg/model"
	"net/http"

	"github.com/twinj/uuid"
)

type UserService struct {
	db               *sql.DB
	userRepoPostgres *user.UserRepoPostgres
	userRepoRedis    *user.UserRepoRedis
	mapError         map[error]int
}

func NewUserService(db *sql.DB, userRepoPostgres *user.UserRepoPostgres, userRepoRedis *user.UserRepoRedis, mapError map[error]int) *UserService {
	return &UserService{
		db:               db,
		userRepoPostgres: userRepoPostgres,
		userRepoRedis:    userRepoRedis,
		mapError:         mapError,
	}
}

func (u *UserService) execTx(ctx context.Context, fn func(*user.UserRepoPostgres) error) error {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	userRepoWithTx := u.userRepoPostgres.WithTx(tx)
	err = fn(userRepoWithTx)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rErr)
		}
		return err
	}

	return tx.Commit()
}

func (u *UserService) GetUser(ctx context.Context, id string) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	var successResponse responseModel.SuccessResponse
	var errorResponse responseModel.ErrorResponse

	userCache, err := u.userRepoRedis.Get(ctx, id)
	if err == nil {
		successResponse.Status = http.StatusOK
		successResponse.Result = model.GetUserResponse{
			ID:          userCache.ID,
			FirstName:   userCache.FirstName,
			LastName:    userCache.LastName,
			Email:       userCache.Email,
			PhoneNumber: userCache.PhoneNumber,
			Avatar:      userCache.Avatar,
			CreatedAt:   userCache.CreatedAt,
			LastUpdated: userCache.LastUpdated,
		}
		return &successResponse, nil
	}

	user, err := u.userRepoPostgres.Get(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			errorResponse.Status = http.StatusNotFound
			errorResponse.ErrorMessage = model.ErrNoUser.Error()
			return nil, &errorResponse
		}
		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = model.ErrInternalServerError.Error()
		return nil, &errorResponse
	}
	if user == nil {
		errorResponse.Status = http.StatusNotFound
		errorResponse.ErrorMessage = model.ErrNoUser.Error()
		return nil, &errorResponse
	}

	go u.userRepoRedis.Create(ctx, user)

	successResponse.Result = model.GetUserResponse{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Avatar:      user.Avatar,
		CreatedAt:   user.CreatedAt,
		LastUpdated: user.LastUpdated,
	}
	successResponse.Status = http.StatusOK

	return &successResponse, nil
}

func (u *UserService) CreateUser(ctx context.Context, createUserRequest model.CreateUserRequest) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	var successResponse responseModel.SuccessResponse
	var errorResponse responseModel.ErrorResponse

	userEmail, eErr := u.userRepoPostgres.GetByEmail(ctx, createUserRequest.Email)
	userByName, uErr := u.userRepoPostgres.GetByUsername(ctx, createUserRequest.Username)
	if (errors.Is(eErr, sql.ErrNoRows) && errors.Is(uErr, sql.ErrNoRows)) ||
		(userByName == nil && userEmail == nil) {
		hashPassword, hashErr := argon2.HashPassword([]byte(createUserRequest.Password))
		if hashErr != nil {
			errorResponse.Status = http.StatusInternalServerError
			errorResponse.ErrorMessage = model.ErrInternalServerError.Error()
			return nil, &errorResponse
		}
		createUserParams := model.CreateUserParams{
			ID:           uuid.NewV4().String(),
			Username:     createUserRequest.Username,
			HashPassword: hashPassword,
			FirstName:    createUserRequest.FirstName,
			LastName:     createUserRequest.LastName,
			Email:        createUserRequest.Email,
			PhoneNumber:  createUserRequest.PhoneNumber,
			Avatar:       createUserRequest.Avatar,
		}
		newUser, createErr := u.userRepoPostgres.Create(ctx, &createUserParams)
		if createErr != nil {
			errorResponse.Status = http.StatusInternalServerError
			errorResponse.ErrorMessage = model.ErrInternalServerError.Error()
			return nil, &errorResponse
		}

		successResponse.Result = newUser
		successResponse.Status = http.StatusCreated
		return &successResponse, nil
	}

	errorResponse.Status = http.StatusBadRequest
	errorResponse.ErrorMessage = model.ErrUserAlreadyExist.Error()

	return nil, &errorResponse
}

func (u *UserService) UpdateUser(ctx context.Context, id string, updateUserRequest model.UpdateUserRequest) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	var successResponse responseModel.SuccessResponse
	var errorResponse responseModel.ErrorResponse

	user, err := u.userRepoPostgres.GetForUpdate(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorResponse.Status = http.StatusNotFound
			errorResponse.ErrorMessage = model.ErrNoUser.Error()
		} else {
			errorResponse.Status = http.StatusInternalServerError
			errorResponse.ErrorMessage = model.ErrInternalServerError.Error()
		}

		return nil, &errorResponse
	}

	if user == nil {
		errorResponse.Status = http.StatusNotFound
		errorResponse.ErrorMessage = model.ErrNoUser.Error()
		return nil, &errorResponse
	}

	var updateUserParams model.UpdateUserParams = model.UpdateUserParams(updateUserRequest)
	if updateUserRequest.Password == "" {
		updateUserParams.Password = user.Password
	} else {
		newHashPassword, hashErr := argon2.HashPassword([]byte(updateUserRequest.Password))
		if hashErr != nil {
			errorResponse.Status = http.StatusInternalServerError
			errorResponse.ErrorMessage = model.ErrInternalServerError.Error()
			return nil, &errorResponse
		}

		updateUserParams.Password = newHashPassword
	}

	if updateUserRequest.FirstName == "" {
		updateUserParams.FirstName = user.FirstName
	}
	if updateUserRequest.LastName == "" {
		updateUserParams.LastName = user.LastName
	}
	if updateUserRequest.Email == "" {
		updateUserParams.Email = user.Email
	}
	if updateUserRequest.PhoneNumber == "" {
		updateUserParams.PhoneNumber = user.PhoneNumber
	}
	if updateUserRequest.Avatar == nil {
		updateUserParams.Avatar = user.Avatar
	}

	uErr := u.userRepoPostgres.Update(ctx, id, updateUserParams)
	if uErr != nil {
		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = model.ErrInternalServerError.Error()
		return nil, &errorResponse
	}

	go u.userRepoRedis.Delete(ctx, id)
	successResponse.Status = http.StatusOK
	return &successResponse, nil
}

func (u *UserService) VerifyCredential(ctx context.Context, loginRequest *model.LoginRequest) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	user, err := u.userRepoPostgres.GetByUsername(ctx, loginRequest.Username)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       u.mapError[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	check, checkErr := argon2.Compare(user.Password, []byte(loginRequest.Password))
	if !check || checkErr != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusUnauthorized,
			ErrorMessage: custom_error.ErrNoPermission.Error(),
		}
		return nil, &errorResponse
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
		Result: user.ID,
	}
	return &successResponse, nil
}
