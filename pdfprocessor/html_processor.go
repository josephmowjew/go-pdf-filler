package pdfprocessor

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/josephmowjew/go-form-processor/types"
)

// HTMLForm represents an HTML form with its fields and configuration
type HTMLForm struct {
	fields   map[string]Field
	inputURL string
	rawHTML  string
	options  Options
	pdfData  []byte // Add this field to store the generated PDF
}

// NewHTMLFormFromURL creates a new HTMLForm instance from a URL
func NewHTMLFormFromURL(url string, opts ...Option) (*HTMLForm, error) {
	// Fetch the HTML content
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch HTML: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTML body: %w", err)
	}

	options := Options{
		Logger: log.Default(),
	}
	for _, opt := range opts {
		opt(&options)
	}

	form := &HTMLForm{
		inputURL: url,
		rawHTML:  string(body),
		fields:   make(map[string]Field),
		options:  options,
	}

	if err := form.loadFields(); err != nil {
		return nil, fmt.Errorf("failed to load form fields: %w", err)
	}

	return form, nil
}

// loadFields reads field information from the HTML document
func (f *HTMLForm) loadFields() error {
	resp, err := http.Get(f.inputURL)
	if err != nil {
		return fmt.Errorf("failed to fetch HTML: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Find ALL form elements, not just those within a form tag
	doc.Find("input, select, textarea").Each(func(i int, s *goquery.Selection) {
		name, exists := s.Attr("name")
		if !exists {
			return
		}

		field := Field{
			Name:     name,
			Required: s.AttrOr("required", "") != "",
			Options:  []string{},
		}

		// Determine field type
		inputType := s.AttrOr("type", "")
		switch {
		case s.Is("select"):
			field.Type = Choice
			s.Find("option").Each(func(i int, opt *goquery.Selection) {
				if value, exists := opt.Attr("value"); exists {
					field.Options = append(field.Options, value)
				}
			})
		case s.Is("input"):
			switch inputType {
			case "checkbox", "radio":
				field.Type = Boolean
			default:
				field.Type = Text
			}
		case s.Is("textarea"):
			field.Type = Text
		}

		f.fields[name] = field
	})

	return nil
}

// GetFields returns all form fields
func (f *HTMLForm) GetFields() map[string]Field {
	fields := make(map[string]Field, len(f.fields))
	for k, v := range f.fields {
		fields[k] = v
	}
	return fields
}

// SetField sets a value for a specific form field
func (f *HTMLForm) SetField(name string, value interface{}) error {
	field, exists := f.fields[name]
	if !exists {
		return fmt.Errorf("field %s not found in form", name)
	}

	// Type validation
	switch field.Type {
	case Text:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field %s requires string value", name)
		}
	case Boolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field %s requires boolean value", name)
		}
	case Choice:
		if strVal, ok := value.(string); ok {
			if !isValidOption(strVal, field.Options) {
				return fmt.Errorf("invalid option for field %s: %s", name, strVal)
			}
		} else {
			return fmt.Errorf("field %s requires string value from options", name)
		}
	}

	field.Value = value
	f.fields[name] = field

	if f.options.ValidateOnSet {
		return f.validateField(field)
	}
	return nil
}

// SetFields sets multiple field values
func (f *HTMLForm) SetFields(fields map[string]interface{}) error {
	var errors []string

	for name, value := range fields {
		if err := f.SetField(name, value); err != nil {
			errors = append(errors, fmt.Sprintf("field '%s': %v", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to set some fields: %s", strings.Join(errors, "; "))
	}

	return nil
}

// Validate checks if all required fields have values
func (f *HTMLForm) Validate() error {
	for _, field := range f.fields {
		if err := f.validateField(field); err != nil {
			return err
		}
	}
	return nil
}

// Upload submits the HTML form
func (f *HTMLForm) Upload(ctx context.Context, config types.UploadConfig) (*types.UploadResponse, error) {
	if f.options.Uploader == nil {
		return nil, fmt.Errorf("uploader service not configured")
	}

	// Use PDF data if available, otherwise use HTML
	var data []byte
	if f.pdfData != nil {
		data = f.pdfData
	} else {
		data = []byte(f.generateFilledHTML())
	}

	// Ensure filename has .pdf extension
	if !strings.HasSuffix(config.FileName, ".pdf") {
		config.FileName = config.FileName + ".pdf"
	}

	// Upload the filled form
	response, err := f.options.Uploader.Upload(ctx, data, config)
	if err != nil {
		return nil, fmt.Errorf("failed to upload form: %w", err)
	}

	return response, nil
}

// PrintFields displays all fields and their properties
func (f *HTMLForm) PrintFields() {
	if f.options.Logger == nil {
		return
	}

	f.options.Logger.Println("HTML Form Fields:")
	f.options.Logger.Println("================")

	for name, field := range f.fields {
		fieldType := "Text"
		switch field.Type {
		case Boolean:
			fieldType = "Boolean"
		case Choice:
			fieldType = "Choice"
		}

		f.options.Logger.Printf("Field: %s\n", name)
		f.options.Logger.Printf("  Type: %s\n", fieldType)
		f.options.Logger.Printf("  Required: %v\n", field.Required)
		if len(field.Options) > 0 {
			f.options.Logger.Printf("  Options: %v\n", field.Options)
		}
		if field.Value != nil {
			f.options.Logger.Printf("  Current Value: %v\n", field.Value)
		}
		f.options.Logger.Println("----------------")
	}
}

// generateFilledHTML creates a filled version of the HTML form
func (f *HTMLForm) generateFilledHTML() string {
	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(f.rawHTML))
	if err != nil {
		if f.options.Logger != nil {
			f.options.Logger.Printf("Error parsing HTML: %v", err)
		}
		return f.rawHTML
	}

	// Fill in form fields
	doc.Find("input, select, textarea").Each(func(i int, s *goquery.Selection) {
		name, exists := s.Attr("name")
		if !exists {
			return
		}

		field, exists := f.fields[name]
		if !exists || field.Value == nil {
			return
		}

		// Handle different input types
		inputType, _ := s.Attr("type")
		switch inputType {
		case "checkbox", "radio":
			if val, ok := field.Value.(bool); ok && val {
				s.SetAttr("checked", "checked")
			}
		default:
			// For text inputs, selects, and textareas
			value := fmt.Sprintf("%v", field.Value)
			if s.Is("select") {
				// For select elements, set the selected attribute on the matching option
				s.Find("option").Each(func(i int, opt *goquery.Selection) {
					if optVal, exists := opt.Attr("value"); exists && optVal == value {
						opt.SetAttr("selected", "selected")
					}
				})
			} else {
				s.SetAttr("value", value)
			}
		}
	})

	// Add necessary styling for PDF generation
	doc.Find("head").AppendHtml(`
		<style>
			body {
				font-family: Arial, sans-serif;
				line-height: 1.6;
				margin: 20px;
			}
			input, select, textarea {
				border: 1px solid #ccc;
				padding: 5px;
				margin: 5px 0;
			}
			input[type="checkbox"], input[type="radio"] {
				margin-right: 5px;
			}
			label {
				display: inline-block;
				margin-right: 10px;
			}
		</style>
	`)

	// Generate the HTML string
	html, err := doc.Html()
	if err != nil {
		if f.options.Logger != nil {
			f.options.Logger.Printf("Error generating HTML: %v", err)
		}
		return f.rawHTML
	}

	// Log the generated HTML for debugging
	if f.options.Logger != nil {
		f.options.Logger.Printf("Generated HTML:\n%s", html)
	}

	return html
}

func (f *HTMLForm) validateField(field Field) error {
	if field.Required && field.Value == nil {
		return fmt.Errorf("required field %s is not set", field.Name)
	}
	return nil
}

// GeneratePDF converts the filled HTML form to PDF format
func (f *HTMLForm) GeneratePDF() error {
	// Create a new Chrome instance
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Set a reasonable timeout
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Generate the filled HTML content
	filledHTML := f.generateFilledHTML()

	// Create a temporary file for the HTML
	tmpHTML, err := os.CreateTemp("", "form-*.html")
	if err != nil {
		return fmt.Errorf("failed to create temporary HTML file: %w", err)
	}
	tmpHTMLPath := tmpHTML.Name()
	defer os.Remove(tmpHTMLPath)

	// Write the filled HTML to the temporary file
	if err := os.WriteFile(tmpHTMLPath, []byte(filledHTML), 0644); err != nil {
		return fmt.Errorf("failed to write HTML to temporary file: %w", err)
	}

	// Convert the file path to a URL
	fileURL := "file://" + tmpHTMLPath

	// PDF generation parameters
	printToPDFParams := page.PrintToPDF().
		WithPrintBackground(true).
		WithPreferCSSPageSize(true).
		WithMarginTop(0.4).
		WithMarginBottom(0.4).
		WithMarginLeft(0.4).
		WithMarginRight(0.4).
		WithPaperWidth(8.5).
		WithPaperHeight(11)

	var pdfData []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfData, _, err = printToPDFParams.Do(ctx)
			return err
		}),
	); err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Store the PDF data in memory for later use by the Upload method
	f.pdfData = pdfData

	if f.options.Logger != nil {
		f.options.Logger.Printf("PDF generated successfully, size: %d bytes", len(pdfData))
	}

	return nil
}
