package shared

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// TODO: use templates

/*
   Utility functions
*/

// GetEnvVar finds the environment variable *varName* and returns its value.
// If it is not found, returns an error.
func GetEnvVar(varName string) (string, error) {
	envVarValue := os.Getenv(varName)
	if envVarValue == "" {
		return "", fmt.Errorf("Env var %s not found", varName)
	}
	return envVarValue, nil
}

// GetJsonResponse makes a GET request to the given *url* and decodes the
// JSON response body into the provided struct *v*.
func GetJsonResponse(url string, v interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(resp.Body)
	if decErr := decoder.Decode(v); decErr != nil {
		return decErr
	}

	return nil
}

/*
   Standardized Item
*/

// StandardizedItem a struct that defines the standard interface that all items
// must be converted to before writing to file or email. Provides some built-in
// methods for converting to the appropriate formats for output.
type StandardizedItem struct {
	Title    string
	Url      string
	Comments string
}

func (si *StandardizedItem) toText() string {
	text := ""

	switch {
	case si.Comments != "":
		text = fmt.Sprintf("Title: %s\nUrl: %s\nComments: %s", si.Title, si.Url, si.Comments)
	default:
		text = fmt.Sprintf("Title: %s\nUrl: %s", si.Title, si.Url)
	}

	return text
}

func (si *StandardizedItem) toHtml() string {
	html := ""

	switch {
	case si.Comments != "":
		html = fmt.Sprintf(`
            <div>
                <a href="%s">%s</a> &#124; <a href="%s">Comments</a>
            </div>
        `, si.Url, si.Title, si.Comments)
	default:
		html = fmt.Sprintf(`
            <div>
                <a href="%s">%s</a><br>
            </div>
        `, si.Url, si.Title)
	}

	return html
}

/*
   Output Section
*/

// OutputSection a struct containing the data for a full section of output items
// to be written to a file or email. Includes some built-in methods for
// converting to the appropriate output format.
type OutputSection struct {
	Name  string
	Items []StandardizedItem
}

func (s *OutputSection) toText() string {
	text := fmt.Sprintf("\n\n=====================================\n"+
		"%s\n=====================================\n\n", s.Name)

	for _, item := range s.Items {
		text += item.toText()
		text += "\n\n"
	}

	return text
}

func (s *OutputSection) toHtml() string {
	html := fmt.Sprintf("<h3>%s</h3>", s.Name)

	for _, item := range s.Items {
		html += item.toHtml()
	}

	return html
}

// CreateRawHtml converts an array of OutputSections to a fully formatted html
// string ready to be written to a file or email.
func CreateRawHtml(sections []OutputSection) string {
	htmlStart := `<!DOCTYPE html>
                  <html>
                    <head><meta charset="UTF-8"></head>
                    <body>
                        <h1>Sponge</h1>`
	htmlEnd := "</body></html>"

	var sectionsHtmlList []string

	for _, section := range sections {
		sectionsHtmlList = append(sectionsHtmlList, section.toHtml())
		fmt.Printf("formatted %d %s items\n", len(section.Items), section.Name)
	}

	sectionsHtml := strings.Join(sectionsHtmlList, "")

	return fmt.Sprintf("%s%s%s", htmlStart, sectionsHtml, htmlEnd)
}

// CreateRawText converts an array of OutputSections to a fully formatted text
// string ready to be written to a file or email.
func CreateRawText(sections []OutputSection) string {
	var rawText string

	for _, section := range sections {
		rawText += section.toText()
		fmt.Printf("Formatted %d %s items in text form\n", len(section.Items), section.Name)
	}

	return rawText
}
