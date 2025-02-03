// Package main provides an example of using the pdfprocessor package.
package main

import (
	"log"

	"github.com/yourusername/pdfprocessor/pdfprocessor"
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

	// Save the filled form
	if err := processor.Save("output.pdf"); err != nil {
		log.Fatalf("Failed to save form: %v", err)
	}

	log.Println("Form processed successfully!")
}
