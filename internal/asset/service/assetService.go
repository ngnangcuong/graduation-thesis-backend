package service

import (
	"context"
	"net/http"
	"os"

	"graduation-thesis/internal/asset/model"
	"graduation-thesis/internal/asset/repository"
	"graduation-thesis/pkg/logger"
	responseModel "graduation-thesis/pkg/model"
)

type AssetService struct {
	assetRepo *repository.AssetRepo
	mapError  map[error]int
	logger    logger.Logger
}

func NewAssetService(assetRepo *repository.AssetRepo, mapError map[error]int, logger logger.Logger) *AssetService {
	return &AssetService{
		assetRepo: assetRepo,
		mapError:  mapError,
		logger:    logger,
	}
}

func getFile(filePath string) (*os.File, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func deleteFile(filePath string) error {
	return os.Remove(filePath)
}

func (a *AssetService) Upload(ctx context.Context, location, dataType string) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	a.logger.Info("Here")
	f, err := getFile(location)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusNotFound,
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}
	defer deleteFile(location)

	updateFile := model.UploadFile{
		FLocation: location,
		FType:     dataType,
		FData:     f,
	}

	source, err := a.assetRepo.Upload(ctx, &updateFile)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       a.mapError[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
		Result: source,
	}

	return &successResponse, nil
}

func (a *AssetService) Get(ctx context.Context) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	return nil, nil
}

func (a *AssetService) Delete(ctx context.Context, fid string) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	if err := a.assetRepo.Delete(ctx, fid); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       a.mapError[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusNoContent,
	}
	return &successResponse, nil
}
