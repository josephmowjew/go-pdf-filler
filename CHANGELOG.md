# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.6] - 2024-02-04

### Fixed
- Fixed syntax error in PDFProcessorConfig initialization in example code

## [0.1.5] - 2024-02-04

### Fixed
- Corrected type mismatch between `PDFProcessorConfig` and `service.Config`
- Fixed variable reassignment in example code
- Properly initialized uploader with configuration values
- Updated example code to demonstrate correct configuration flow

## [0.1.4] - 2024-02-04

### Fixed
- Resolved type mismatch in PDF form field iteration
- Corrected variable shadowing in field inspection example 
- Updated documentation to reflect proper field type handling
- Fixed example code in README for form field inspection

## [0.1.3] - 2024-02-04

### Fixed
- Resolved type mismatch in PDF form field iteration
- Corrected variable shadowing in field inspection example
- Updated documentation to reflect proper field type handling
- Fixed example code in README for form field inspection

## [0.1.2] - 2024-02-04

### Added
- Enhanced error handling with custom `HTTPError` type
- Detailed error messages for authentication and upload failures
- Automatic temporary file cleanup using runtime finalizer
- Manual `Cleanup()` method for explicit file cleanup
- Better logging for upload responses and errors

### Changed
- Improved error messages for authentication failures
- Updated example code to use environment variables for bearer token
- Enhanced temporary file management in `NewFormFromURL`
- Better handling of HTTP response status codes
- Updated documentation with security best practices

### Security
- Removed hardcoded bearer token from example code
- Added recommendations for secure token management
- Improved cleanup of temporary files
- Enhanced error handling for authentication failures

## [0.1.1] - 2024-02-03

### Added
- Upload functionality for filled PDF forms
- HTTP upload service integration
- Context support for upload operations
- Configurable uploader via options pattern
- Temporary file handling with proper cleanup

### Changed
- Replace local file saving with direct upload functionality
- Make uploader service constructor public
- Update example code to demonstrate upload usage

### Removed
- Local file saving functionality (breaking change)

### Security
- Added proper temporary file cleanup
- Context support for request cancellation
- Secure file handling practices

## [0.1.0] - 2024-02-03

### Added
- Initial release of the Dynamic PDF Form Filler package
- Core PDF form processing functionality
  - Dynamic field detection using pdftk
  - Support for text, boolean, and choice field types
  - Field validation system
  - Type-safe field setting
- Configuration options
  - Validation on field set
  - Custom logging support
- Example implementation in `example/main.go`
- Comprehensive documentation
- Error handling for common scenarios
- Integration with fillpdf library for PDF manipulation

### Security
- Implemented safe PDF processing using pdftk
- Type validation for form fields
- Error handling for file operations 
