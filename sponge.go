package main

import "net/http"
import "encoding/json"
import "fmt"
import "time"
import "os"

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
   - filter Hacker News if no url
*/

type HackerNewsItem struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

func (h HackerNewsItem) Format() string {
	return fmt.Sprintf("Title: %s\nUrl: %s", h.Title, h.Url)
}

var hackerNewsListUrl = "https://hacker-news.firebaseio.com/v0/topstories.json"
var hackerNewsItemUrl = "https://hacker-news.firebaseio.com/v0/item/%d.json"
var outputLocation = "/tmp/sponge_out.txt"

var itemsToFetch = 10

func main() {

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(hackerNewsListUrl)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	hnl := make([]int, 0)
	decoder.Decode(&hnl)

	hnTopItems := make([]int, itemsToFetch)
	copy(hnTopItems, hnl[:itemsToFetch])

	hnTopItemsDetails := make([]HackerNewsItem, 0)
	itemChannel := make(chan HackerNewsItem)
	receivedChannel := make(chan int)

	for _, id := range hnTopItems {
		go func(id int) {
			resp, err := client.Get(fmt.Sprintf(hackerNewsItemUrl, id))
			if err != nil {
				fmt.Println("error!", err)
			}
			defer resp.Body.Close()

			item := HackerNewsItem{}

			decoder := json.NewDecoder(resp.Body)
			dec_err := decoder.Decode(&item)
			if dec_err != nil {
				print("error!", err)
			}

			itemChannel <- item
			itemNumber := <-receivedChannel

			if itemNumber == itemsToFetch {
				close(itemChannel)
			}
		}(id)
	}

	itemsReceived := 0
	for elem := range itemChannel {
		hnTopItemsDetails = append(hnTopItemsDetails, elem)
		itemsReceived++

		receivedChannel <- itemsReceived
	}

	f, err := os.Create(outputLocation)
	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()

	for _, item := range hnTopItemsDetails {
		f.WriteString(item.Format())
		f.WriteString("\n\n")
	}

	fmt.Printf("Done writing output to %s\n", outputLocation)
}
