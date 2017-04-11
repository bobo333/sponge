package shared

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// WriteToFile creates a file at *outputFilePath* and writes the provided
// *text* to that file. If *textOutput* is true, the *text* is assumed to be
// plain text, and the file extension will be ".txt". Otherwise it is assumed
// to be html and the extension will be ".html"
func WriteToFile(outputFilePath, text string, textOutput bool) {
	extension := "html"
	if textOutput {
		extension = "txt"
	}
	outputFilePath = fmt.Sprintf("%s.%s", outputFilePath, extension)

	fmt.Println(outputFilePath)

	f, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	f.WriteString(text)

	fmt.Printf("Done writing output to %s\n", outputFilePath)
}

// WriteToEmail takes the given *text* and sends it to *emailAddress* via
// the Mailgun API. If *textOutput* is true, the email is sent as plaintext.
// If *textOutput* is false, the email is sent as html.
func WriteToEmail(emailAddress string, text string, textOutput bool) {
	fieldName := "html"
	if textOutput {
		fieldName = "text"
	}

	// get key env var
	key, err1 := GetEnvVar("MAILGUN_API_KEY")
	if err1 != nil {
		fmt.Printf("%#v\n", err1)
		return
	}

	mailgunDomain, err2 := GetEnvVar("MAILGUN_DOMAIN")
	if err2 != nil {
		fmt.Printf("%#v\n", err2)
		return
	}

	baseUrl := "https://api.mailgun.net/v3"
	emailUrl := fmt.Sprintf("%s/%s/messages", baseUrl, mailgunDomain)

	date := time.Now().Format("2006/01/02")
	subject := fmt.Sprintf("Sponge %s", date)

	// add fields to form (including email body)
	form := url.Values{}
	form.Add("to", emailAddress)
	form.Add("from", "sponge@stevencipriano.com")
	form.Add("subject", subject)
	form.Add(fieldName, text)

	req, reqErr := http.NewRequest(http.MethodPost, emailUrl, strings.NewReader(form.Encode()))
	if reqErr != nil {
		fmt.Printf("%#v", reqErr)
		return
	}

	req.SetBasicAuth("api", key)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// make client
	client := &http.Client{Timeout: 10 * time.Second}

	// send request, check for errors
	resp, respErr := client.Do(req)
	if respErr != nil {
		fmt.Printf("%#v", respErr)
		return
	}

	// check response for errors or success message
	defer resp.Body.Close()

	body, bodyErr := ioutil.ReadAll(resp.Body)

	if bodyErr != nil {
		fmt.Println(bodyErr)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Println("Status code:", resp.StatusCode)
		fmt.Println(string(body))
		return
	}

	fmt.Printf("Done sending email to %s\n", emailAddress)
}
