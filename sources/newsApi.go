package sources

import (
	"fmt"
	shared "github.com/bobo333/sponge/shared"
)

type NewsApiItem struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

func (n *NewsApiItem) standardize() shared.StandardizedItem {
	return shared.StandardizedItem{
		Title: n.Title,
		Url:   n.Url,
	}
}

type NewsApiResponse struct {
	Status   string        `json:"status"`
	Message  string        `json:"message"`
	Articles []NewsApiItem `json:"articles"`
}

func GetNewsApiItems(newsSource string, numItems int) (shared.OutputSection, error) {
	newsApiKey, newsApiKeyErr := shared.GetEnvVar("NEWS_API_KEY")
	if newsApiKeyErr != nil {
		return shared.OutputSection{}, newsApiKeyErr
	}

	newsApiUrl := fmt.Sprintf("https://newsapi.org/v1/articles?source=%s&apiKey=%s&sortBy=top", newsSource, newsApiKey)
	apiResponse := NewsApiResponse{}
	err := shared.GetJsonResponse(newsApiUrl, &apiResponse)
	if err != nil {
		return shared.OutputSection{}, err
	}

	var newsApiItems []shared.StandardizedItem
	for i := 0; i < numItems && i < len(apiResponse.Articles); i++ {
		newsApiItems = append(newsApiItems, apiResponse.Articles[i].standardize())
	}

	output := shared.OutputSection{
		Name:  newsSource,
		Items: newsApiItems}

	return output, nil
}
