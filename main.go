package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/desertbit/fillpdf"
)

// FormField represents a unified form field structure
type FormField struct {
	Name         string      // PDF field name
	Value        interface{} // Field value (string/bool)
	Type         string      // PDF field type (Text, Button, Choice)
	DefaultValue string      // Default value if any
	Options      []string    // Available options for Choice fields
}

// PDFForm represents a PDF form with its fields and paths
type PDFForm struct {
	fields     map[string]FormField
	inputPath  string
	outputPath string
}

// NewPDFForm creates a new PDFForm instance
func NewPDFForm(inputPath, outputPath string) (*PDFForm, error) {
	form := &PDFForm{
		inputPath:  inputPath,
		outputPath: outputPath,
		fields:     make(map[string]FormField),
	}

	// Initialize fields from PDF
	if err := form.initializeFields(); err != nil {
		return nil, fmt.Errorf("failed to initialize fields: %v", err)
	}

	return form, nil
}

// parseFieldInfo parses a block of field information from pdftk output
func parseFieldInfo(block string) FormField {
	lines := strings.Split(block, "\n")
	field := FormField{
		Options: make([]string, 0),
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
			field.Type = value
		case "FieldValue":
			field.DefaultValue = value
		case "FieldStateOption":
			field.Options = append(field.Options, value)
		}
	}

	return field
}

// initializeFields reads and initializes fields from the PDF
func (f *PDFForm) initializeFields() error {
	cmd := exec.Command("pdftk", f.inputPath, "dump_data_fields")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list fields: %v", err)
	}

	blocks := strings.Split(string(output), "---")
	for _, block := range blocks {
		if strings.TrimSpace(block) == "" {
			continue
		}
		field := parseFieldInfo(block)
		if field.Name != "" {
			f.fields[field.Name] = field
		}
	}

	return nil
}

// PrintFieldInfo prints the structured field information
func (f *PDFForm) PrintFieldInfo() {
	fmt.Println("\nPDF Form Fields:")
	fmt.Println("================")

	// Group fields by type
	typeGroups := make(map[string][]FormField)
	for _, field := range f.fields {
		typeGroups[field.Type] = append(typeGroups[field.Type], field)
	}

	for fieldType, fields := range typeGroups {
		fmt.Printf("\n%s Fields:\n", fieldType)
		for _, field := range fields {
			printField(field)
		}
	}
}

func printField(field FormField) {
	fmt.Printf("  - %s\n", field.Name)
	fmt.Printf("    Type: %s\n", field.Type)
	if field.DefaultValue != "" {
		fmt.Printf("    Default: %s\n", field.DefaultValue)
	}
	if len(field.Options) > 0 {
		fmt.Printf("    Options: %v\n", field.Options)
	}
	fmt.Println()
}

// SetField sets a field value with type validation
func (f *PDFForm) SetField(name string, value interface{}) error {
	field, exists := f.fields[name]
	if !exists {
		return fmt.Errorf("field not found: %s", name)
	}

	switch field.Type {
	case "Text":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field %s requires string value", name)
		}
	case "Button":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field %s requires boolean value", name)
		}
	default:
		return fmt.Errorf("unsupported field type: %s", field.Type)
	}

	field.Value = value
	f.fields[name] = field
	return nil
}

// Fill fills the PDF form with the provided field values
func (f *PDFForm) Fill() error {
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
		default:
			formData[name] = fmt.Sprint(field.Value)
		}
	}

	return fillpdf.Fill(formData, f.inputPath, f.outputPath)
}

func main() {
	// Create a new form
	form, err := NewPDFForm("form.pdf", "filled_form.pdf")
	if err != nil {
		log.Fatalf("Failed to create form: %v", err)
	}

	// Print available fields
	//form.PrintFieldInfo()

	// Example field values
	fieldsToSet := map[string]interface{}{
		"1 Vehicle Identification Number": "ABC123XYZ456789",
		"2 Year":                          "2024",
		"3 Make":                          "Toyota",
		"4 Body Style":                    "Sedan",
		"5 Model":                         "Camry",
		"6 Major Color":                   "Silver",
		"7 Minor Color":                   "Black",
		"8 Texas License Plate No":        "ABC1234",
		"9 Odometer Reading no tenths":    "50000",
		"Title Only":                      true,
		"Registration Purposes Only":      true,
		"Individual":                      true,
	}

	// Set all fields
	for name, value := range fieldsToSet {
		if err := form.SetField(name, value); err != nil {
			log.Printf("Warning: Failed to set field %s: %v", name, err)
		}
	}

	// Fill the form
	if err := form.Fill(); err != nil {
		log.Fatalf("Failed to fill form: %v", err)
	}

	fmt.Println("Form filled successfully! Output saved to", form.outputPath)
}
