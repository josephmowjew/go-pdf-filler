# go-pdf-filler

A Go package for programmatically filling PDF forms with support for validation, logging, and multiple field types. This package provides a robust solution for automated PDF form filling with type safety, validation, and direct upload capabilities.

## Features

- Dynamic PDF form field detection and manipulation
- Support for multiple field types:
  - Text fields
  - Boolean fields (checkboxes, radio buttons)
  - Choice fields (dropdowns, lists)
- Built-in field validation
- Configurable logging
- Type-safe field setting
- Batch field updates
- Direct upload functionality with multipart form support
- Field value normalization and type conversion
- Fuzzy field name matching
- Automatic temporary file cleanup
- Enhanced error handling with custom error types
- Configurable upload service with metadata support
- Context-aware operations
- PDF Field Analysis capabilities:
  - Field type detection
  - Required field identification
  - Field options enumeration
  - Current value inspection
  - Detailed field properties export

## Requirements

### Go Version
- Go 1.22 or later

### Required System Dependencies

#### PDFtk Server (Required)
PDFtk Server is required for PDF form field detection and manipulation:
- For macOS: `brew install pdftk-java`
- For Ubuntu/Debian: `sudo apt-get install pdftk`
- For Windows: Download and install [PDFtk Server](https://www.pdflabs.com/tools/pdftk-server/)
- For other Linux distributions: Check your package manager or install from [PDFtk Server](https://www.pdflabs.com/tools/pdftk-server/)

#### Required Go Packages
These will be automatically installed when you run `go get`:
- github.com/desertbit/fillpdf - For PDF form filling
- github.com/unidoc/unipdf/v3 - For PDF processing
- Other dependencies will be handled automatically by Go modules

### Optional Dependencies
- GhostScript (recommended for better PDF handling)
  - For macOS: `brew install ghostscript`
  - For Ubuntu/Debian: `sudo apt-get install ghostscript`
  - For Windows: Download from [Ghostscript Downloads](https://www.ghostscript.com/releases/gsdnld.html)

### System Requirements
- Sufficient disk space for temporary file operations
- Network access for remote PDF fetching and uploading
- Write permissions in the working directory

## Installation

```bash
go get github.com/josephmowjew/go-pdf-filler@v0.1.7
```

## Usage

Here's a complete example demonstrating the main features:

```go
package main

import (
    "context"
    "log"
    "os"
    "github.com/josephmowjew/go-pdf-filler/pdfprocessor"
    "github.com/josephmowjew/go-pdf-filler/types"
)

func main() {
    // Initialize with configuration
    config := pdfprocessor.PDFProcessorConfig{
        UploadBaseURL: "https://your-upload-service.com/api/upload",
        BearerToken:   os.Getenv("PDF_UPLOADER_TOKEN"),
        ValidateOnSet: true,
        Logger:        log.Default(),
    }

    // Create a new processor
    processor, err := pdfprocessor.NewPDFProcessor(config)
    if err != nil {
        log.Fatalf("Failed to initialize processor: %v", err)
    }

    // Create a form from URL with options
    form, err := pdfprocessor.NewFormFromURL("https://example.com/form.pdf",
        pdfprocessor.WithValidation(),
        pdfprocessor.WithLogger(log.Default()),
    )
    if err != nil {
        log.Fatalf("Failed to create form: %v", err)
    }

    // Print available fields
    form.PrintFields()

    // Set multiple fields with automatic type conversion
    fields := map[string]interface{}{
        "Name":        "John Doe",
        "Age":         "30",
        "IsEmployed":  true,
        "Department":  "Engineering",
        "Title Only": true,
    }

    if err := form.SetFields(fields); err != nil {
        log.Fatalf("Error setting fields: %v", err)
    }

    // Configure upload with metadata
    uploadConfig := types.UploadConfig{
        FileName:        "filled_form.pdf",
        OrganizationID:  "org123",
        BranchID:        "branch456",
        CreatedBy:       "system",
    }

    // Upload with context and handle response
    ctx := context.Background()
    response, err := form.Upload(ctx, uploadConfig)
    if err != nil {
        switch e := err.(type) {
        case *service.HTTPError:
            log.Fatalf("Upload failed: %s", e.Error())
        default:
            log.Fatalf("Unexpected error: %v", err)
        }
    }

    log.Printf("Form uploaded successfully! Download URL: %s", response.FileDownloadUri)

    // Analyze PDF fields and export to file
    if err := dumpPDFFields(processor, "pdf_fields_analysis.txt"); err != nil {
        log.Printf("Warning: Failed to dump fields analysis: %v", err)
    }

    // Read and display field analysis
    if analysis, err := os.ReadFile("pdf_fields_analysis.txt"); err == nil {
        log.Printf("PDF Fields Analysis:\n%s", string(analysis))
    }
}

// Helper function to analyze PDF fields
func dumpPDFFields(form *pdfprocessor.PDFForm, outputPath string) error {
    var sb strings.Builder
    fields := form.GetFields()
    
    for name, field := range fields {
        sb.WriteString(fmt.Sprintf("Field: %s\n", name))
        sb.WriteString(fmt.Sprintf("Type: %v\n", field.Type))
        sb.WriteString(fmt.Sprintf("Required: %v\n", field.Required))
        if len(field.Options) > 0 {
            sb.WriteString(fmt.Sprintf("Options: %v\n", field.Options))
        }
        if field.Value != nil {
            sb.WriteString(fmt.Sprintf("Current Value: %v\n", field.Value))
        }
        sb.WriteString("\n-------------------\n\n")
    }
    
    return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}
```

## Field Analysis Output

The field analysis functionality generates a detailed report containing:

```text
Field: Name of Dependent
Type: 0
Required: false
Current Value: John Doe
-------------------

Field: Age of Dependent
Type: 0
Required: false
Current Value: 25
-------------------

Field: Dropdown2
Type: 2
Required: false
Options: [Choice 1 Choice 2 Choice 3]
Current Value: Choice 1
-------------------
```

Field Types:
- Type 0: Text Field
- Type 1: Boolean Field (Checkbox/Radio)
- Type 2: Choice Field (Dropdown/List)

## Security Considerations

- Store bearer tokens in environment variables
- Use HTTPS for remote PDF form URLs
- Implement proper error handling
- Use context for request cancellation
- Clean up temporary files using `Cleanup()`
- Validate upload configurations
- Handle sensitive metadata appropriately

## API Documentation

### Processor Configuration

```go
type PDFProcessorConfig struct {
    UploadBaseURL string       // Base URL for file uploads
    BearerToken   string       // Authentication token
    ValidateOnSet bool         // Enable validation on field set
    Logger        *log.Logger  // Custom logger
}
```

### Form Creation Options

```go
// Create from URL with options
form, err := pdfprocessor.NewFormFromURL(url,
    pdfprocessor.WithValidation(),
    pdfprocessor.WithLogger(logger),
)

// Create from local file
form, err := pdfprocessor.NewForm(filepath,
    pdfprocessor.WithValidation(),
    pdfprocessor.WithLogger(logger),
)
```

### Field Operations

- `GetFields() map[string]Field`: Get all form fields
- `SetField(name string, value interface{}) error`: Set single field
- `SetFields(fields map[string]interface{}) error`: Set multiple fields
- `PrintFields()`: Display all fields and properties
- `FindMatchingField(searchName string) (string, bool)`: Fuzzy field search
- `ConvertFieldValue(name string, value interface{}) (interface{}, error)`: Type conversion
- `Validate() error`: Validate all fields

### Upload Configuration

```go
type UploadConfig struct {
    FileName        string
    OrganizationID  string
    BranchID        string
    CreatedBy       string
}
```

### Upload Response

```go
type UploadResponse struct {
    FileName        string
    FileDownloadUri string
    FileType        string
    Size            int64
}
```

## Error Handling

The package provides custom error types for better error handling:
- `ErrInvalidConfig`: Configuration validation errors
- `HTTPError`: Upload and network-related errors
- Field validation errors
- Type conversion errors

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Uses [fillpdf](https://github.com/desertbit/fillpdf) for PDF manipulation
- Requires PDFtk for PDF processing 
