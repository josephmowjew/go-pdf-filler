// Package main provides an example of using the pdfprocessor package.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/josephmowjew/go-pdf-filler/pdfprocessor"
	service "github.com/josephmowjew/go-pdf-filler/pdfprocessor/services"
	"github.com/josephmowjew/go-pdf-filler/types"
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
	// Set up a more verbose logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Create processor configuration
	config := pdfprocessor.PDFProcessorConfig{
		UploadBaseURL: "https://staging-storage.lyvepulse.com/storage/files",
		BearerToken:   "eyJhbGciOiJIUzM4NCJ9.eyJzdWIiOiJuby1yZXBseUBvcHNzaWZ5LmNvbSIsInVzZXJuYW1lIjoibm8tcmVwbHlAb3Bzc2lmeS5jb20iLCJlbXBsb3llZUlkIjoibm8tcmVwbHlAb3Bzc2lmeS5jb20iLCJmaXJzdE5hbWUiOiJTWVNURU0iLCJsYXN0TmFtZSI6IlNZU1RFTSIsInBob25lTnVtYmVyIjoiODc2NzUyMzQyIiwiZW5hYmxlZCI6dHJ1ZSwicGVuZGluZ1Jlc2V0IjpmYWxzZSwicm9sZXMiOlt7InJvbGVJZCI6IlNZU19BRE1JTiIsImJyYW5jaElkIjoiQlItMTAwMiIsIm9yZ2FuaXNhdGlvbmFsSWQiOiI1NDMyMSJ9XSwiaWF0IjoxNzM4ODMzOTcxLCJleHAiOjE3Mzg4NjI3NzF9.kB8IZG3o7Oq9d49A2QzpVOjxWZBd1HxL4e_vHQUKhMwpgS1G1RDOuHIUz_wiV7Ut",
		ValidateOnSet: true,
		Logger:        logger,
	}

	// Create an uploader from the processor config
	uploader := service.NewUploader(service.Config{
		UploadBaseURL: config.UploadBaseURL,
		BearerToken:   config.BearerToken,
	})

	// Example 1: PDF Form Processing
	pdfForm, err := pdfprocessor.NewFormFromURL("https://www.txdmv.gov/sites/default/files/form_files/130-U.pdf",
		pdfprocessor.WithValidation(),
		pdfprocessor.WithLogger(log.Default()),
		pdfprocessor.WithUploader(uploader),
	)
	if err != nil {
		log.Fatalf("Failed to create PDF form: %v", err)
	}

	// Example 2: HTML Form Processing
	htmlForm, err := pdfprocessor.NewHTMLFormFromURL("https://staging-storage.lyvepulse.com/storage/files/internal-files?objectName=metadata-files/d5ad01dd-1053-4271-8a00-2adf257ec34b-bc1722425fad-c018-d104-e1bc-34302a4f/application.html",
		pdfprocessor.WithValidation(),
		pdfprocessor.WithLogger(logger),
		pdfprocessor.WithUploader(uploader),
	)
	if err != nil {
		logger.Fatalf("Failed to create HTML form: %v", err)
	}

	// Get and print all fields immediately after creation
	fields := htmlForm.GetFields()
	logger.Printf("Found %d fields in total", len(fields))
	for name, field := range fields {
		logger.Printf("Field details - Name: %s, Type: %v, Required: %v",
			name, field.Type, field.Required)
		if len(field.Options) > 0 {
			logger.Printf("  Options for %s: %v", name, field.Options)
		}
	}

	// Process both forms
	forms := []pdfprocessor.FormProcessor{pdfForm}

	for _, form := range forms {
		// Print available fields
		form.PrintFields()

		// Set form fields
		// fields := map[string]interface{}{
		// 	"Name":       "John Doe",
		// 	"Age":        "30",
		// 	"IsEmployed": true,
		// 	"Department": "Engineering",
		// }

		fields := map[string]interface{}{
			"firstName":      "John",
			"middleName":     "Robert",
			"lastName":       "Doe",
			"mailingAddress": "123 Main Street",
			"city":           "Austin",
			"state":          "TX",
			"zip":            "78701",
			"email":          "john.doe@example.com",
			"phone":          "512-555-0123",

			// Vehicle Information
			"vin":        "1HGCM82633A123456",
			"year":       "2022",
			"make":       "Toyota",
			"model":      "Camry",
			"bodyStyle":  "Sedan",
			"majorColor": "Silver",
			"minorColor": "Black",
			"odometer":   "15000",

			// License and Registration
			"licensePlate": "ABC1234",
			"county":       "Travis",

			// Sales Information
			"salesPrice":       true,
			"salesPriceAmount": "25000",
			"taxableAmount":    true,

			// Additional Options
			"electronicTitleRequest": true,
			"eReminder":              true,

			// ID Information
			"applicantID":         "DL12345678",
			"idType":              true,
			"driverLicenseIssuer": "TX",
		}

		if err := pdfForm.SetFields(fields); err != nil {
			logger.Printf("Warning: Some fields could not be set: %v", err)
		}

		// Generate PDF from HTML and validate
		if err := htmlForm.GeneratePDF(); err != nil {
			logger.Fatalf("Failed to generate PDF: %v", err)
		}

		// Validate the form
		if err := pdfForm.Validate(); err != nil {
			logger.Printf("Validation failed: %v", err)
			return
		}

		// Upload configuration
		uploadConfig := types.UploadConfig{
			FileName:       "vehicle_registration_form.pdf", // Added .pdf extension
			OrganizationID: "54321",
			BranchID:       "BR-1002",
			CreatedBy:      "system",
		}

		// Upload the form
		response, err := pdfForm.Upload(context.Background(), uploadConfig)
		if err != nil {
			logger.Printf("Failed to upload form: %v", err)
			return
		}

		logger.Printf("Form uploaded successfully! Download URL: %s", response.FileDownloadUri)
	}
}
