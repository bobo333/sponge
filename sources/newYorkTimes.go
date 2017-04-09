package sources

import (
	"fmt"
	shared "github.com/bobo333/sponge/shared"
)

/*
   NEW YORK TIMES
*/

type NytItem struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

func (n NytItem) standardize() shared.StandardizedItem {
	return shared.StandardizedItem{
		Title: n.Title,
		Url:   n.Url}
}

type NytList struct {
	Results []NytItem `json:"results"`
}

func GetNyt(numItems int) (shared.OutputSection, error) {
	nytApiKey, apiKeyErr := shared.GetEnvVar("NYT_API_KEY")
	if apiKeyErr != nil {
		return shared.OutputSection{}, apiKeyErr
	}

	nytUrl := fmt.Sprintf("https://api.nytimes.com/svc/topstories/v2/home.json?api-key=%s", nytApiKey)
	nytList := NytList{}
	err := shared.GetJsonResponse(nytUrl, &nytList)
	if err != nil {
		return shared.OutputSection{}, err
	}

	var nytItems []shared.StandardizedItem
	for i := 0; i < numItems; i++ {
		item := nytList.Results[i]
		nytItems = append(nytItems, item.standardize())
	}

	output := shared.OutputSection{
		Name:  "New York Times",
		Items: nytItems}

	return output, nil
}
