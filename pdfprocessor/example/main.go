// Package main provides an example of using the pdfprocessor package.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"gitlab.lyvepulse.com/lyvepulse/go-pdf-filler/pdfprocessor"
	service "gitlab.lyvepulse.com/lyvepulse/go-pdf-filler/pdfprocessor/services"
	"gitlab.lyvepulse.com/lyvepulse/go-pdf-filler/types"
)

func dumpPDFFields(processor *pdfprocessor.PDFForm, outputPath string) error {
	formFields := processor.GetFields()

	var sb strings.Builder
	sb.WriteString("PDF Form Fields Analysis\n")
	sb.WriteString("=======================\n\n")

	if len(formFields) == 0 {
		sb.WriteString("WARNING: No fields were detected in the PDF!\n\n")
		sb.WriteString("Debug Information:\n")
		sb.WriteString("1. Number of fields: 0\n")
		sb.WriteString("2. Please check if:\n")
		sb.WriteString("   - The PDF actually contains form fields\n")
		sb.WriteString("   - PDFtk is properly installed\n")
		sb.WriteString("   - The PDF file is accessible and not corrupted\n")
	} else {
		sb.WriteString(fmt.Sprintf("Found %d fields:\n\n", len(formFields)))
		for name, field := range formFields {
			sb.WriteString(fmt.Sprintf("Field Name: '%s'\n", name))
			sb.WriteString(fmt.Sprintf("Length: %d characters\n", len(name)))

			// Character analysis
			sb.WriteString("Character Analysis:\n")
			for i, char := range name {
				sb.WriteString(fmt.Sprintf("  Position %d: '%c' (ASCII: %d)\n", i, char, char))
			}

			// Field properties
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
	}

	// Write to file
	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}

func main() {
	// Create processor configuration
	config := pdfprocessor.PDFProcessorConfig{
		UploadBaseURL: "https://staging-storage.lyvepulse.com/storage/files",
		BearerToken:   "bearer_token-goes-here",
		ValidateOnSet: true,
		Logger:        log.Default(),
	}

	// Initialize the processor
	processor, err := pdfprocessor.NewPDFProcessor(config)
	if err != nil {
		log.Fatalf("Failed to initialize processor: %v", err)
	}

	// Create an uploader from the processor config
	uploader := service.NewUploader(service.Config{
		UploadBaseURL: config.UploadBaseURL,
		BearerToken:   config.BearerToken,
	})

	// Example 1: Using a local file
	processorLocal, err := pdfprocessor.NewForm("Sample-Fillable-PDF.pdf",
		pdfprocessor.WithValidation(),
		pdfprocessor.WithLogger(log.Default()),
		pdfprocessor.WithUploader(uploader),
	)
	if err != nil {
		log.Fatalf("Failed to create form from local file: %v", err)
	}

	// Example 2: Using a URL
	// processorURL, err := pdfprocessor.NewFormFromURL("https://www.txdmv.gov/sites/default/files/form_files/130-U.pdf",
	// 	pdfprocessor.WithValidation(),
	// 	pdfprocessor.WithLogger(log.Default()),
	// 	pdfprocessor.WithUploader(uploader),
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to create form from URL: %v", err)
	// }

	// // Dump fields to analysis file
	// if err := dumpPDFFields(processorLocal, "pdf_fields_analysis.txt"); err != nil {
	// 	log.Printf("Warning: Failed to dump fields analysis: %v", err)
	// }

	// First, let's print exact field names with their lengths
	log.Println("Available fields with exact names:")
	formFields := processor.GetFields()
	for name := range formFields {
		log.Printf("Field name: '%s' (length: %d)", name, len(name))
	}

	// Define field values with exact field names as shown in the console
	rawFields := map[string]interface{}{
		"Name of Dependent":    "John Doe",
		"Age     of Dependent": "25",
		"Name":                 "Jane Smith",
		"Dropdown2":            "Choice 1",
		"Option 1":             true,  // Changed from "option1"
		"Option 2":             true,  // Changed from "option2"
		"Option 3":             false, // Changed from "option3"
	}

	// Using the URL-based processor for this example
	processor = processorLocal

	// Print all available fields before setting values
	processor.PrintFields()

	// Alternatively, you can get the fields and process them yourself
	formFields = processor.GetFields()
	for name, field := range formFields {
		log.Printf("Found field: %s (Type: %v, Required: %v)\n",
			name, field.Type, field.Required)
	}

	// Set all fields with smart matching
	if err := processorLocal.SetFields(rawFields); err != nil {
		log.Printf("Warning: Some fields could not be set: %v", err)
	}

	// Validate the form
	if err := processor.Validate(); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	// Create upload configuration using the types package
	uploadConfig := types.UploadConfig{
		FileName:       "vehicle_registration.pdf",
		OrganizationID: "54321",
		BranchID:       "BR-1002",
		CreatedBy:      "system",
	}

	// Upload the filled form
	ctx := context.Background()
	response, err := processor.Upload(ctx, uploadConfig)
	if err != nil {
		log.Fatalf("Failed to upload form: %v", err)
	}

	log.Printf("Form uploaded successfully! Download URL: %s", response.FileDownloadUri)

	// Example of saving to a local file instead of uploading
	if err := processor.Save("filled_form.pdf"); err != nil {
		log.Fatalf("Failed to save form locally: %v", err)
	}
	log.Println("Form saved successfully to filled_form.pdf")

	// Read the analysis file and print it
	// if analysis, err := os.ReadFile("pdf_fields_analysis.txt"); err == nil {
	// 	log.Printf("PDF Fields Analysis:\n%s", string(analysis))
	// }
}
