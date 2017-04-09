package shared

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

/*
   Utility functions
*/

func GetEnvVar(varName string) (string, error) {
	envVarValue := os.Getenv(varName)
	if envVarValue == "" {
		return "", errors.New(fmt.Sprintf("Env var %s not found", varName))
	}
	return envVarValue, nil
}

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

// Standardized Item
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

// Output Section
type OutputSection struct {
	Name  string
	Items []StandardizedItem
}

func (s *OutputSection) ToText() string {
	text := fmt.Sprintf("\n\n=====================================\n"+
		"%s\n=====================================\n\n", s.Name)

	for _, item := range s.Items {
		text += item.toText()
		text += "\n\n"
	}

	return text
}

func (s *OutputSection) ToHtml() string {
	html := fmt.Sprintf("<h3>%s</h3>", s.Name)

	for _, item := range s.Items {
		html += item.toHtml()
	}

	return html
}

func CreateRawHtml(sections []OutputSection) string {
	htmlStart := `<!DOCTYPE html>
                  <html>
                    <head><meta charset="UTF-8"></head>
                    <body>
                        <h1>Sponge</h1>`
	htmlEnd := "</body></html>"

	var sectionsHtmlList []string

	for _, section := range sections {
		sectionsHtmlList = append(sectionsHtmlList, section.ToHtml())
		fmt.Printf("formatted %d %s items\n", len(section.Items), section.Name)
	}

	sectionsHtml := strings.Join(sectionsHtmlList, "")

	return fmt.Sprintf("%s%s%s", htmlStart, sectionsHtml, htmlEnd)
}

func CreateRawText(sections []OutputSection) string {
	var rawText string

	for _, section := range sections {
		rawText += section.ToText()
		fmt.Printf("Formatted %d %s items in text form\n", len(section.Items), section.Name)
	}

	return rawText
}
