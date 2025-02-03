package service

import (
	"fmt"
)

// Config holds the service configuration
type Config struct {
	UploadBaseURL string
	BearerToken   string
}

// UploadConfig contains configuration for a single upload
type UploadConfig struct {
	OrganizationalID string
	BranchID         string
	CreatedBy        string
	FileName         string
}

// Config validation
func (c Config) Validate() error {
	if c.UploadBaseURL == "" {
		return fmt.Errorf("upload base URL is required")
	}
	if c.BearerToken == "" {
		return fmt.Errorf("bearer token is required")
	}
	return nil
}

// UploadConfig validation
func (c UploadConfig) Validate() error {
	if c.OrganizationalID == "" {
		return fmt.Errorf("organizational ID is required")
	}
	if c.BranchID == "" {
		return fmt.Errorf("branch ID is required")
	}
	if c.CreatedBy == "" {
		return fmt.Errorf("creator is required")
	}
	if c.FileName == "" {
		return fmt.Errorf("filename is required")
	}
	return nil
}
