package main

import "net/http"
import "encoding/json"
import "fmt"
import "time"
import "os"
import "sync"

/*
   TODO:
   - parameterize items to get
   - parameterize output location
   - general cleanup
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
var itemsToFetch = 10
var outputLocation = "/tmp/sponge_out.txt"

type Formatted struct {
	Body string
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

func getHackerNews() []Formatted {
	hackerNewsListUrl := "https://hacker-news.firebaseio.com/v0/topstories.json"
	hackerNewsItemUrl := "https://hacker-news.firebaseio.com/v0/item/%d.json"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(hackerNewsListUrl)
	if err != nil {
		fmt.Println(err)
		return make([]Formatted, 0)
	}
	defer resp.Body.Close()

	hnl := make([]int, 0)
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&hnl)

	// take only top items (returns 500 initially)
	hnTopItems := make([]int, itemsToFetch)
	copy(hnTopItems, hnl[:itemsToFetch])

	hnTopItemsDetails := make([]Formatted, 0)

	var wg sync.WaitGroup
	wg.Add(itemsToFetch)

	for _, id := range hnTopItems {
		go func(id int) {
			defer wg.Done()

			resp, err := client.Get(fmt.Sprintf(hackerNewsItemUrl, id))
			if err != nil {
				fmt.Println("error!", err)
				return
			}
			defer resp.Body.Close()

			item := HackerNewsItem{}
			decoder := json.NewDecoder(resp.Body)
			dec_err := decoder.Decode(&item)
			if dec_err != nil {
				print("error!", err)
			} else {
				hnTopItemsDetails = append(hnTopItemsDetails, item.getFormatted())

			}
		}(id)
	}

	wg.Wait()

	return hnTopItemsDetails
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

func getRedditGolang() []Formatted {
	golangListUrl := fmt.Sprintf("https://www.reddit.com/r/golang/top.json?raw_json=1&t=day&limit=%d", itemsToFetch)

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", golangListUrl, nil)
	req.Header.Set("User-Agent", "golang Sponge:0.0.1 (by /u/bobo333)") // required or reddit API will return 429 code
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return make([]Formatted, 0)
	}
	if resp.StatusCode != 200 {
		fmt.Println("Non 200 response", resp.StatusCode)
		return make([]Formatted, 0)
	}
	defer resp.Body.Close()

	redditList := RedditList{}
	decoder := json.NewDecoder(resp.Body)
	decodeErr := decoder.Decode(&redditList)
	if decodeErr != nil {
		fmt.Println(decodeErr)
		return make([]Formatted, 0)
	}

	redditItems := make([]Formatted, itemsToFetch)
	for i, item := range redditList.Data.Children {
		redditItems[i] = item.Data.getFormatted()
	}

	return redditItems
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

func getNyt() []Formatted {
	apiKeyName := "NYT_API_KEY"

	nytApiKey := os.Getenv(apiKeyName)
	if nytApiKey == "" {
		fmt.Printf("%s not found!\n", apiKeyName)
		return make([]Formatted, 0)
	}

	nytUrl := fmt.Sprintf("https://api.nytimes.com/svc/topstories/v2/home.json?api-key=%s", nytApiKey)

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(nytUrl)

	if err != nil {
		fmt.Println(err)
		return make([]Formatted, 0)
	}
	defer resp.Body.Close()

	nytList := NytList{}
	decoder := json.NewDecoder(resp.Body)
	decodeErr := decoder.Decode(&nytList)
	if decodeErr != nil {
		fmt.Println(err)
		return make([]Formatted, 0)
	}

	nytItems := make([]Formatted, itemsToFetch)
	for i := 0; i < itemsToFetch; i++ {
		item := nytList.Results[i]
		nytItems[i] = item.getFormatted()
	}

	return nytItems
}

/*
   File creation
*/

func writeSection(f *os.File, sectionName string, items []Formatted) {
	heading := fmt.Sprintf("\n\n=====================================\n"+
		"%s\n=====================================\n\n", sectionName)
	f.WriteString(heading)

	for _, item := range items {
		f.WriteString(item.Body)
		f.WriteString("\n\n")
	}

	fmt.Printf("wrote %d %s items\n", len(items), sectionName)
}

func main() {

	redditItems := getRedditGolang()
	hackerNewsItems := getHackerNews()
	nytItems := getNyt()

	f, err := os.Create(outputLocation)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	writeSection(f, "New York Times", nytItems)
	writeSection(f, "Hacker News", hackerNewsItems)
	writeSection(f, "Reddit Golang", redditItems)

	fmt.Printf("Done writing output to %s\n", outputLocation)
}
