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
- Direct upload functionality
- Customizable configuration options
- Automatic temporary file cleanup
- Enhanced error handling for upload operations

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
go get gitlab.lyvepulse.com:lyvepulse/go-pdf-filler@v0.1.1
```

## Usage

Here's a basic example of how to use the package with upload functionality:

```go
package main

import (
    "context"
    "log"
    "os"
    "gitlab.lyvepulse.com:lyvepulse/go-pdf-filler/pdfprocessor"
    service "gitlab.lyvepulse.com:lyvepulse/go-pdf-filler/pdfprocessor/services"
)

func main() {
    // Create an uploader instance with bearer token from environment
    uploaderConfig := service.Config{
        UploadBaseURL: "https://your-upload-service.com/api/upload",
        BearerToken:   os.Getenv("PDF_UPLOADER_TOKEN"), // Get token from environment variable
    }
    uploader := service.NewUploader(uploaderConfig)

    // Initialize the PDF form processor with uploader
    processor, err := pdfprocessor.NewFormFromURL("https://example.com/form.pdf",
        pdfprocessor.WithValidation(),
        pdfprocessor.WithLogger(log.Default()),
        pdfprocessor.WithUploader(uploader),
    )
    if err != nil {
        log.Fatalf("Failed to create form: %v", err)
    }
    // Ensure cleanup of temporary files
    defer processor.Cleanup()

    // Define field values to be set
    fields := map[string]interface{}{
        "Name": "John Doe",
        "Age": "30",
        "IsEmployed": true,
        "Department": "Engineering",
    }

    // Set all fields
    if err := processor.SetFields(fields); err != nil {
        log.Fatalf("Error setting fields: %v", err)
    }

    // Validate the form
    if err := processor.Validate(); err != nil {
        log.Fatalf("Validation failed: %v", err)
    }

    // Create upload configuration
    uploadConfig := service.UploadConfig{
        FileName:         "filled_form.pdf",
        OrganizationalID: "org123",
        BranchID:         "branch456",
        CreatedBy:        "system",
    }

    // Upload with better error handling
    ctx := context.Background()
    response, err := processor.Upload(ctx, uploadConfig)
    if err != nil {
        switch e := err.(type) {
        case *service.HTTPError:
            log.Fatalf("Upload failed: %s", e.Error())
        default:
            log.Fatalf("Unexpected error during upload: %v", err)
        }
    }

    log.Printf("Form uploaded successfully! Download URL: %s", response.FileDownloadUri)
}
```

## Security Considerations

- Never hardcode bearer tokens in your code
- Use environment variables or secure configuration management for sensitive credentials
- Always use HTTPS for remote PDF form URLs
- Implement proper error handling for authentication failures
- Clean up temporary files using the provided `Cleanup()` method

## API Documentation

### Creating a New Form Processor

```go
processor, err := pdfprocessor.NewForm(inputPath string, opts ...Option)
```

### Configuration Options

- `WithValidation()`: Enables validation when setting field values
- `WithLogger(logger *log.Logger)`: Sets a custom logger for the form processor
- `WithUploader(uploader service.Uploader)`: Sets the uploader service for direct upload functionality

### Main Methods

- `SetField(name string, value interface{}) error`: Set a single field value
- `SetFields(fields map[string]interface{}) error`: Set multiple field values
- `Validate() error`: Validate all form fields
- `Upload(ctx context.Context, config service.UploadConfig) (*service.UploadResponse, error)`: Upload the filled form

### Field Types

The package supports three types of form fields:
- `Text`: For text input fields
- `Boolean`: For checkboxes and radio buttons
- `Choice`: For dropdown menus and list selections

### Upload Configuration

The `UploadConfig` struct allows you to specify:
- `FileName`: Name of the file to be uploaded
- `OrganizationalID`: Organization identifier
- `BranchID`: Branch identifier
- `CreatedBy`: User or system identifier

## Error Handling

The package provides detailed error messages for:
- Field not found
- Invalid field type
- Required field validation
- Invalid choice options
- Upload failures
- Configuration errors

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
