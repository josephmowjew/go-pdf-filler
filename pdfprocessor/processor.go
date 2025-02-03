// Package pdfprocessor provides functionality for dynamically filling PDF forms
// with support for various field types, validation, and configuration options.
package pdfprocessor

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/desertbit/fillpdf"
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
	options   Options
}

// Options configures the behavior of the PDF form processor.
type Options struct {
	ValidateOnSet bool        // Whether to validate fields when they are set
	Logger        *log.Logger // Logger for processing information
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
