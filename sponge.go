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
       - nyt
       - wash post
       - wsj
       - reddit?
       - techcrunch?
       - economist
   - filter Hacker News if no url
*/
var itemsToFetch = 10
var outputLocation = "/tmp/sponge_out.txt"

type formattable interface {
	Format() string
}

/*
   HACKER NEWS
*/

type HackerNewsItem struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

func (h HackerNewsItem) Format() string {
	return fmt.Sprintf("Title: %s\nUrl: %s", h.Title, h.Url)
}

func getHackerNews() []HackerNewsItem {
	hackerNewsListUrl := "https://hacker-news.firebaseio.com/v0/topstories.json"
	hackerNewsItemUrl := "https://hacker-news.firebaseio.com/v0/item/%d.json"

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(hackerNewsListUrl)
	if err != nil {
		fmt.Println(err)
		return make([]HackerNewsItem, 0)
	}
	defer resp.Body.Close()

	hnl := make([]int, 0)
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(&hnl)

	// take only top items (returns 500 initially)
	hnTopItems := make([]int, itemsToFetch)
	copy(hnTopItems, hnl[:itemsToFetch])

	hnTopItemsDetails := make([]HackerNewsItem, 0)

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
				fmt.Println("appending item")
				hnTopItemsDetails = append(hnTopItemsDetails, item)

			}
		}(id)
	}

	wg.Wait()

	return hnTopItemsDetails
}

/*
   Reddit
*/

func main() {

	hn := getHackerNews()

	f, err := os.Create(outputLocation)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()

	for _, item := range hn {
		fmt.Println("writing item")
		f.WriteString(item.Format())
		f.WriteString("\n\n")
	}

	fmt.Printf("Done writing output to %s\n", outputLocation)
}
