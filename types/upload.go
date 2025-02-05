package types

import "fmt"

// UploadConfig represents the configuration for uploading a filled PDF
type UploadConfig struct {
	FileName         string
	OrganizationalID string
	BranchID         string
	CreatedBy        string
}

// Validate checks if the upload configuration is valid
func (c UploadConfig) Validate() error {
	if c.FileName == "" {
		return fmt.Errorf("filename is required")
	}
	if c.OrganizationalID == "" {
		return fmt.Errorf("organizational ID is required")
	}
	if c.BranchID == "" {
		return fmt.Errorf("branch ID is required")
	}
	if c.CreatedBy == "" {
		return fmt.Errorf("creator is required")
	}
	return nil
}

// UploadResponse represents the response from the upload service
type UploadResponse struct {
	FileName        string `json:"fileName"`
	FileDownloadUri string `json:"fileDownloadUri"`
	FileType        string `json:"fileType"`
	Size            int64  `json:"size"`
}
