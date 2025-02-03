# go-pdf-filler

A Go package for programmatically filling PDF forms with support for validation, logging, and multiple field types. This package provides a robust solution for automated PDF form filling with type safety and validation.

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
- Customizable configuration options

## Requirements

- Go 1.22 or later
- pdftk (PDF Toolkit) installed on your system
  - For macOS: `brew install pdftk-java`
  - For Ubuntu/Debian: `sudo apt-get install pdftk`
  - For other systems: [PDFtk Server](https://www.pdflabs.com/tools/pdftk-server/)

## Installation

```bash
go get github.com/yourusername/go-pdf-filler
```

## Usage

Here's a basic example of how to use the package:

```go
package main

import (
    "log"
    "github.com/yourusername/go-pdf-filler/pdfprocessor"
)

func main() {
    // Initialize the PDF form processor
    processor, err := pdfprocessor.NewForm("form.pdf",
        pdfprocessor.WithValidation(),
        pdfprocessor.WithLogger(log.Default()),
    )
    if err != nil {
        log.Fatalf("Failed to create form: %v", err)
    }

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

    // Save the filled form
    if err := processor.Save("output.pdf"); err != nil {
        log.Fatalf("Failed to save form: %v", err)
    }
}
```

## API Documentation

### Creating a New Form Processor

```go
processor, err := pdfprocessor.NewForm(inputPath string, opts ...Option)
```

### Configuration Options

- `WithValidation()`: Enables validation when setting field values
- `WithLogger(logger *log.Logger)`: Sets a custom logger for the form processor

### Main Methods

- `SetField(name string, value interface{}) error`: Set a single field value
- `SetFields(fields map[string]interface{}) error`: Set multiple field values
- `Validate() error`: Validate all form fields
- `Save(outputPath string) error`: Save the filled form to a new PDF

### Field Types

The package supports three types of form fields:
- `Text`: For text input fields
- `Boolean`: For checkboxes and radio buttons
- `Choice`: For dropdown menus and list selections

## Error Handling

The package provides detailed error messages for:
- Field not found
- Invalid field type
- Required field validation
- Invalid choice options
- File operations

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