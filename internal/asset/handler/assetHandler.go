package handler

import (
	"fmt"
	"graduation-thesis/internal/asset/service"
	responseModel "graduation-thesis/pkg/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AssetHandler struct {
	localDir     string
	assetService *service.AssetService
}

func NewAssetHandler(localDir string, assetService *service.AssetService) *AssetHandler {
	return &AssetHandler{
		localDir:     localDir,
		assetService: assetService,
	}
}

func (a *AssetHandler) Upload(c *gin.Context) {
	// userID := c.MustGet("user_id")
	file, err := c.FormFile("file")
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}
	location := fmt.Sprintf("%s\\%s_%s", a.localDir, "userID", file.Filename) // Handle name
	c.SaveUploadedFile(file, location)

	successResponse, errorResponse := a.assetService.Upload(c, location, file.Header.Get("Content-Type"))
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, gin.H{
		"location": successResponse.Result,
	})
}

func (a *AssetHandler) Get(c *gin.Context) {
	successResponse, errorResponse := a.assetService.Get(c)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (a *AssetHandler) Delete(c *gin.Context) {
	fid := c.Param("fid")
	successResponse, errorResponse := a.assetService.Delete(c, fid)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}
