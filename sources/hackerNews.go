package sources

import (
	"fmt"
	"github.com/bobo333/sponge/shared"
	"sync"
)

/*
   HACKER NEWS
*/

type hackerNewsItem struct {
	Title string `json:"title"`
	Url   string `json:"url"`
	Id    int    `json:"id"`
}

func (h hackerNewsItem) standardize() shared.StandardizedItem {
	return shared.StandardizedItem{
		Title:    h.Title,
		Url:      h.Url,
		Comments: fmt.Sprintf("https://news.ycombinator.com/item?id=%d", h.Id)}
}

// GetHackerNews gets *numItems* Hacker News items, standardizes them, and
// compiles them into shared.OutputSection.
func GetHackerNews(numItems int) (shared.OutputSection, error) {
	hackerNewsListUrl := "https://hacker-news.firebaseio.com/v0/topstories.json"
	hackerNewsItemUrl := "https://hacker-news.firebaseio.com/v0/item/%d.json"

	var hnl []int
	shared.GetJsonResponse(hackerNewsListUrl, &hnl)

	// take only top items (returns 500 initially)
	var hnTopItems []int
	hnTopItems = append(hnTopItems, hnl[:numItems]...)
	var hnTopItemsDetails []shared.StandardizedItem
	collectorChan := make(chan shared.StandardizedItem)

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
			item := hackerNewsItem{}
			err := shared.GetJsonResponse(hnItemUrl, &item)
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

	hackerNewsSection := shared.OutputSection{
		Name:  "Hacker News",
		Items: hnTopItemsDetails}

	return hackerNewsSection, nil
}
