// Package main provides an example of using the pdfprocessor package.
package main

import (
	"context"
	"log"

	"github.com/josephmowjew/go-pdf-filler/pdfprocessor"
	service "github.com/josephmowjew/go-pdf-filler/pdfprocessor/services"
)

func main() {
	// Create an uploader instance
	uploaderConfig := service.Config{
		UploadBaseURL: "https://staging-storage.lyvepulse.com/storage/files",
		BearerToken:   "eyJhbGciOiJIUzM4NCJ9.eyJzdWIiOiJuby1yZXBseUBvcHNzaWZ5LmNvbSIsInVzZXJuYW1lIjoibm8tcmVwbHlAb3Bzc2lmeS5jb20iLCJlbXBsb3llZUlkIjoibm8tcmVwbHlAb3Bzc2lmeS5jb20iLCJmaXJzdE5hbWUiOiJTWVNURU0iLCJsYXN0TmFtZSI6IlNZU1RFTSIsInBob25lTnVtYmVyIjoiODc2NzUyMzQyIiwiZW5hYmxlZCI6dHJ1ZSwicGVuZGluZ1Jlc2V0IjpmYWxzZSwicm9sZXMiOlt7InJvbGVJZCI6IlNZU19BRE1JTiIsImJyYW5jaElkIjoiQlItMTAwMiIsIm9yZ2FuaXNhdGlvbmFsSWQiOiI1NDMyMSJ9XSwiaWF0IjoxNzM4NjEwNTg2LCJleHAiOjE3Mzg2MzkzODZ9.KKVqII6bXB01aX2QVmV2eGO9c9Ec3nK-MoB6Jq4GxDXB-w-EZLGl2O3Xyrth9RjU",
	}
	uploader := service.NewUploader(uploaderConfig)

	// Initialize the PDF form processor with uploader
	processor, err := pdfprocessor.NewForm("form.pdf",
		pdfprocessor.WithValidation(),
		pdfprocessor.WithLogger(log.Default()),
		pdfprocessor.WithUploader(uploader),
	)
	if err != nil {
		log.Fatalf("Failed to create form: %v", err)
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
