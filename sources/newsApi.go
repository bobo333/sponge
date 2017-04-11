package sources

import (
	"fmt"
	shared "github.com/bobo333/sponge/shared"
)

type newsApiItem struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

func (n *newsApiItem) standardize() shared.StandardizedItem {
	return shared.StandardizedItem{
		Title: n.Title,
		Url:   n.Url,
	}
}

type newsApiResponse struct {
	Status   string        `json:"status"`
	Message  string        `json:"message"`
	Articles []newsApiItem `json:"articles"`
}

// GetNewsApiItems retrieves *numItems* of *newsSource* type from NewsApi,
// standardizes them, and compiles them into shared.OutputSection
func GetNewsApiItems(newsSource string, numItems int) (shared.OutputSection, error) {
	newsApiKey, newsApiKeyErr := shared.GetEnvVar("NEWS_API_KEY")
	if newsApiKeyErr != nil {
		return shared.OutputSection{}, newsApiKeyErr
	}

	newsApiUrl := fmt.Sprintf("https://newsapi.org/v1/articles?source=%s&apiKey=%s&sortBy=top", newsSource, newsApiKey)
	apiResponse := newsApiResponse{}
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
