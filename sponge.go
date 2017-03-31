package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

/*
   TODO:
   - make submodules
   - other sources
       - wash post
       - wsj
       - reddit
            - python
            - programming
            - sysadmin
            - others?
       - techcrunch?
       - economist
   - filter Hacker News if no url
*/

var defaultOutput = filepath.Join(os.TempDir(), "sponge_out.txt")
var itemsToFetch = flag.Int("count", 10, "Number of items to fetch from each source")
var outputLocation = flag.String("out", defaultOutput, "Output file")

type Formatted struct {
	Body string
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
}

func (h HackerNewsItem) getFormatted() Formatted {
	return Formatted{
		Body: fmt.Sprintf("Title: %s\nUrl: %s", h.Title, h.Url)}
}

func getHackerNews() ([]Formatted, error) {
	hackerNewsListUrl := "https://hacker-news.firebaseio.com/v0/topstories.json"
	hackerNewsItemUrl := "https://hacker-news.firebaseio.com/v0/item/%d.json"

	var hnl []int
	getJsonResponse(hackerNewsListUrl, &hnl)

	// take only top items (returns 500 initially)
	var hnTopItems []int
	hnTopItems = append(hnTopItems, hnl[:*itemsToFetch]...)
	var hnTopItemsDetails []Formatted
	collectorChan := make(chan Formatted)

	var wg sync.WaitGroup
	wg.Add(len(hnTopItems))

	go func() {
		wg.Wait()
		close(collectorChan)
	}()

	for _, id := range hnTopItems {
		id := id // need this or will only use LAST value of id for all goroutines
		go func() {
			defer wg.Done()

			hnItemUrl := fmt.Sprintf(hackerNewsItemUrl, id)
			item := HackerNewsItem{}
			err := getJsonResponse(hnItemUrl, &item)
			if err != nil {
				fmt.Printf("%#v\n", err)
			} else {
				collectorChan <- item.getFormatted()
			}
		}()
	}

	for item := range collectorChan {
		hnTopItemsDetails = append(hnTopItemsDetails, item)
	}

	return hnTopItemsDetails, nil
}

/*
   REDDIT GOLANG
*/

type RedditItem struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

func (r RedditItem) getFormatted() Formatted {
	return Formatted{
		Body: fmt.Sprintf("Title: %s\nUrl: %s", r.Title, r.Url)}
}

type RedditList struct {
	Data struct {
		Children []struct {
			Data RedditItem `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func getRedditGolang() ([]Formatted, error) {
	redditUsernameEnvName := "REDDIT_USERNAME"
	redditUsername, envVarErr := getEnvVar(redditUsernameEnvName)
	if envVarErr != nil {
		return nil, envVarErr
	}

	userAgent := fmt.Sprintf("golang Sponge:0.0.1 (by /u/%s)", redditUsername)
	golangListUrl := fmt.Sprintf("https://www.reddit.com/r/golang/top.json?raw_json=1&t=day&limit=%d", *itemsToFetch)

	// TODO: factor out client and json parsing
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", golangListUrl, nil)
	req.Header.Set("User-Agent", userAgent) // required or reddit API will return 429 code
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Non 200 response %d", resp.StatusCode))
	}
	defer resp.Body.Close()

	redditList := RedditList{}
	decoder := json.NewDecoder(resp.Body)
	if decodeErr := decoder.Decode(&redditList); decodeErr != nil {
		return nil, decodeErr
	}

	var redditItems []Formatted
	for _, item := range redditList.Data.Children {
		redditItems = append(redditItems, item.Data.getFormatted())
	}

	return redditItems, nil
}

/*
   NEW YORK TIMES
*/

type NytItem struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

func (n NytItem) getFormatted() Formatted {
	return Formatted{
		Body: fmt.Sprintf("Title: %s\nUrl: %s", n.Title, n.Url)}
}

type NytList struct {
	Results []NytItem `json:"results"`
}

func getNyt() ([]Formatted, error) {
	nytApiKey, apiKeyErr := getEnvVar("NYT_API_KEY")
	if apiKeyErr != nil {
		return nil, apiKeyErr
	}

	nytUrl := fmt.Sprintf("https://api.nytimes.com/svc/topstories/v2/home.json?api-key=%s", nytApiKey)
	nytList := NytList{}
	err := getJsonResponse(nytUrl, &nytList)
	if err != nil {
		return nil, err
	}

	var nytItems []Formatted
	for i := 0; i < *itemsToFetch; i++ {
		item := nytList.Results[i]
		nytItems = append(nytItems, item.getFormatted())
	}

	return nytItems, nil
}

/*
   File creation
*/

func writeSection(f *os.File, section outputSection) {
	heading := fmt.Sprintf("\n\n=====================================\n"+
		"%s\n=====================================\n\n", section.Name)
	f.WriteString(heading)

	for _, item := range section.Data {
		f.WriteString(item.Body)
		f.WriteString("\n\n")
	}

	fmt.Printf("wrote %d %s items\n", len(section.Data), section.Name)
}

type outputSection struct {
	Name string
	Data []Formatted
}

func main() {
	flag.Parse()

	mappings := map[string]func() ([]Formatted, error){
		"Hacker News":    getHackerNews,
		"Reddit Golang":  getRedditGolang,
		"New York Times": getNyt,
	}

	var wg sync.WaitGroup
	wg.Add(3)
	receiverChannel := make(chan outputSection)

	go func() {
		wg.Wait()
		close(receiverChannel)
	}()

	for name, fxn := range mappings {
		name := name
		fxn := fxn

		go func() {
			defer wg.Done()

			itemsList, err := fxn()
			if err != nil {
				fmt.Println(fmt.Sprintf("%s error: ", name), err)
			} else {
				receiverChannel <- outputSection{
					Name: name,
					Data: itemsList}
			}
		}()
	}

	f, err := os.Create(*outputLocation)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	for section := range receiverChannel {
		writeSection(f, section)
	}

	fmt.Printf("Done writing output to %s\n", *outputLocation)
}
