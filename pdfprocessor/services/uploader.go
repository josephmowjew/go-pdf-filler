package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

// Update Response structure to match the actual server response
type UploadResponse struct {
	FileName        string `json:"fileName"`
	FileDownloadUri string `json:"fileDownloadUri"`
	FileType        string `json:"fileType"`
	Size            int64  `json:"size"`
}

// Update the interface to return the full response
type Uploader interface {
	Upload(ctx context.Context, data []byte, config UploadConfig) (*UploadResponse, error)
}

type httpUploader struct {
	baseURL     string
	bearerToken string
	client      *http.Client
}

// NewUploader creates a new instance of the HTTP uploader with the given configuration.
func NewUploader(config Config) Uploader {
	return &httpUploader{
		baseURL:     config.UploadBaseURL,
		bearerToken: config.BearerToken,
		client:      &http.Client{},
	}
}

// Custom error types for better error handling
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	switch e.StatusCode {
	case 401:
		return "Authentication failed: Bearer token is invalid or expired"
	case 403:
		return "Authorization failed: Insufficient permissions to upload file"
	case 520:
		return "Server error: Unable to process upload request (possibly due to invalid or expired authentication)"
	default:
		return fmt.Sprintf("Upload failed with status %d: %s", e.StatusCode, e.Message)
	}
}

// Update the Upload method to return the full response
func (u *httpUploader) Upload(ctx context.Context, data []byte, config UploadConfig) (*UploadResponse, error) {
	if err := config.Validate(); err != nil {
		return nil, &ErrInvalidConfig{Message: err.Error()}
	}

	log.Printf("Uploading file %s for org %s", config.FileName, config.OrganizationalID)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", config.FileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	// Add metadata
	metadata := map[string]string{
		"organizationalId": config.OrganizationalID,
		"branchId":         config.BranchID,
		"createdBy":        config.CreatedBy,
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := writer.WriteField("metadata", string(metadataJSON)); err != nil {
		return nil, fmt.Errorf("failed to write metadata field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	uploadURL := fmt.Sprintf("%s?organisationalId=%s&branchId=%s&createdBy=%s&authenticate=false",
		u.baseURL,
		config.OrganizationalID,
		config.BranchID,
		config.CreatedBy,
	)

	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+u.bearerToken)

	// Send request
	resp, err := u.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error while uploading: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log the raw response for debugging
	log.Printf("Server response status: %d", resp.StatusCode)
	log.Printf("Server response body: %s", string(respBody))

	// Handle different status codes
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorMessage string
		// Try to parse error message from response body
		var errorResponse struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if err := json.Unmarshal(respBody, &errorResponse); err == nil {
			if errorResponse.Message != "" {
				errorMessage = errorResponse.Message
			} else if errorResponse.Error != "" {
				errorMessage = errorResponse.Error
			}
		}
		if errorMessage == "" {
			errorMessage = string(respBody)
		}

		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    errorMessage,
		}
	}

	// Parse successful response
	var result UploadResponse
	if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode successful response: %w\nResponse body: %s", err, string(respBody))
	}

	return &result, nil
}
