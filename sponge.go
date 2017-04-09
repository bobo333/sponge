package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

/*
   TODO:
   - better env var inits and checks
   - make submodules
   - other sources
       - wash post
       - wsj
       - techcrunch?
       - economist
   - filter Hacker News if no url
*/

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

func getJsonResponse(url string, v interface{}) error {
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

func getEnvVar(varName string) (string, error) {
	envVarValue := os.Getenv(varName)
	if envVarValue == "" {
		return "", errors.New(fmt.Sprintf("Env var %s not found", varName))
	}
	return envVarValue, nil
}

/*
   HACKER NEWS
*/

type HackerNewsItem struct {
	Title string `json:"title"`
	Url   string `json:"url"`
	Id    int    `json:"id"`
}

func (h HackerNewsItem) standardize() StandardizedItem {
	return StandardizedItem{
		Title:    h.Title,
		Url:      h.Url,
		Comments: fmt.Sprintf("https://news.ycombinator.com/item?id=%d", h.Id)}
}

func getHackerNews(numItems int) (outputSection, error) {
	hackerNewsListUrl := "https://hacker-news.firebaseio.com/v0/topstories.json"
	hackerNewsItemUrl := "https://hacker-news.firebaseio.com/v0/item/%d.json"

	var hnl []int
	getJsonResponse(hackerNewsListUrl, &hnl)

	// take only top items (returns 500 initially)
	var hnTopItems []int
	hnTopItems = append(hnTopItems, hnl[:numItems]...)
	var hnTopItemsDetails []StandardizedItem
	collectorChan := make(chan StandardizedItem)

	var wg sync.WaitGroup

	go func() {
		wg.Wait()
		close(collectorChan)
	}()

	for _, id := range hnTopItems {
		wg.Add(1)
		id := id // need this or will only use LAST value of id for all goroutines
		go func() {
			defer wg.Done()

			hnItemUrl := fmt.Sprintf(hackerNewsItemUrl, id)
			item := HackerNewsItem{}
			err := getJsonResponse(hnItemUrl, &item)
			if err != nil {
				fmt.Printf("%#v\n", err)
			} else {
				collectorChan <- item.standardize()
			}
		}()
	}

	for item := range collectorChan {
		hnTopItemsDetails = append(hnTopItemsDetails, item)
	}

	hackerNewsSection := outputSection{
		Name:  "Hacker News",
		Items: hnTopItemsDetails}

	return hackerNewsSection, nil
}

/*
   REDDIT GOLANG
*/

type RedditItem struct {
	Title     string `json:"title"`
	Url       string `json:"url"`
	Permalink string `json:"permalink"`
}

func (r RedditItem) standardize() StandardizedItem {
	return StandardizedItem{
		Title:    r.Title,
		Url:      r.Url,
		Comments: fmt.Sprintf("https://reddit.com%s", r.Permalink)}
}

type RedditList struct {
	Data struct {
		Children []struct {
			Data RedditItem `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func getReddit(subName string, numItems int) (outputSection, error) {
	redditUsernameEnvName := "REDDIT_USERNAME"
	redditUsername, envVarErr := getEnvVar(redditUsernameEnvName)
	if envVarErr != nil {
		return outputSection{}, envVarErr
	}

	userAgent := fmt.Sprintf("golang Sponge:0.0.1 (by /u/%s)", redditUsername)
	golangListUrl := fmt.Sprintf("https://www.reddit.com/r/%s/top.json?raw_json=1&t=day&limit=%d", subName, numItems)

	// TODO: factor out client and json parsing
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", golangListUrl, nil)
	req.Header.Set("User-Agent", userAgent) // required or reddit API will return 429 code
	resp, err := client.Do(req)
	if err != nil {
		return outputSection{}, err
	}
	if resp.StatusCode != 200 {
		return outputSection{}, errors.New(fmt.Sprintf("Non 200 response %d", resp.StatusCode))
	}
	defer resp.Body.Close()

	redditList := RedditList{}
	decoder := json.NewDecoder(resp.Body)
	if decodeErr := decoder.Decode(&redditList); decodeErr != nil {
		return outputSection{}, decodeErr
	}

	var redditItems []StandardizedItem
	for _, item := range redditList.Data.Children {
		redditItems = append(redditItems, item.Data.standardize())
	}

	output := outputSection{
		Name:  fmt.Sprintf("Reddit r/%s", subName),
		Items: redditItems}

	return output, nil
}

/*
   NEW YORK TIMES
*/

type NytItem struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

func (n NytItem) standardize() StandardizedItem {
	return StandardizedItem{
		Title: n.Title,
		Url:   n.Url}
}

type NytList struct {
	Results []NytItem `json:"results"`
}

func getNyt(numItems int) (outputSection, error) {
	nytApiKey, apiKeyErr := getEnvVar("NYT_API_KEY")
	if apiKeyErr != nil {
		return outputSection{}, apiKeyErr
	}

	nytUrl := fmt.Sprintf("https://api.nytimes.com/svc/topstories/v2/home.json?api-key=%s", nytApiKey)
	nytList := NytList{}
	err := getJsonResponse(nytUrl, &nytList)
	if err != nil {
		return outputSection{}, err
	}

	var nytItems []StandardizedItem
	for i := 0; i < numItems; i++ {
		item := nytList.Results[i]
		nytItems = append(nytItems, item.standardize())
	}

	output := outputSection{
		Name:  "New York Times",
		Items: nytItems}

	return output, nil
}

/*
   File creation
*/
func createRawText(sections []outputSection) string {
	var rawText string

	for _, section := range sections {
		rawText += section.toText()
		fmt.Printf("Formatted %d %s items in text form\n", len(section.Items), section.Name)
	}

	return rawText
}

func writeToFile(outputFilePath, text string, textOutput bool) {
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

/*
   Email creation
*/
func writeToEmail(emailAddress string, text string, textOutput bool) {
	fieldName := "html"
	if textOutput {
		fieldName = "text"
	}

	// get key env var
	key, err1 := getEnvVar("MAILGUN_API_KEY")
	if err1 != nil {
		fmt.Printf("%#v", err1)
		return
	}

	mailgunDomain, err2 := getEnvVar("MAILGUN_DOMAIN")
	if err2 != nil {
		fmt.Printf("%#v, err2")
		return
	}

	baseUrl := "https://api.mailgun.net/v3"
	email_url := fmt.Sprintf("%s/%s/messages", baseUrl, mailgunDomain)

	date := time.Now().Format("2006/01/02")
	subject := fmt.Sprintf("Sponge %s", date)

	// add fields to form (including email body)
	form := url.Values{}
	form.Add("to", emailAddress)
	form.Add("from", "sponge@stevencipriano.com")
	form.Add("subject", subject)
	form.Add(fieldName, text)

	req, reqErr := http.NewRequest(http.MethodPost, email_url, strings.NewReader(form.Encode()))
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

func createRawHtml(sections []outputSection) string {
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

/*
   Output Section
*/

type outputSection struct {
	Name  string
	Items []StandardizedItem
}

func (s *outputSection) toText() string {
	text := fmt.Sprintf("\n\n=====================================\n"+
		"%s\n=====================================\n\n", s.Name)

	for _, item := range s.Items {
		text += item.toText()
		text += "\n\n"
	}

	return text
}

func (s *outputSection) toHtml() string {
	html := fmt.Sprintf("<h3>%s</h3>", s.Name)

	for _, item := range s.Items {
		html += item.toHtml()
	}

	return html
}

/*
   main
*/

func main() {
	defaultOutput := filepath.Join(os.TempDir(), "sponge_out")
	numItems := flag.Int("count", 10, "Number of items to fetch from each source")
	outputLocation := flag.String("out", defaultOutput, "Output file")
	emailDestination := flag.String("email", "", "The email address to send results to")
	textOutput := flag.Bool("text", false, "Output in text form (default is html)")

	flag.Parse()

	sectionsToGet := []func(int) (outputSection, error){
		getHackerNews,
		getNyt,
	}

	subredditsToGet := []string{
		"golang",
		"python",
		"sysadmin",
		"programming",
		"liverpoolfc",
	}

	var wg sync.WaitGroup
	receiverChannel := make(chan outputSection)

	for _, fxn := range sectionsToGet {
		wg.Add(1)
		fxn := fxn

		go func() {
			defer wg.Done()

			output, err := fxn(*numItems)
			if err != nil {
				fmt.Printf("%s\n", err)
			} else {
				receiverChannel <- output
			}
		}()
	}

	for _, subName := range subredditsToGet {
		wg.Add(1)
		subName := subName

		go func() {
			defer wg.Done()

			output, err := getReddit(subName, *numItems)
			if err != nil {
				fmt.Printf("%s\n", err)
			} else {
				receiverChannel <- output
			}
		}()
	}

	go func() {
		wg.Wait()
		close(receiverChannel)
	}()

	var sections []outputSection
	for section := range receiverChannel {
		sections = append(sections, section)
	}

	// turn sections into text or html output
	var rawOutput string
	if *textOutput {
		rawOutput = createRawText(sections)
	} else {
		rawOutput = createRawHtml(sections)
	}

	// write output to destination (file or email)
	if *emailDestination != "" {
		writeToEmail(*emailDestination, rawOutput, *textOutput)
	} else {
		writeToFile(*outputLocation, rawOutput, *textOutput)
	}

}
