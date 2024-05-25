package request

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"graduation-thesis/pkg/custom_error"
	"graduation-thesis/pkg/model"
	"io"
	"net/http"
	"os"
	"time"
)

func HTTPRequestCall(url, method, apiKey string, body io.Reader, timeout time.Duration) (interface{}, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, custom_error.ErrInternalServerError
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := http.Client{
		Timeout: timeout,
	}

	res, rErr := client.Do(req)
	if rErr != nil {
		if errors.Is(rErr, os.ErrDeadlineExceeded) || errors.Is(rErr, context.DeadlineExceeded) {
			return nil, custom_error.ErrTimeout
		}
		return nil, custom_error.ErrConnectionErr
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, custom_error.ErrInternalServerError
	}

	if isSuccessResponse(res.StatusCode) {
		var successResponse model.SuccessResponse
		if err := json.Unmarshal(resBody, &successResponse); err != nil {
			return nil, custom_error.ErrInternalServerError
		}

		return successResponse.Result, nil
	}

	var errorResponse model.ErrorResponse
	if err := json.Unmarshal(resBody, &errorResponse); err != nil {
		return nil, custom_error.ErrInternalServerError
	}

	return nil, custom_error.MappingStatusError()[errorResponse.Status]
}

func isSuccessResponse(statusCode int) bool {
	return statusCode < 400
}
