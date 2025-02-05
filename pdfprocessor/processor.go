// Package pdfprocessor provides functionality for dynamically filling PDF forms
// with support for various field types, validation, and configuration options.
package pdfprocessor

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/desertbit/fillpdf"
	service "gitlab.lyvepulse.com/lyvepulse/go-pdf-filler/pdfprocessor/services"
	"gitlab.lyvepulse.com/lyvepulse/go-pdf-filler/types"
)

// FieldType represents the type of form field in a PDF document.
type FieldType int

const (
	// Text represents a text input field.
	Text FieldType = iota
	// Boolean represents a checkbox or radio button field.
	Boolean
	// Choice represents a dropdown or list selection field.
	Choice
)

// Field represents a single form field in a PDF document.
type Field struct {
	Name     string      // Name of the field in the PDF
	Type     FieldType   // Type of the field
	Options  []string    // Available options for Choice fields
	Required bool        // Whether the field is required
	Value    interface{} // Current value of the field
}

// PDFForm represents a PDF form with its fields and configuration.
type PDFForm struct {
	fields    map[string]Field
	inputPath string
	inputURL  string
	options   Options
}

// Options configures the behavior of the PDF form processor.
type Options struct {
	ValidateOnSet bool             // Whether to validate fields when they are set
	Logger        *log.Logger      // Logger for processing information
	Uploader      service.Uploader // Uploader service for direct PDF uploads
}

// Option is a function that configures Options.
type Option func(*Options)

// WithValidation enables validation when setting field values.
func WithValidation() Option {
	return func(o *Options) {
		o.ValidateOnSet = true
	}
}

// WithLogger sets a custom logger for the form processor.
func WithLogger(logger *log.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// WithUploader sets the uploader service for the form processor.
func WithUploader(uploader service.Uploader) Option {
	return func(o *Options) {
		o.Uploader = uploader
	}
}

// NewForm creates a new PDFForm instance with the specified input path and options.
func NewForm(inputPath string, opts ...Option) (*PDFForm, error) {
	options := Options{
		Logger: log.Default(),
	}
	for _, opt := range opts {
		opt(&options)
	}

	form := &PDFForm{
		inputPath: inputPath,
		fields:    make(map[string]Field),
		options:   options,
	}

	if err := form.loadFields(); err != nil {
		return nil, fmt.Errorf("failed to load form fields: %w", err)
	}

	return form, nil
}

// NewFormFromURL creates a new PDFForm instance from a URL with the specified options.
func NewFormFromURL(url string, opts ...Option) (*PDFForm, error) {
	// Download the file to a temporary location
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download PDF: %w", err)
	}
	defer resp.Body.Close()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "pdf-form-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	// Copy the response body to the temporary file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to save PDF to temporary file: %w", err)
	}
	tmpFile.Close()

	options := Options{
		Logger: log.Default(),
	}
	for _, opt := range opts {
		opt(&options)
	}

	form := &PDFForm{
		inputPath: tmpFile.Name(),
		inputURL:  url,
		fields:    make(map[string]Field),
		options:   options,
	}

	if err := form.loadFields(); err != nil {
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to load form fields: %w", err)
	}

	// Add cleanup function to the form
	runtime.SetFinalizer(form, func(f *PDFForm) {
		if f.inputURL != "" && f.inputPath != "" {
			os.Remove(f.inputPath)
		}
	})

	return form, nil
}

// loadFields reads field information from the PDF using pdftk.
func (f *PDFForm) loadFields() error {
	cmd := exec.Command("pdftk", f.inputPath, "dump_data_fields")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pdftk error: %w", err)
	}

	blocks := strings.Split(string(output), "---")
	for _, block := range blocks {
		field := parseFieldBlock(block)
		if field.Name != "" {
			f.fields[field.Name] = field
		}
	}
	return nil
}

// parseFieldBlock parses a single field block from pdftk output.
func parseFieldBlock(block string) Field {
	lines := strings.Split(block, "\n")
	field := Field{
		Options: []string{},
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]
		switch key {
		case "FieldName":
			field.Name = value
		case "FieldType":
			field.Type = mapFieldType(value)
		case "FieldStateOption":
			field.Options = append(field.Options, value)
		case "FieldFlags":
			if strings.Contains(value, "Required") {
				field.Required = true
			}
		}
	}
	return field
}

// mapFieldType converts pdftk field type to internal FieldType.
func mapFieldType(pdftkType string) FieldType {
	switch pdftkType {
	case "Text":
		return Text
	case "Button":
		return Boolean
	case "Choice":
		return Choice
	default:
		return Text
	}
}

// SetField sets a value for a specific form field with type validation.
func (f *PDFForm) SetField(name string, value interface{}) error {
	field, exists := f.fields[name]
	if !exists {
		return fmt.Errorf("field %s not found in form", name)
	}

	// Type validation
	switch field.Type {
	case Text:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field %s requires string value", name)
		}
	case Boolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field %s requires boolean value", name)
		}
	case Choice:
		if strVal, ok := value.(string); ok {
			if !isValidOption(strVal, field.Options) {
				return fmt.Errorf("invalid option for field %s: %s", name, strVal)
			}
		} else {
			return fmt.Errorf("field %s requires string value from options", name)
		}
	}

	field.Value = value
	f.fields[name] = field

	if f.options.ValidateOnSet {
		return f.validateField(field)
	}
	return nil
}

// SetFields sets multiple field values at once.
func (f *PDFForm) SetFields(fields map[string]interface{}) error {
	for name, value := range fields {
		if err := f.SetField(name, value); err != nil {
			return fmt.Errorf("error setting field %s: %w", name, err)
		}
	}
	return nil
}

// Validate checks if all required fields have values.
func (f *PDFForm) Validate() error {
	for _, field := range f.fields {
		if field.Required && field.Value == nil {
			return fmt.Errorf("required field %s is missing", field.Name)
		}
	}
	return nil
}

// Save writes the filled form to the specified output path.
func (f *PDFForm) Save(outputPath string) error {
	formData := make(fillpdf.Form)

	for name, field := range f.fields {
		if field.Value == nil {
			continue
		}

		switch v := field.Value.(type) {
		case bool:
			if v {
				formData[name] = "On"
			} else {
				formData[name] = "Off"
			}
		case time.Time:
			formData[name] = v.Format(time.RFC3339)
		default:
			formData[name] = fmt.Sprint(v)
		}
	}

	if err := fillpdf.Fill(formData, f.inputPath, outputPath); err != nil {
		return fmt.Errorf("fillpdf error: %w", err)
	}
	return nil
}

// isValidOption checks if a value is in the list of allowed options.
func isValidOption(value string, options []string) bool {
	for _, opt := range options {
		if opt == value {
			return true
		}
	}
	return false
}

// validateField checks if a field meets validation requirements.
func (f *PDFForm) validateField(field Field) error {
	if field.Required && field.Value == nil {
		return fmt.Errorf("required field %s is not set", field.Name)
	}
	return nil
}

// Upload generates the filled PDF and uploads it using the configured uploader service.
func (f *PDFForm) Upload(ctx context.Context, config types.UploadConfig) (*types.UploadResponse, error) {
	if f.options.Uploader == nil {
		return nil, fmt.Errorf("uploader service not configured")
	}

	// Convert form data to fillpdf.Form
	formData := make(fillpdf.Form)
	for name, field := range f.fields {
		if field.Value == nil {
			continue
		}

		switch v := field.Value.(type) {
		case bool:
			if v {
				formData[name] = "On"
			} else {
				formData[name] = "Off"
			}
		case time.Time:
			formData[name] = v.Format(time.RFC3339)
		default:
			formData[name] = fmt.Sprint(v)
		}
	}

	// Create a temporary file for fillpdf (it requires file paths)
	tempOutput := "temp_output.pdf"
	if err := fillpdf.Fill(formData, f.inputPath, tempOutput); err != nil {
		return nil, fmt.Errorf("failed to fill PDF: %w", err)
	}

	// Read the temporary file
	data, err := os.ReadFile(tempOutput)
	if err != nil {
		os.Remove(tempOutput) // Clean up
		return nil, fmt.Errorf("failed to read filled PDF: %w", err)
	}

	// Clean up the temporary file
	os.Remove(tempOutput)

	// Upload the filled PDF
	response, err := f.options.Uploader.Upload(ctx, data, config)
	if err != nil {
		return nil, fmt.Errorf("failed to upload PDF: %w", err)
	}

	return response, nil
}

// GetFields returns a map of all fields in the PDF form.
func (f *PDFForm) GetFields() map[string]Field {
	// Return a copy of the fields map to prevent modification of internal state
	fields := make(map[string]Field, len(f.fields))
	for k, v := range f.fields {
		fields[k] = v
	}
	return fields
}

// PrintFields prints all fields and their properties to the configured logger.
func (f *PDFForm) PrintFields() {
	if f.options.Logger == nil {
		return
	}

	f.options.Logger.Println("PDF Form Fields:")
	f.options.Logger.Println("================")

	for name, field := range f.fields {
		fieldType := "Text"
		switch field.Type {
		case Boolean:
			fieldType = "Boolean"
		case Choice:
			fieldType = "Choice"
		}

		f.options.Logger.Printf("Field: %s\n", name)
		f.options.Logger.Printf("  Type: %s\n", fieldType)
		f.options.Logger.Printf("  Required: %v\n", field.Required)
		if len(field.Options) > 0 {
			f.options.Logger.Printf("  Options: %v\n", field.Options)
		}
		if field.Value != nil {
			f.options.Logger.Printf("  Current Value: %v\n", field.Value)
		}
		f.options.Logger.Println("----------------")
	}
}

// PDFProcessorConfig represents the configuration for the PDF processor
type PDFProcessorConfig struct {
	// Upload configuration
	UploadBaseURL string
	BearerToken   string

	// Optional configurations
	ValidateOnSet bool
	Logger        *log.Logger
}

// NewPDFProcessor creates a new PDF processor with the given configuration
func NewPDFProcessor(config PDFProcessorConfig) (*PDFForm, error) {
	uploader := service.NewUploader(service.Config{
		UploadBaseURL: config.UploadBaseURL,
		BearerToken:   config.BearerToken,
	})

	options := Options{
		ValidateOnSet: config.ValidateOnSet,
		Logger:        config.Logger,
		Uploader:      uploader,
	}

	return &PDFForm{
		options: options,
		fields:  make(map[string]Field),
	}, nil
}

// UploadConfig represents the configuration for uploading a filled PDF
type UploadConfig struct {
	FileName       string
	OrganizationID string
	BranchID       string
	CreatedBy      string
}

// Validate checks if the upload configuration is valid
func (c UploadConfig) Validate() error {
	if c.FileName == "" {
		return fmt.Errorf("filename is required")
	}
	if c.OrganizationID == "" {
		return fmt.Errorf("organizational ID is required")
	}
	if c.BranchID == "" {
		return fmt.Errorf("branch ID is required")
	}
	if c.CreatedBy == "" {
		return fmt.Errorf("creator is required")
	}
	return nil
}

// NormalizeFieldName normalizes a field name for comparison
func (f *PDFForm) NormalizeFieldName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	// Remove extra spaces
	name = strings.TrimSpace(name)
	// Replace multiple spaces with single space
	name = strings.Join(strings.Fields(name), " ")
	// Remove common suffixes
	name = strings.TrimSuffix(name, " optional")
	name = strings.TrimSuffix(name, " required")
	// Remove special characters
	name = strings.Map(func(r rune) rune {
		if r == ' ' || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, name)
	return name
}

// ConvertFieldValue converts a value to the appropriate type based on the field type
func (f *PDFForm) ConvertFieldValue(name string, value interface{}) (interface{}, error) {
	field, exists := f.fields[name]
	if !exists {
		return nil, fmt.Errorf("field %s not found", name)
	}

	switch field.Type {
	case Boolean:
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			v = strings.ToLower(strings.TrimSpace(v))
			if v == "true" || v == "yes" || v == "1" || v == "on" {
				return true, nil
			}
			if v == "false" || v == "no" || v == "0" || v == "off" {
				return false, nil
			}
			return false, fmt.Errorf("invalid boolean value for field %s: %v", name, value)
		default:
			return false, fmt.Errorf("unsupported value type for boolean field %s: %T", name, value)
		}
	case Text:
		switch v := value.(type) {
		case string:
			return v, nil
		default:
			return fmt.Sprintf("%v", value), nil
		}
	case Choice:
		strVal := fmt.Sprintf("%v", value)
		if !isValidOption(strVal, field.Options) {
			return nil, fmt.Errorf("invalid option for field %s: %s", name, strVal)
		}
		return strVal, nil
	default:
		return fmt.Sprintf("%v", value), nil
	}
}

// FindMatchingField attempts to find a matching field name using exact or fuzzy matching
func (f *PDFForm) FindMatchingField(searchName string) (string, bool) {
	normalized := f.NormalizeFieldName(searchName)

	// Try exact match first
	for name := range f.fields {
		if f.NormalizeFieldName(name) == normalized {
			return name, true
		}
	}

	// Try fuzzy match
	for name := range f.fields {
		normalizedField := f.NormalizeFieldName(name)
		if strings.Contains(normalizedField, normalized) ||
			strings.Contains(normalized, normalizedField) {
			return name, true
		}
	}

	return "", false
}
