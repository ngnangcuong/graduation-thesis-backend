package service

import (
	"context"
	"database/sql"
	"graduation-thesis/internal/group/model"
	"graduation-thesis/internal/group/repository"
	"graduation-thesis/pkg/custom_error"
	responseModel "graduation-thesis/pkg/model"
	"net/http"
	"time"

	"github.com/twinj/uuid"
)

type GroupService struct {
	db               *sql.DB
	groupRepo        *repository.GroupRepo
	conversationRepo *repository.ConversationRepo
	errorMap         map[error]int
}

func NewGroupService(db *sql.DB, groupRepo *repository.GroupRepo, conversationRepo *repository.ConversationRepo, errorMap map[error]int) *GroupService {
	return &GroupService{
		db:               db,
		groupRepo:        groupRepo,
		conversationRepo: conversationRepo,
		errorMap:         errorMap,
	}
}

func (g *GroupService) execTx(ctx context.Context, fn func(*repository.GroupRepo, *repository.ConversationRepo) error) error {
	tx, err := g.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	groupRepoWithTx := g.groupRepo.WithTx(tx)
	conversationRepoWithTx := g.conversationRepo.WithTx(tx)
	err = fn(groupRepoWithTx, conversationRepoWithTx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (g *GroupService) isInGroup(userID string, usersInGroup []string) bool {
	for _, user := range usersInGroup {
		if user == userID {
			return true
		}
	}
	return false
}

func (g *GroupService) GetGroup(ctx context.Context, userID, groupID, groupName string) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	if groupName != "" && groupID != "" {
		errorResponse := responseModel.ErrorResponse{
			Status:       g.errorMap[custom_error.ErrInvalidParameter],
			ErrorMessage: custom_error.ErrInvalidParameter.Error(),
		}
		return nil, &errorResponse
	}

	var (
		group *model.Group
		err   error
	)

	queryContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if groupName != "" {
		group, err = g.groupRepo.GetByName(queryContext, groupName)
	} else {
		group, err = g.groupRepo.Get(queryContext, groupID)
	}

	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       g.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
		Result: group,
	}

	return &successResponse, nil
}

func (g *GroupService) CreateGroup(ctx context.Context, request *model.CreateGroupRequest, userID string) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	if len(request.Members) <= 2 {
		errorResponse := responseModel.ErrorResponse{
			Status:       g.errorMap[custom_error.ErrInvalidParameter],
			ErrorMessage: custom_error.ErrInvalidParameter.Error(),
		}
		return nil, &errorResponse
	}

	if !g.isInGroup(userID, request.Members) {
		errorResponse := responseModel.ErrorResponse{
			Status:       g.errorMap[custom_error.ErrNoPermission],
			ErrorMessage: custom_error.ErrNoPermission.Error(),
		}
		return nil, &errorResponse
	}

	var createGroupResponse model.CreateGroupResponse
	queryContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := g.execTx(ctx, func(gr *repository.GroupRepo, cr *repository.ConversationRepo) error {
		conversationID := uuid.NewV4().String()
		cErr := cr.Create(queryContext, conversationID)
		if cErr != nil {
			return cErr
		}

		addErr := cr.AddMembers(queryContext, conversationID, request.Members)
		if addErr != nil {
			return addErr
		}

		group := model.Group{
			ID:             uuid.NewV4().String(),
			GroupName:      request.GroupName,
			CreatedAt:      time.Now(),
			LastUpdated:    time.Now(),
			ConversationID: conversationID,
			Deleted:        false,
			Admins:         []string{userID},
		}
		gErr := gr.Create(queryContext, group)
		if gErr != nil {
			return gErr
		}

		createGroupResponse.GroupID = group.ID
		createGroupResponse.ConversationID = conversationID
		return nil
	})

	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       g.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusCreated,
		Result: createGroupResponse,
	}
	return &successResponse, nil
}

func (g *GroupService) UpdateGroup(ctx context.Context, request *model.UpdateGroupRequest, groupID, userID string) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	if request.GroupName == "" && len(request.Members) == 0 && len(request.Admins) == 0 {
		errorResponse := responseModel.ErrorResponse{
			Status:       g.errorMap[custom_error.ErrInvalidParameter],
			ErrorMessage: custom_error.ErrInvalidParameter.Error(),
		}
		return nil, &errorResponse
	}

	queryContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err := g.execTx(ctx, func(gr *repository.GroupRepo, cr *repository.ConversationRepo) error {
		group, gErr := gr.Get(queryContext, groupID)
		if gErr != nil {
			return gErr
		}

		if !g.isInGroup(userID, group.Admins) {
			return custom_error.ErrNoPermission
		}

		updateGroupParams := model.UpdateGroupParams{
			ID:          groupID,
			LastUpdated: time.Now(),
			Deleted:     false,
		}
		if request.GroupName != "" {
			updateGroupParams.GroupName = request.GroupName
		} else {
			updateGroupParams.GroupName = group.GroupName
		}

		if len(request.Members) > 0 {
			for _, r := range request.Members {
				if r.Action == "add" {
					if err := cr.AddMembers(queryContext, group.ConversationID, r.Users); err != nil {
						return err
					}
				}
				if r.Action == "remove" {
					if err := cr.RemoveMembers(queryContext, group.ConversationID, r.Users); err != nil {
						return err
					}
				}
			}
		}

		updateGroupParams.Admins = group.Admins
		if _, err := g.groupRepo.Update(queryContext, updateGroupParams); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       g.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusNoContent,
	}
	return &successResponse, nil
}

func (g *GroupService) DeleteGroup(ctx context.Context, groupID, userID string) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	queryContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	group, err := g.groupRepo.Get(queryContext, groupID)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       g.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	if !g.isInGroup(userID, group.Admins) {
		errorResponse := responseModel.ErrorResponse{
			Status:       g.errorMap[custom_error.ErrNoPermission],
			ErrorMessage: custom_error.ErrNoPermission.Error(),
		}
		return nil, &errorResponse
	}

	params := model.UpdateGroupParams{
		ID:          groupID,
		GroupName:   group.GroupName,
		LastUpdated: time.Now(),
		Deleted:     true,
		Admins:      group.Admins,
	}

	_, err = g.groupRepo.Update(queryContext, params)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       g.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusNoContent,
	}
	return &successResponse, nil
}

func (g *GroupService) createNewUserList(users []string, request []model.ChangeUser) []string {
	currentAdminUsers := make(map[string]int, len(users))
	for index, user := range users {
		currentAdminUsers[user] = index
	}
	var result []string
	copy(result, users)

	for _, r := range request {
		if r.Action == "add" {
			for _, user := range r.Users {
				if _, ok := currentAdminUsers[user]; !ok {
					result = append(result, user)
				}
			}
		}

		if r.Action == "remove" {
			for _, user := range r.Users {
				if i, ok := currentAdminUsers[user]; ok {
					result = append(result[:i], result[i+1:]...)
				}
			}
		}
	}
	return result
}
