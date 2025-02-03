package service

import "fmt"

// ErrInvalidConfig represents configuration validation errors
type ErrInvalidConfig struct {
	Message string
}

func (e ErrInvalidConfig) Error() string {
	return fmt.Sprintf("invalid configuration: %s", e.Message)
}

// ErrUpload represents upload-related errors
type ErrUpload struct {
	StatusCode int
	Message    string
}

func (e ErrUpload) Error() string {
	return fmt.Sprintf("upload failed (status %d): %s", e.StatusCode, e.Message)
}
