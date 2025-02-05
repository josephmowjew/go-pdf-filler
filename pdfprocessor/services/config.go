package service

import (
	"fmt"
)

// Config holds the service configuration
type Config struct {
	UploadBaseURL string
	BearerToken   string
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
