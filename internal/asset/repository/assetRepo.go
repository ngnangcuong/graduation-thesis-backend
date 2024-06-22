package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"graduation-thesis/internal/asset/model"
	"graduation-thesis/pkg/custom_error"
)

type AssetRepo struct {
	seaweedMasterUrl string
	seaweedVolumeUrl string
	baseUrl          string
}

func NewAssetRepo(seaweedMasterUrl, seaweedVolumeUrl, baseUrl string) *AssetRepo {
	return &AssetRepo{
		seaweedMasterUrl: seaweedMasterUrl,
		seaweedVolumeUrl: seaweedVolumeUrl,
		baseUrl:          baseUrl,
	}
}

func (a *AssetRepo) getAssign() (string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/dir/assign", a.seaweedMasterUrl),
		nil,
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	res, rErr := client.Do(req)
	if rErr != nil {
		if errors.Is(rErr, os.ErrDeadlineExceeded) || errors.Is(rErr, context.DeadlineExceeded) {
			return "", custom_error.ErrTimeout
		}
		return "", custom_error.ErrConnectionErr
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var response map[string]interface{}

	if err := json.Unmarshal(resBody, &response); err != nil {
		return "", err
	}

	return response["fid"].(string), nil
}

func (a *AssetRepo) uploadToS3(fid string, file *model.UploadFile) error {
	var (
		buf = new(bytes.Buffer)
		w   = multipart.NewWriter(buf)
	)

	part, err := w.CreateFormFile(file.FType, filepath.Base(file.FLocation))
	if err != nil {
		return err
	}

	if _, err := io.Copy(part, file.FData); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", a.seaweedVolumeUrl, fid),
		buf,
	)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", w.FormDataContentType())

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	res, rErr := client.Do(req)
	if rErr != nil {
		if errors.Is(rErr, os.ErrDeadlineExceeded) || errors.Is(rErr, context.DeadlineExceeded) {
			return custom_error.ErrTimeout
		}
		return custom_error.ErrConnectionErr
	}
	defer res.Body.Close()
	return nil
}

func (a *AssetRepo) Upload(ctx context.Context, file *model.UploadFile) (string, error) {
	defer file.FData.Close()
	fid, err := a.getAssign()
	if err != nil {
		return "", err
	}

	if err := a.uploadToS3(fid, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", a.baseUrl, fid), nil
}

func (a *AssetRepo) Delete(ctx context.Context, fid string) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/%s", a.seaweedVolumeUrl, fid),
		nil,
	)
	if err != nil {
		return err
	}

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	res, rErr := client.Do(req)
	if rErr != nil {
		if errors.Is(rErr, os.ErrDeadlineExceeded) || errors.Is(rErr, context.DeadlineExceeded) {
			return custom_error.ErrTimeout
		}
		return custom_error.ErrConnectionErr
	}
	defer res.Body.Close()
	return nil
}

func (a *AssetRepo) Get(ctx context.Context, fid string) ([]byte, string, error) {
	url := fmt.Sprintf("%s/%s", a.seaweedVolumeUrl, fid)
	res, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()
	content, rErr := io.ReadAll(res.Body)
	if rErr != nil {
		return nil, "", rErr
	}

	contentType := http.DetectContentType(content)
	return content, contentType, nil
}
