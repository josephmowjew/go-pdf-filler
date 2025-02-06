package pdfprocessor

import (
	"context"

	"github.com/josephmowjew/go-form-processor/types"
)

// FormProcessor defines the common interface for both PDF and HTML form processing
type FormProcessor interface {
	// GetFields returns all form fields
	GetFields() map[string]Field
	// SetField sets a single field value
	SetField(name string, value interface{}) error
	// SetFields sets multiple field values
	SetFields(fields map[string]interface{}) error
	// Validate checks if all required fields are set
	Validate() error
	// Upload uploads the filled form
	Upload(ctx context.Context, config types.UploadConfig) (*types.UploadResponse, error)
	// PrintFields displays all fields and their properties
	PrintFields()
}
