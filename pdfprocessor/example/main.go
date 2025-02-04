// Package main provides an example of using the pdfprocessor package.
package main

import (
	"context"
	"log"

	"gitlab.lyvepulse.com/lyvepulse/go-pdf-filler/pdfprocessor"
	service "gitlab.lyvepulse.com/lyvepulse/go-pdf-filler/pdfprocessor/services"
)

func main() {
	// Create processor configuration
	config := pdfprocessor.PDFProcessorConfig{
		UploadBaseURL: "https://staging-storage.lyvepulse.com/storage/files",
		BearerToken:   "your-bearer-token",
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
	// processorLocal, err := pdfprocessor.NewForm("form.pdf",
	// 	pdfprocessor.WithValidation(),
	// 	pdfprocessor.WithLogger(log.Default()),
	// 	pdfprocessor.WithUploader(uploader),
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to create form from local file: %v", err)
	// }

	// Example 2: Using a URL
	processorURL, err := pdfprocessor.NewFormFromURL("https://www.txdmv.gov/sites/default/files/form_files/130-U.pdf",
		pdfprocessor.WithValidation(),
		pdfprocessor.WithLogger(log.Default()),
		pdfprocessor.WithUploader(uploader),
	)
	if err != nil {
		log.Fatalf("Failed to create form from URL: %v", err)
	}

	// Define field values to be set
	fields := map[string]interface{}{
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

	// Using the URL-based processor for this example
	processor = processorURL

	// Print all available fields before setting values
	//processor.PrintFields()

	// Alternatively, you can get the fields and process them yourself
	formFields := processor.GetFields()
	for name, field := range formFields {
		log.Printf("Found field: %s (Type: %v, Required: %v)\n",
			name, field.Type, field.Required)
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
		FileName:         "vehicle_registration.pdf",
		OrganizationalID: "54321",
		BranchID:         "BR-1002",
		CreatedBy:        "system",
	}

	// Upload the filled form
	ctx := context.Background()
	response, err := processor.Upload(ctx, uploadConfig)
	if err != nil {
		log.Fatalf("Failed to upload form: %v", err)
	}

	log.Printf("Form uploaded successfully! Download URL: %s", response.FileDownloadUri)
}
