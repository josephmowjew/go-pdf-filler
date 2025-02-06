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

	"github.com/josephmowjew/go-form-processor/types"
)

// Uploader interface defines the contract for uploading PDFs
type Uploader interface {
	Upload(ctx context.Context, data []byte, config types.UploadConfig) (*types.UploadResponse, error)
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

// Update the Upload method to return the full response
func (u *httpUploader) Upload(ctx context.Context, data []byte, config types.UploadConfig) (*types.UploadResponse, error) {
	if err := config.Validate(); err != nil {
		return nil, &ErrInvalidConfig{Message: err.Error()}
	}

	log.Printf("Uploading file %s for org %s", config.FileName, config.OrganizationID)

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
		"organizationalId": config.OrganizationID,
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

	// Create request with properly formatted URL - remove /upload from path
	uploadURL := fmt.Sprintf("%s?organisationalId=%s&branchId=%s&createdBy=%s&authenticate=false",
		u.baseURL,
		config.OrganizationID,
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
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read and log the raw response for debugging
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log the raw response
	fmt.Printf("Raw server response: %s\n", string(respBody))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Create new reader from the response body we read
	var result types.UploadResponse
	if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w\nResponse body: %s", err, string(respBody))
	}

	return &result, nil
}
