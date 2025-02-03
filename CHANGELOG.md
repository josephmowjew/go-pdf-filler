# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-02-03

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